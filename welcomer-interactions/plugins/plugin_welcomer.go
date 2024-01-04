package plugins

import (
	"context"
	"errors"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4"
	jsoniter "github.com/json-iterator/go"
)

func NewWelcomerCog() *WelcomerCog {
	return &WelcomerCog{
		InteractionCommands: subway.SetupInteractionCommandable(&subway.InteractionCommandable{}),
	}
}

type WelcomerCog struct {
	InteractionCommands *subway.InteractionCommandable
}

// Assert types.

var (
	_ subway.Cog                        = (*WelcomerCog)(nil)
	_ subway.CogWithInteractionCommands = (*WelcomerCog)(nil)
)

var WelcomerModule string

const (
	WelcomerModuleAll    = "all"
	WelcomerModuleText   = "text"
	WelcomerModuleImages = "images"
	WelcomerModuleDMs    = "dms"
)

func (p *WelcomerCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "Welcomer",
		Description: "Provides the functionality for the 'Welcomer' feature",
	}
}

func (p *WelcomerCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return p.InteractionCommands
}

func (p *WelcomerCog) RegisterCog(sub *subway.Subway) error {
	welcomerGroup := subway.NewSubcommandGroup(
		"welcomer",
		"Welcome new users to your server with fancy images, text or send them a direct message.",
	)

	welcomerGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "test",
		Description: "Tests the Welcomer functionality.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     false,
				ArgumentType: subway.ArgumentTypeMember,
				Name:         "user",
				Description:  "The user you would like to send the welcome message for.",
			},
		},

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				member := subway.MustGetArgument(ctx, "user").MustMember()
				if member == nil {
					member = interaction.Member
				}

				queries := welcomer.GetQueriesFromContext(ctx)

				// Fetch guild settings.

				guildSettingsWelcomerText, err := queries.GetWelcomerTextGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to fetch welcomer text guild settings")

					return nil, err
				}

				guildSettingsWelcomerImages, err := queries.GetWelcomerImagesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to fetch welcomer image guild settings")

					return nil, err
				}

				guildSettingsWelcomerDMs, err := queries.GetWelcomerDMsGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to fetch welcomer dm guild settings")

					return nil, err
				}

				// If no modules are enabled, let the user know.
				if !guildSettingsWelcomerText.ToggleEnabled && !guildSettingsWelcomerImages.ToggleEnabled && !guildSettingsWelcomerDMs.ToggleEnabled {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("No modules are enabled. Please use `/welcomer enable`", welcomer.EmbedColourError),
						},
					}, nil
				}

				// If text or images are enabled, but no channel is set, let the user know.
				if !guildSettingsWelcomerDMs.ToggleEnabled && guildSettingsWelcomerText.Channel == 0 {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("No channel is set. Please use `/welcomer channel`", welcomer.EmbedColourError),
						},
					}, nil
				}

				// GuildID may be missing, fill it in.
				member.GuildID = interaction.GuildID

				data, err := jsoniter.Marshal(welcomer.CustomEventInvokeWelcomerStructure{
					Interaction: &interaction,
					Member:      member,
				})
				if err != nil {
					return nil, err
				}

				_, err = sub.SandwichClient.RelayMessage(ctx, &sandwich.RelayMessageRequest{
					Manager: welcomer.GetManagerNameFromContext(ctx),
					Type:    welcomer.CustomEventInvokeWelcomer,
					Data:    data,
				})
				if err != nil {
					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeDeferredChannelMessageSource,
				}, nil
			})
		},
	})

	welcomerGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "enable",
		Description: "Enables the Welcomer functionality.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeString,
				Name:         "module",
				Description:  "The module you would like to enable.",

				Choices: []*discord.ApplicationCommandOptionChoice{
					{Name: WelcomerModuleAll, Value: welcomer.S2J(WelcomerModuleAll)},
					{Name: WelcomerModuleText, Value: welcomer.S2J(WelcomerModuleText)},
					{Name: WelcomerModuleImages, Value: welcomer.S2J(WelcomerModuleImages)},
					{Name: WelcomerModuleDMs, Value: welcomer.S2J(WelcomerModuleDMs)},
				},
			},
		},

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				module := subway.MustGetArgument(ctx, "module").MustString()

				switch module {
				case WelcomerModuleAll:
					if err := toggleWelcomerTextGuildSetting(ctx, sub, int64(*interaction.GuildID), true); err != nil {
						return nil, err
					}

					if err := toggleWelcomerImagesGuildSetting(ctx, sub, int64(*interaction.GuildID), true); err != nil {
						return nil, err
					}

					if err := toggleWelcomerDMsGuildSetting(ctx, sub, int64(*interaction.GuildID), true); err != nil {
						return nil, err
					}

					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Enabled all modules", welcomer.EmbedColourSuccess),
						},
					}, nil
				case WelcomerModuleText:
					if err := toggleWelcomerTextGuildSetting(ctx, sub, int64(*interaction.GuildID), true); err != nil {
						return nil, err
					}

					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Enabled text module", welcomer.EmbedColourSuccess),
						},
					}, nil
				case WelcomerModuleImages:
					if err := toggleWelcomerImagesGuildSetting(ctx, sub, int64(*interaction.GuildID), true); err != nil {
						return nil, err
					}

					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Enabled images module", welcomer.EmbedColourSuccess),
						},
					}, nil
				case WelcomerModuleDMs:
					if err := toggleWelcomerDMsGuildSetting(ctx, sub, int64(*interaction.GuildID), true); err != nil {
						return nil, err
					}

					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Enabled DMs module", welcomer.EmbedColourSuccess),
						},
					}, nil
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Unknown module: "+module, welcomer.EmbedColourError),
					},
				}, nil
			})
		},
	})

	welcomerGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "disable",
		Description: "Disables the Welcomer functionality.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeString,
				Name:         "module",
				Description:  "The module you would like to disable.",

				Choices: []*discord.ApplicationCommandOptionChoice{
					{Name: WelcomerModuleAll, Value: welcomer.S2J(WelcomerModuleAll)},
					{Name: WelcomerModuleText, Value: welcomer.S2J(WelcomerModuleText)},
					{Name: WelcomerModuleImages, Value: welcomer.S2J(WelcomerModuleImages)},
					{Name: WelcomerModuleDMs, Value: welcomer.S2J(WelcomerModuleDMs)},
				},
			},
		},

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				module := subway.MustGetArgument(ctx, "module").MustString()

				switch module {
				case WelcomerModuleAll:
					if err := toggleWelcomerTextGuildSetting(ctx, sub, int64(*interaction.GuildID), false); err != nil {
						return nil, err
					}

					if err := toggleWelcomerImagesGuildSetting(ctx, sub, int64(*interaction.GuildID), false); err != nil {
						return nil, err
					}

					if err := toggleWelcomerDMsGuildSetting(ctx, sub, int64(*interaction.GuildID), false); err != nil {
						return nil, err
					}

					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Disabled all modules", welcomer.EmbedColourSuccess),
						},
					}, nil
				case WelcomerModuleText:
					if err := toggleWelcomerTextGuildSetting(ctx, sub, int64(*interaction.GuildID), false); err != nil {
						return nil, err
					}

					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Disabled text module", welcomer.EmbedColourSuccess),
						},
					}, nil
				case WelcomerModuleImages:
					if err := toggleWelcomerImagesGuildSetting(ctx, sub, int64(*interaction.GuildID), false); err != nil {
						return nil, err
					}

					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Disabled images module", welcomer.EmbedColourSuccess),
						},
					}, nil
				case WelcomerModuleDMs:
					if err := toggleWelcomerDMsGuildSetting(ctx, sub, int64(*interaction.GuildID), false); err != nil {
						return nil, err
					}

					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Disabled DMs module", welcomer.EmbedColourSuccess),
						},
					}, nil
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Unknown module: "+module, welcomer.EmbedColourError),
					},
				}, nil
			})
		},
	})

	welcomerGroup.AddInteractionCommand(&subway.InteractionCommandable{
		Name:        "channel",
		Description: "Sets the channel to send the welcome message to.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     false,
				ArgumentType: subway.ArgumentTypeTextChannel,
				Name:         "channel",
				Description:  "The channel you would like to send the welcome message to.",
			},
		},

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				channel := subway.MustGetArgument(ctx, "channel").MustChannel()

				queries := welcomer.GetQueriesFromContext(ctx)
				guildSettingsWelcomerText, err := getWelcomerTextGuildSettings(ctx, sub, queries, int64(*interaction.GuildID))
				if err != nil {
					return nil, err
				}

				if channel != nil {
					guildSettingsWelcomerText.Channel = int64(channel.ID)
				} else {
					guildSettingsWelcomerText.Channel = 0
				}

				_, err = queries.UpdateWelcomerTextGuildSettings(ctx, &database.UpdateWelcomerTextGuildSettingsParams{
					GuildID:       guildSettingsWelcomerText.GuildID,
					ToggleEnabled: guildSettingsWelcomerText.ToggleEnabled,
					Channel:       guildSettingsWelcomerText.Channel,
					MessageFormat: guildSettingsWelcomerText.MessageFormat,
				})
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update welcomer text guild settings")

					return nil, err
				}

				if channel != nil {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Set channel to: <#"+channel.ID.String()+">", welcomer.EmbedColourSuccess),
						},
					}, nil
				} else {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Unset channel. Welcomer text and image features will not work, if enabled.", welcomer.EmbedColourWarn),
						},
					}, nil
				}
			})
		},
	})

	p.InteractionCommands.MustAddInteractionCommand(welcomerGroup)

	return nil
}

