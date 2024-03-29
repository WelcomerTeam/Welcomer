package plugins

import (
	"context"
	"errors"
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4"
	jsoniter "github.com/json-iterator/go"
)

func NewTempChannelsCog() *TempChannelsCog {
	return &TempChannelsCog{
		InteractionCommands: subway.SetupInteractionCommandable(&subway.InteractionCommandable{}),
	}
}

type TempChannelsCog struct {
	InteractionCommands *subway.InteractionCommandable
}

// Assert types.

var (
	_ subway.Cog                        = (*TempChannelsCog)(nil)
	_ subway.CogWithInteractionCommands = (*TempChannelsCog)(nil)
)

const (
	TempChannelsModuleTempChannels = "tempchannels"
	TempChannelsModuleAutoPurge    = "autopurge"
)

func (w *TempChannelsCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "TempChannels",
		Description: "Provides the functionality for the 'TempChannels' feature",
	}
}

func (w *TempChannelsCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return w.InteractionCommands
}

func (w *TempChannelsCog) RegisterCog(sub *subway.Subway) error {
	tempchannelsGroup := subway.NewSubcommandGroup(
		"tempchannels",
		"Welcome new users to your server with fancy images, text or send them a direct message.",
	)

	// Disable the tempchannels module for DM channels.
	tempchannelsGroup.DMPermission = &welcomer.False

	tempchannelsGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "give",
		Description: "Gives you a tempchannel.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission: &welcomer.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsTempChannels, err := queries.GetTempChannelsGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get tempchannels guild settings")
				}

				if !guildSettingsTempChannels.ToggleEnabled || guildSettingsTempChannels.ChannelCategory == 0 {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Tempchannels is not enabled on this server.", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// GuildID may be missing, fill it in.
				interaction.Member.GuildID = interaction.GuildID

				data, err := jsoniter.Marshal(welcomer.CustomEventInvokeTempChannelsStructure{
					Interaction: &interaction,
					Member:      *interaction.Member,
				})
				if err != nil {
					return nil, err
				}

				_, err = sub.SandwichClient.RelayMessage(ctx, &sandwich.RelayMessageRequest{
					Manager: welcomer.GetManagerNameFromContext(ctx),
					Type:    welcomer.CustomEventInvokeTempChannels,
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

	tempchannelsGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "remove",
		Description: "Removes your tempchannel.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission: &welcomer.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
				// GuildID may be missing, fill it in.
				interaction.Member.GuildID = interaction.GuildID

				data, err := jsoniter.Marshal(welcomer.CustomEventInvokeTempChannelsRemoveStructure{
					Interaction: &interaction,
					Member:      *interaction.Member,
				})
				if err != nil {
					return nil, err
				}

				_, err = sub.SandwichClient.RelayMessage(ctx, &sandwich.RelayMessageRequest{
					Manager: welcomer.GetManagerNameFromContext(ctx),
					Type:    welcomer.CustomEventInvokeTempChannelsRemove,
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

	tempchannelsGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "enable",
		Description: "Enables a tempchannels module for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeString,
				Name:         "module",
				Description:  "The module to enable.",

				Choices: []*discord.ApplicationCommandOptionChoice{
					{Name: TempChannelsModuleTempChannels, Value: welcomer.StringToJsonLiteral(TempChannelsModuleTempChannels)},
					{Name: TempChannelsModuleAutoPurge, Value: welcomer.StringToJsonLiteral(TempChannelsModuleAutoPurge)},
				},
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				module := subway.MustGetArgument(ctx, "module").MustString()

				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsTempChannels, err := queries.GetTempChannelsGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get tempchannels guild settings")
				}

				switch module {
				case TempChannelsModuleTempChannels:
					guildSettingsTempChannels.ToggleEnabled = true
				case TempChannelsModuleAutoPurge:
					guildSettingsTempChannels.ToggleAutopurge = true
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

				_, err = queries.UpdateTempChannelsGuildSettings(ctx, &database.UpdateTempChannelsGuildSettingsParams{
					GuildID:          int64(*interaction.GuildID),
					ToggleEnabled:    guildSettingsTempChannels.ToggleEnabled,
					ToggleAutopurge:  guildSettingsTempChannels.ToggleAutopurge,
					ChannelLobby:     guildSettingsTempChannels.ChannelLobby,
					ChannelCategory:  guildSettingsTempChannels.ChannelCategory,
					DefaultUserCount: guildSettingsTempChannels.DefaultUserCount,
				})
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update tempchannels text guild settings")

					return nil, err
				}

				switch module {
				case TempChannelsModuleTempChannels:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Enabled tempchannels. Users can now use `/tempchannels give` or join the lobby channel, if set.", welcomer.EmbedColourSuccess),
						},
					}, nil
				case TempChannelsModuleAutoPurge:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Enabled autopurge. Tempchannels will now be cleared when they are left empty.", welcomer.EmbedColourSuccess),
						},
					}, nil
				default:
					return nil, nil // Unreachable
				}
			})
		},
	})

	tempchannelsGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "disable",
		Description: "Disables a tempchannels module for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeString,
				Name:         "module",
				Description:  "The module to disable.",

				Choices: []*discord.ApplicationCommandOptionChoice{
					{Name: TempChannelsModuleTempChannels, Value: welcomer.StringToJsonLiteral(TempChannelsModuleTempChannels)},
					{Name: TempChannelsModuleAutoPurge, Value: welcomer.StringToJsonLiteral(TempChannelsModuleAutoPurge)},
				},
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				module := subway.MustGetArgument(ctx, "module").MustString()

				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsTempChannels, err := queries.GetTempChannelsGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get tempchannels guild settings")
				}

				switch module {
				case TempChannelsModuleTempChannels:
					guildSettingsTempChannels.ToggleEnabled = false
				case TempChannelsModuleAutoPurge:
					guildSettingsTempChannels.ToggleAutopurge = false
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

				_, err = queries.UpdateTempChannelsGuildSettings(ctx, &database.UpdateTempChannelsGuildSettingsParams{
					GuildID:          int64(*interaction.GuildID),
					ToggleEnabled:    guildSettingsTempChannels.ToggleEnabled,
					ToggleAutopurge:  guildSettingsTempChannels.ToggleAutopurge,
					ChannelLobby:     guildSettingsTempChannels.ChannelLobby,
					ChannelCategory:  guildSettingsTempChannels.ChannelCategory,
					DefaultUserCount: guildSettingsTempChannels.DefaultUserCount,
				})
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update tempchannels text guild settings")

					return nil, err
				}

				switch module {
				case TempChannelsModuleTempChannels:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Disabled tempchannels.", welcomer.EmbedColourSuccess),
						},
					}, nil
				case TempChannelsModuleAutoPurge:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Disabled autopurge.", welcomer.EmbedColourSuccess),
						},
					}, nil
				default:
					return nil, nil // Unreachable
				}
			})
		},
	})

	tempchannelsGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "setcategory",
		Description: "Sets the category to create tempchannels in.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeCategoryChannel,
				Name:         "category",
				Description:  "The category you would like to create tempchannels in.",
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				channel := subway.MustGetArgument(ctx, "category").MustChannel()

				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsTempChannels, err := queries.GetTempChannelsGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get tempchannels guild settings")
				}

				guildSettingsTempChannels.ChannelCategory = int64(channel.ID)

				_, err = queries.UpdateTempChannelsGuildSettings(ctx, &database.UpdateTempChannelsGuildSettingsParams{
					GuildID:          int64(*interaction.GuildID),
					ToggleEnabled:    guildSettingsTempChannels.ToggleEnabled,
					ToggleAutopurge:  guildSettingsTempChannels.ToggleAutopurge,
					ChannelLobby:     guildSettingsTempChannels.ChannelLobby,
					ChannelCategory:  guildSettingsTempChannels.ChannelCategory,
					DefaultUserCount: guildSettingsTempChannels.DefaultUserCount,
				})
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update tempchannels guild settings")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Set tempchannels category to: <#"+channel.ID.String()+">.", welcomer.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	tempchannelsGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "setlobby",
		Description: "Sets the tempchannels lobby in.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     false,
				ArgumentType: subway.ArgumentTypeVoiceChannel,
				Name:         "channel",
				Description:  "The voice channel you would like to make the tempchannels lobby.",
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				channel := subway.MustGetArgument(ctx, "channel").MustChannel()

				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsTempChannels, err := queries.GetTempChannelsGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get tempchannels guild settings")
				}

				if channel != nil {
					guildSettingsTempChannels.ChannelLobby = int64(channel.ID)
				} else {
					guildSettingsTempChannels.ChannelLobby = 0
				}

				_, err = queries.UpdateTempChannelsGuildSettings(ctx, &database.UpdateTempChannelsGuildSettingsParams{
					GuildID:          int64(*interaction.GuildID),
					ToggleEnabled:    guildSettingsTempChannels.ToggleEnabled,
					ToggleAutopurge:  guildSettingsTempChannels.ToggleAutopurge,
					ChannelLobby:     guildSettingsTempChannels.ChannelLobby,
					ChannelCategory:  guildSettingsTempChannels.ChannelCategory,
					DefaultUserCount: guildSettingsTempChannels.DefaultUserCount,
				})
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update tempchannels guild settings")

					return nil, err
				}

				if channel != nil {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed(fmt.Sprintf(
								"Set tempchannels lobby to: <#%d>.\n\nWhen tempchannels is enabled, users will be able to join <#%d> to be automatically moved to a tempchannel, without running `/tempchannels give`.",
								channel.ID, channel.ID,
							), welcomer.EmbedColourSuccess),
						},
					}, nil
				} else {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Removed tempchannels lobby.", welcomer.EmbedColourSuccess),
						},
					}, nil
				}
			})
		},
	})

	w.InteractionCommands.MustAddInteractionCommand(tempchannelsGroup)

	return nil
}
