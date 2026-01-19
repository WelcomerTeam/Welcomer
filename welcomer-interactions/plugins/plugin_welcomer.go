package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
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
	welcomerGroup.DMPermission = &welcomer.False

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

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(welcomer.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return core.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				member := subway.MustGetArgument(ctx, "user").MustMember()
				if member.User == nil || member.User.ID.IsNil() {
					member = *interaction.Member
				}

				guildSettingsWelcomerText, err := welcomer.Queries.GetWelcomerTextGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsWelcomerText = &database.GuildSettingsWelcomerText{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultWelcomerText.ToggleEnabled,
							Channel:       welcomer.DefaultWelcomerText.Channel,
							MessageFormat: welcomer.DefaultWelcomerText.MessageFormat,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get welcomer text guild settings")

						return nil, err
					}
				}

				guildSettingsWelcomerImages, err := welcomer.Queries.GetWelcomerImagesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsWelcomerImages = &database.GuildSettingsWelcomerImages{
							GuildID:                int64(*interaction.GuildID),
							ToggleEnabled:          welcomer.DefaultWelcomerImages.ToggleEnabled,
							ToggleImageBorder:      welcomer.DefaultWelcomerImages.ToggleImageBorder,
							ToggleShowAvatar:       welcomer.DefaultWelcomerImages.ToggleShowAvatar,
							BackgroundName:         welcomer.DefaultWelcomerImages.BackgroundName,
							ColourText:             welcomer.DefaultWelcomerImages.ColourText,
							ColourTextBorder:       welcomer.DefaultWelcomerImages.ColourTextBorder,
							ColourImageBorder:      welcomer.DefaultWelcomerImages.ColourImageBorder,
							ColourProfileBorder:    welcomer.DefaultWelcomerImages.ColourProfileBorder,
							ImageAlignment:         welcomer.DefaultWelcomerImages.ImageAlignment,
							ImageTheme:             welcomer.DefaultWelcomerImages.ImageTheme,
							ImageMessage:           welcomer.DefaultWelcomerImages.ImageMessage,
							ImageProfileBorderType: welcomer.DefaultWelcomerImages.ImageProfileBorderType,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get welcomer.image guild settings")

						return nil, err
					}
				}

				guildSettingsWelcomerDMs, err := welcomer.Queries.GetWelcomerDMsGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsWelcomerDMs = &database.GuildSettingsWelcomerDms{
							GuildID:             int64(*interaction.GuildID),
							ToggleEnabled:       welcomer.DefaultWelcomerDms.ToggleEnabled,
							ToggleUseTextFormat: welcomer.DefaultWelcomerDms.ToggleUseTextFormat,
							ToggleIncludeImage:  welcomer.DefaultWelcomerDms.ToggleIncludeImage,
							MessageFormat:       welcomer.DefaultWelcomerDms.MessageFormat,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get welcomer DMs guild settings")

						return nil, err
					}
				}

				// If no modules are enabled, let the user know.
				if !guildSettingsWelcomerText.ToggleEnabled && !guildSettingsWelcomerImages.ToggleEnabled && !guildSettingsWelcomerDMs.ToggleEnabled {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("No modules are enabled. Please use `/welcomer enable`", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// If text or images are enabled, but no channel is set, let the user know.
				if (guildSettingsWelcomerText.ToggleEnabled || guildSettingsWelcomerImages.ToggleEnabled) && guildSettingsWelcomerText.Channel == 0 {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("No channel is set. Please use `/welcomer setchannel`", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// GuildID may be missing, fill it in.
				member.GuildID = interaction.GuildID

				data, err := json.Marshal(core.CustomEventInvokeWelcomerStructure{
					Interaction:  &interaction,
					Member:       member,
					IgnoreDedupe: true,
				})
				if err != nil {
					return nil, err
				}

				_, err = sub.SandwichClient.RelayMessage(ctx, &sandwich.RelayMessageRequest{
					Identifier: core.GetManagerNameFromContext(ctx),
					Type:       core.CustomEventInvokeWelcomer,
					Data:       data,
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
					{Name: WelcomerModuleAll, Value: welcomer.StringToJsonLiteral(WelcomerModuleAll)},
					{Name: WelcomerModuleText, Value: welcomer.StringToJsonLiteral(WelcomerModuleText)},
					{Name: WelcomerModuleImages, Value: welcomer.StringToJsonLiteral(WelcomerModuleImages)},
					{Name: WelcomerModuleDMs, Value: welcomer.StringToJsonLiteral(WelcomerModuleDMs)},
				},
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(welcomer.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return core.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				module := subway.MustGetArgument(ctx, "module").MustString()

				guildSettingsWelcomerText, err := welcomer.Queries.GetWelcomerTextGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsWelcomerText = &database.GuildSettingsWelcomerText{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultWelcomerText.ToggleEnabled,
							Channel:       welcomer.DefaultWelcomerText.Channel,
							MessageFormat: welcomer.DefaultWelcomerText.MessageFormat,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get welcomer text guild settings")

						return nil, err
					}
				}

				guildSettingsWelcomerDMs, err := welcomer.Queries.GetWelcomerDMsGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsWelcomerDMs = &database.GuildSettingsWelcomerDms{
							GuildID:             int64(*interaction.GuildID),
							ToggleEnabled:       welcomer.DefaultWelcomerDms.ToggleEnabled,
							ToggleUseTextFormat: welcomer.DefaultWelcomerDms.ToggleUseTextFormat,
							ToggleIncludeImage:  welcomer.DefaultWelcomerDms.ToggleIncludeImage,
							MessageFormat:       welcomer.DefaultWelcomerDms.MessageFormat,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get welcomer DMs guild settings")

						return nil, err
					}
				}

				guildSettingsWelcomerImages, err := welcomer.Queries.GetWelcomerImagesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsWelcomerImages = &database.GuildSettingsWelcomerImages{
							GuildID:                int64(*interaction.GuildID),
							ToggleEnabled:          welcomer.DefaultWelcomerImages.ToggleEnabled,
							ToggleImageBorder:      welcomer.DefaultWelcomerImages.ToggleImageBorder,
							ToggleShowAvatar:       welcomer.DefaultWelcomerImages.ToggleShowAvatar,
							BackgroundName:         welcomer.DefaultWelcomerImages.BackgroundName,
							ColourText:             welcomer.DefaultWelcomerImages.ColourText,
							ColourTextBorder:       welcomer.DefaultWelcomerImages.ColourTextBorder,
							ColourImageBorder:      welcomer.DefaultWelcomerImages.ColourImageBorder,
							ColourProfileBorder:    welcomer.DefaultWelcomerImages.ColourProfileBorder,
							ImageAlignment:         welcomer.DefaultWelcomerImages.ImageAlignment,
							ImageTheme:             welcomer.DefaultWelcomerImages.ImageTheme,
							ImageMessage:           welcomer.DefaultWelcomerImages.ImageMessage,
							ImageProfileBorderType: welcomer.DefaultWelcomerImages.ImageProfileBorderType,
							UseCustomBuilder:       welcomer.DefaultWelcomerImages.UseCustomBuilder,
							CustomBuilderData:      welcomer.DefaultWelcomerImages.CustomBuilderData,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get welcomer.image guild settings")

						return nil, err
					}
				}

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
							Embeds: welcomer.NewEmbed("Unknown module: "+module, welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// Update database.

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.CreateOrUpdateWelcomerTextGuildSettingsWithAudit(ctx, database.CreateOrUpdateWelcomerTextGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsWelcomerText.ToggleEnabled,
							Channel:       guildSettingsWelcomerText.Channel,
							MessageFormat: guildSettingsWelcomerText.MessageFormat,
						}, interaction.GetUser().ID)

						return err
					},
					func() error {
						return core.EnsureGuild(ctx, discord.Snowflake(*interaction.GuildID))
					},
					nil,
				)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update welcomer text guild settings")

					return nil, err
				}

				_, err = welcomer.CreateOrUpdateWelcomerImagesGuildSettingsWithAudit(ctx, database.CreateOrUpdateWelcomerImagesGuildSettingsParams{
					GuildID:                int64(*interaction.GuildID),
					ToggleEnabled:          guildSettingsWelcomerImages.ToggleEnabled,
					ToggleImageBorder:      guildSettingsWelcomerImages.ToggleImageBorder,
					ToggleShowAvatar:       guildSettingsWelcomerImages.ToggleShowAvatar,
					BackgroundName:         guildSettingsWelcomerImages.BackgroundName,
					ColourText:             guildSettingsWelcomerImages.ColourText,
					ColourTextBorder:       guildSettingsWelcomerImages.ColourTextBorder,
					ColourImageBorder:      guildSettingsWelcomerImages.ColourImageBorder,
					ColourProfileBorder:    guildSettingsWelcomerImages.ColourProfileBorder,
					ImageAlignment:         guildSettingsWelcomerImages.ImageAlignment,
					ImageTheme:             guildSettingsWelcomerImages.ImageTheme,
					ImageMessage:           guildSettingsWelcomerImages.ImageMessage,
					ImageProfileBorderType: guildSettingsWelcomerImages.ImageProfileBorderType,
					UseCustomBuilder:       guildSettingsWelcomerImages.UseCustomBuilder,
					CustomBuilderData:      guildSettingsWelcomerImages.CustomBuilderData,
				}, interaction.GetUser().ID)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update welcomer images guild settings")

					return nil, err
				}

				_, err = welcomer.CreateOrUpdateWelcomerDMsGuildSettingsWithAudit(ctx, database.CreateOrUpdateWelcomerDMsGuildSettingsParams{
					GuildID:             int64(*interaction.GuildID),
					ToggleEnabled:       guildSettingsWelcomerDMs.ToggleEnabled,
					ToggleUseTextFormat: guildSettingsWelcomerDMs.ToggleUseTextFormat,
					ToggleIncludeImage:  guildSettingsWelcomerDMs.ToggleIncludeImage,
					MessageFormat:       guildSettingsWelcomerDMs.MessageFormat,
				}, interaction.GetUser().ID)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update welcomer DMs guild settings")

					return nil, err
				}

				switch module {
				case WelcomerModuleAll:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Enabled all modules.  Run `/welcomer test` to see the message that is sent.", welcomer.EmbedColourSuccess),
						},
					}, nil
				case WelcomerModuleText:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Enabled welcomer text messages.  Run `/welcomer test` to see the message that is sent.", welcomer.EmbedColourSuccess),
						},
					}, nil
				case WelcomerModuleImages:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Enabled welcomer images.  Run `/welcomer test` to see the message that is sent.", welcomer.EmbedColourSuccess),
						},
					}, nil
				case WelcomerModuleDMs:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Enabled welcomer direct messages.  Run `/welcomer test` to see the message that is sent.", welcomer.EmbedColourSuccess),
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
					{Name: WelcomerModuleAll, Value: welcomer.StringToJsonLiteral(WelcomerModuleAll)},
					{Name: WelcomerModuleText, Value: welcomer.StringToJsonLiteral(WelcomerModuleText)},
					{Name: WelcomerModuleImages, Value: welcomer.StringToJsonLiteral(WelcomerModuleImages)},
					{Name: WelcomerModuleDMs, Value: welcomer.StringToJsonLiteral(WelcomerModuleDMs)},
				},
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(welcomer.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return core.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				module := subway.MustGetArgument(ctx, "module").MustString()

				guildSettingsWelcomerText, err := welcomer.Queries.GetWelcomerTextGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsWelcomerText = &database.GuildSettingsWelcomerText{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultWelcomerText.ToggleEnabled,
							Channel:       welcomer.DefaultWelcomerText.Channel,
							MessageFormat: welcomer.DefaultWelcomerText.MessageFormat,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get welcomer text guild settings")

						return nil, err
					}
				}

				guildSettingsWelcomerDMs, err := welcomer.Queries.GetWelcomerDMsGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsWelcomerDMs = &database.GuildSettingsWelcomerDms{
							GuildID:             int64(*interaction.GuildID),
							ToggleEnabled:       welcomer.DefaultWelcomerDms.ToggleEnabled,
							ToggleUseTextFormat: welcomer.DefaultWelcomerDms.ToggleUseTextFormat,
							ToggleIncludeImage:  welcomer.DefaultWelcomerDms.ToggleIncludeImage,
							MessageFormat:       welcomer.DefaultWelcomerDms.MessageFormat,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get welcomer DMs guild settings")

						return nil, err
					}
				}

				guildSettingsWelcomerImages, err := welcomer.Queries.GetWelcomerImagesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsWelcomerImages = &database.GuildSettingsWelcomerImages{
							GuildID:                int64(*interaction.GuildID),
							ToggleEnabled:          welcomer.DefaultWelcomerImages.ToggleEnabled,
							ToggleImageBorder:      welcomer.DefaultWelcomerImages.ToggleImageBorder,
							ToggleShowAvatar:       welcomer.DefaultWelcomerImages.ToggleShowAvatar,
							BackgroundName:         welcomer.DefaultWelcomerImages.BackgroundName,
							ColourText:             welcomer.DefaultWelcomerImages.ColourText,
							ColourTextBorder:       welcomer.DefaultWelcomerImages.ColourTextBorder,
							ColourImageBorder:      welcomer.DefaultWelcomerImages.ColourImageBorder,
							ColourProfileBorder:    welcomer.DefaultWelcomerImages.ColourProfileBorder,
							ImageAlignment:         welcomer.DefaultWelcomerImages.ImageAlignment,
							ImageTheme:             welcomer.DefaultWelcomerImages.ImageTheme,
							ImageMessage:           welcomer.DefaultWelcomerImages.ImageMessage,
							ImageProfileBorderType: welcomer.DefaultWelcomerImages.ImageProfileBorderType,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get welcomer.image guild settings")

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
							Embeds: welcomer.NewEmbed("Unknown module: "+module, welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// Update database.

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.CreateOrUpdateWelcomerTextGuildSettingsWithAudit(ctx, database.CreateOrUpdateWelcomerTextGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsWelcomerText.ToggleEnabled,
							Channel:       guildSettingsWelcomerText.Channel,
							MessageFormat: guildSettingsWelcomerText.MessageFormat,
						}, interaction.GetUser().ID)

						return err
					},
					func() error {
						return core.EnsureGuild(ctx, discord.Snowflake(*interaction.GuildID))
					},
					nil,
				)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update welcomer text guild settings")

					return nil, err
				}

				_, err = welcomer.CreateOrUpdateWelcomerImagesGuildSettingsWithAudit(ctx, database.CreateOrUpdateWelcomerImagesGuildSettingsParams{
					GuildID:                int64(*interaction.GuildID),
					ToggleEnabled:          guildSettingsWelcomerImages.ToggleEnabled,
					ToggleImageBorder:      guildSettingsWelcomerImages.ToggleImageBorder,
					ToggleShowAvatar:       guildSettingsWelcomerImages.ToggleShowAvatar,
					BackgroundName:         guildSettingsWelcomerImages.BackgroundName,
					ColourText:             guildSettingsWelcomerImages.ColourText,
					ColourTextBorder:       guildSettingsWelcomerImages.ColourTextBorder,
					ColourImageBorder:      guildSettingsWelcomerImages.ColourImageBorder,
					ColourProfileBorder:    guildSettingsWelcomerImages.ColourProfileBorder,
					ImageAlignment:         guildSettingsWelcomerImages.ImageAlignment,
					ImageTheme:             guildSettingsWelcomerImages.ImageTheme,
					ImageMessage:           guildSettingsWelcomerImages.ImageMessage,
					ImageProfileBorderType: guildSettingsWelcomerImages.ImageProfileBorderType,
				}, interaction.GetUser().ID)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update welcomer images guild settings")

					return nil, err
				}

				_, err = welcomer.CreateOrUpdateWelcomerDMsGuildSettingsWithAudit(ctx, database.CreateOrUpdateWelcomerDMsGuildSettingsParams{
					GuildID:             int64(*interaction.GuildID),
					ToggleEnabled:       guildSettingsWelcomerDMs.ToggleEnabled,
					ToggleUseTextFormat: guildSettingsWelcomerDMs.ToggleUseTextFormat,
					ToggleIncludeImage:  guildSettingsWelcomerDMs.ToggleIncludeImage,
					MessageFormat:       guildSettingsWelcomerDMs.MessageFormat,
				}, interaction.GetUser().ID)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update welcomer DMs guild settings")

					return nil, err
				}

				switch module {
				case WelcomerModuleAll:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Disabled all modules.", welcomer.EmbedColourSuccess),
						},
					}, nil
				case WelcomerModuleText:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Disabled welcomer text messages.", welcomer.EmbedColourSuccess),
						},
					}, nil
				case WelcomerModuleImages:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Disabled welcomer images", welcomer.EmbedColourSuccess),
						},
					}, nil
				case WelcomerModuleDMs:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Disabled welcomer direct messages", welcomer.EmbedColourSuccess),
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

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(welcomer.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return core.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				channel := subway.MustGetArgument(ctx, "channel").MustChannel()

				guildSettingsWelcomerText, err := welcomer.Queries.GetWelcomerTextGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsWelcomerText = &database.GuildSettingsWelcomerText{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultWelcomerText.ToggleEnabled,
							Channel:       welcomer.DefaultWelcomerText.Channel,
							MessageFormat: welcomer.DefaultWelcomerText.MessageFormat,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get welcomer text guild settings")

						return nil, err
					}
				}

				if !channel.ID.IsNil() {
					guildSettingsWelcomerText.Channel = int64(channel.ID)
				} else {
					guildSettingsWelcomerText.Channel = 0
				}

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.CreateOrUpdateWelcomerTextGuildSettingsWithAudit(ctx, database.CreateOrUpdateWelcomerTextGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsWelcomerText.ToggleEnabled,
							Channel:       guildSettingsWelcomerText.Channel,
							MessageFormat: guildSettingsWelcomerText.MessageFormat,
						}, interaction.GetUser().ID)

						return err
					},
					func() error {
						return core.EnsureGuild(ctx, discord.Snowflake(*interaction.GuildID))
					},
					nil,
				)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update welcomer text guild settings")

					return nil, err
				}

				if !channel.ID.IsNil() {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Set welcomer channel to: <#"+channel.ID.String()+">.  Run `/welcomer test` to see the message that is sent.", welcomer.EmbedColourSuccess),
						},
					}, nil
				} else {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Removed welcomer channel. Welcomer text and image features will not work, if they are enabled.", welcomer.EmbedColourWarn),
						},
					}, nil
				}
			})
		},
	})

	welcomerGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "setmessage",
		Description: "Configure the welcomer messages on the welcomer dashboard",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(welcomer.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: []discord.Embed{
							{
								Description: fmt.Sprintf("Configure your welcomer text messages, dm messages and image messages on our dashboard [**here**](%s).", welcomer.WebsiteURL+"/dashboard/"+interaction.GuildID.String()+"/welcomer"),
								Color:       welcomer.EmbedColourInfo,
							},
						},
					},
				}, nil
			})
		},
	})

	w.InteractionCommands.MustAddInteractionCommand(welcomerGroup)

	return nil
}
