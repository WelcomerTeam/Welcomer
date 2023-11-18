package backend

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

// Route GET /api/guild/:guildID
func getGuild(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireMutualGuild(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			welcomerPresence, discordGuild, guildMembers, err := hasWelcomerPresence(guildID, true)
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to check welcomer presence")
			}

			if !welcomerPresence {
				ctx.JSON(http.StatusForbidden, BaseResponse{
					Ok:    false,
					Error: ErrWelcomerMissing.Error(),
					Data:  nil,
				})

				return
			}

			grpcContext := backend.GetBasicEventContext().ToGRPCContext()

			channels, err := backend.GRPCInterface.FetchChannelsByName(grpcContext, guildID, "")
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to fetch guild channels")
			}

			sort.SliceStable(channels, func(i, j int) bool {
				return channels[i].Position < channels[j].Position
			})

			roles, err := backend.GRPCInterface.FetchRolesByName(grpcContext, guildID, "")
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to fetch guild roles")
			}

			sort.SliceStable(roles, func(i, j int) bool {
				return roles[i].Position < roles[j].Position
			})

			emojis, err := backend.GRPCInterface.FetchEmojisByName(grpcContext, guildID, "")
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to fetch guild emojis")
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

			hasWelcomerPro, hasCustomBackgrounds, err := getGuildMembership(discordGuild.ID)
			if err != nil {
				backend.Logger.Warn().Err(err).Int("guildID", int(discordGuild.ID)).Msg("Exception getting welcomer membership")
			}

			guildConfig, err := backend.Database.GetGuild(backend.ctx, int64(discordGuild.ID))
			if err != nil {
				backend.Logger.Warn().Err(err).Int("guildID", int(discordGuild.ID)).Msg("Exception getting guild settings")
			}

			partialGuild := GuildToPartial(discordGuild)
			partialGuild.Roles = CalculateRoleValues(partialGuild.Roles, guildMembers)

			guild := Guild{
				Guild: partialGuild,

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
