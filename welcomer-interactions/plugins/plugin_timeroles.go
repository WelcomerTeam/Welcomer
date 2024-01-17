package plugins

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4"
)

func NewTimeRolesCog() *TimeRolesCog {
	return &TimeRolesCog{
		InteractionCommands: subway.SetupInteractionCommandable(&subway.InteractionCommandable{}),
	}
}

type TimeRolesCog struct {
	InteractionCommands *subway.InteractionCommandable
}

// Assert types.

var (
	_ subway.Cog                        = (*TimeRolesCog)(nil)
	_ subway.CogWithInteractionCommands = (*TimeRolesCog)(nil)
)

func (r *TimeRolesCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "TimeRoles",
		Description: "Provides the cog for the 'TimeRoles' feature.",
	}
}

func (r *TimeRolesCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return r.InteractionCommands
}

func (r *TimeRolesCog) RegisterCog(sub *subway.Subway) error {
	ruleGroup := subway.NewSubcommandGroup(
		"timeroles",
		"Provides timeroles for the server.",
	)

	// Disable the TimeRoles module for DM channels.
	ruleGroup.DMPermission = &welcomer.False

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "enable",
		Description: "Enable timerole for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.IntToInt64Pointer(discord.PermissionElevated),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsTimeRoles, err := queries.GetTimeRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get TimeRoles guild settings.")
				}

				guildSettingsTimeRoles.ToggleEnabled = true

				_, err = queries.UpdateTimeRolesGuildSettings(ctx, &database.UpdateTimeRolesGuildSettingsParams{
					GuildID:       int64(*interaction.GuildID),
					ToggleEnabled: guildSettingsTimeRoles.ToggleEnabled,
					Timeroles:     guildSettingsTimeRoles.Timeroles,
				})
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update TimeRoles guild settings.")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Enabled TimeRoles. Run `/TimeRoles list` to see the list of TimeRoles configured.", welcomer.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "disable",
		Description: "Disables timerole for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.IntToInt64Pointer(discord.PermissionElevated),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsTimeRoles, err := queries.GetTimeRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get TimeRoles guild settings.")
				}

				guildSettingsTimeRoles.ToggleEnabled = false

				_, err = queries.UpdateTimeRolesGuildSettings(ctx, &database.UpdateTimeRolesGuildSettingsParams{
					GuildID:       int64(*interaction.GuildID),
					ToggleEnabled: guildSettingsTimeRoles.ToggleEnabled,
					Timeroles:     guildSettingsTimeRoles.Timeroles,
				})
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update TimeRoles guild settings.")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Disabled TimeRoles.", welcomer.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "check",
		Description: "Check your TimeRoles progress on the server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission: &welcomer.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsTimeRoles, err := queries.GetTimeRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get TimeRoles guild settings.")

					return nil, err
				}

				embeds := []*discord.Embed{}
				embed := &discord.Embed{Title: "TimeRoles", Color: welcomer.EmbedColourInfo}

				timeRoleList := welcomer.UnmarshalTimeRolesJSON(guildSettingsTimeRoles.Timeroles.Bytes)

				timeRoleList, err = welcomer.FilterAssignableTimeRoles(ctx, sub, interaction, timeRoleList)
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to filter assignable roles.")

					return nil, err
				}

				if len(timeRoleList) == 0 || !guildSettingsTimeRoles.ToggleEnabled {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("There are no TimeRoles set for this server.", welcomer.EmbedColourInfo),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// Sort timeRoleList by Seconds in ascending order
				sort.Slice(timeRoleList, func(i, j int) bool {
					return timeRoleList[i].Seconds < timeRoleList[j].Seconds
				})

				now := time.Now()

				hasTimeRoleRemaining := false
				var timeRoleRemainingPercent int

				// Check if the user has any time roles remaining.
				// If they do, then set the embed description to the first one.
				for _, role := range timeRoleList {
					roleGivenAt := interaction.Member.JoinedAt.Add(time.Second * time.Duration(role.Seconds))

					if roleGivenAt.After(now) {
						embed.Description = fmt.Sprintf(
							"You joined this server <t:%d:R>!\n\nNext role: <@&%d>\nTime until next role: <t:%d:R>\n\n",
							interaction.Member.JoinedAt.Unix(),
							role.Role,
							interaction.Member.JoinedAt.Add(time.Second*time.Duration(role.Seconds)).Unix(),
						)

						hasTimeRoleRemaining = true
						timeRoleRemainingPercent = int((float64(time.Since(interaction.Member.JoinedAt).Seconds()) /
							float64(role.Seconds)) * 100)

						break
					}
				}

				// If the user has no time roles remaining then let them know.
				if !hasTimeRoleRemaining {
					embed.Description = fmt.Sprintf(
						"You joined this server <t:%d:R>!\n\nThere are no more roles left!\n\n",
						interaction.Member.JoinedAt.Unix(),
					)
				}

				// List all the time roles.
				for _, role := range timeRoleList {
					roleGivenAt := interaction.Member.JoinedAt.Add(time.Second * time.Duration(role.Seconds))

					var roleMessage string

					if roleGivenAt.After(now) {
						roleMessage = fmt.Sprintf(welcomer.EmojiNeutral+" <@&%d> <t:%d:R>\n", role.Role, roleGivenAt.Unix())
					} else {
						roleMessage = fmt.Sprintf(welcomer.EmojiCheck+" <@&%d>\n", role.Role)
					}

					// If the embed content will go over 4000 characters then create a new embed and continue from that one.
					if len(embed.Description)+len(roleMessage) > 4000 {
						embeds = append(embeds, embed)
						embed = &discord.Embed{Color: welcomer.EmbedColourInfo}
					}

					embed.Description += roleMessage
				}

				embeds = append(embeds, embed)

				if hasTimeRoleRemaining {
					embeds[len(embeds)-1].SetImage(&discord.EmbedImage{
						URL: fmt.Sprintf("https://cdn.welcomer.gg/minecraftxp.png?percent=%d", timeRoleRemainingPercent),
					})
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: embeds,
						Flags:  uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			})
		},
	})

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "list",
		Description: "List the TimeRoles for the server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission: &welcomer.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsTimeRoles, err := queries.GetTimeRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get TimeRoles guild settings.")

					return nil, err
				}

				embeds := []*discord.Embed{}
				embed := &discord.Embed{Title: "TimeRoles", Color: welcomer.EmbedColourInfo}

				timeRoleList := welcomer.UnmarshalTimeRolesJSON(guildSettingsTimeRoles.Timeroles.Bytes)

				timeRoleList, err = welcomer.FilterAssignableTimeRoles(ctx, sub, interaction, timeRoleList)
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to filter assignable roles.")

					return nil, err
				}

				if len(timeRoleList) == 0 || !guildSettingsTimeRoles.ToggleEnabled {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("There are no TimeRoles set for this server.", welcomer.EmbedColourInfo),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// Sort timeRoleList by Seconds in ascending order
				sort.Slice(timeRoleList, func(i, j int) bool {
					return timeRoleList[i].Seconds < timeRoleList[j].Seconds
				})

				for _, role := range timeRoleList {
					roleMessage := fmt.Sprintf("- <@&%d> - `%s`\n", role.Role, welcomer.HumanizeDuration(role.Seconds))

					// If the embed content will go over 4000 characters then create a new embed and continue from that one.
					if len(embed.Description)+len(roleMessage) > 4000 {
						embeds = append(embeds, embed)
						embed = &discord.Embed{Color: welcomer.EmbedColourInfo}
					}

					embed.Description += roleMessage
				}

				embeds = append(embeds, embed)

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: embeds,
						Flags:  uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			})
		},
	})

	r.InteractionCommands.MustAddInteractionCommand(ruleGroup)

	return nil
}
