package backend

import (
	"errors"
	"net/http"
	"sort"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
)

// Route GET /api/guild/:guildID.
func getGuild(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireMutualGuild(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			welcomerPresence, discordGuild, guildMembers, err := hasWelcomerPresence(ctx, guildID, true)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to check welcomer presence")
			}

			if !welcomerPresence {
				ctx.JSON(http.StatusForbidden, BaseResponse{
					Ok:    false,
					Error: ErrWelcomerMissing.Error(),
					Data:  nil,
				})

				return
			}

			grpcContext := backend.GetBasicEventContext(ctx).ToGRPCContext()

			channels, err := welcomer.GRPCInterface.FetchChannelsByName(grpcContext, guildID, "")
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to fetch guild channels")
			}

			sort.SliceStable(channels, func(i, j int) bool {
				return channels[i].Position < channels[j].Position
			})

			roles, err := welcomer.GRPCInterface.FetchRolesByName(grpcContext, guildID, "")
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to fetch guild roles")
			}

			sort.SliceStable(roles, func(i, j int) bool {
				return roles[i].Position < roles[j].Position
			})

			emojis, err := welcomer.GRPCInterface.FetchEmojisByName(grpcContext, guildID, "")
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to fetch guild emojis")
			}

			for index, emoji := range emojis {
				emoji.User = nil
				emojis[index] = emoji
			}

			sort.SliceStable(emojis, func(i, j int) bool {
				return emojis[i].ID < emojis[j].ID
			})

			discordGuild.Channels = channels
			discordGuild.Roles = roles
			discordGuild.Emojis = emojis

			hasWelcomerPro, hasCustomBackgrounds, err := getGuildMembership(ctx, discordGuild.ID)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int("guildID", int(discordGuild.ID)).Msg("Exception getting welcomer membership")
			}

			guildConfig, err := welcomer.Queries.GetGuild(ctx, int64(discordGuild.ID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					guildConfig = &database.Guilds{
						GuildID:          int64(discordGuild.ID),
						EmbedColour:      welcomer.DefaultGuild.EmbedColour,
						SiteSplashUrl:    welcomer.DefaultGuild.SiteSplashUrl,
						SiteStaffVisible: welcomer.DefaultGuild.SiteStaffVisible,
						SiteGuildVisible: welcomer.DefaultGuild.SiteGuildVisible,
						SiteAllowInvites: welcomer.DefaultGuild.SiteAllowInvites,
					}
				} else {
					welcomer.Logger.Warn().Err(err).Int("guildID", int(discordGuild.ID)).Msg("Exception getting guild settings")

					ctx.JSON(http.StatusInternalServerError, BaseResponse{
						Ok: false,
					})

					return
				}
			}

			partialGuild := GuildToPartial(discordGuild)
			partialGuild.Roles = welcomer.CalculateRoleValues(roles, guildMembers)

			guild := Guild{
				Guild: &partialGuild,

				HasWelcomerPro:       hasWelcomerPro,
				HasCustomBackgrounds: hasCustomBackgrounds,

				EmbedColour:  int(guildConfig.EmbedColour),
				SplashURL:    guildConfig.SiteSplashUrl,
				StaffVisible: guildConfig.SiteStaffVisible,
				GuildVisible: guildConfig.SiteGuildVisible,
				AllowInvites: guildConfig.SiteAllowInvites,
			}

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: guild,
			})
		})
	})
}

func registerGuildRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID", getGuild)
}
