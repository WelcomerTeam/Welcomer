package backend

import (
	"errors"
	"net/http"

	discord "github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
)

// Route GET /api/guild/:guildID/reactionroles.
func getGuildSettingsReactionRoles(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			reactionroles, err := welcomer.Queries.GetReactionRolesGuildSettings(ctx, int64(guildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					reactionroles = &database.GuildSettingsReactionRoles{
						GuildID:       int64(guildID),
						ToggleEnabled: welcomer.DefaultReactionRoles.ToggleEnabled,
						ReactionRoles: welcomer.DefaultReactionRoles.ReactionRoles,
					}
				} else {
					welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild reaction roles settings")

					ctx.JSON(http.StatusInternalServerError, BaseResponse{
						Ok: false,
					})

					return
				}
			}

			partial := GuildSettingsReactionRolesSettingsToPartial(reactionroles)

			partial.ReactionRoles = append(partial.ReactionRoles, welcomer.GuildSettingsReactionRole{
				Enabled:         false,
				ChannelID:       1,
				MessageID:       2,
				IsSystemMessage: true,
				MessageEmbed: &discord.Embed{
					Title: "Select a role",
				},
				Type: welcomer.ReactionRoleTypeEmoji,
				Roles: []welcomer.ReactionRoleOption{
					{
						RoleID: 3,
						Emoji:  "😀",
						Name:   "Example Role",
					},
					{
						RoleID: 4,
						Emoji:  "🔥",
						Name:   "Example Role 2",
					},
				},
			})
			partial.ReactionRoles = append(partial.ReactionRoles, welcomer.GuildSettingsReactionRole{
				Enabled:         true,
				ChannelID:       1,
				MessageID:       2,
				IsSystemMessage: false,
				MessageEmbed:    &discord.Embed{},
				Type:            welcomer.ReactionRoleTypeDropdown,
				Roles: []welcomer.ReactionRoleOption{
					{
						RoleID:      3,
						Emoji:       "😀",
						Name:        "Example Role",
						Description: "This is an example role description.",
					},
					{
						RoleID: 4,
						Emoji:  "🔥",
						Name:   "Example Role 2",
					},
				},
			})
			partial.ReactionRoles = append(partial.ReactionRoles, welcomer.GuildSettingsReactionRole{
				Enabled:         true,
				ChannelID:       1,
				MessageID:       2,
				IsSystemMessage: false,
				MessageEmbed:    &discord.Embed{},
				Type:            welcomer.ReactionRoleTypeButtons,
				Roles: []welcomer.ReactionRoleOption{
					{
						Style:       discord.InteractionComponentStylePrimary,
						RoleID:      3,
						Emoji:       "😀",
						Name:        "Example Role",
						Description: "This is an example role description.",
					},
					{
						Style:  discord.InteractionComponentStyleSuccess,
						RoleID: 4,
						Emoji:  "🔥",
						Name:   "Example Role 2",
					},
					{
						RoleID: 5,
						Emoji:  "🚀",
						Name:   "Example Role 3",
					},
				},
			})

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: partial,
			})
		})
	})
}

// Route POST /api/guild/:guildID/reactionroles.
func setGuildSettingsReactionRoles(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			partial := &GuildSettingsReactionRoles{}

			var err error

			err = ctx.BindJSON(partial)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			err = doValidateReactionRoles(partial)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			guildID := tryGetGuildID(ctx)

			reactionroles := PartialToGuildSettingsReactionRolesSettings(int64(guildID), partial)

			databaseReactionRolesGuildSettings := database.CreateOrUpdateReactionRolesGuildSettingsParams(*reactionroles)

			user := tryGetUser(ctx)
			welcomer.Logger.Info().Int64("guild_id", int64(guildID)).Interface("obj", *reactionroles).Int64("user_id", int64(user.ID)).Msg("Creating or updating reaction roles guild settings")

			err = welcomer.RetryWithFallback(
				func() error {
					_, err = welcomer.CreateOrUpdateReactionRolesGuildSettingsWithAudit(ctx, databaseReactionRolesGuildSettings, user.ID)

					return err
				},
				func() error {
					return welcomer.EnsureGuild(ctx, guildID)
				},
				nil,
			)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Int64("user_id", int64(user.ID)).Msg("Failed to create or update guild reaction roles settings")

				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok: false,
				})

				return
			}

			getGuildSettingsReactionRoles(ctx)
		})
	})
}

func tryGetSnowflakeFromCtx(ctx *gin.Context, key string) discord.Snowflake {
	rawSnowflake := ctx.Param(key)

	snowflakeInt, err := welcomer.Atoi(rawSnowflake)
	if err != nil {
		return 0
	}

	return discord.Snowflake(snowflakeInt)
}

// Route GET /api/guild/:guildID/checkmessage/:channelID/:messageID.
func checkGuildSettingsReactionRolesMessage(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)
			channelID := tryGetSnowflakeFromCtx(ctx, "channelID")
			messageID := tryGetSnowflakeFromCtx(ctx, "messageID")

			channel, err := welcomer.SandwichClient.FetchGuildChannel(ctx, &sandwich.FetchGuildChannelRequest{
				GuildId:    int64(guildID),
				ChannelIds: []int64{int64(channelID)},
			})
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Int64("channel_id", int64(channelID)).Msg("Failed to fetch channel for reaction roles")

				ctx.JSON(http.StatusInternalServerError, NewGenericErrorWithLineNumber())

				return
			}

			if len(channel.Channels) == 0 {
				welcomer.Logger.Warn().Int64("guild_id", int64(guildID)).Int64("channel_id", int64(channelID)).Msg("Channel not found for reaction roles")

				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: "The channel specified was not found in this server.",
				})

				return
			}

			message, err := discord.GetChannelMessage(ctx, backend.BotSession, channelID, messageID)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Int64("channel_id", int64(channelID)).Int64("message_id", int64(messageID)).Msg("Failed to get message for reaction roles")

				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: "Failed to get message. Make sure the message exists and the bot has access to it.",
				})

				return
			}

			emoji := "🎉"

			err = message.AddReaction(ctx, backend.BotSession, emoji)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Int64("channel_id", int64(channelID)).Int64("message_id", int64(messageID)).Msg("Failed to add reaction to message for reaction roles")

				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: "Failed to add reaction to message. Make sure the bot has permission to add reactions in the specified channel.",
				})

				return
			}

			_ = discord.DeleteOwnReaction(ctx, backend.BotSession, message.ChannelID, message.ID, emoji)

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok: true,
			})
		})
	})
}

// Validate reaction role settings.
func doValidateReactionRoles(partial *GuildSettingsReactionRoles) error {
	// TODO: validate reaction roles

	return nil
}

func registerGuildSettingsReactionRolesRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/reactionroles", getGuildSettingsReactionRoles)
	g.POST("/api/guild/:guildID/reactionroles", setGuildSettingsReactionRoles)

	g.POST("/api/guild/:guildID/checkmessage/:channelID/:messageID", checkGuildSettingsReactionRolesMessage)
}
