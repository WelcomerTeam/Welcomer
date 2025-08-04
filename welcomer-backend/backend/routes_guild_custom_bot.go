package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	sandwich_protobuf "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
)

func GetGuildCustomBotCount(ctx context.Context, guildID discord.Snowflake) (int, error) {
	customBots, err := welcomer.Queries.GetCustomBotsByGuildId(ctx, int64(guildID))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get custom bots")

		return 0, fmt.Errorf("failed to get custom bots for guild %d: %w", guildID, err)
	}

	return len(customBots), nil
}

// Route GET /api/guild/:guildID/custom-bot
func getGuildCustomBots(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			customBots, err := welcomer.Queries.GetCustomBotsByGuildId(ctx, int64(guildID))
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get custom bots")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			bots := make([]GuildCustomBot, 0)

			applicationsPb, err := welcomer.SandwichClient.FetchApplication(ctx, &sandwich_protobuf.ApplicationIdentifier{})
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to fetch applications for custom bots")
			}

			applications := applicationsPb.GetApplications()

			for _, bot := range customBots {
				shards := make([]GetStatusResponseShard, 0)

				application, ok := applications[welcomer.GetCustomBotKey(bot.CustomBotUuid)]
				if ok {
					for _, shard := range application.GetShards() {
						shards = append(shards, GetStatusResponseShard{
							ShardID: int(shard.GetId()),
							Status:  int(shard.GetStatus()),
							Latency: int(shard.GetGatewayLatency()),
							Guilds:  int(shard.GetGuilds()),
							Uptime:  int(time.Since(time.Unix(shard.GetStartedAt(), 0)).Seconds()),
						})
					}

					slices.SortFunc(shards, func(a, b GetStatusResponseShard) int {
						return a.ShardID - b.ShardID
					})
				}

				bots = append(bots, GuildCustomBot{
					UUID:              bot.CustomBotUuid.String(),
					IsActive:          bot.IsActive,
					PublicKey:         bot.PublicKey,
					ApplicationID:     discord.Snowflake(bot.ApplicationID),
					ApplicationName:   bot.ApplicationName,
					ApplicationAvatar: bot.ApplicationAvatar,
					Shards:            shards,
					Environment:       bot.Environment,
				})
			}

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok: true,
				Data: GuildCustomBotResponse{
					Limit: welcomer.GetGuildCustomBotLimit(ctx, guildID),
					Bots:  bots,
				},
			})
		})
	})
}

