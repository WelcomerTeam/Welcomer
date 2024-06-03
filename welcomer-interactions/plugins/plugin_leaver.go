package plugins

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/jackc/pgx/v4"
)

func NewLeaverCog() *LeaverCog {
	return &LeaverCog{
		InteractionCommands: subway.SetupInteractionCommandable(&subway.InteractionCommandable{}),
	}
}

type LeaverCog struct {
	InteractionCommands *subway.InteractionCommandable
}

// Assert types.

var (
	_ subway.Cog                        = (*LeaverCog)(nil)
	_ subway.CogWithInteractionCommands = (*LeaverCog)(nil)
)

const (
	LeaverModuleLeaver = "leaver"
	LeaverModuleDMs    = "dms"
)

func (w *LeaverCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "Leaver",
		Description: "Provides the functionality for the 'Leaver' feature",
	}
}

func (w *LeaverCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return w.InteractionCommands
}

func (w *LeaverCog) RegisterCog(sub *subway.Subway) error {
	leaverGroup := subway.NewSubcommandGroup(
		"leaver",
		"Welcome new users to your server with fancy images, text or send them a direct message.",
	)

	// Disable the leaver module for DM channels.
	leaverGroup.DMPermission = &utils.False

	leaverGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "test",
		Description: "Tests the Leaver functionality.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     false,
				ArgumentType: subway.ArgumentTypeMember,
				Name:         "user",
				Description:  "The user you would like to send the leaver message for.",
			},
		},

		DMPermission:            &utils.False,
		DefaultMemberPermission: utils.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				member := subway.MustGetArgument(ctx, "user").MustMember()
				if member == nil {
					member = interaction.Member
				}

				queries := welcomer.GetQueriesFromContext(ctx)

				// Fetch guild settings.

				guildSettingsLeaver, err := queries.GetLeaverGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsLeaver = &database.GuildSettingsLeaver{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: database.DefaultLeaver.ToggleEnabled,
							Channel:       database.DefaultLeaver.Channel,
							MessageFormat: database.DefaultLeaver.MessageFormat,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get leaver guild settings")

						return nil, err
					}
				}

				// If no modules are enabled, let the user know.
				if !guildSettingsLeaver.ToggleEnabled {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Leaver is not enabled. Please use `/leaver enable`", utils.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// If text or images are enabled, but no channel is set, let the user know.
				if !guildSettingsLeaver.ToggleEnabled && guildSettingsLeaver.Channel == 0 {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("No channel is set. Please use `/leaver channel`", utils.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				data, err := json.Marshal(core.CustomEventInvokeLeaverStructure{
					Interaction: &interaction,
					User:        *member.User,
					GuildID:     *interaction.GuildID,
				})
				if err != nil {
					return nil, err
				}

				_, err = sub.SandwichClient.RelayMessage(ctx, &sandwich.RelayMessageRequest{
					Manager: welcomer.GetManagerNameFromContext(ctx),
					Type:    core.CustomEventInvokeLeaver,
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

	leaverGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "enable",
		Description: "Enables leaver for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &utils.False,
		DefaultMemberPermission: utils.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsLeaver, err := queries.GetLeaverGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsLeaver = &database.GuildSettingsLeaver{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: database.DefaultLeaver.ToggleEnabled,
							Channel:       database.DefaultLeaver.Channel,
							MessageFormat: database.DefaultLeaver.MessageFormat,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get leaver guild settings")

						return nil, err
					}
				}

				guildSettingsLeaver.MessageFormat = utils.SetupJSONB(guildSettingsLeaver.MessageFormat)
				guildSettingsLeaver.ToggleEnabled = true

				err = utils.RetryWithFallback(
					func() error {
						_, err = queries.CreateOrUpdateLeaverGuildSettings(ctx, database.CreateOrUpdateLeaverGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsLeaver.ToggleEnabled,
							Channel:       guildSettingsLeaver.Channel,
							MessageFormat: guildSettingsLeaver.MessageFormat,
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
						Msg("Failed to update leaver guild settings")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: utils.NewEmbed("Enabled leaver direct messages.", utils.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	leaverGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "disable",
		Description: "Disables leaver for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &utils.False,
		DefaultMemberPermission: utils.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsLeaver, err := queries.GetLeaverGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsLeaver = &database.GuildSettingsLeaver{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: database.DefaultLeaver.ToggleEnabled,
							Channel:       database.DefaultLeaver.Channel,
							MessageFormat: database.DefaultLeaver.MessageFormat,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get leaver guild settings")

						return nil, err
					}
				}

				guildSettingsLeaver.MessageFormat = utils.SetupJSONB(guildSettingsLeaver.MessageFormat)
				guildSettingsLeaver.ToggleEnabled = false

				err = utils.RetryWithFallback(
					func() error {
						_, err = queries.CreateOrUpdateLeaverGuildSettings(ctx, database.CreateOrUpdateLeaverGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsLeaver.ToggleEnabled,
							Channel:       guildSettingsLeaver.Channel,
							MessageFormat: guildSettingsLeaver.MessageFormat,
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
						Msg("Failed to update leaver guild settings")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: utils.NewEmbed("Disabled leaver direct messages.", utils.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	leaverGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "setchannel",
		Description: "Sets the channel to send leaver messages to.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     false,
				ArgumentType: subway.ArgumentTypeTextChannel,
				Name:         "channel",
				Description:  "The channel you would like to send the leave message to.",
			},
		},

		DMPermission:            &utils.False,
		DefaultMemberPermission: utils.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				channel := subway.MustGetArgument(ctx, "channel").MustChannel()

				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsLeaver, err := queries.GetLeaverGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsLeaver = &database.GuildSettingsLeaver{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: database.DefaultLeaver.ToggleEnabled,
							Channel:       database.DefaultLeaver.Channel,
							MessageFormat: database.DefaultLeaver.MessageFormat,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get leaver guild settings")

						return nil, err
					}
				}

				guildSettingsLeaver.MessageFormat = utils.SetupJSONB(guildSettingsLeaver.MessageFormat)
				guildSettingsLeaver.Channel = utils.If(channel != nil, int64(channel.ID), 0)

				err = utils.RetryWithFallback(
					func() error {
						_, err = queries.CreateOrUpdateLeaverGuildSettings(ctx, database.CreateOrUpdateLeaverGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsLeaver.ToggleEnabled,
							Channel:       guildSettingsLeaver.Channel,
							MessageFormat: guildSettingsLeaver.MessageFormat,
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
						Msg("Failed to update leaver guild settings")

					return nil, err
				}

				if channel != nil {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Set leaver channel to: <#"+channel.ID.String()+">.", utils.EmbedColourSuccess),
						},
					}, nil
				} else {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Removed leaver channel. Leaver will not work, if enabled.", utils.EmbedColourWarn),
						},
					}, nil
				}
			})
		},
	})

	w.InteractionCommands.MustAddInteractionCommand(leaverGroup)

	return nil
}
