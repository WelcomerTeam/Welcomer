package plugins

import (
	"context"
	"errors"

	"github.com/WelcomerTeam/Discord/discord"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
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
	borderwallGroup.DMPermission = &utils.False

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

				Choices: []*discord.ApplicationCommandOptionChoice{
					{Name: BorderwallModuleBorderwall, Value: utils.StringToJsonLiteral(BorderwallModuleBorderwall)},
					{Name: BorderwallModuleDMs, Value: utils.StringToJsonLiteral(BorderwallModuleDMs)},
				},
			},
		},

		DMPermission:            &utils.False,
		DefaultMemberPermission: utils.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				module := subway.MustGetArgument(ctx, "module").MustString()

				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsBorderwall, err := queries.GetBorderwallGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsBorderwall = &database.GuildSettingsBorderwall{
							GuildID:         int64(*interaction.GuildID),
							ToggleEnabled:   database.DefaultBorderwall.ToggleEnabled,
							ToggleSendDm:    database.DefaultBorderwall.ToggleSendDm,
							Channel:         database.DefaultBorderwall.Channel,
							MessageVerify:   database.DefaultBorderwall.MessageVerify,
							MessageVerified: database.DefaultBorderwall.MessageVerified,
							RolesOnJoin:     database.DefaultBorderwall.RolesOnJoin,
							RolesOnVerify:   database.DefaultBorderwall.RolesOnVerify,
						}
					} else {
						sub.Logger.Error().Err(err).
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
							Embeds: utils.NewEmbed("Unknown module: "+module, utils.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// A channel must be set if direct messages are disabled.
				if guildSettingsBorderwall.Channel == 0 && guildSettingsBorderwall.ToggleEnabled && !guildSettingsBorderwall.ToggleSendDm {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("A channel must be selected if you are not sending borderwall messages via direct message. Please set a channel with `/borderwall channel` before enabling borderwall.", utils.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				err = utils.RetryWithFallback(
					func() error {
						_, err = queries.CreateOrUpdateBorderwallGuildSettings(ctx, database.CreateOrUpdateBorderwallGuildSettingsParams{
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
						return welcomer.EnsureGuild(ctx, queries, discord.Snowflake(*interaction.GuildID))
					},
					nil,
				)
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update borderwall guild settings")

					return nil, err
				}

				switch module {
				case BorderwallModuleBorderwall:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Enabled borderwall. Users will now have to verify when joining the server.", utils.EmbedColourSuccess),
						},
					}, nil
				case BorderwallModuleDMs:
					if guildSettingsBorderwall.ToggleEnabled {
						return &discord.InteractionResponse{
							Type: discord.InteractionCallbackTypeChannelMessageSource,
							Data: &discord.InteractionCallbackData{
								Embeds: utils.NewEmbed("Enabled borderwall direct messages. Users will now receive instructions on how to verify with borderwall when joining the server.", utils.EmbedColourSuccess),
							},
						}, nil
					} else {
						return &discord.InteractionResponse{
							Type: discord.InteractionCallbackTypeChannelMessageSource,
							Data: &discord.InteractionCallbackData{
								Embeds: utils.NewEmbed("Enabled borderwall direct messages. Borderwall is not enabled, users won't have to verify when joining the server.", utils.EmbedColourWarn),
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

				Choices: []*discord.ApplicationCommandOptionChoice{
					{Name: BorderwallModuleBorderwall, Value: utils.StringToJsonLiteral(BorderwallModuleBorderwall)},
					{Name: BorderwallModuleDMs, Value: utils.StringToJsonLiteral(BorderwallModuleDMs)},
				},
			},
		},

		DMPermission:            &utils.False,
		DefaultMemberPermission: utils.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				module := subway.MustGetArgument(ctx, "module").MustString()

				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsBorderwall, err := queries.GetBorderwallGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsBorderwall = &database.GuildSettingsBorderwall{
							GuildID:         int64(*interaction.GuildID),
							ToggleEnabled:   database.DefaultBorderwall.ToggleEnabled,
							ToggleSendDm:    database.DefaultBorderwall.ToggleSendDm,
							Channel:         database.DefaultBorderwall.Channel,
							MessageVerify:   database.DefaultBorderwall.MessageVerify,
							MessageVerified: database.DefaultBorderwall.MessageVerified,
							RolesOnJoin:     database.DefaultBorderwall.RolesOnJoin,
							RolesOnVerify:   database.DefaultBorderwall.RolesOnVerify,
						}
					} else {
						sub.Logger.Error().Err(err).
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
							Embeds: utils.NewEmbed("Unknown module: "+module, utils.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				err = utils.RetryWithFallback(
					func() error {
						_, err = queries.CreateOrUpdateBorderwallGuildSettings(ctx, database.CreateOrUpdateBorderwallGuildSettingsParams{
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
						return welcomer.EnsureGuild(ctx, queries, discord.Snowflake(*interaction.GuildID))
					},
					nil,
				)
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update borderwall guild settings")

					return nil, err
				}

				switch module {
				case BorderwallModuleBorderwall:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Disabled borderwall.", utils.EmbedColourSuccess),
						},
					}, nil
				case BorderwallModuleDMs:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Disabled borderwall direct messages.", utils.EmbedColourSuccess),
						},
					}, nil
				}

				return nil, nil // Unreachable
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

		DMPermission:            &utils.False,
		DefaultMemberPermission: utils.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				channel := subway.MustGetArgument(ctx, "channel").MustChannel()

				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsBorderwall, err := queries.GetBorderwallGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsBorderwall = &database.GuildSettingsBorderwall{
							GuildID:         int64(*interaction.GuildID),
							ToggleEnabled:   database.DefaultBorderwall.ToggleEnabled,
							ToggleSendDm:    database.DefaultBorderwall.ToggleSendDm,
							Channel:         database.DefaultBorderwall.Channel,
							MessageVerify:   database.DefaultBorderwall.MessageVerify,
							MessageVerified: database.DefaultBorderwall.MessageVerified,
							RolesOnJoin:     database.DefaultBorderwall.RolesOnJoin,
							RolesOnVerify:   database.DefaultBorderwall.RolesOnVerify,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get borderwall guild settings")

						return nil, err
					}
				}

				if channel != nil {
					guildSettingsBorderwall.Channel = int64(channel.ID)
				} else {
					guildSettingsBorderwall.Channel = 0
				}

				// A channel must be set if direct messages are disabled.
				if guildSettingsBorderwall.Channel == 0 && guildSettingsBorderwall.ToggleEnabled && !guildSettingsBorderwall.ToggleSendDm {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("A channel must be selected if you are not sending borderwall messages via direct message.", utils.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				err = utils.RetryWithFallback(
					func() error {
						_, err = queries.CreateOrUpdateBorderwallGuildSettings(ctx, database.CreateOrUpdateBorderwallGuildSettingsParams{
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
						return welcomer.EnsureGuild(ctx, queries, discord.Snowflake(*interaction.GuildID))
					},
					nil,
				)
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update borderwall guild settings")

					return nil, err
				}

				if channel != nil {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Set borderwall channel to: <#"+channel.ID.String()+">.", utils.EmbedColourSuccess),
						},
					}, nil
				} else {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Removed borderwall channel. Borderwall will only send direct messages to users.", utils.EmbedColourWarn),
						},
					}, nil
				}
			})
		},
	})

	b.InteractionCommands.MustAddInteractionCommand(borderwallGroup)

	return nil
}
