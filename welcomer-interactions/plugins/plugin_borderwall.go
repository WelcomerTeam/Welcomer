package plugins

import (
	"context"
	"errors"

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

				Choices: []*discord.ApplicationCommandOptionChoice{
					{Name: BorderwallModuleBorderwall, Value: welcomer.S2J(BorderwallModuleBorderwall)},
					{Name: BorderwallModuleDMs, Value: welcomer.S2J(BorderwallModuleDMs)},
				},
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: discord.PermissionElevated,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				module := subway.MustGetArgument(ctx, "module").MustString()

				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsBorderwall, err := queries.GetBorderwallGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get borderwall guild settings.")
				}

				guildSettingsBorderwall.GuildID = int64(*interaction.GuildID)

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

				_, err = queries.UpdateBorderwallGuildSettings(ctx, &database.UpdateBorderwallGuildSettingsParams{
					GuildID:         guildSettingsBorderwall.GuildID,
					ToggleEnabled:   guildSettingsBorderwall.ToggleEnabled,
					ToggleSendDm:    guildSettingsBorderwall.ToggleSendDm,
					Channel:         guildSettingsBorderwall.Channel,
					MessageVerify:   guildSettingsBorderwall.MessageVerify,
					MessageVerified: guildSettingsBorderwall.MessageVerified,
					RolesOnJoin:     guildSettingsBorderwall.RolesOnJoin,
					RolesOnVerify:   guildSettingsBorderwall.RolesOnVerify,
				})
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update borderwall guild settings.")

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

				Choices: []*discord.ApplicationCommandOptionChoice{
					{Name: BorderwallModuleBorderwall, Value: welcomer.S2J(BorderwallModuleBorderwall)},
					{Name: BorderwallModuleDMs, Value: welcomer.S2J(BorderwallModuleDMs)},
				},
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: discord.PermissionElevated,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				module := subway.MustGetArgument(ctx, "module").MustString()

				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsBorderwall, err := queries.GetBorderwallGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get borderwall guild settings.")
				}

				guildSettingsBorderwall.GuildID = int64(*interaction.GuildID)

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

				_, err = queries.UpdateBorderwallGuildSettings(ctx, &database.UpdateBorderwallGuildSettingsParams{
					GuildID:         guildSettingsBorderwall.GuildID,
					ToggleEnabled:   guildSettingsBorderwall.ToggleEnabled,
					ToggleSendDm:    guildSettingsBorderwall.ToggleSendDm,
					Channel:         guildSettingsBorderwall.Channel,
					MessageVerify:   guildSettingsBorderwall.MessageVerify,
					MessageVerified: guildSettingsBorderwall.MessageVerified,
					RolesOnJoin:     guildSettingsBorderwall.RolesOnJoin,
					RolesOnVerify:   guildSettingsBorderwall.RolesOnVerify,
				})
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update borderwall guild settings.")

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

	b.InteractionCommands.MustAddInteractionCommand(borderwallGroup)

	return nil
}
