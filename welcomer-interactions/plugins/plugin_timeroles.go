package plugins

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
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
		"Automatically assign roles to users depending on how long they have been in the server.",
	)

	// Disable the TimeRoles module for DM channels.
	ruleGroup.DMPermission = &welcomer.False

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "enable",
		Description: "Enable timeroles for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				guildSettingsTimeRoles, err := welcomer.Queries.GetTimeRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsTimeRoles = &database.GuildSettingsTimeroles{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultTimeRoles.ToggleEnabled,
							Timeroles:     welcomer.DefaultTimeRoles.Timeroles,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get timeroles guild settings")

						return nil, err
					}
				}

				guildSettingsTimeRoles.Timeroles = welcomer.SetupJSONB(guildSettingsTimeRoles.Timeroles)
				guildSettingsTimeRoles.ToggleEnabled = true

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.Queries.CreateOrUpdateTimeRolesGuildSettings(ctx, database.CreateOrUpdateTimeRolesGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsTimeRoles.ToggleEnabled,
							Timeroles:     guildSettingsTimeRoles.Timeroles,
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
						Msg("Failed to update timeroles guild settings")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Enabled timeroles. Run `/timeroles list` to see the list of timeroles configured.", welcomer.EmbedColourSuccess),
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
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				guildSettingsTimeRoles, err := welcomer.Queries.GetTimeRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsTimeRoles = &database.GuildSettingsTimeroles{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultTimeRoles.ToggleEnabled,
							Timeroles:     welcomer.DefaultTimeRoles.Timeroles,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get TimeRoles guild settings")

						return nil, err
					}
				}
				guildSettingsTimeRoles.Timeroles = welcomer.SetupJSONB(guildSettingsTimeRoles.Timeroles)
				guildSettingsTimeRoles.ToggleEnabled = false

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.Queries.CreateOrUpdateTimeRolesGuildSettings(ctx, database.CreateOrUpdateTimeRolesGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsTimeRoles.ToggleEnabled,
							Timeroles:     guildSettingsTimeRoles.Timeroles,
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
						Msg("Failed to update timeroles guild settings")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Disabled timeroles.", welcomer.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "check",
		Description: "Check your timeroles progress on the server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission: &welcomer.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
				guildSettingsTimeRoles, err := welcomer.Queries.GetTimeRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsTimeRoles = &database.GuildSettingsTimeroles{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultTimeRoles.ToggleEnabled,
							Timeroles:     welcomer.DefaultTimeRoles.Timeroles,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get timeroles guild settings")

						return nil, err
					}
				}

				embeds := []discord.Embed{}
				embed := discord.Embed{Title: "TimeRoles", Color: welcomer.EmbedColourInfo}

				timeRoleList := welcomer.UnmarshalTimeRolesJSON(guildSettingsTimeRoles.Timeroles.Bytes)

				timeRoleList, err = welcomer.FilterAssignableTimeRoles(ctx, sub.SandwichClient, int64(*interaction.GuildID), int64(interaction.ApplicationID), timeRoleList)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to filter assignable roles")

					return nil, err
				}

				if !guildSettingsTimeRoles.ToggleEnabled {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Timeroles are disabled for this server.", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				if len(timeRoleList) == 0 {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("There are no timeroles set for this server.", welcomer.EmbedColourInfo),
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
						embed = discord.Embed{Color: welcomer.EmbedColourInfo}
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
		Description: "List the timeroles for the server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission: &welcomer.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				guildSettingsTimeRoles, err := welcomer.Queries.GetTimeRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsTimeRoles = &database.GuildSettingsTimeroles{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultTimeRoles.ToggleEnabled,
							Timeroles:     welcomer.DefaultTimeRoles.Timeroles,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get timeroles guild settings")

						return nil, err
					}
				}

				embeds := []discord.Embed{}
				embed := discord.Embed{Title: "TimeRoles", Color: welcomer.EmbedColourInfo}

				timeRoleList := welcomer.UnmarshalTimeRolesJSON(guildSettingsTimeRoles.Timeroles.Bytes)

				timeRoleList, err = welcomer.FilterAssignableTimeRoles(ctx, sub.SandwichClient, int64(*interaction.GuildID), int64(interaction.ApplicationID), timeRoleList)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to filter assignable roles")

					return nil, err
				}

				if !guildSettingsTimeRoles.ToggleEnabled && !welcomer.IsInterationAuthorElevated(sub, interaction) {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Timeroles are disabled for this server.", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				if len(timeRoleList) == 0 {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("There are no timeroles set for this server.", welcomer.EmbedColourInfo),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// Sort timeRoleList by Seconds in ascending order
				sort.Slice(timeRoleList, func(i, j int) bool {
					return timeRoleList[i].Seconds < timeRoleList[j].Seconds
				})

				for _, role := range timeRoleList {
					roleMessage := fmt.Sprintf("- <@&%d> - `%s`\n", role.Role, welcomer.HumanizeDuration(role.Seconds, true))

					// If the embed content will go over 4000 characters then create a new embed and continue from that one.
					if len(embed.Description)+len(roleMessage) > 4000 {
						embeds = append(embeds, embed)
						embed = discord.Embed{Color: welcomer.EmbedColourInfo}
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

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "addrole",
		Description: "Add a timerole to the server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Name:         "role",
				Description:  "The role to assign.",
				ArgumentType: subway.ArgumentTypeRole,
				Required:     true,
			},
			{
				Name:         "duration",
				Description:  "The duration after which the role will be assigned (e.g., `1h`, `30m`).",
				ArgumentType: subway.ArgumentTypeString,
				Required:     true,
			},
			{
				Required:     false,
				ArgumentType: subway.ArgumentTypeBool,
				Name:         "ignore-permissions",
				Description:  "Ignores role permissions.",
			},
		},

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				role := subway.MustGetArgument(ctx, "role").MustRole()
				durationString := subway.MustGetArgument(ctx, "duration").MustString()
				ignoreRolePermissions := subway.MustGetArgument(ctx, "ignore-permissions").MustBool()

				var seconds int

				// Check if the durationString is only a number
				if secondsValue, err := strconv.Atoi(durationString); err == nil && secondsValue >= 0 {
					seconds = secondsValue
				} else {
					var err error

					seconds, err = welcomer.ParseDurationAsSeconds(durationString)
					if err != nil || seconds < 0 {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Str("duration", durationString).
							Msg("Failed to parse duration")

						return &discord.InteractionResponse{
							Type: discord.InteractionCallbackTypeChannelMessageSource,
							Data: &discord.InteractionCallbackData{
								Embeds: welcomer.NewEmbed("Invalid duration. It must be a positive number in a valid format (e.g., `5y`, `30d`, `1h`, `30m`, `3600s`, `3600`). Only years, days, hours, minutes and seconds are supported.", welcomer.EmbedColourError),
								Flags:  uint32(discord.MessageFlagEphemeral),
							},
						}, nil
					}
				}

				canAssignRoles, isRoleAssignable, isRoleElevated, err := welcomer.Accelerator_CanAssignRole(ctx, *interaction.GuildID, role)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to check if welcomer can assign role")

					return nil, err
				}

				if !canAssignRoles {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Welcomer is missing permissions to assign roles", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// Check if the role is assignable by welcomer using the guild roles and roles Welcomer has.
				if !isRoleAssignable {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("### This role is not assignable\nWelcomer cannot assign users this role as Welcomer's highest role is below this role's position. Please rearrange your roles in the server settings to move Welcomer's role above this role.", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				if !ignoreRolePermissions && isRoleElevated {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("### This role is elevated\nThis role has elevated permissions. If you are sure you want to use this role, please run the command again with ignore-permissions set to true.\n\nPermissions:\n"+welcomer.GetRolePermissionListAsString(int(role.Permissions)), welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				guildSettingsTempChannels, err := welcomer.Queries.GetTimeRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsTempChannels = &database.GuildSettingsTimeroles{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultTimeRoles.ToggleEnabled,
							Timeroles:     welcomer.DefaultTimeRoles.Timeroles,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get timeroles guild settings")

						return nil, err
					}
				}

				timeRoles := welcomer.UnmarshalTimeRolesJSON(guildSettingsTempChannels.Timeroles.Bytes)

				// Check if the role already exists in the list
				if slices.ContainsFunc(timeRoles, func(tr welcomer.GuildSettingsTimeRolesRole) bool { return tr.Role == role.ID }) {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("This role is already in the list.", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				timeRoles = append(timeRoles, welcomer.GuildSettingsTimeRolesRole{role.ID, seconds})

				// Update the guild settings with the new timeRoles
				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.Queries.CreateOrUpdateTimeRolesGuildSettings(ctx, database.CreateOrUpdateTimeRolesGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsTempChannels.ToggleEnabled,
							Timeroles:     welcomer.BytesToJSONB(welcomer.MarshalTimeRolesJSON(timeRoles)),
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
						Msg("Failed to update timeroles guild settings")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed(fmt.Sprintf("Added timerole <@&%d> with duration `%s`. Run `/timeroles list` to see the list of timeroles configured.", role.ID, welcomer.HumanizeDuration(seconds, true)), welcomer.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "removerole",
		Description: "Remove a timerole from the server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Name:         "role",
				Description:  "The role to remove.",
				ArgumentType: subway.ArgumentTypeRole,
				Required:     true,
			},
		},

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				role := subway.MustGetArgument(ctx, "role").MustRole()

				guildSettingsTimeRoles, err := welcomer.Queries.GetTimeRolesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsTimeRoles = &database.GuildSettingsTimeroles{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: welcomer.DefaultTimeRoles.ToggleEnabled,
							Timeroles:     welcomer.DefaultTimeRoles.Timeroles,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get timeroles guild settings")

						return nil, err
					}
				}

				timeRoles := welcomer.UnmarshalTimeRolesJSON(guildSettingsTimeRoles.Timeroles.Bytes)

				// Check if the role exists in the list.
				if !slices.ContainsFunc(timeRoles, func(tr welcomer.GuildSettingsTimeRolesRole) bool { return tr.Role == role.ID }) {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("This role is not in the list.", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				// Remove the role from the list.
				timeRoles = slices.DeleteFunc(timeRoles, func(tr welcomer.GuildSettingsTimeRolesRole) bool { return tr.Role == role.ID })

				// Update the guild settings with the new timeRoles
				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.Queries.CreateOrUpdateTimeRolesGuildSettings(ctx, database.CreateOrUpdateTimeRolesGuildSettingsParams{
							GuildID:       int64(*interaction.GuildID),
							ToggleEnabled: guildSettingsTimeRoles.ToggleEnabled,
							Timeroles:     welcomer.BytesToJSONB(welcomer.MarshalTimeRolesJSON(timeRoles)),
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
						Msg("Failed to update timerole guild settings")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed(fmt.Sprintf("Removed timerole <@&%d>. Run `/timeroles list` to see the list of timeroles configured.", role.ID), welcomer.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	r.InteractionCommands.MustAddInteractionCommand(ruleGroup)

	return nil
}
