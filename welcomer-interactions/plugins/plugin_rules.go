package plugins

import (
	"context"
	"errors"
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/jackc/pgx/v4"
)

func NewRulesCog() *RulesCog {
	return &RulesCog{
		InteractionCommands: subway.SetupInteractionCommandable(&subway.InteractionCommandable{}),
	}
}

type RulesCog struct {
	InteractionCommands *subway.InteractionCommandable
}

// Assert types.

var (
	_ subway.Cog                        = (*RulesCog)(nil)
	_ subway.CogWithInteractionCommands = (*RulesCog)(nil)
)

const (
	RuleModuleRules = "rules"
	RuleModuleDMs   = "dms"
)

func (r *RulesCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "Rules",
		Description: "Provides the cog for the 'Rules' feature.",
	}
}

func (r *RulesCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return r.InteractionCommands
}

func (r *RulesCog) RegisterCog(sub *subway.Subway) error {
	ruleGroup := subway.NewSubcommandGroup(
		"rules",
		"Provides rules for the server.",
	)

	// Disable the rules module for DM channels.
	ruleGroup.DMPermission = &utils.False

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "enable",
		Description: "Enable a rule module for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeString,
				Name:         "module",
				Description:  "The module to enable.",

				Choices: []discord.ApplicationCommandOptionChoice{
					{Name: RuleModuleRules, Value: utils.StringToJsonLiteral(RuleModuleRules)},
					{Name: RuleModuleDMs, Value: utils.StringToJsonLiteral(RuleModuleDMs)},
				},
			},
		},

		DMPermission:            &utils.False,
		DefaultMemberPermission: utils.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				module := subway.MustGetArgument(ctx, "module").MustString()

				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsRules, err := queries.GetRulesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsRules = &database.GuildSettingsRules{
							GuildID:          int64(*interaction.GuildID),
							ToggleEnabled:    database.DefaultRules.ToggleEnabled,
							ToggleDmsEnabled: database.DefaultRules.ToggleDmsEnabled,
							Rules:            database.DefaultRules.Rules,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get rules guild settings")

						return nil, err
					}
				}

				switch module {
				case RuleModuleRules:
					guildSettingsRules.ToggleEnabled = true
				case RuleModuleDMs:
					guildSettingsRules.ToggleDmsEnabled = true
				default:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Unknown module: "+module, utils.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				err = utils.RetryWithFallback(
					func() error {
						_, err = queries.CreateOrUpdateRulesGuildSettings(ctx, database.CreateOrUpdateRulesGuildSettingsParams{
							GuildID:          int64(*interaction.GuildID),
							ToggleEnabled:    guildSettingsRules.ToggleEnabled,
							ToggleDmsEnabled: guildSettingsRules.ToggleDmsEnabled,
							Rules:            guildSettingsRules.Rules,
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
						Msg("Failed to update rules guild settings")

					return nil, err
				}

				switch module {
				case RuleModuleRules:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Enabled rules. Run `/rules list` to see the list of rules configured.", utils.EmbedColourSuccess),
						},
					}, nil
				case RuleModuleDMs:
					if guildSettingsRules.ToggleEnabled {
						return &discord.InteractionResponse{
							Type: discord.InteractionCallbackTypeChannelMessageSource,
							Data: &discord.InteractionCallbackData{
								Embeds: utils.NewEmbed("Enabled rule direct messages. Users will now receive a list of rules when joining the server.", utils.EmbedColourSuccess),
							},
						}, nil
					} else {
						return &discord.InteractionResponse{
							Type: discord.InteractionCallbackTypeChannelMessageSource,
							Data: &discord.InteractionCallbackData{
								Embeds: utils.NewEmbed("Enabled rule direct messages. Rules are not enabled, users will not receive a list of rules when joining the server.", utils.EmbedColourWarn),
							},
						}, nil
					}
				}

				return nil, nil // Unreachable
			})
		},
	})

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "disable",
		Description: "Disables a rule module for this server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeString,
				Name:         "module",
				Description:  "The module to disable.",

				Choices: []discord.ApplicationCommandOptionChoice{
					{Name: RuleModuleRules, Value: utils.StringToJsonLiteral(RuleModuleRules)},
					{Name: RuleModuleDMs, Value: utils.StringToJsonLiteral(RuleModuleDMs)},
				},
			},
		},

		DMPermission:            &utils.False,
		DefaultMemberPermission: utils.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				module := subway.MustGetArgument(ctx, "module").MustString()

				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsRules, err := queries.GetRulesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsRules = &database.GuildSettingsRules{
							GuildID:          int64(*interaction.GuildID),
							ToggleEnabled:    database.DefaultRules.ToggleEnabled,
							ToggleDmsEnabled: database.DefaultRules.ToggleDmsEnabled,
							Rules:            database.DefaultRules.Rules,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get rules guild settings")

						return nil, err
					}
				}

				switch module {
				case RuleModuleRules:
					guildSettingsRules.ToggleEnabled = false
				case RuleModuleDMs:
					guildSettingsRules.ToggleDmsEnabled = false
				default:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Unknown module: "+module, utils.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				err = utils.RetryWithFallback(
					func() error {
						_, err = queries.CreateOrUpdateRulesGuildSettings(ctx, database.CreateOrUpdateRulesGuildSettingsParams{
							GuildID:          int64(*interaction.GuildID),
							ToggleEnabled:    guildSettingsRules.ToggleEnabled,
							ToggleDmsEnabled: guildSettingsRules.ToggleDmsEnabled,
							Rules:            guildSettingsRules.Rules,
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
						Msg("Failed to update rules guild settings")

					return nil, err
				}

				switch module {
				case RuleModuleRules:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Disabled rules.", utils.EmbedColourSuccess),
						},
					}, nil
				case RuleModuleDMs:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("Disabled rule direct messages.", utils.EmbedColourSuccess),
						},
					}, nil
				}

				return nil, nil // Unreachable
			})
		},
	})

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "list",
		Description: "List the rules for the server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission: &utils.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsRules, err := queries.GetRulesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsRules = &database.GuildSettingsRules{
							GuildID:          int64(*interaction.GuildID),
							ToggleEnabled:    database.DefaultRules.ToggleEnabled,
							ToggleDmsEnabled: database.DefaultRules.ToggleDmsEnabled,
							Rules:            database.DefaultRules.Rules,
						}
					} else {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get rules guild settings")

						return nil, err
					}
				}

				if len(guildSettingsRules.Rules) == 0 || !guildSettingsRules.ToggleEnabled {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("There are no rules set for this server.", utils.EmbedColourInfo),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				embeds := []discord.Embed{}
				embed := discord.Embed{Title: "Rules", Color: utils.EmbedColourInfo}

				for ruleNumber, rule := range guildSettingsRules.Rules {
					ruleWithNumber := fmt.Sprintf("%d. %s\n", ruleNumber, rule)

					// If the embed content will go over 4000 characters then create a new embed and continue from that one.
					if len(embed.Description)+len(ruleWithNumber) > 4000 {
						embeds = append(embeds, embed)
						embed = discord.Embed{Color: utils.EmbedColourInfo}
					}

					embed.Description += ruleWithNumber
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