// Route POST /api/guild/:guildID/custom-bot/:botID
func postGuildCustomBot(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			payload := &GuildCustomBotPayload{}

			var err error

			err = ctx.BindJSON(&payload)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			guildID := tryGetGuildID(ctx)

			var customBotUUID uuid.UUID

			var currentUser *discord.User

			var encryptedToken string

			if payload.PublicKey != "" {
				if !welcomer.IsValidPublicKey(payload.PublicKey) {
					welcomer.Logger.Warn().Msg("Invalid public key format")

					ctx.JSON(http.StatusBadRequest, NewBaseResponse(NewInvalidParameterError("publicKey"), nil))

					return
				}
			}

			if payload.Token != "" {
				if !welcomer.IsValidDiscordToken(payload.Token) {
					welcomer.Logger.Warn().Msg("Invalid Discord token format")

					ctx.JSON(http.StatusBadRequest, NewBaseResponse(NewInvalidParameterError("token"), nil))

					return
				}

				currentUser, err = discord.GetCurrentUser(ctx, discord.NewSession("Bot "+payload.Token, welcomer.RESTInterface))
				if err != nil {
					welcomer.Logger.Error().Err(err).Str("botID", customBotUUID.String()).Msg("Failed to fetch current user")

					ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

					return
				}

				encryptedToken, err = welcomer.EncryptBotToken(payload.Token, customBotUUID)
				if err != nil {
					welcomer.Logger.Error().Err(err).Str("botID", customBotUUID.String()).Msg("Failed to encrypt bot token")

					ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

					return
				}
			}

			var existingCustomBot database.GetCustomBotByIdRow

			botID := ctx.Param("botID")
			if botID != "" {
				customBotUUID, err = uuid.FromString(botID)
				if err != nil {
					ctx.JSON(http.StatusBadRequest, NewBaseResponse(NewInvalidParameterError("botID"), nil))

					return
				}

				existingCustomBotPtr, err := welcomer.Queries.GetCustomBotById(ctx, database.GetCustomBotByIdParams{
					CustomBotUuid: customBotUUID,
					GuildID:       int64(guildID),
				})

				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					welcomer.Logger.Warn().Err(err).Str("botID", customBotUUID.String()).Int64("guild_id", int64(guildID)).Msg("Failed to get custom bot")

					ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

					return
				}

				if errors.Is(err, pgx.ErrNoRows) {
					welcomer.Logger.Warn().Str("botID", customBotUUID.String()).Int64("guild_id", int64(guildID)).Msg("Custom bot not found")

					ctx.JSON(http.StatusNotFound, NewBaseResponse(nil, nil))

					return
				}

				existingCustomBot = *existingCustomBotPtr
			} else {
				customBotCount, err := GetGuildCustomBotCount(ctx, guildID)
				if err != nil {
					welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get custom bot count")

					ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

					return
				}

				if customBotCount >= welcomer.GetGuildCustomBotLimit(ctx, guildID) {
					welcomer.Logger.Warn().Int64("guild_id", int64(guildID)).Msg("Custom bot limit reached")

					ctx.JSON(http.StatusBadRequest, NewBaseResponse(ErrCustomBotLimitReached, nil))

					return
				}

				customBotUUID = uuid.Must(gen.NewV7())

				_, err = welcomer.Queries.CreateCustomBot(ctx, database.CreateCustomBotParams{
					CustomBotUuid:     customBotUUID,
					GuildID:           int64(guildID),
					Token:             encryptedToken,
					IsActive:          true,
					ApplicationID:     int64(currentUser.ID),
					ApplicationName:   welcomer.GetUserDisplayName(currentUser),
					ApplicationAvatar: currentUser.Avatar,
					Environment:       welcomer.GetCustomBotEnvironmentType(),
				})
				if err != nil {
					welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create custom bot")

					ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

					return
				}
			}

			err = UpdateIntegrationPublicKeys(ctx)
			if err != nil {
				welcomer.Logger.Error().Err(err).Msg("Failed to update integration public keys")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			if payload.Token != "" {
				_, err = welcomer.Queries.UpdateCustomBotToken(ctx, database.UpdateCustomBotTokenParams{
					CustomBotUuid:     customBotUUID,
					PublicKey:         payload.PublicKey,
					Token:             encryptedToken,
					IsActive:          true,
					ApplicationID:     int64(currentUser.ID),
					ApplicationName:   welcomer.GetUserDisplayName(currentUser),
					ApplicationAvatar: currentUser.Avatar,
					Environment:       welcomer.GetCustomBotEnvironmentType(),
				})
				if err != nil {
					welcomer.Logger.Warn().Err(err).Str("botID", customBotUUID.String()).Int64("guild_id", int64(guildID)).Msg("Failed to update custom bot")

					ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

					return
				}
			} else {
				_, err = welcomer.Queries.UpdateCustomBot(ctx, database.UpdateCustomBotParams{
					CustomBotUuid:     customBotUUID,
					PublicKey:         payload.PublicKey,
					IsActive:          true,
					ApplicationID:     existingCustomBot.ApplicationID,
					ApplicationName:   existingCustomBot.ApplicationName,
					ApplicationAvatar: existingCustomBot.ApplicationAvatar,
					Environment:       welcomer.GetCustomBotEnvironmentType(),
				})
				if err != nil {
					welcomer.Logger.Warn().Err(err).Str("botID", customBotUUID.String()).Int64("guild_id", int64(guildID)).Msg("Failed to update custom bot")

					ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

					return
				}
			}

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok: true,
			})
		})
	})
}

