package plugins

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

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
	freerolesGroup := subway.NewSubcommandGroup(
		"freeroles",
		"Provides a set of roles that users can assign to themselves at any time.",
	)

	// Disable the freeroles module for DM channels.
	freerolesGroup.DMPermission = &welcomer.False

	freerolesGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "enable",
		Description: "Enable freerole for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(welcomer.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				guildSettingsFreeRoles, err := welcomer.Queries.GetFreeRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsFreeRoles = &database.GuildSettingsFreeroles{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultFreeRoles.ToggleEnabled,
							Roles:         welcomer.DefaultFreeRoles.Roles,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get freeroles guild settings")

						return nil, err
					}
				}

				guildSettingsFreeRoles.ToggleEnabled = true

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.CreateOrUpdateFreeRolesGuildSettingsWithAudit(ctx, database.CreateOrUpdateFreeRolesGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsFreeRoles.ToggleEnabled,
							Roles:         guildSettingsFreeRoles.Roles,
						}, interaction.GetUser().ID)

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
						Msg("Failed to update freeroles guild settings")

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

	freerolesGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "disable",
		Description: "Disables freerole for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(welcomer.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				guildSettingsFreeRoles, err := welcomer.Queries.GetFreeRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsFreeRoles = &database.GuildSettingsFreeroles{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultFreeRoles.ToggleEnabled,
							Roles:         welcomer.DefaultFreeRoles.Roles,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get freeroles guild settings")

						return nil, err
					}
				}

				guildSettingsFreeRoles.ToggleEnabled = false

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.CreateOrUpdateFreeRolesGuildSettingsWithAudit(ctx, database.CreateOrUpdateFreeRolesGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsFreeRoles.ToggleEnabled,
							Roles:         guildSettingsFreeRoles.Roles,
						}, interaction.GetUser().ID)

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
						Msg("Failed to update freeroles guild settings")

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

	freerolesGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "list",
		Description: "List the freeroles for the server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission: &welcomer.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
				guildSettingsFreeRoles, err := welcomer.Queries.GetFreeRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsFreeRoles = &database.GuildSettingsFreeroles{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultFreeRoles.ToggleEnabled,
							Roles:         welcomer.DefaultFreeRoles.Roles,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get freeroles guild settings")

						return nil, err
					}
				}

				roleList, err := welcomer.FilterAssignableRolesAsSnowflakes(ctx, sub.SandwichClient, int64(*interaction.GuildID), int64(interaction.ApplicationID), guildSettingsFreeRoles.Roles)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to filter assignable roles")

					return nil, err
				}

				if !guildSettingsFreeRoles.ToggleEnabled {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Freeroles are not enabled for this server.", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				if len(roleList) == 0 {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("There are no freeroles set for this server.", welcomer.EmbedColourInfo),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				embeds := []discord.Embed{}
				embed := discord.Embed{Title: "FreeRoles", Color: welcomer.EmbedColourInfo}

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

	freerolesGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "give",
		Description: "Gives a freerole.",

		AutocompleteHandler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) ([]discord.ApplicationCommandOptionChoice, error) {
			autocompleteRole := subway.MustGetArgument(ctx, "role").MustString()

			guildSettingsFreeRoles, err := welcomer.Queries.GetFreeRolesGuildSettings(ctx, int64(*interaction.GuildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					guildSettingsFreeRoles = &database.GuildSettingsFreeroles{
						GuildID:       int64(*interaction.GuildID),
						ToggleEnabled: welcomer.DefaultFreeRoles.ToggleEnabled,
						Roles:         welcomer.DefaultFreeRoles.Roles,
					}
				} else {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get freeroles guild settings")

					return nil, err
				}
			}

			// Check if freeroles are enabled.
			if !guildSettingsFreeRoles.ToggleEnabled {
				return nil, nil
			}

			roleList, err := welcomer.FilterAssignableRoles(ctx, sub.SandwichClient, int64(*interaction.GuildID), int64(interaction.ApplicationID), guildSettingsFreeRoles.Roles)
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(*interaction.GuildID)).
					Msg("Failed to filter assignable roles")

				return nil, err
			}

			choices := make([]discord.ApplicationCommandOptionChoice, 0, len(roleList))

			isAutocompleteRoleNumber := false
			if _, err := strconv.Atoi(autocompleteRole); err == nil {
				isAutocompleteRoleNumber = true
			}

			for _, role := range roleList {
				if autocompleteRole != "" {
					if isAutocompleteRoleNumber {
						if !strings.Contains(role.ID.String(), autocompleteRole) {
							continue
						}
					} else {
						if !welcomer.CompareStrings(role.Name, autocompleteRole) {
							continue
						}
					}
				}

				choices = append(choices, discord.ApplicationCommandOptionChoice{
					Name:  welcomer.Overflow(role.Name, 100),
					Value: welcomer.StringToJsonLiteral(role.ID.String()),
				})
			}

			if len(choices) > 25 {
				choices = choices[:25]
			}

			return choices, nil
		},

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeString,
				Name:         "role",
				Description:  "The role to give.",
				Autocomplete: &welcomer.True,
			},
		},

		DMPermission: &welcomer.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			roleID := subway.MustGetArgument(ctx, "role").MustString()

			roleIDInt64, err := welcomer.Atoi(roleID)
			if err != nil {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Invalid role ID.", welcomer.EmbedColourError),
						Flags:  uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			}

			role := discord.Role{ID: discord.Snowflake(roleIDInt64)}

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

			guildSettingsFreeRoles, err := welcomer.Queries.GetFreeRolesGuildSettings(ctx, int64(*interaction.GuildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					guildSettingsFreeRoles = &database.GuildSettingsFreeroles{
						GuildID:       int64(*interaction.GuildID),
						ToggleEnabled: welcomer.DefaultFreeRoles.ToggleEnabled,
						Roles:         welcomer.DefaultFreeRoles.Roles,
					}
				} else {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get freeroles guild settings")

					return nil, err
				}
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

			roleList, err := welcomer.FilterAssignableRolesAsSnowflakes(ctx, sub.SandwichClient, int64(*interaction.GuildID), int64(interaction.ApplicationID), guildSettingsFreeRoles.Roles)
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(*interaction.GuildID)).
					Msg("Failed to filter assignable roles")

				return nil, err
			}

			// Check if role.ID is in roleList
			if !slices.Contains(roleList, role.ID) {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("This role is not assignable.", welcomer.EmbedColourError),
						Flags:  uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			}

			session, err := welcomer.AcquireSession(ctx, welcomer.GetManagerNameFromContext(ctx))
			if err != nil {
				return nil, err
			}

			// GuildID may be missing, fill it in.
			interaction.Member.GuildID = interaction.GuildID

			err = interaction.Member.AddRoles(ctx, session,
				[]discord.Snowflake{role.ID},
				welcomer.ToPointer("Assigned with FreeRoles"),
				true,
			)
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

	freerolesGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "remove",
		Description: "Removes your freerole.",

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

			if !slices.Contains(interaction.Member.Roles, role.ID) {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("You do not have this role.", welcomer.EmbedColourInfo),
						Flags:  uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			}

			guildSettingsFreeRoles, err := welcomer.Queries.GetFreeRolesGuildSettings(ctx, int64(*interaction.GuildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					guildSettingsFreeRoles = &database.GuildSettingsFreeroles{
						GuildID:       int64(*interaction.GuildID),
						ToggleEnabled: welcomer.DefaultFreeRoles.ToggleEnabled,
						Roles:         welcomer.DefaultFreeRoles.Roles,
					}
				} else {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get freeroles guild settings")

					return nil, err
				}
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

			roleList, err := welcomer.FilterAssignableRolesAsSnowflakes(ctx, sub.SandwichClient, int64(*interaction.GuildID), int64(interaction.ApplicationID), guildSettingsFreeRoles.Roles)
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(*interaction.GuildID)).
					Msg("Failed to filter assignable roles")

				return nil, err
			}

			// Check if role.ID is in roleList
			if !slices.Contains(roleList, role.ID) {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("This role is not assignable.", welcomer.EmbedColourError),
						Flags:  uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			}

			session, err := welcomer.AcquireSession(ctx, welcomer.GetManagerNameFromContext(ctx))
			if err != nil {
				return nil, err
			}

			// GuildID may be missing, fill it in.
			interaction.Member.GuildID = interaction.GuildID

			err = interaction.Member.RemoveRoles(ctx, session,
				[]discord.Snowflake{role.ID},
				welcomer.ToPointer("Unassigned with FreeRoles"),
				false,
			)
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

	freerolesGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "addrole",
		Description: "Adds a role to the list of freeroles.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeRole,
				Name:         "role",
				Description:  "The role to add.",
			},
			{
				Required:     false,
				ArgumentType: subway.ArgumentTypeBool,
				Name:         "ignore-permissions",
				Description:  "Ignores role permissions.",
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(welcomer.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				role := subway.MustGetArgument(ctx, "role").MustRole()
				ignoreRolePermissions := subway.MustGetArgument(ctx, "ignore-permissions").MustBool()

				canAssignRoles, isRoleAssignable, isRoleElevated, err := welcomer.Accelerator_CanAssignRole(ctx, *interaction.GuildID, &role)
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
							Embeds: welcomer.NewEmbed("### This role is not assignable\nWelcomer cannot assign users this role as it does not have permission to manage roles or Welcomer's highest role is below this role's position. Please rearrange your roles in the server settings to move Welcomer's role above this role.", welcomer.EmbedColourError),
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

				guildSettingsFreeRoles, err := welcomer.Queries.GetFreeRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsFreeRoles = &database.GuildSettingsFreeroles{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultFreeRoles.ToggleEnabled,
							Roles:         welcomer.DefaultFreeRoles.Roles,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get freeroles guild settings")

						return nil, err
					}
				}

				// Add the role to the list of freeroles if not already present.
				for slices.Contains(guildSettingsFreeRoles.Roles, int64(role.ID)) {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("This role is already in the list.", welcomer.EmbedColourInfo),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				guildSettingsFreeRoles.Roles = append(guildSettingsFreeRoles.Roles, int64(role.ID))

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.CreateOrUpdateFreeRolesGuildSettingsWithAudit(ctx, database.CreateOrUpdateFreeRolesGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsFreeRoles.ToggleEnabled,
							Roles:         guildSettingsFreeRoles.Roles,
						}, interaction.GetUser().ID)

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
						Msg("Failed to update freeroles guild settings")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed(fmt.Sprintf("The role <@&%d> has been added to the list of freeroles.", role.ID), welcomer.EmbedColourSuccess),
						Flags:  uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			})
		},
	})

	freerolesGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "removerole",
		Description: "Removes a role from the list of freeroles.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeRole,
				Name:         "role",
				Description:  "The role to remove.",
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(welcomer.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				role := subway.MustGetArgument(ctx, "role").MustRole()

				guildSettingsFreeRoles, err := welcomer.Queries.GetFreeRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsFreeRoles = &database.GuildSettingsFreeroles{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultFreeRoles.ToggleEnabled,
							Roles:         welcomer.DefaultFreeRoles.Roles,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get freeroles guild settings")

						return nil, err
					}
				}

				// Check if the role is in the list of freeroles.
				if !slices.Contains(guildSettingsFreeRoles.Roles, int64(role.ID)) {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("This role is not in the list.", welcomer.EmbedColourInfo),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// Remove the role from the list of freeroles.
				guildSettingsFreeRoles.Roles = slices.DeleteFunc(guildSettingsFreeRoles.Roles, func(r int64) bool {
					return r == int64(role.ID)
				})

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.CreateOrUpdateFreeRolesGuildSettingsWithAudit(ctx, database.CreateOrUpdateFreeRolesGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsFreeRoles.ToggleEnabled,
							Roles:         guildSettingsFreeRoles.Roles,
						}, interaction.GetUser().ID)

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
						Msg("Failed to update freeroles guild settings")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed(fmt.Sprintf("The role <@&%d> has been removed from the list of freeroles.", role.ID), welcomer.EmbedColourSuccess),
						Flags:  uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			})
		},
	})

	r.InteractionCommands.MustAddInteractionCommand(freerolesGroup)

	return nil
}
