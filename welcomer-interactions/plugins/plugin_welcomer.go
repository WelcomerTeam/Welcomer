package plugins

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	subway "github.com/WelcomerTeam/Subway/subway"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
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

const (
	WelcomerModuleAll    = "all"
	WelcomerModuleText   = "text"
	WelcomerModuleImages = "images"
	WelcomerModuleDMs    = "dms"
)

func (w *WelcomerCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "Welcomer",
		Description: "Provides the functionality for the 'Welcomer' feature",
	}
}

func (w *WelcomerCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return w.InteractionCommands
}

func (w *WelcomerCog) RegisterCog(sub *subway.Subway) error {
	welcomerGroup := subway.NewSubcommandGroup(
		"welcomer",
		"Welcome new users to your server with fancy images, text or send them a direct message.",
	)

	// Disable the welcomer module for DM channels.
	welcomerGroup.DMPermission = &utils.False

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

		DMPermission:            &utils.False,
		DefaultMemberPermission: utils.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return core.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				member := subway.MustGetArgument(ctx, "user").MustMember()
				if member.User == nil || member.User.ID.IsNil() {
					member = *interaction.Member
				}

				queries := core.GetQueriesFromContext(ctx)

				guildSettingsWelcomerText, err := queries.GetWelcomerTextGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsWelcomerText = &database.GuildSettingsWelcomerText{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: database.DefaultWelcomerText.ToggleEnabled,
							Channel:       database.DefaultWelcomerText.Channel,
							MessageFormat: database.DefaultWelcomerText.MessageFormat,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get welcomer text guild settings")

						return nil, err
					}
				}

				guildSettingsWelcomerImages, err := queries.GetWelcomerImagesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsWelcomerImages = &database.GuildSettingsWelcomerImages{
							GuildID:                int64(*interaction.GuildID),
							ToggleEnabled:          database.DefaultWelcomerImages.ToggleEnabled,
							ToggleImageBorder:      database.DefaultWelcomerImages.ToggleImageBorder,
							BackgroundName:         database.DefaultWelcomerImages.BackgroundName,
							ColourText:             database.DefaultWelcomerImages.ColourText,
							ColourTextBorder:       database.DefaultWelcomerImages.ColourTextBorder,
							ColourImageBorder:      database.DefaultWelcomerImages.ColourImageBorder,
							ColourProfileBorder:    database.DefaultWelcomerImages.ColourProfileBorder,
							ImageAlignment:         database.DefaultWelcomerImages.ImageAlignment,
							ImageTheme:             database.DefaultWelcomerImages.ImageTheme,
							ImageMessage:           database.DefaultWelcomerImages.ImageMessage,
							ImageProfileBorderType: database.DefaultWelcomerImages.ImageProfileBorderType,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get utils.image guild settings")

						return nil, err
					}
				}

				guildSettingsWelcomerDMs, err := queries.GetWelcomerDMsGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsWelcomerDMs = &database.GuildSettingsWelcomerDms{
							GuildID:             int64(*interaction.GuildID),
							ToggleEnabled:       database.DefaultWelcomerDms.ToggleEnabled,
							ToggleUseTextFormat: database.DefaultWelcomerDms.ToggleUseTextFormat,
							ToggleIncludeImage:  database.DefaultWelcomerDms.ToggleIncludeImage,
							MessageFormat:       database.DefaultWelcomerDms.MessageFormat,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get welcomer DMs guild settings")

						return nil, err
					}
				}

				guildSettingsWelcomerText.MessageFormat = utils.SetupJSONB(guildSettingsWelcomerText.MessageFormat)
				guildSettingsWelcomerDMs.MessageFormat = utils.SetupJSONB(guildSettingsWelcomerDMs.MessageFormat)

				// If no modules are enabled, let the user know.
				if !guildSettingsWelcomerText.ToggleEnabled && !guildSettingsWelcomerImages.ToggleEnabled && !guildSettingsWelcomerDMs.ToggleEnabled {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("No modules are enabled. Please use `/welcomer enable`", utils.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// If text or images are enabled, but no channel is set, let the user know.
				if !guildSettingsWelcomerDMs.ToggleEnabled && guildSettingsWelcomerText.Channel == 0 {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("No channel is set. Please use `/welcomer channel`", utils.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// GuildID may be missing, fill it in.
				member.GuildID = interaction.GuildID

				data, err := json.Marshal(core.CustomEventInvokeWelcomerStructure{
					Interaction: &interaction,
					Member:      member,
				})
				if err != nil {
					return nil, err
				}

				_, err = sub.SandwichClient.RelayMessage(ctx, &sandwich.RelayMessageRequest{
					Manager: core.GetManagerNameFromContext(ctx),
					Type:    core.CustomEventInvokeWelcomer,
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
		Description: "Enables a welcomer module for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeString,
				Name:         "module",
				Description:  "The module to enable.",

				Choices: []discord.ApplicationCommandOptionChoice{
					{Name: WelcomerModuleAll, Value: utils.StringToJsonLiteral(WelcomerModuleAll)},
					{Name: WelcomerModuleText, Value: utils.StringToJsonLiteral(WelcomerModuleText)},
					{Name: WelcomerModuleImages, Value: utils.StringToJsonLiteral(WelcomerModuleImages)},
					{Name: WelcomerModuleDMs, Value: utils.StringToJsonLiteral(WelcomerModuleDMs)},
				},
			},
		},

		DMPermission:            &utils.False,
		DefaultMemberPermission: utils.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return core.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				module := subway.MustGetArgument(ctx, "module").MustString()

				queries := core.GetQueriesFromContext(ctx)

				guildSettingsWelcomerText, err := queries.GetWelcomerTextGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsWelcomerText = &database.GuildSettingsWelcomerText{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: database.DefaultWelcomerText.ToggleEnabled,
							Channel:       database.DefaultWelcomerText.Channel,
							MessageFormat: database.DefaultWelcomerText.MessageFormat,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get welcomer text guild settings")

						return nil, err
					}
				}

				guildSettingsWelcomerDMs, err := queries.GetWelcomerDMsGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsWelcomerDMs = &database.GuildSettingsWelcomerDms{
							GuildID:             int64(*interaction.GuildID),
							ToggleEnabled:       database.DefaultWelcomerDms.ToggleEnabled,
							ToggleUseTextFormat: database.DefaultWelcomerDms.ToggleUseTextFormat,
							ToggleIncludeImage:  database.DefaultWelcomerDms.ToggleIncludeImage,
							MessageFormat:       database.DefaultWelcomerDms.MessageFormat,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get welcomer DMs guild settings")

						return nil, err
					}
				}

				guildSettingsWelcomerImages, err := queries.GetWelcomerImagesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsWelcomerImages = &database.GuildSettingsWelcomerImages{
							GuildID:                int64(*interaction.GuildID),
							ToggleEnabled:          database.DefaultWelcomerImages.ToggleEnabled,
							ToggleImageBorder:      database.DefaultWelcomerImages.ToggleImageBorder,
							BackgroundName:         database.DefaultWelcomerImages.BackgroundName,
							ColourText:             database.DefaultWelcomerImages.ColourText,
							ColourTextBorder:       database.DefaultWelcomerImages.ColourTextBorder,
							ColourImageBorder:      database.DefaultWelcomerImages.ColourImageBorder,
							ColourProfileBorder:    database.DefaultWelcomerImages.ColourProfileBorder,
							ImageAlignment:         database.DefaultWelcomerImages.ImageAlignment,
							ImageTheme:             database.DefaultWelcomerImages.ImageTheme,
							ImageMessage:           database.DefaultWelcomerImages.ImageMessage,
							ImageProfileBorderType: database.DefaultWelcomerImages.ImageProfileBorderType,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get utils.image guild settings")

						return nil, err
					}
				}

				guildSettingsWelcomerText.MessageFormat = utils.SetupJSONB(guildSettingsWelcomerText.MessageFormat)
				guildSettingsWelcomerDMs.MessageFormat = utils.SetupJSONB(guildSettingsWelcomerDMs.MessageFormat)

				switch module {
				case WelcomerModuleAll:
					guildSettingsWelcomerText.ToggleEnabled = true
					guildSettingsWelcomerImages.ToggleEnabled = true
					guildSettingsWelcomerDMs.ToggleEnabled = true
				case WelcomerModuleText:
					guildSettingsWelcomerText.ToggleEnabled = true
				case WelcomerModuleImages:
					guildSettingsWelcomerImages.ToggleEnabled = true
				case WelcomerModuleDMs:
					guildSettingsWelcomerDMs.ToggleEnabled = true
				default:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Unknown module: "+module, utils.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// Update database.

				err = utils.RetryWithFallback(
					func() error {
						_, err = queries.CreateOrUpdateWelcomerTextGuildSettings(ctx, database.CreateOrUpdateWelcomerTextGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsWelcomerText.ToggleEnabled,
							Channel:       guildSettingsWelcomerText.Channel,
							MessageFormat: guildSettingsWelcomerText.MessageFormat,
						})

						return err
					},
					func() error {
						return core.EnsureGuild(ctx, queries, discord.Snowflake(*interaction.GuildID))
					},
					nil,
				)
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update welcomer text guild settings")

					return nil, err
				}

				_, err = queries.CreateOrUpdateWelcomerImagesGuildSettings(ctx, database.CreateOrUpdateWelcomerImagesGuildSettingsParams{
					GuildID:                int64(*interaction.GuildID),
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
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update welcomer images guild settings")

					return nil, err
				}

				_, err = queries.CreateOrUpdateWelcomerDMsGuildSettings(ctx, database.CreateOrUpdateWelcomerDMsGuildSettingsParams{
					GuildID:             int64(*interaction.GuildID),
					ToggleEnabled:       guildSettingsWelcomerDMs.ToggleEnabled,
					ToggleUseTextFormat: guildSettingsWelcomerDMs.ToggleUseTextFormat,
					ToggleIncludeImage:  guildSettingsWelcomerDMs.ToggleIncludeImage,
					MessageFormat:       guildSettingsWelcomerDMs.MessageFormat,
				})
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update welcomer DMs guild settings")

					return nil, err
				}

				switch module {
				case WelcomerModuleAll:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Enabled all modules.  Run `/welcomer test` to see the message that is sent.", utils.EmbedColourSuccess),
						},
					}, nil
				case WelcomerModuleText:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Enabled welcomer text messages.  Run `/welcomer test` to see the message that is sent.", utils.EmbedColourSuccess),
						},
					}, nil
				case WelcomerModuleImages:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Enabled welcomer images.  Run `/welcomer test` to see the message that is sent.", utils.EmbedColourSuccess),
						},
					}, nil
				case WelcomerModuleDMs:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Enabled welcomer direct messages.  Run `/welcomer test` to see the message that is sent.", utils.EmbedColourSuccess),
						},
					}, nil
				default:
					return nil, nil // Unreachable
				}
			})
		},
	})

	welcomerGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "disable",
		Description: "Disables a welcomer module for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeString,
				Name:         "module",
				Description:  "The module to disable.",

				Choices: []discord.ApplicationCommandOptionChoice{
					{Name: WelcomerModuleAll, Value: utils.StringToJsonLiteral(WelcomerModuleAll)},
					{Name: WelcomerModuleText, Value: utils.StringToJsonLiteral(WelcomerModuleText)},
					{Name: WelcomerModuleImages, Value: utils.StringToJsonLiteral(WelcomerModuleImages)},
					{Name: WelcomerModuleDMs, Value: utils.StringToJsonLiteral(WelcomerModuleDMs)},
				},
			},
		},

		DMPermission:            &utils.False,
		DefaultMemberPermission: utils.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return core.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				module := subway.MustGetArgument(ctx, "module").MustString()

				queries := core.GetQueriesFromContext(ctx)

				guildSettingsWelcomerText, err := queries.GetWelcomerTextGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsWelcomerText = &database.GuildSettingsWelcomerText{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: database.DefaultWelcomerText.ToggleEnabled,
							Channel:       database.DefaultWelcomerText.Channel,
							MessageFormat: database.DefaultWelcomerText.MessageFormat,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get welcomer text guild settings")

						return nil, err
					}
				}

				guildSettingsWelcomerDMs, err := queries.GetWelcomerDMsGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsWelcomerDMs = &database.GuildSettingsWelcomerDms{
							GuildID:             int64(*interaction.GuildID),
							ToggleEnabled:       database.DefaultWelcomerDms.ToggleEnabled,
							ToggleUseTextFormat: database.DefaultWelcomerDms.ToggleUseTextFormat,
							ToggleIncludeImage:  database.DefaultWelcomerDms.ToggleIncludeImage,
							MessageFormat:       database.DefaultWelcomerDms.MessageFormat,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get welcomer DMs guild settings")

						return nil, err
					}
				}

				guildSettingsWelcomerImages, err := queries.GetWelcomerImagesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsWelcomerImages = &database.GuildSettingsWelcomerImages{
							GuildID:                int64(*interaction.GuildID),
							ToggleEnabled:          database.DefaultWelcomerImages.ToggleEnabled,
							ToggleImageBorder:      database.DefaultWelcomerImages.ToggleImageBorder,
							BackgroundName:         database.DefaultWelcomerImages.BackgroundName,
							ColourText:             database.DefaultWelcomerImages.ColourText,
							ColourTextBorder:       database.DefaultWelcomerImages.ColourTextBorder,
							ColourImageBorder:      database.DefaultWelcomerImages.ColourImageBorder,
							ColourProfileBorder:    database.DefaultWelcomerImages.ColourProfileBorder,
							ImageAlignment:         database.DefaultWelcomerImages.ImageAlignment,
							ImageTheme:             database.DefaultWelcomerImages.ImageTheme,
							ImageMessage:           database.DefaultWelcomerImages.ImageMessage,
							ImageProfileBorderType: database.DefaultWelcomerImages.ImageProfileBorderType,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get utils.image guild settings")

						return nil, err
					}
				}

				if guildSettingsWelcomerText.MessageFormat.Status == pgtype.Undefined {
					guildSettingsWelcomerText.MessageFormat.Status = pgtype.Null
				}

				if guildSettingsWelcomerDMs.MessageFormat.Status == pgtype.Undefined {
					guildSettingsWelcomerDMs.MessageFormat.Status = pgtype.Null
				}

				switch module {
				case WelcomerModuleAll:
					guildSettingsWelcomerText.ToggleEnabled = false
					guildSettingsWelcomerImages.ToggleEnabled = false
					guildSettingsWelcomerDMs.ToggleEnabled = false
				case WelcomerModuleText:
					guildSettingsWelcomerText.ToggleEnabled = false
				case WelcomerModuleImages:
					guildSettingsWelcomerImages.ToggleEnabled = false
				case WelcomerModuleDMs:
					guildSettingsWelcomerDMs.ToggleEnabled = false
				default:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Unknown module: "+module, utils.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// Update database.

				err = utils.RetryWithFallback(
					func() error {
						_, err = queries.CreateOrUpdateWelcomerTextGuildSettings(ctx, database.CreateOrUpdateWelcomerTextGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsWelcomerText.ToggleEnabled,
							Channel:       guildSettingsWelcomerText.Channel,
							MessageFormat: guildSettingsWelcomerText.MessageFormat,
						})
						return err
					},
					func() error {
						return core.EnsureGuild(ctx, queries, discord.Snowflake(*interaction.GuildID))
					},
					nil,
				)
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update welcomer text guild settings")

					return nil, err
				}

				_, err = queries.CreateOrUpdateWelcomerImagesGuildSettings(ctx, database.CreateOrUpdateWelcomerImagesGuildSettingsParams{
					GuildID:                int64(*interaction.GuildID),
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
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update welcomer images guild settings")

					return nil, err
				}

				_, err = queries.CreateOrUpdateWelcomerDMsGuildSettings(ctx, database.CreateOrUpdateWelcomerDMsGuildSettingsParams{
					GuildID:             int64(*interaction.GuildID),
					ToggleEnabled:       guildSettingsWelcomerDMs.ToggleEnabled,
					ToggleUseTextFormat: guildSettingsWelcomerDMs.ToggleUseTextFormat,
					ToggleIncludeImage:  guildSettingsWelcomerDMs.ToggleIncludeImage,
					MessageFormat:       guildSettingsWelcomerDMs.MessageFormat,
				})
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update welcomer DMs guild settings")

					return nil, err
				}

				switch module {
				case WelcomerModuleAll:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Disabled all modules.", utils.EmbedColourSuccess),
						},
					}, nil
				case WelcomerModuleText:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Disabled welcomer text messages.", utils.EmbedColourSuccess),
						},
					}, nil
				case WelcomerModuleImages:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Disabled welcomer images", utils.EmbedColourSuccess),
						},
					}, nil
				case WelcomerModuleDMs:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Disabled welcomer direct messages", utils.EmbedColourSuccess),
						},
					}, nil
				default:
					return nil, nil
				}
			})
		},
	})

	welcomerGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "setchannel",
		Description: "Sets the channel to send welcome messages to.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     false,
				ArgumentType: subway.ArgumentTypeTextChannel,
				Name:         "channel",
				Description:  "The channel you would like to send the welcome message to.",
			},
		},

		DMPermission:            &utils.False,
		DefaultMemberPermission: utils.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return core.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				channel := subway.MustGetArgument(ctx, "channel").MustChannel()

				queries := core.GetQueriesFromContext(ctx)

				guildSettingsWelcomerText, err := queries.GetWelcomerTextGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsWelcomerText = &database.GuildSettingsWelcomerText{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: database.DefaultWelcomerText.ToggleEnabled,
							Channel:       database.DefaultWelcomerText.Channel,
							MessageFormat: database.DefaultWelcomerText.MessageFormat,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get welcomer text guild settings")

						return nil, err
					}
				}
				if guildSettingsWelcomerText.MessageFormat.Status == pgtype.Undefined {
					guildSettingsWelcomerText.MessageFormat.Status = pgtype.Null
				}

				if !channel.ID.IsNil() {
					guildSettingsWelcomerText.Channel = int64(channel.ID)
				} else {
					guildSettingsWelcomerText.Channel = 0
				}

				err = utils.RetryWithFallback(
					func() error {
						_, err = queries.CreateOrUpdateWelcomerTextGuildSettings(ctx, database.CreateOrUpdateWelcomerTextGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsWelcomerText.ToggleEnabled,
							Channel:       guildSettingsWelcomerText.Channel,
							MessageFormat: guildSettingsWelcomerText.MessageFormat,
						})
						return err
					},
					func() error {
						return core.EnsureGuild(ctx, queries, discord.Snowflake(*interaction.GuildID))
					},
					nil,
				)
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update welcomer text guild settings")

					return nil, err
				}

				if !channel.ID.IsNil() {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Set welcomer channel to: <#"+channel.ID.String()+">.  Run `/welcomer test` to see the message that is sent.", utils.EmbedColourSuccess),
						},
					}, nil
				} else {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Removed welcomer channel. Welcomer text and image features will not work, if they are enabled.", utils.EmbedColourWarn),
						},
					}, nil
				}
			})
		},
	})

	w.InteractionCommands.MustAddInteractionCommand(welcomerGroup)

	return nil
}
