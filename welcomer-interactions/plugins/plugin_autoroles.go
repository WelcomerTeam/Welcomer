package plugins

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/WelcomerTeam/Discord/discord"
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
	autorolesGroup := subway.NewSubcommandGroup(
		"autoroles",
		"Automatically give roles to users when they join the server.",
	)

	// Disable the autoroles module for DM channels.
	autorolesGroup.DMPermission = &welcomer.False

	autorolesGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "enable",
		Description: "Enable autorole for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				guildSettingsAutoRoles, err := welcomer.Queries.GetAutoRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsAutoRoles = &database.GuildSettingsAutoroles{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultAutoroles.ToggleEnabled,
							Roles:         welcomer.DefaultAutoroles.Roles,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get autoroles guild settings")

						return nil, err
					}
				}

				guildSettingsAutoRoles.ToggleEnabled = true

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.Queries.CreateOrUpdateAutoRolesGuildSettings(ctx, database.CreateOrUpdateAutoRolesGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsAutoRoles.ToggleEnabled,
							Roles:         guildSettingsAutoRoles.Roles,
						})

						return err
					},
					func() error {
						return welcomer.EnsureGuild(ctx, discord.Snowflake(*interaction.GuildID))
					},
					nil,
				)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update autoroles guild settings")

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

	autorolesGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "disable",
		Description: "Disables autorole for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				guildSettingsAutoRoles, err := welcomer.Queries.GetAutoRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsAutoRoles = &database.GuildSettingsAutoroles{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultAutoroles.ToggleEnabled,
							Roles:         welcomer.DefaultAutoroles.Roles,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get autoroles guild settings")

						return nil, err
					}
				}

				guildSettingsAutoRoles.ToggleEnabled = false

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.Queries.CreateOrUpdateAutoRolesGuildSettings(ctx, database.CreateOrUpdateAutoRolesGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsAutoRoles.ToggleEnabled,
							Roles:         guildSettingsAutoRoles.Roles,
						})

						return err
					},
					func() error {
						return welcomer.EnsureGuild(ctx, discord.Snowflake(*interaction.GuildID))
					},
					nil,
				)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update autoroles guild settings")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Disabled autoroles.", welcomer.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	autorolesGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "list",
		Description: "List the autoroles for the server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission: &welcomer.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				guildSettingsAutoRoles, err := welcomer.Queries.GetAutoRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsAutoRoles = &database.GuildSettingsAutoroles{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultAutoroles.ToggleEnabled,
							Roles:         welcomer.DefaultAutoroles.Roles,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get autoroles guild settings")

						return nil, err
					}
				}

				roleList, err := welcomer.FilterAssignableRolesAsSnowflakes(ctx, sub.SandwichClient, int64(*interaction.GuildID), int64(interaction.ApplicationID), guildSettingsAutoRoles.Roles)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to filter assignable roles")

					return nil, err
				}

				if len(roleList) == 0 {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("There are no autoroles set for this server.", welcomer.EmbedColourInfo),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				embeds := []discord.Embed{}
				embed := discord.Embed{Title: "AutoRoles", Color: welcomer.EmbedColourInfo}

				for _, roleID := range roleList {
					roleMessage := fmt.Sprintf("- <@&%d>\n", roleID)

					// If the embed content will go over 4000 characters then create a new embed and continue from that one.
					if len(embed.Description)+len(roleMessage) > 4000 {
						embeds = append(embeds, embed)
						embed = discord.Embed{Color: welcomer.EmbedColourInfo}
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

	autorolesGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "addrole",
		Description: "Add a role to the autoroles list.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Name:         "role",
				Description:  "The role to add to the autoroles list.",
				ArgumentType: subway.ArgumentTypeRole,
				Required:     true,
			},
			{
				Required:     false,
				ArgumentType: subway.ArgumentTypeBool,
				Name:         "ignore-permissions",
				Description:  "Ignores role permissions.",
			},
		},

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				role := subway.MustGetArgument(ctx, "role").MustRole()
				ignoreRolePermissions := subway.MustGetArgument(ctx, "ignore-permissions").MustBool()

				canAssignRoles, isRoleAssignable, isRoleElevated, err := welcomer.Accelerator_CanAssignRole(ctx, *interaction.GuildID, role)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to check if welcomer can assign role")

					return nil, err
				}

				if !canAssignRoles {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Welcomer is missing permissions to assign roles", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// Check if the role is assignable by welcomer using the guild roles and roles Welcomer has.
				if !isRoleAssignable {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("### This role is not assignable\nWelcomer cannot assign users this role as Welcomer's highest role is below this role's position. Please rearrange your roles in the server settings to move Welcomer's role above this role.", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				if !ignoreRolePermissions && isRoleElevated {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("### This role is elevated\nThis role has elevated permissions. If you are sure you want to use this role, please run the command again with ignore-permissions set to true.\n\nPermissions:\n"+welcomer.GetRolePermissionListAsString(int(role.Permissions)), welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				guildSettingsAutoRoles, err := welcomer.Queries.GetAutoRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsAutoRoles = &database.GuildSettingsAutoroles{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultAutoroles.ToggleEnabled,
							Roles:         welcomer.DefaultAutoroles.Roles,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get autoroles guild settings")

						return nil, err
					}
				}

				// Check if the role is already in the list.
				for _, r := range guildSettingsAutoRoles.Roles {
					if discord.Snowflake(r) == role.ID {
						return &discord.InteractionResponse{
							Type: discord.InteractionCallbackTypeChannelMessageSource,
							Data: &discord.InteractionCallbackData{
								Embeds: welcomer.NewEmbed("Role already in the autoroles list.", welcomer.EmbedColourError),
								Flags:  uint32(discord.MessageFlagEphemeral),
							},
						}, nil
					}
				}

				guildSettingsAutoRoles.Roles = append(guildSettingsAutoRoles.Roles, int64(role.ID))

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.Queries.CreateOrUpdateAutoRolesGuildSettings(ctx, database.CreateOrUpdateAutoRolesGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsAutoRoles.ToggleEnabled,
							Roles:         guildSettingsAutoRoles.Roles,
						})

						return err
					},
					func() error {
						return welcomer.EnsureGuild(ctx, discord.Snowflake(*interaction.GuildID))
					},
					nil,
				)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update autoroles guild settings")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed(fmt.Sprintf("Added <@&%d> to the autoroles list.", role.ID), welcomer.EmbedColourSuccess),
						Flags:  uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			})
		},
	})

	autorolesGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "removerole",
		Description: "Remove a role from the autoroles list.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Name:         "role",
				Description:  "The role to remove from the autoroles list.",
				ArgumentType: subway.ArgumentTypeRole,
				Required:     true,
			},
		},

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				role := subway.MustGetArgument(ctx, "role").MustRole()

				guildSettingsAutoRoles, err := welcomer.Queries.GetAutoRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsAutoRoles = &database.GuildSettingsAutoroles{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultAutoroles.ToggleEnabled,
							Roles:         welcomer.DefaultAutoroles.Roles,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get autoroles guild settings")

						return nil, err
					}
				}

				if !slices.Contains(guildSettingsAutoRoles.Roles, int64(role.ID)) {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("This role is not in the list.", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// Remove the role from the list.
				guildSettingsAutoRoles.Roles = slices.DeleteFunc(guildSettingsAutoRoles.Roles, func(r int64) bool {
					return r == int64(role.ID)
				})

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.Queries.CreateOrUpdateAutoRolesGuildSettings(ctx, database.CreateOrUpdateAutoRolesGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsAutoRoles.ToggleEnabled,
							Roles:         guildSettingsAutoRoles.Roles,
						})

						return err
					},
					func() error {
						return welcomer.EnsureGuild(ctx, discord.Snowflake(*interaction.GuildID))
					},
					nil,
				)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update autoroles guild settings")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed(fmt.Sprintf("Removed <@&%d> from the autoroles list.", role.ID), welcomer.EmbedColourSuccess),
						Flags:  uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			})
		},
	})

	r.InteractionCommands.MustAddInteractionCommand(autorolesGroup)

	return nil
}
