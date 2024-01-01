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
							Content: "Enabled all modules.",
						},
					}, nil
				case WelcomerModuleText:
					if err := toggleWelcomerTextGuildSetting(ctx, sub, int64(*interaction.GuildID), true); err != nil {
						return nil, err
					}

					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Content: "Enabled text module.",
						},
					}, nil
				case WelcomerModuleImages:
					if err := toggleWelcomerImagesGuildSetting(ctx, sub, int64(*interaction.GuildID), true); err != nil {
						return nil, err
					}

					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Content: "Enabled images module.",
						},
					}, nil
				case WelcomerModuleDMs:
					if err := toggleWelcomerDMsGuildSetting(ctx, sub, int64(*interaction.GuildID), true); err != nil {
						return nil, err
					}

					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Content: "Enabled DMs module.",
						},
					}, nil
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Content: "Unknown module: " + module,
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
							Content: "Disabled all modules.",
						},
					}, nil
				case WelcomerModuleText:
					if err := toggleWelcomerTextGuildSetting(ctx, sub, int64(*interaction.GuildID), false); err != nil {
						return nil, err
					}

					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Content: "Disabled text module.",
						},
					}, nil
				case WelcomerModuleImages:
					if err := toggleWelcomerImagesGuildSetting(ctx, sub, int64(*interaction.GuildID), false); err != nil {
						return nil, err
					}

					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Content: "Disabled images module.",
						},
					}, nil
				case WelcomerModuleDMs:
					if err := toggleWelcomerDMsGuildSetting(ctx, sub, int64(*interaction.GuildID), false); err != nil {
						return nil, err
					}

					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Content: "Disabled DMs module.",
						},
					}, nil
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Content: "Unknown module: " + module,
					},
				}, nil
			})
		},
	})

	p.InteractionCommands.MustAddInteractionCommand(welcomerGroup)

	return nil
}

func toggleWelcomerTextGuildSetting(ctx context.Context, sub *subway.Subway, guildID int64, value bool) error {
	queries := welcomer.GetQueriesFromContext(ctx)
	guildSettingsWelcomerText, err := queries.GetWelcomerTextGuildSettings(ctx, guildID)
	if err != nil {
		sub.Logger.Error().Err(err).
			Int64("guild_id", guildID).
			Msg("Failed to fetch welcomer text guild settings")

		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}

	if guildSettingsWelcomerText == nil {
		guildSettingsWelcomerText = &database.GuildSettingsWelcomerText{
			GuildID: guildID,
		}
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
