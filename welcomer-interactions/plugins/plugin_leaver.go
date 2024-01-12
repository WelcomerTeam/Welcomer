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
	leaverGroup.DMPermission = &welcomer.False

	leaverGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "enable",
		Description: "Enables leaver for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.IntToInt64Pointer(discord.PermissionElevated),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsLeaver, err := queries.GetLeaverGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to fetch leaver guild settings")

					if !errors.Is(err, pgx.ErrNoRows) {
						return nil, err
					}
				}

				guildSettingsLeaver.ToggleEnabled = true

				_, err = queries.UpdateLeaverGuildSettings(ctx, &database.UpdateLeaverGuildSettingsParams{
					GuildID:       int64(*interaction.GuildID),
					ToggleEnabled: guildSettingsLeaver.ToggleEnabled,
					Channel:       guildSettingsLeaver.Channel,
					MessageFormat: guildSettingsLeaver.MessageFormat,
				})
				if err != nil {
					sub.Logger.Error().Err(err).
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
		DefaultMemberPermission: welcomer.IntToInt64Pointer(discord.PermissionElevated),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsLeaver, err := queries.GetLeaverGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to fetch leaver guild settings")

					if !errors.Is(err, pgx.ErrNoRows) {
						return nil, err
					}
				}

				guildSettingsLeaver.ToggleEnabled = false

				_, err = queries.UpdateLeaverGuildSettings(ctx, &database.UpdateLeaverGuildSettingsParams{
					GuildID:       int64(*interaction.GuildID),
					ToggleEnabled: guildSettingsLeaver.ToggleEnabled,
					Channel:       guildSettingsLeaver.Channel,
					MessageFormat: guildSettingsLeaver.MessageFormat,
				})
				if err != nil {
					sub.Logger.Error().Err(err).
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

	leaverGroup.AddInteractionCommand(&subway.InteractionCommandable{
		Name:        "channel",
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
		DefaultMemberPermission: welcomer.IntToInt64Pointer(discord.PermissionElevated),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				channel := subway.MustGetArgument(ctx, "channel").MustChannel()

				queries := welcomer.GetQueriesFromContext(ctx)
				guildSettingsLeaver, err := queries.GetLeaverGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					return nil, err
				}

				if channel != nil {
					guildSettingsLeaver.Channel = int64(channel.ID)
				} else {
					guildSettingsLeaver.Channel = 0
				}

				_, err = queries.UpdateLeaverGuildSettings(ctx, &database.UpdateLeaverGuildSettingsParams{
					GuildID:       int64(*interaction.GuildID),
					ToggleEnabled: guildSettingsLeaver.ToggleEnabled,
					Channel:       guildSettingsLeaver.Channel,
					MessageFormat: guildSettingsLeaver.MessageFormat,
				})
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
							Embeds: welcomer.NewEmbed("Set leaver channel to: <#"+channel.ID.String()+">.", welcomer.EmbedColourSuccess),
						},
					}, nil
				} else {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Unset leaver channel. Leaver will not work, if enabled.", welcomer.EmbedColourWarn),
						},
					}, nil
				}
			})
		},
	})

	w.InteractionCommands.MustAddInteractionCommand(leaverGroup)

	return nil
}