// ROUTE DELETE /api/guild/:guildID/custom-bot/:botID
func deleteGuildCustomBot(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			botID := ctx.Param("botID")
			if botID == "" {
				ctx.JSON(http.StatusBadRequest, NewMissingParameterError("botID"))

				return
			}

			customBotUUID, err := uuid.FromString(botID)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, NewBaseResponse(NewInvalidParameterError("botID"), nil))

				return
			}

			_, err = welcomer.Queries.GetCustomBotById(ctx, database.GetCustomBotByIdParams{
				CustomBotUuid: customBotUUID,
				GuildID:       int64(guildID),
			})

			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				welcomer.Logger.Warn().Err(err).Str("botID", customBotUUID.String()).Int64("guild_id", int64(guildID)).Msg("Failed to get custom bot")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			err = welcomer.StopCustomBot(ctx, customBotUUID, true)
			if err != nil && err.Error() != "application not found" {
				welcomer.Logger.Error().Err(err).Str("botID", customBotUUID.String()).Int64("guild_id", int64(guildID)).Msg("Failed to stop custom bot")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			_, err = welcomer.Queries.DeleteCustomBot(ctx, customBotUUID)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Str("botID", customBotUUID.String()).Int64("guild_id", int64(guildID)).Msg("Failed to delete custom bot")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			err = UpdateIntegrationPublicKeys(ctx)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Msg("Failed to update integration public keys")
			}

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok: true,
			})
		})
	})
}

// Route POST /api/guild/:guildID/custom-bot/start
func postGuildCustomBotStart(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			botID := ctx.Param("botID")
			if botID == "" {
				ctx.JSON(http.StatusBadRequest, NewMissingParameterError("botID"))

				return
			}

			customBotUUID, err := uuid.FromString(botID)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, NewBaseResponse(NewInvalidParameterError("botID"), nil))

				return
			}

			_, err = welcomer.Queries.GetCustomBotById(ctx, database.GetCustomBotByIdParams{
				CustomBotUuid: customBotUUID,
				GuildID:       int64(guildID),
			})

			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				welcomer.Logger.Warn().Err(err).Str("botID", customBotUUID.String()).Int64("guild_id", int64(guildID)).Msg("Failed to get custom bot")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			customBot, err := welcomer.Queries.GetCustomBotByIdWithToken(ctx, database.GetCustomBotByIdWithTokenParams{
				CustomBotUuid: customBotUUID,
				GuildID:       int64(guildID),
			})
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				welcomer.Logger.Error().Err(err).Str("botID", customBotUUID.String()).Int64("guild_id", int64(guildID)).Msg("Failed to get custom bot")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			if customBot == nil {
				welcomer.Logger.Error().Str("botID", customBotUUID.String()).Int64("guild_id", int64(guildID)).Msg("Custom bot not found")

				ctx.JSON(http.StatusNotFound, NewBaseResponse(nil, nil))

				return
			}

			if customBot.Token == "" {
				welcomer.Logger.Error().Str("botID", customBotUUID.String()).Int64("guild_id", int64(guildID)).Msg("Custom bot token is empty")

				ctx.JSON(http.StatusBadRequest, NewBaseResponse(ErrInvalidToken, nil))

				return
			}

			decryptedBotToken, err := welcomer.DecryptBotToken(customBot.Token, customBotUUID)
			if err != nil {
				welcomer.Logger.Error().Err(err).Str("botID", customBotUUID.String()).Int64("guild_id", int64(guildID)).Msg("Failed to decrypt custom bot token")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			err = UpdateInteractionCommands(ctx, decryptedBotToken, discord.Snowflake(customBot.ApplicationID))
			if err != nil {
				welcomer.Logger.Error().Err(err).Str("botID", customBotUUID.String()).Int64("guild_id", int64(guildID)).Msg("Failed to update interaction commands")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			err = welcomer.StartCustomBot(ctx, customBotUUID, decryptedBotToken, guildID, true)
			if err != nil {
				welcomer.Logger.Error().Err(err).Str("botID", customBotUUID.String()).Int64("guild_id", int64(guildID)).Msg("Failed to start custom bot")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok: true,
			})
		})
	})
}

