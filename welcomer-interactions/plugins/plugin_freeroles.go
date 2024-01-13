package plugins

import (
	"context"
	"errors"
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4"
)

func NewFreeRolesCog() *FreeRolesCog {
	return &FreeRolesCog{
		InteractionCommands: subway.SetupInteractionCommandable(&subway.InteractionCommandable{}),
	}
}

type FreeRolesCog struct {
	InteractionCommands *subway.InteractionCommandable
}

// Assert types.

var (
	_ subway.Cog                        = (*FreeRolesCog)(nil)
	_ subway.CogWithInteractionCommands = (*FreeRolesCog)(nil)
)

func (r *FreeRolesCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "FreeRoles",
		Description: "Provides the cog for the 'FreeRoles' feature.",
	}
}

func (r *FreeRolesCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return r.InteractionCommands
}

func (r *FreeRolesCog) RegisterCog(sub *subway.Subway) error {
	ruleGroup := subway.NewSubcommandGroup(
		"freeroles",
		"Provides freeroles for the server.",
	)

	// Disable the freeroles module for DM channels.
	ruleGroup.DMPermission = &welcomer.False

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "enable",
		Description: "Enable freerole for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.IntToInt64Pointer(discord.PermissionElevated),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsFreeRoles, err := queries.GetFreeRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get freeroles guild settings.")
				}

				guildSettingsFreeRoles.ToggleEnabled = true

				_, err = queries.UpdateFreeRolesGuildSettings(ctx, &database.UpdateFreeRolesGuildSettingsParams{
					GuildID:       int64(*interaction.GuildID),
					ToggleEnabled: guildSettingsFreeRoles.ToggleEnabled,
					Roles:         guildSettingsFreeRoles.Roles,
				})
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update freeroles guild settings.")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Enabled freeroles. Run `/freeroles list` to see the list of freeroles configured.", welcomer.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "disable",
		Description: "Disables freerole for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.IntToInt64Pointer(discord.PermissionElevated),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsFreeRoles, err := queries.GetFreeRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get freeroles guild settings.")
				}

				guildSettingsFreeRoles.ToggleEnabled = false

				_, err = queries.UpdateFreeRolesGuildSettings(ctx, &database.UpdateFreeRolesGuildSettingsParams{
					GuildID:       int64(*interaction.GuildID),
					ToggleEnabled: guildSettingsFreeRoles.ToggleEnabled,
					Roles:         guildSettingsFreeRoles.Roles,
				})
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update freeroles guild settings.")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Disabled freeroles.", welcomer.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "list",
		Description: "List the freeroles for the server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission: &welcomer.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsFreeRoles, err := queries.GetFreeRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get freeroles guild settings.")

					return nil, err
				}

				embeds := []*discord.Embed{}
				embed := &discord.Embed{Title: "FreeRoles", Color: welcomer.EmbedColourInfo}

				roleList, err := welcomer.FilterAssignableRoles(ctx, sub, interaction, guildSettingsFreeRoles.Roles)
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to filter assignable roles.")

					return nil, err
				}

				if len(roleList) == 0 || !guildSettingsFreeRoles.ToggleEnabled {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("There are no freeroles set for this server.", welcomer.EmbedColourInfo),
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

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "give",
		Description: "Gives a freerole.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeRole,
				Name:         "role",
				Description:  "The role to give.",
			},
		},

		DMPermission: &welcomer.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			role := subway.MustGetArgument(ctx, "role").MustRole()

			// Check if the user already has the role.
			for _, roleID := range interaction.Member.Roles {
				if roleID == role.ID {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("You already have this role.", welcomer.EmbedColourInfo),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}
			}

			queries := welcomer.GetQueriesFromContext(ctx)

			guildSettingsFreeRoles, err := queries.GetFreeRolesGuildSettings(ctx, int64(*interaction.GuildID))
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				sub.Logger.Error().Err(err).
					Int64("guild_id", int64(*interaction.GuildID)).
					Msg("Failed to get freeroles guild settings.")

				return nil, err
			}

			// Check if freeroles are enabled.
			if !guildSettingsFreeRoles.ToggleEnabled {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Freeroles are not enabled for this server.", welcomer.EmbedColourError),
						Flags:  uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			}

			roleList, err := welcomer.FilterAssignableRoles(ctx, sub, interaction, guildSettingsFreeRoles.Roles)
			if err != nil {
				sub.Logger.Error().Err(err).
					Int64("guild_id", int64(*interaction.GuildID)).
					Msg("Failed to filter assignable roles.")

				return nil, err
			}

			// Check if role.ID is in roleList
			found := false
			for _, roleID := range roleList {
				if int64(role.ID) == roleID {
					found = true

					break
				}
			}

			if !found {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("This role is not assignable.", welcomer.EmbedColourError),
						Flags:  uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			}

			session, err := welcomer.AcquireSession(ctx, sub, welcomer.GetManagerNameFromContext(ctx))
			if err != nil {
				return nil, err
			}

			// GuildID may be missing, fill it in.
			interaction.Member.GuildID = interaction.GuildID

			err = interaction.Member.AddRoles(session, []discord.Snowflake{role.ID}, welcomer.StringToPointer("Assigned with FreeRoles"), false)
			if err != nil {
				return nil, err
			}

			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Embeds: welcomer.NewEmbed(fmt.Sprintf("You have been assigned the role <@&%d>.", role.ID), welcomer.EmbedColourSuccess),
					Flags:  uint32(discord.MessageFlagEphemeral),
				},
			}, nil
		},
	})

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "remove",
		Description: "Removes a freerole.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeRole,
				Name:         "role",
				Description:  "The role to remove.",
			},
		},

		DMPermission: &welcomer.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			role := subway.MustGetArgument(ctx, "role").MustRole()

			hasRole := false

			// Check if the user has the role.
			for _, roleID := range interaction.Member.Roles {
				if roleID == role.ID {
					hasRole = true

					break
				}
			}

			if !hasRole {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("You do not have this role.", welcomer.EmbedColourInfo),
						Flags:  uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			}

			queries := welcomer.GetQueriesFromContext(ctx)

			guildSettingsFreeRoles, err := queries.GetFreeRolesGuildSettings(ctx, int64(*interaction.GuildID))
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				sub.Logger.Error().Err(err).
					Int64("guild_id", int64(*interaction.GuildID)).
					Msg("Failed to get freeroles guild settings.")

				return nil, err
			}

			// Check if freeroles are enabled.
			if !guildSettingsFreeRoles.ToggleEnabled {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Freeroles are not enabled for this server.", welcomer.EmbedColourError),
						Flags:  uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			}

			roleList, err := welcomer.FilterAssignableRoles(ctx, sub, interaction, guildSettingsFreeRoles.Roles)
			if err != nil {
				sub.Logger.Error().Err(err).
					Int64("guild_id", int64(*interaction.GuildID)).
					Msg("Failed to filter assignable roles.")

				return nil, err
			}

			// Check if role.ID is in roleList
			found := false
			for _, roleID := range roleList {
				if int64(role.ID) == roleID {
					found = true

					break
				}
			}

			if !found {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("This role is not assignable.", welcomer.EmbedColourError),
						Flags:  uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			}

			session, err := welcomer.AcquireSession(ctx, sub, welcomer.GetManagerNameFromContext(ctx))
			if err != nil {
				return nil, err
			}

			// GuildID may be missing, fill it in.
			interaction.Member.GuildID = interaction.GuildID

			err = interaction.Member.RemoveRoles(session, []discord.Snowflake{role.ID}, welcomer.StringToPointer("Unassigned with FreeRoles"), false)
			if err != nil {
				return nil, err
			}

			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Embeds: welcomer.NewEmbed(fmt.Sprintf("You have unassigned the role <@&%d>.", role.ID), welcomer.EmbedColourSuccess),
					Flags:  uint32(discord.MessageFlagEphemeral),
				},
			}, nil
		},
	})

	r.InteractionCommands.MustAddInteractionCommand(ruleGroup)

	return nil
}
