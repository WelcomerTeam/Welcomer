package plugins

import (
	"context"
	"errors"
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4"
)

func NewAutoRolesCog() *AutoRolesCog {
	return &AutoRolesCog{
		InteractionCommands: subway.SetupInteractionCommandable(&subway.InteractionCommandable{}),
	}
}

type AutoRolesCog struct {
	InteractionCommands *subway.InteractionCommandable
}

// Assert types.

var (
	_ subway.Cog                        = (*AutoRolesCog)(nil)
	_ subway.CogWithInteractionCommands = (*AutoRolesCog)(nil)
)

func (r *AutoRolesCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "AutoRoles",
		Description: "Provides the cog for the 'AutoRoles' feature.",
	}
}

func (r *AutoRolesCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return r.InteractionCommands
}

func (r *AutoRolesCog) RegisterCog(sub *subway.Subway) error {
	ruleGroup := subway.NewSubcommandGroup(
		"autoroles",
		"Provide autoroles for the server.",
	)

	// Disable the autoroles module for DM channels.
	ruleGroup.DMPermission = &welcomer.False

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "enable",
		Description: "Enable a rule module for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.IntToInt64Pointer(discord.PermissionElevated),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsAutoRoles, err := queries.GetAutoRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get autoroles guild settings.")
				}

				guildSettingsAutoRoles.GuildID = int64(*interaction.GuildID)
				guildSettingsAutoRoles.ToggleEnabled = true

				_, err = queries.UpdateAutoRolesGuildSettings(ctx, &database.UpdateAutoRolesGuildSettingsParams{
					GuildID:       guildSettingsAutoRoles.GuildID,
					ToggleEnabled: guildSettingsAutoRoles.ToggleEnabled,
					Roles:         guildSettingsAutoRoles.Roles,
				})
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update autoroles guild settings.")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Enabled autoroles. Run `/autoroles list` to see the list of autoroles configured.", welcomer.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "disable",
		Description: "Disables a rule module for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.IntToInt64Pointer(discord.PermissionElevated),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsAutoRoles, err := queries.GetAutoRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get autoroles guild settings.")
				}

				guildSettingsAutoRoles.GuildID = int64(*interaction.GuildID)

				guildSettingsAutoRoles.ToggleEnabled = false

				_, err = queries.UpdateAutoRolesGuildSettings(ctx, &database.UpdateAutoRolesGuildSettingsParams{
					GuildID:       guildSettingsAutoRoles.GuildID,
					ToggleEnabled: guildSettingsAutoRoles.ToggleEnabled,
					Roles:         guildSettingsAutoRoles.Roles,
				})
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update autoroles guild settings.")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Disabled auto roles.", welcomer.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "list",
		Description: "List the autoroles for the server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission: &welcomer.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
				queries := welcomer.GetQueriesFromContext(ctx)
				guildSettingsAutoRoles, err := queries.GetAutoRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get autoroles guild settings.")

					return nil, err
				}

				embeds := []*discord.Embed{}
				embed := &discord.Embed{Title: "AutoRoles", Color: welcomer.EmbedColourInfo}

				roleList := make([]int64, 0, len(guildSettingsAutoRoles.Roles))

				guildRoles, err := sub.SandwichClient.FetchGuildRoles(ctx, &sandwich.FetchGuildRolesRequest{
					GuildID: int64(*interaction.GuildID),
				})
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to fetch guild roles.")

					return nil, err
				}

				guildMember, err := sub.SandwichClient.FetchGuildMembers(ctx, &sandwich.FetchGuildMembersRequest{
					GuildID: int64(*interaction.GuildID),
					UserIDs: []int64{int64(interaction.ApplicationID)},
				})
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Int64("user_id", int64(interaction.ApplicationID)).
						Msg("Failed to fetch application guild member.")
				}

				// Get the guild member of the application.
				applicationUser, ok := guildMember.GuildMembers[int64(interaction.ApplicationID)]
				if !ok {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Int64("user_id", int64(interaction.ApplicationID)).
						Msg("Application guild member not present in response.")

					return nil, welcomer.ErrMissingApplicationUser
				}

				// Get the top role position of the application user.
				var applicationUserTopRolePosition int32
				for _, roleID := range applicationUser.Roles {
					role, ok := guildRoles.GuildRoles[roleID]
					println("User Roles", roleID, ok)
					if ok && role.Position > applicationUserTopRolePosition {
						applicationUserTopRolePosition = role.Position
					}
				}

				println("Top Role", applicationUserTopRolePosition)

				// Filter out any roles that are not in cache or are above the application user's top role position.
				for _, roleID := range guildSettingsAutoRoles.Roles {
					role, ok := guildRoles.GuildRoles[roleID]
					println("Lookup", roleID, ok)
					if ok {
						println(role.Name, role.Position, applicationUserTopRolePosition)
						if role.Position < applicationUserTopRolePosition {
							roleList = append(roleList, roleID)
						}
					}
				}

				if len(roleList) == 0 || !guildSettingsAutoRoles.ToggleEnabled {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("There are no autoroles set for this server.", welcomer.EmbedColourInfo),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				for _, roleID := range roleList {
					roleMessage := fmt.Sprintf("- <@&%d>\n", roleID)

					// If the embed content will go over 4000 characters then create a new embed and continue from that one.
					if len(embed.Description)+len(roleMessage) > 4000 {
						embeds = append(embeds, embed)
						embed = &discord.Embed{Color: welcomer.EmbedColourInfo}
					}

					embed.Description += roleMessage
				}

				embeds = append(embeds, embed)

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: embeds,
						Flags:  uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			})
		},
	})

	r.InteractionCommands.MustAddInteractionCommand(ruleGroup)

	return nil
}