func getWelcomerTextGuildSettings(ctx context.Context, sub *subway.Subway, queries *database.Queries, guildID int64) (*database.GuildSettingsWelcomerText, error) {
	guildSettingsWelcomerText, err := queries.GetWelcomerTextGuildSettings(ctx, guildID)
	if err != nil {
		sub.Logger.Error().Err(err).
			Int64("guild_id", guildID).
			Msg("Failed to fetch welcomer text guild settings")

		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}

	if guildSettingsWelcomerText == nil {
		guildSettingsWelcomerText = &database.GuildSettingsWelcomerText{
			GuildID: guildID,
		}
	}

	return guildSettingsWelcomerText, err
}

func toggleWelcomerTextGuildSetting(ctx context.Context, sub *subway.Subway, guildID int64, value bool) error {
	queries := welcomer.GetQueriesFromContext(ctx)

	guildSettingsWelcomerText, err := getWelcomerTextGuildSettings(ctx, sub, queries, guildID)
	if err != nil {
		return err
	}

	guildSettingsWelcomerText.ToggleEnabled = value

	_, err = queries.UpdateWelcomerTextGuildSettings(ctx, &database.UpdateWelcomerTextGuildSettingsParams{
		GuildID:       guildSettingsWelcomerText.GuildID,
		ToggleEnabled: guildSettingsWelcomerText.ToggleEnabled,
		Channel:       guildSettingsWelcomerText.Channel,
		MessageFormat: guildSettingsWelcomerText.MessageFormat,
	})
	if err != nil {
		sub.Logger.Error().Err(err).
			Int64("guild_id", guildID).
			Msg("Failed to update welcomer text guild settings")

		return err
	}

	return nil
}

