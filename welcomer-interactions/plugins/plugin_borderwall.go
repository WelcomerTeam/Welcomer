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

func NewBorderwallCog() *BorderwallCog {
	return &BorderwallCog{
		InteractionCommands: subway.SetupInteractionCommandable(&subway.InteractionCommandable{}),
	}
}

type BorderwallCog struct {
	InteractionCommands *subway.InteractionCommandable
}

// Assert types.

var (
	_ subway.Cog                        = (*BorderwallCog)(nil)
	_ subway.CogWithInteractionCommands = (*BorderwallCog)(nil)
)

const (
	BorderwallModuleBorderwall = "borderwall"
	BorderwallModuleDMs        = "dms"

	BorderwallRoleTypeJoin   = "join"
	BorderwallRoleTypeVerify = "verify"
)

func (b *BorderwallCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "Borderwall",
		Description: "Provides the functionality for the 'Borderwall' feature",
	}
}

func (b *BorderwallCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return b.InteractionCommands
}

func (b *BorderwallCog) RegisterCog(sub *subway.Subway) error {
	borderwallGroup := subway.NewSubcommandGroup(
		"borderwall",
		"Protect your server by verifying users when they join.",
	)

	// Disable the borderwall module for DM channels.
	borderwallGroup.DMPermission = &welcomer.False

	borderwallGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "enable",
		Description: "Enable a borderwall module for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeString,
				Name:         "module",
				Description:  "The module to enable.",

				Choices: []discord.ApplicationCommandOptionChoice{
					{Name: BorderwallModuleBorderwall, Value: welcomer.StringToJsonLiteral(BorderwallModuleBorderwall)},
					{Name: BorderwallModuleDMs, Value: welcomer.StringToJsonLiteral(BorderwallModuleDMs)},
				},
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				module := subway.MustGetArgument(ctx, "module").MustString()

				guildSettingsBorderwall, err := welcomer.Queries.GetBorderwallGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsBorderwall = &database.GuildSettingsBorderwall{
							GuildID:         int64(*interaction.GuildID),
							ToggleEnabled:   welcomer.DefaultBorderwall.ToggleEnabled,
							ToggleSendDm:    welcomer.DefaultBorderwall.ToggleSendDm,
							Channel:         welcomer.DefaultBorderwall.Channel,
							MessageVerify:   welcomer.DefaultBorderwall.MessageVerify,
							MessageVerified: welcomer.DefaultBorderwall.MessageVerified,
							RolesOnJoin:     welcomer.DefaultBorderwall.RolesOnJoin,
							RolesOnVerify:   welcomer.DefaultBorderwall.RolesOnVerify,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get borderwall guild settings")

						return nil, err
					}
				}

				switch module {
				case BorderwallModuleBorderwall:
					guildSettingsBorderwall.ToggleEnabled = true
				case BorderwallModuleDMs:
					guildSettingsBorderwall.ToggleSendDm = true
				default:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Unknown module: "+module, welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// A channel must be set if direct messages are disabled.
				if guildSettingsBorderwall.Channel == 0 && guildSettingsBorderwall.ToggleEnabled && !guildSettingsBorderwall.ToggleSendDm {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("A channel must be selected if you are not sending borderwall messages via direct message. Please set a channel with `/borderwall channel` before enabling borderwall.", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.Queries.CreateOrUpdateBorderwallGuildSettings(ctx, database.CreateOrUpdateBorderwallGuildSettingsParams{
							GuildID:         int64(*interaction.GuildID),
							ToggleEnabled:   guildSettingsBorderwall.ToggleEnabled,
							ToggleSendDm:    guildSettingsBorderwall.ToggleSendDm,
							Channel:         guildSettingsBorderwall.Channel,
							MessageVerify:   guildSettingsBorderwall.MessageVerify,
							MessageVerified: guildSettingsBorderwall.MessageVerified,
							RolesOnJoin:     guildSettingsBorderwall.RolesOnJoin,
							RolesOnVerify:   guildSettingsBorderwall.RolesOnVerify,
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
						Msg("Failed to update borderwall guild settings")

					return nil, err
				}

				switch module {
				case BorderwallModuleBorderwall:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Enabled borderwall. Users will now have to verify when joining the server.", welcomer.EmbedColourSuccess),
						},
					}, nil
				case BorderwallModuleDMs:
					if guildSettingsBorderwall.ToggleEnabled {
						return &discord.InteractionResponse{
							Type: discord.InteractionCallbackTypeChannelMessageSource,
							Data: &discord.InteractionCallbackData{
								Embeds: welcomer.NewEmbed("Enabled borderwall direct messages. Users will now receive instructions on how to verify with borderwall when joining the server.", welcomer.EmbedColourSuccess),
							},
						}, nil
					} else {
						return &discord.InteractionResponse{
							Type: discord.InteractionCallbackTypeChannelMessageSource,
							Data: &discord.InteractionCallbackData{
								Embeds: welcomer.NewEmbed("Enabled borderwall direct messages. Borderwall is not enabled, users won't have to verify when joining the server.", welcomer.EmbedColourWarn),
							},
						}, nil
					}
				}

				return nil, nil // Unreachable
			})
		},
	})

	borderwallGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "disable",
		Description: "Disables a borderwall module for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeString,
				Name:         "module",
				Description:  "The module to disable.",

				Choices: []discord.ApplicationCommandOptionChoice{
					{Name: BorderwallModuleBorderwall, Value: welcomer.StringToJsonLiteral(BorderwallModuleBorderwall)},
					{Name: BorderwallModuleDMs, Value: welcomer.StringToJsonLiteral(BorderwallModuleDMs)},
				},
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				module := subway.MustGetArgument(ctx, "module").MustString()

				guildSettingsBorderwall, err := welcomer.Queries.GetBorderwallGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsBorderwall = &database.GuildSettingsBorderwall{
							GuildID:         int64(*interaction.GuildID),
							ToggleEnabled:   welcomer.DefaultBorderwall.ToggleEnabled,
							ToggleSendDm:    welcomer.DefaultBorderwall.ToggleSendDm,
							Channel:         welcomer.DefaultBorderwall.Channel,
							MessageVerify:   welcomer.DefaultBorderwall.MessageVerify,
							MessageVerified: welcomer.DefaultBorderwall.MessageVerified,
							RolesOnJoin:     welcomer.DefaultBorderwall.RolesOnJoin,
							RolesOnVerify:   welcomer.DefaultBorderwall.RolesOnVerify,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get borderwall guild settings")

						return nil, err
					}
				}

				switch module {
				case BorderwallModuleBorderwall:
					guildSettingsBorderwall.ToggleEnabled = false
				case BorderwallModuleDMs:
					guildSettingsBorderwall.ToggleSendDm = false
				default:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Unknown module: "+module, welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.Queries.CreateOrUpdateBorderwallGuildSettings(ctx, database.CreateOrUpdateBorderwallGuildSettingsParams{
							GuildID:         int64(*interaction.GuildID),
							ToggleEnabled:   guildSettingsBorderwall.ToggleEnabled,
							ToggleSendDm:    guildSettingsBorderwall.ToggleSendDm,
							Channel:         guildSettingsBorderwall.Channel,
							MessageVerify:   guildSettingsBorderwall.MessageVerify,
							MessageVerified: guildSettingsBorderwall.MessageVerified,
							RolesOnJoin:     guildSettingsBorderwall.RolesOnJoin,
							RolesOnVerify:   guildSettingsBorderwall.RolesOnVerify,
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
						Msg("Failed to update borderwall guild settings")

					return nil, err
				}

				switch module {
				case BorderwallModuleBorderwall:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Disabled borderwall.", welcomer.EmbedColourSuccess),
						},
					}, nil
				case BorderwallModuleDMs:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Disabled borderwall direct messages.", welcomer.EmbedColourSuccess),
						},
					}, nil
				}

				return nil, nil // Unreachable
			})
		},
	})

	borderwallGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "setmessage",
		Description: "Configure the borderwall messages on the welcomer dashboard",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: []discord.Embed{
							{
								Description: fmt.Sprintf("Configure your borderwall verify and verified messages on our dashboard [**here**](%s).", welcomer.WebsiteURL+"/dashboard/"+interaction.GuildID.String()+"/borderwall"),
								Color:       welcomer.EmbedColourInfo,
							},
						},
					},
				}, nil
			})
		},
	})

	borderwallGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "setchannel",
		Description: "Sets the channel to send borderwall messages to.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     false,
				ArgumentType: subway.ArgumentTypeTextChannel,
				Name:         "channel",
				Description:  "The channel you would like to send the borderwall message to.",
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				channel := subway.MustGetArgument(ctx, "channel").MustChannel()

				guildSettingsBorderwall, err := welcomer.Queries.GetBorderwallGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsBorderwall = &database.GuildSettingsBorderwall{
							GuildID:         int64(*interaction.GuildID),
							ToggleEnabled:   welcomer.DefaultBorderwall.ToggleEnabled,
							ToggleSendDm:    welcomer.DefaultBorderwall.ToggleSendDm,
							Channel:         welcomer.DefaultBorderwall.Channel,
							MessageVerify:   welcomer.DefaultBorderwall.MessageVerify,
							MessageVerified: welcomer.DefaultBorderwall.MessageVerified,
							RolesOnJoin:     welcomer.DefaultBorderwall.RolesOnJoin,
							RolesOnVerify:   welcomer.DefaultBorderwall.RolesOnVerify,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get borderwall guild settings")

						return nil, err
					}
				}

				if !channel.ID.IsNil() {
					guildSettingsBorderwall.Channel = int64(channel.ID)
				} else {
					guildSettingsBorderwall.Channel = 0
				}

				// A channel must be set if direct messages are disabled.
				if guildSettingsBorderwall.Channel == 0 && guildSettingsBorderwall.ToggleEnabled && !guildSettingsBorderwall.ToggleSendDm {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("A channel must be selected if you are not sending borderwall messages via direct message.", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.Queries.CreateOrUpdateBorderwallGuildSettings(ctx, database.CreateOrUpdateBorderwallGuildSettingsParams{
							GuildID:         int64(*interaction.GuildID),
							ToggleEnabled:   guildSettingsBorderwall.ToggleEnabled,
							ToggleSendDm:    guildSettingsBorderwall.ToggleSendDm,
							Channel:         guildSettingsBorderwall.Channel,
							MessageVerify:   guildSettingsBorderwall.MessageVerify,
							MessageVerified: guildSettingsBorderwall.MessageVerified,
							RolesOnJoin:     guildSettingsBorderwall.RolesOnJoin,
							RolesOnVerify:   guildSettingsBorderwall.RolesOnVerify,
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
						Msg("Failed to update borderwall guild settings")

					return nil, err
				}

				if !channel.ID.IsNil() {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Set borderwall channel to: <#"+channel.ID.String()+">.", welcomer.EmbedColourSuccess),
						},
					}, nil
				} else {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Removed borderwall channel. Borderwall will only send direct messages to users.", welcomer.EmbedColourWarn),
						},
					}, nil
				}
			})
		},
	})

	borderwallGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "listroles",
		Description: "Lists roles given when joining or verifying with borderwall.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeString,
				Name:         "type",
				Description:  "The type of role to list.",

				Choices: []discord.ApplicationCommandOptionChoice{
					{Name: BorderwallRoleTypeJoin, Value: welcomer.StringToJsonLiteral(BorderwallRoleTypeJoin)},
					{Name: BorderwallRoleTypeVerify, Value: welcomer.StringToJsonLiteral(BorderwallRoleTypeVerify)},
				},
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				roleType := subway.MustGetArgument(ctx, "type").MustString()

				guildSettingsBorderwall, err := welcomer.Queries.GetBorderwallGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsBorderwall = &database.GuildSettingsBorderwall{
							GuildID:         int64(*interaction.GuildID),
							ToggleEnabled:   welcomer.DefaultBorderwall.ToggleEnabled,
							ToggleSendDm:    welcomer.DefaultBorderwall.ToggleSendDm,
							Channel:         welcomer.DefaultBorderwall.Channel,
							MessageVerify:   welcomer.DefaultBorderwall.MessageVerify,
							MessageVerified: welcomer.DefaultBorderwall.MessageVerified,
							RolesOnJoin:     welcomer.DefaultBorderwall.RolesOnJoin,
							RolesOnVerify:   welcomer.DefaultBorderwall.RolesOnVerify,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get borderwall guild settings")

						return nil, err
					}
				}

				var roles []int64

				switch roleType {
				case BorderwallRoleTypeJoin:
					roles = guildSettingsBorderwall.RolesOnJoin
				case BorderwallRoleTypeVerify:
					roles = guildSettingsBorderwall.RolesOnVerify
				default:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Unknown role type: "+roleType, welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				if len(roles) == 0 {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed(fmt.Sprintf("There are no borderwall %s roles set for this server.", roleType), welcomer.EmbedColourInfo),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				embeds := []discord.Embed{}
				embed := discord.Embed{Title: "Borderwall " + roleType + " roles", Color: welcomer.EmbedColourInfo}

				for _, roleID := range roles {
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
					},
				}, nil
			})
		},
	})

	borderwallGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "addroles",
		Description: "Adds roles to be given when joining or verifying with borderwall.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Name:         "type",
				Description:  "The type of role to add.",
				ArgumentType: subway.ArgumentTypeString,
				Required:     true,

				Choices: []discord.ApplicationCommandOptionChoice{
					{Name: BorderwallRoleTypeJoin, Value: welcomer.StringToJsonLiteral(BorderwallRoleTypeJoin)},
					{Name: BorderwallRoleTypeVerify, Value: welcomer.StringToJsonLiteral(BorderwallRoleTypeVerify)},
				},
			},
			{
				Name:         "role",
				Description:  "The role to add.",
				ArgumentType: subway.ArgumentTypeRole,
				Required:     true,
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				roleType := subway.MustGetArgument(ctx, "type").MustString()
				role := subway.MustGetArgument(ctx, "role").MustRole()

				canAssignRoles, isRoleAssignable, _, err := welcomer.Accelerator_CanAssignRole(ctx, *interaction.GuildID, role)
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
							Embeds: welcomer.NewEmbed("### This role is not assignable\nWelcomer will not be able to assign this role to users as Welcomer's highest role is below this role's position.", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				guildSettingsBorderwall, err := welcomer.Queries.GetBorderwallGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsBorderwall = &database.GuildSettingsBorderwall{
							GuildID:         int64(*interaction.GuildID),
							ToggleEnabled:   welcomer.DefaultBorderwall.ToggleEnabled,
							ToggleSendDm:    welcomer.DefaultBorderwall.ToggleSendDm,
							Channel:         welcomer.DefaultBorderwall.Channel,
							MessageVerify:   welcomer.DefaultBorderwall.MessageVerify,
							MessageVerified: welcomer.DefaultBorderwall.MessageVerified,
							RolesOnJoin:     welcomer.DefaultBorderwall.RolesOnJoin,
							RolesOnVerify:   welcomer.DefaultBorderwall.RolesOnVerify,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get borderwall guild settings")

						return nil, err
					}
				}

				switch roleType {
				case BorderwallRoleTypeJoin:
					// Check if the role is already in the list.
					if slices.Contains(guildSettingsBorderwall.RolesOnJoin, int64(role.ID)) {
						return &discord.InteractionResponse{
							Type: discord.InteractionCallbackTypeChannelMessageSource,
							Data: &discord.InteractionCallbackData{
								Embeds: welcomer.NewEmbed("This role is already in the list.", welcomer.EmbedColourError),
								Flags:  uint32(discord.MessageFlagEphemeral),
							},
						}, nil
					}

					guildSettingsBorderwall.RolesOnJoin = append(guildSettingsBorderwall.RolesOnJoin, int64(role.ID))
				case BorderwallRoleTypeVerify:
					// Check if the role is already in the list.
					if slices.Contains(guildSettingsBorderwall.RolesOnVerify, int64(role.ID)) {
						return &discord.InteractionResponse{
							Type: discord.InteractionCallbackTypeChannelMessageSource,
							Data: &discord.InteractionCallbackData{
								Embeds: welcomer.NewEmbed("This role is already in the list.", welcomer.EmbedColourError),
								Flags:  uint32(discord.MessageFlagEphemeral),
							},
						}, nil
					}

					guildSettingsBorderwall.RolesOnVerify = append(guildSettingsBorderwall.RolesOnVerify, int64(role.ID))
				default:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Unknown role type: "+roleType, welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.Queries.CreateOrUpdateBorderwallGuildSettings(ctx, database.CreateOrUpdateBorderwallGuildSettingsParams{
							GuildID:         int64(*interaction.GuildID),
							ToggleEnabled:   guildSettingsBorderwall.ToggleEnabled,
							ToggleSendDm:    guildSettingsBorderwall.ToggleSendDm,
							Channel:         guildSettingsBorderwall.Channel,
							MessageVerify:   guildSettingsBorderwall.MessageVerify,
							MessageVerified: guildSettingsBorderwall.MessageVerified,
							RolesOnJoin:     guildSettingsBorderwall.RolesOnJoin,
							RolesOnVerify:   guildSettingsBorderwall.RolesOnVerify,
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
						Msg("Failed to update borderwall guild settings")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed(fmt.Sprintf("Added role <@&%d> to borderwall %s roles.", role.ID, roleType), welcomer.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	borderwallGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "removeroles",
		Description: "Removes roles from being given when joining or verifying with borderwall.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Name:         "type",
				Description:  "The type of role to remove.",
				ArgumentType: subway.ArgumentTypeString,
				Required:     true,

				Choices: []discord.ApplicationCommandOptionChoice{
					{Name: BorderwallRoleTypeJoin, Value: welcomer.StringToJsonLiteral(BorderwallRoleTypeJoin)},
					{Name: BorderwallRoleTypeVerify, Value: welcomer.StringToJsonLiteral(BorderwallRoleTypeVerify)},
				},
			},
			{
				Name:         "role",
				Description:  "The role to remove.",
				ArgumentType: subway.ArgumentTypeRole,
				Required:     true,
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				roleType := subway.MustGetArgument(ctx, "type").MustString()
				role := subway.MustGetArgument(ctx, "role").MustRole()

				guildSettingsBorderwall, err := welcomer.Queries.GetBorderwallGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsBorderwall = &database.GuildSettingsBorderwall{
							GuildID:         int64(*interaction.GuildID),
							ToggleEnabled:   welcomer.DefaultBorderwall.ToggleEnabled,
							ToggleSendDm:    welcomer.DefaultBorderwall.ToggleSendDm,
							Channel:         welcomer.DefaultBorderwall.Channel,
							MessageVerify:   welcomer.DefaultBorderwall.MessageVerify,
							MessageVerified: welcomer.DefaultBorderwall.MessageVerified,
							RolesOnJoin:     welcomer.DefaultBorderwall.RolesOnJoin,
							RolesOnVerify:   welcomer.DefaultBorderwall.RolesOnVerify,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get borderwall guild settings")

						return nil, err
					}
				}

				switch roleType {
				case BorderwallRoleTypeJoin:
					// Check if the role exists in the list.
					if !slices.Contains(guildSettingsBorderwall.RolesOnJoin, int64(role.ID)) {
						return &discord.InteractionResponse{
							Type: discord.InteractionCallbackTypeChannelMessageSource,
							Data: &discord.InteractionCallbackData{
								Embeds: welcomer.NewEmbed("This role is not in the list.", welcomer.EmbedColourError),
								Flags:  uint32(discord.MessageFlagEphemeral),
							},
						}, nil
					}

					// Remove the role from the list.
					guildSettingsBorderwall.RolesOnJoin = slices.DeleteFunc(guildSettingsBorderwall.RolesOnJoin, func(r int64) bool {
						return r == int64(role.ID)
					})
				case BorderwallRoleTypeVerify:
					// Check if the role exists in the list.
					if !slices.Contains(guildSettingsBorderwall.RolesOnVerify, int64(role.ID)) {
						return &discord.InteractionResponse{
							Type: discord.InteractionCallbackTypeChannelMessageSource,
							Data: &discord.InteractionCallbackData{
								Embeds: welcomer.NewEmbed("This role is not in the list.", welcomer.EmbedColourError),
								Flags:  uint32(discord.MessageFlagEphemeral),
							},
						}, nil
					}

					// Remove the role from the list.
					guildSettingsBorderwall.RolesOnVerify = slices.DeleteFunc(guildSettingsBorderwall.RolesOnVerify, func(r int64) bool {
						return r == int64(role.ID)
					})
				default:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Unknown role type: "+roleType, welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.Queries.CreateOrUpdateBorderwallGuildSettings(ctx, database.CreateOrUpdateBorderwallGuildSettingsParams{
							GuildID:         int64(*interaction.GuildID),
							ToggleEnabled:   guildSettingsBorderwall.ToggleEnabled,
							ToggleSendDm:    guildSettingsBorderwall.ToggleSendDm,
							Channel:         guildSettingsBorderwall.Channel,
							MessageVerify:   guildSettingsBorderwall.MessageVerify,
							MessageVerified: guildSettingsBorderwall.MessageVerified,
							RolesOnJoin:     guildSettingsBorderwall.RolesOnJoin,
							RolesOnVerify:   guildSettingsBorderwall.RolesOnVerify,
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
						Msg("Failed to update borderwall guild settings")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed(fmt.Sprintf("Removed role <@&%d> from borderwall %s roles.", role.ID, roleType), welcomer.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	b.InteractionCommands.MustAddInteractionCommand(borderwallGroup)

	return nil
}