// ROUTE POST /api/guild/:guildID/custom-bot/stop
func postGuildCustomBotStop(ctx *gin.Context) {
	requireGuildElevation(ctx, func(ctx *gin.Context) {
		guildID := tryGetGuildID(ctx)

		botID := ctx.Param("botID")
		if botID == "" {
			ctx.JSON(http.StatusBadRequest, NewMissingParameterError("botID"))

			return
		}

		customBotUUID, err := uuid.FromString(botID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, NewBaseResponse(NewInvalidParameterError("botID"), nil))

			return
		}

		_, err = welcomer.Queries.GetCustomBotById(ctx, database.GetCustomBotByIdParams{
			CustomBotUuid: customBotUUID,
			GuildID:       int64(guildID),
		})

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			welcomer.Logger.Warn().Err(err).Str("botID", customBotUUID.String()).Int64("guild_id", int64(guildID)).Msg("Failed to get custom bot")

			ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

			return
		}

		err = welcomer.StopCustomBot(ctx, customBotUUID, true)
		if err != nil {
			welcomer.Logger.Error().Err(err).Str("botID", customBotUUID.String()).Int64("guild_id", int64(guildID)).Msg("Failed to stop custom bot")

			ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

			return
		}

		ctx.JSON(http.StatusOK, BaseResponse{
			Ok: true,
		})
	})
}

func registerGuildCustomBotRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/custom-bots", getGuildCustomBots)
	g.POST("/api/guild/:guildID/custom-bots/", postGuildCustomBot)
	g.DELETE("/api/guild/:guildID/custom-bots/:botID", deleteGuildCustomBot)
	g.POST("/api/guild/:guildID/custom-bots/:botID", postGuildCustomBot)
	g.POST("/api/guild/:guildID/custom-bots/:botID/start", postGuildCustomBotStart)
	g.POST("/api/guild/:guildID/custom-bots/:botID/stop", postGuildCustomBotStop)
}

func UpdateInteractionCommands(ctx context.Context, token string, applicationID discord.Snowflake) error {
	body := struct {
		Token         string            `json:"token"`
		ApplicationID discord.Snowflake `json:"application_id"`
	}{token, applicationID}

	bodyBuffer := bytes.NewBuffer(nil)

	err := json.NewEncoder(bodyBuffer).Encode(body)
	if err != nil {
		return fmt.Errorf("failed to encode request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, os.Getenv("WELCOMER_INTEGRATIONS_ADDRESS")+"/internal/sync-commands", bodyBuffer)
	if err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to update integration commands")

		return fmt.Errorf("failed to update integration commands: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to update integration commands")

		return fmt.Errorf("failed to update integration commands: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		welcomer.Logger.Error().Str("status", resp.Status).Str("body", string(body)).Msg("Failed to update integration commands")

		return fmt.Errorf("failed to update integration commands: %s", string(body))
	}

	println(resp.StatusCode)

	welcomer.Logger.Info().Str("application_id", applicationID.String()).Msg("Successfully updated integration commands")

	return nil
}

func UpdateIntegrationPublicKeys(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, os.Getenv("WELCOMER_INTEGRATIONS_ADDRESS")+"/internal/fetch-public-keys", nil)
	if err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to update integration public keys")

		return fmt.Errorf("failed to update integration public keys: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to update integration public keys")

		return fmt.Errorf("failed to update integration public keys: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		welcomer.Logger.Error().Str("status", resp.Status).Str("body", string(body)).Msg("Failed to update integration public keys")

		return fmt.Errorf("failed to update integration public keys: %s", string(body))
	}

	welcomer.Logger.Info().Msg("Successfully updated integration public keys")

	return nil
}