func toggleWelcomerImagesGuildSetting(ctx context.Context, sub *subway.Subway, guildID int64, value bool) error {
	queries := welcomer.GetQueriesFromContext(ctx)
	guildSettingsWelcomerImages, err := queries.GetWelcomerImagesGuildSettings(ctx, guildID)
	if err != nil {
		sub.Logger.Error().Err(err).
			Int64("guild_id", guildID).
			Msg("Failed to fetch welcomer images guild settings")

		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}

	if guildSettingsWelcomerImages == nil {
		guildSettingsWelcomerImages = &database.GuildSettingsWelcomerImages{
			GuildID: guildID,
		}
	}

	guildSettingsWelcomerImages.ToggleEnabled = value

	_, err = queries.UpdateWelcomerImagesGuildSettings(ctx, &database.UpdateWelcomerImagesGuildSettingsParams{
		GuildID:                guildSettingsWelcomerImages.GuildID,
		ToggleEnabled:          guildSettingsWelcomerImages.ToggleEnabled,
		ToggleImageBorder:      guildSettingsWelcomerImages.ToggleImageBorder,
		BackgroundName:         guildSettingsWelcomerImages.BackgroundName,
		ColourText:             guildSettingsWelcomerImages.ColourText,
		ColourTextBorder:       guildSettingsWelcomerImages.ColourTextBorder,
		ColourImageBorder:      guildSettingsWelcomerImages.ColourImageBorder,
		ColourProfileBorder:    guildSettingsWelcomerImages.ColourProfileBorder,
		ImageAlignment:         guildSettingsWelcomerImages.ImageAlignment,
		ImageTheme:             guildSettingsWelcomerImages.ImageTheme,
		ImageMessage:           guildSettingsWelcomerImages.ImageMessage,
		ImageProfileBorderType: guildSettingsWelcomerImages.ImageProfileBorderType,
	})
	if err != nil {
		sub.Logger.Error().Err(err).
			Int64("guild_id", guildID).
			Msg("Failed to update welcomer images guild settings")

		return err
	}

	return nil
}

func toggleWelcomerDMsGuildSetting(ctx context.Context, sub *subway.Subway, guildID int64, value bool) error {
	queries := welcomer.GetQueriesFromContext(ctx)
	guildSettingsWelcomerDMs, err := queries.GetWelcomerDMsGuildSettings(ctx, guildID)
	if err != nil {
		sub.Logger.Error().Err(err).
			Int64("guild_id", guildID).
			Msg("Failed to fetch welcomer DMs guild settings")

		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}

	if guildSettingsWelcomerDMs == nil {
		guildSettingsWelcomerDMs = &database.GuildSettingsWelcomerDms{
			GuildID: guildID,
		}
	}

	guildSettingsWelcomerDMs.ToggleEnabled = value

	_, err = queries.UpdateWelcomerDMsGuildSettings(ctx, &database.UpdateWelcomerDMsGuildSettingsParams{
		GuildID:             guildSettingsWelcomerDMs.GuildID,
		ToggleEnabled:       guildSettingsWelcomerDMs.ToggleEnabled,
		ToggleUseTextFormat: guildSettingsWelcomerDMs.ToggleUseTextFormat,
		ToggleIncludeImage:  guildSettingsWelcomerDMs.ToggleIncludeImage,
		MessageFormat:       guildSettingsWelcomerDMs.MessageFormat,
	})
	if err != nil {
		sub.Logger.Error().Err(err).
			Int64("guild_id", guildID).
			Msg("Failed to update welcomer DMs guild settings")

		return err
	}

	return nil
}
