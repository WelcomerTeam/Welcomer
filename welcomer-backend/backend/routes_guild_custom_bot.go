package backend

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
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

var regexDiscordToken = regexp.MustCompile(`^[A-Za-z0-9_\-]{24,28}\.[A-Za-z0-9_\-]{6}\.[A-Za-z0-9_\-]{27,38}$`)

func isValidDiscordToken(token string) bool {
	return regexDiscordToken.MatchString(token)
}

func GetGuildCustomBotLimit(ctx context.Context, guildID discord.Snowflake) int {
	hasWelcomerPro, _, err := getGuildMembership(ctx, guildID)
	if err != nil {
		welcomer.Logger.Warn().Err(err).Int("guildID", int(guildID)).Msg("Exception getting welcomer membership")
	}

	if hasWelcomerPro {
		return 1
	}

	return 0
}

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
				var shards []GetStatusResponseShard

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
					ApplicationID:     discord.Snowflake(bot.ApplicationID),
					ApplicationName:   bot.ApplicationName,
					ApplicationAvatar: bot.ApplicationAvatar,
					Shards:            shards,
				})
			}

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok: true,
				Data: GuildCustomBotResponse{
					Limit: GetGuildCustomBotLimit(ctx, guildID),
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

			if payload.Token != "" && !isValidDiscordToken(payload.Token) {
				welcomer.Logger.Warn().Msg("Invalid Discord token format")

				ctx.JSON(http.StatusBadRequest, NewInvalidParameterError("token"))

				return
			}

			guildID := tryGetGuildID(ctx)

			var customBotUUID uuid.UUID

			currentUser, err := discord.GetCurrentUser(ctx, discord.NewSession("Bot "+payload.Token, welcomer.RESTInterface))
			if err != nil {
				welcomer.Logger.Error().Err(err).Str("botID", customBotUUID.String()).Msg("Failed to fetch current user")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			encryptedToken, err := welcomer.EncryptBotToken(payload.Token, customBotUUID)
			if err != nil {
				welcomer.Logger.Error().Err(err).Str("botID", customBotUUID.String()).Msg("Failed to encrypt bot token")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			botID := ctx.Param("botID")
			if botID != "" {
				customBotUUID, err = uuid.FromString(botID)
				if err != nil {
					ctx.JSON(http.StatusBadRequest, NewInvalidParameterError("botID"))

					return
				}
			} else {
				customBotCount, err := GetGuildCustomBotCount(ctx, guildID)
				if err != nil {
					welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get custom bot count")

					ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

					return
				}

				if customBotCount >= GetGuildCustomBotLimit(ctx, guildID) {
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
				})
				if err != nil {
					welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create custom bot")

					ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

					return
				}
			}

			// TODO: tell integrations to pull new public keys

			_, err = welcomer.Queries.UpdateCustomBotToken(ctx, database.UpdateCustomBotTokenParams{
				CustomBotUuid:     customBotUUID,
				PublicKey:         payload.PublicKey,
				Token:             encryptedToken,
				IsActive:          true,
				ApplicationID:     int64(currentUser.ID),
				ApplicationName:   welcomer.GetUserDisplayName(currentUser),
				ApplicationAvatar: currentUser.Avatar,
			})
			if err != nil {
				welcomer.Logger.Warn().Err(err).Str("botID", customBotUUID.String()).Int64("guild_id", int64(guildID)).Msg("Failed to update custom bot")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
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
				ctx.JSON(http.StatusBadRequest, NewInvalidParameterError("botID"))

				return
			}

			customBot, err := welcomer.Queries.GetCustomBotByIdWithToken(ctx, customBotUUID)
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

			err = welcomer.StartCustomBot(ctx, customBotUUID, decryptedBotToken, true)
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
			ctx.JSON(http.StatusBadRequest, NewInvalidParameterError("botID"))

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
	g.POST("/api/guild/:guildID/custom-bots/:botID", postGuildCustomBot)
	g.POST("/api/guild/:guildID/custom-bots/start", postGuildCustomBotStart)
	g.POST("/api/guild/:guildID/custom-bots/stop", postGuildCustomBotStop)
}
