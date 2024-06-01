package plugins

import (
	"context"
	"errors"
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
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
		"Provides autoroles for the server.",
	)

	// Disable the autoroles module for DM channels.
	ruleGroup.DMPermission = &utils.False

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "enable",
		Description: "Enable autorole for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &utils.False,
		DefaultMemberPermission: utils.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsAutoRoles, err := queries.GetAutoRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsAutoRoles = &database.GuildSettingsAutoroles{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: database.DefaultAutoroles.ToggleEnabled,
							Roles:         database.DefaultAutoroles.Roles,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get autoroles guild settings")

						return nil, err
					}
				}

				guildSettingsAutoRoles.ToggleEnabled = true

				err = utils.RetryWithFallback(
					func() error {
						_, err = queries.CreateOrUpdateAutoRolesGuildSettings(ctx, database.CreateOrUpdateAutoRolesGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsAutoRoles.ToggleEnabled,
							Roles:         guildSettingsAutoRoles.Roles,
						})
						return err
					},
					func() error {
						return welcomer.EnsureGuild(ctx, queries, discord.Snowflake(*interaction.GuildID))
					},
					nil,
				)
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update autoroles guild settings")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: utils.NewEmbed("Enabled autoroles. Run `/autoroles list` to see the list of autoroles configured.", utils.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "disable",
		Description: "Disables autorole for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &utils.False,
		DefaultMemberPermission: utils.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsAutoRoles, err := queries.GetAutoRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsAutoRoles = &database.GuildSettingsAutoroles{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: database.DefaultAutoroles.ToggleEnabled,
							Roles:         database.DefaultAutoroles.Roles,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get autoroles guild settings")

						return nil, err
					}
				}

				guildSettingsAutoRoles.ToggleEnabled = false

				err = utils.RetryWithFallback(
					func() error {
						_, err = queries.CreateOrUpdateAutoRolesGuildSettings(ctx, database.CreateOrUpdateAutoRolesGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsAutoRoles.ToggleEnabled,
							Roles:         guildSettingsAutoRoles.Roles,
						})
						return err
					},
					func() error {
						return welcomer.EnsureGuild(ctx, queries, discord.Snowflake(*interaction.GuildID))
					},
					nil,
				)
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update autoroles guild settings")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: utils.NewEmbed("Disabled autoroles.", utils.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "list",
		Description: "List the autoroles for the server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission: &utils.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				queries := welcomer.GetQueriesFromContext(ctx)
				guildSettingsAutoRoles, err := queries.GetAutoRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsAutoRoles = &database.GuildSettingsAutoroles{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: database.DefaultAutoroles.ToggleEnabled,
							Roles:         database.DefaultAutoroles.Roles,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get autoroles guild settings")

						return nil, err
					}
				}

				embeds := []*discord.Embed{}
				embed := &discord.Embed{Title: "AutoRoles", Color: utils.EmbedColourInfo}

				roleList, err := welcomer.FilterAssignableRoles(ctx, sub.SandwichClient, sub.Logger, int64(*interaction.GuildID), int64(interaction.ApplicationID), guildSettingsAutoRoles.Roles)
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to filter assignable roles")

					return nil, err
				}

				if len(roleList) == 0 || !guildSettingsAutoRoles.ToggleEnabled {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("There are no autoroles set for this server.", utils.EmbedColourInfo),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				for _, roleID := range roleList {
					roleMessage := fmt.Sprintf("- <@&%d>\n", roleID)

					// If the embed content will go over 4000 characters then create a new embed and continue from that one.
					if len(embed.Description)+len(roleMessage) > 4000 {
						embeds = append(embeds, embed)
						embed = &discord.Embed{Color: utils.EmbedColourInfo}
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
