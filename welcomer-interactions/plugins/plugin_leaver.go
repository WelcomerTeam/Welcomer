package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
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
		"Say farewell to users to your server with a unique message.",
	)

	// Disable the leaver module for DM channels.
	leaverGroup.DMPermission = &welcomer.False

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

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				member := subway.MustGetArgument(ctx, "user").MustMember()
				if member.User == nil || member.User.ID.IsNil() {
					member = *interaction.Member
				}

				// Fetch guild settings.

				guildSettingsLeaver, err := welcomer.Queries.GetLeaverGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsLeaver = &database.GuildSettingsLeaver{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultLeaver.ToggleEnabled,
							Channel:       welcomer.DefaultLeaver.Channel,
							MessageFormat: welcomer.DefaultLeaver.MessageFormat,
						}
					} else {
						welcomer.Logger.Error().Err(err).
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
							Embeds: welcomer.NewEmbed("Leaver is not enabled. Please use `/leaver enable`", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// If text or images are enabled, but no channel is set, let the user know.
				if !guildSettingsLeaver.ToggleEnabled && guildSettingsLeaver.Channel == 0 {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("No channel is set. Please use `/leaver channel`", welcomer.EmbedColourError),
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

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				guildSettingsLeaver, err := welcomer.Queries.GetLeaverGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsLeaver = &database.GuildSettingsLeaver{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultLeaver.ToggleEnabled,
							Channel:       welcomer.DefaultLeaver.Channel,
							MessageFormat: welcomer.DefaultLeaver.MessageFormat,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get leaver guild settings")

						return nil, err
					}
				}

				guildSettingsLeaver.MessageFormat = welcomer.SetupJSONB(guildSettingsLeaver.MessageFormat)
				guildSettingsLeaver.ToggleEnabled = true

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.Queries.CreateOrUpdateLeaverGuildSettings(ctx, database.CreateOrUpdateLeaverGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsLeaver.ToggleEnabled,
							Channel:       guildSettingsLeaver.Channel,
							MessageFormat: guildSettingsLeaver.MessageFormat,
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
						Msg("Failed to update leaver guild settings")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Enabled leaver direct messages.", welcomer.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	leaverGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "disable",
		Description: "Disables leaver for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				guildSettingsLeaver, err := welcomer.Queries.GetLeaverGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsLeaver = &database.GuildSettingsLeaver{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultLeaver.ToggleEnabled,
							Channel:       welcomer.DefaultLeaver.Channel,
							MessageFormat: welcomer.DefaultLeaver.MessageFormat,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get leaver guild settings")

						return nil, err
					}
				}

				guildSettingsLeaver.MessageFormat = welcomer.SetupJSONB(guildSettingsLeaver.MessageFormat)
				guildSettingsLeaver.ToggleEnabled = false

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.Queries.CreateOrUpdateLeaverGuildSettings(ctx, database.CreateOrUpdateLeaverGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsLeaver.ToggleEnabled,
							Channel:       guildSettingsLeaver.Channel,
							MessageFormat: guildSettingsLeaver.MessageFormat,
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
						Msg("Failed to update leaver guild settings")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Disabled leaver direct messages.", welcomer.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	leaverGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "setmessage",
		Description: "Configure the leaver messages on the welcomer dashboard",

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
								Description: fmt.Sprintf("Configure your leaver message on our dashboard [**here**](%s).", welcomer.WebsiteURL+"/dashboard/"+interaction.GuildID.String()+"/leaver"),
								Color:       welcomer.EmbedColourInfo,
							},
						},
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

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				channel := subway.MustGetArgument(ctx, "channel").MustChannel()

				guildSettingsLeaver, err := welcomer.Queries.GetLeaverGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsLeaver = &database.GuildSettingsLeaver{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultLeaver.ToggleEnabled,
							Channel:       welcomer.DefaultLeaver.Channel,
							MessageFormat: welcomer.DefaultLeaver.MessageFormat,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get leaver guild settings")

						return nil, err
					}
				}

				guildSettingsLeaver.MessageFormat = welcomer.SetupJSONB(guildSettingsLeaver.MessageFormat)
				guildSettingsLeaver.Channel = welcomer.If(!channel.ID.IsNil(), int64(channel.ID), 0)

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.Queries.CreateOrUpdateLeaverGuildSettings(ctx, database.CreateOrUpdateLeaverGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsLeaver.ToggleEnabled,
							Channel:       guildSettingsLeaver.Channel,
							MessageFormat: guildSettingsLeaver.MessageFormat,
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
						Msg("Failed to update leaver guild settings")

					return nil, err
				}

				if !channel.ID.IsNil() {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Set leaver channel to: <#"+channel.ID.String()+">.", welcomer.EmbedColourSuccess),
						},
					}, nil
				} else {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Removed leaver channel. Leaver will not work, if enabled.", welcomer.EmbedColourWarn),
						},
					}, nil
				}
			})
		},
	})

	w.InteractionCommands.MustAddInteractionCommand(leaverGroup)

	return nil
}
