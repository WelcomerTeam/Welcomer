package plugins

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/WelcomerTeam/Discord/discord"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
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
		"Provide a set of rules for your server, easily accessible and automatically sent on join.",
	)

	// Disable the rules module for DM channels.
	ruleGroup.DMPermission = &welcomer.False

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
					{Name: RuleModuleRules, Value: welcomer.StringToJsonLiteral(RuleModuleRules)},
					{Name: RuleModuleDMs, Value: welcomer.StringToJsonLiteral(RuleModuleDMs)},
				},
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				module := subway.MustGetArgument(ctx, "module").MustString()

				guildSettingsRules, err := welcomer.Queries.GetRulesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsRules = &database.GuildSettingsRules{
							GuildID:          int64(*interaction.GuildID),
							ToggleEnabled:    welcomer.DefaultRules.ToggleEnabled,
							ToggleDmsEnabled: welcomer.DefaultRules.ToggleDmsEnabled,
							Rules:            welcomer.DefaultRules.Rules,
						}
					} else {
						welcomer.Logger.Error().Err(err).
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
							Embeds: welcomer.NewEmbed("Unknown module: "+module, welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.Queries.CreateOrUpdateRulesGuildSettings(ctx, database.CreateOrUpdateRulesGuildSettingsParams{
							GuildID:          int64(*interaction.GuildID),
							ToggleEnabled:    guildSettingsRules.ToggleEnabled,
							ToggleDmsEnabled: guildSettingsRules.ToggleDmsEnabled,
							Rules:            guildSettingsRules.Rules,
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
						Msg("Failed to update rules guild settings")

					return nil, err
				}

				switch module {
				case RuleModuleRules:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Enabled rules. Run `/rules list` to see the list of rules configured.", welcomer.EmbedColourSuccess),
						},
					}, nil
				case RuleModuleDMs:
					if guildSettingsRules.ToggleEnabled {
						return &discord.InteractionResponse{
							Type: discord.InteractionCallbackTypeChannelMessageSource,
							Data: &discord.InteractionCallbackData{
								Embeds: welcomer.NewEmbed("Enabled rule direct messages. Users will now receive a list of rules when joining the server.", welcomer.EmbedColourSuccess),
							},
						}, nil
					} else {
						return &discord.InteractionResponse{
							Type: discord.InteractionCallbackTypeChannelMessageSource,
							Data: &discord.InteractionCallbackData{
								Embeds: welcomer.NewEmbed("Enabled rule direct messages. Rules are not enabled, users will not receive a list of rules when joining the server.", welcomer.EmbedColourWarn),
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
					{Name: RuleModuleRules, Value: welcomer.StringToJsonLiteral(RuleModuleRules)},
					{Name: RuleModuleDMs, Value: welcomer.StringToJsonLiteral(RuleModuleDMs)},
				},
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				module := subway.MustGetArgument(ctx, "module").MustString()

				guildSettingsRules, err := welcomer.Queries.GetRulesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsRules = &database.GuildSettingsRules{
							GuildID:          int64(*interaction.GuildID),
							ToggleEnabled:    welcomer.DefaultRules.ToggleEnabled,
							ToggleDmsEnabled: welcomer.DefaultRules.ToggleDmsEnabled,
							Rules:            welcomer.DefaultRules.Rules,
						}
					} else {
						welcomer.Logger.Error().Err(err).
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
							Embeds: welcomer.NewEmbed("Unknown module: "+module, welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.Queries.CreateOrUpdateRulesGuildSettings(ctx, database.CreateOrUpdateRulesGuildSettingsParams{
							GuildID:          int64(*interaction.GuildID),
							ToggleEnabled:    guildSettingsRules.ToggleEnabled,
							ToggleDmsEnabled: guildSettingsRules.ToggleDmsEnabled,
							Rules:            guildSettingsRules.Rules,
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
						Msg("Failed to update rules guild settings")

					return nil, err
				}

				switch module {
				case RuleModuleRules:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Disabled rules.", welcomer.EmbedColourSuccess),
						},
					}, nil
				case RuleModuleDMs:
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Disabled rule direct messages.", welcomer.EmbedColourSuccess),
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

		DMPermission: &welcomer.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
				guildSettingsRules, err := welcomer.Queries.GetRulesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						guildSettingsRules = &database.GuildSettingsRules{
							GuildID:          int64(*interaction.GuildID),
							ToggleEnabled:    welcomer.DefaultRules.ToggleEnabled,
							ToggleDmsEnabled: welcomer.DefaultRules.ToggleDmsEnabled,
							Rules:            welcomer.DefaultRules.Rules,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get rules guild settings")

						return nil, err
					}
				}

				if !guildSettingsRules.ToggleEnabled {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Rules are not enabled for this server.", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				if len(guildSettingsRules.Rules) == 0 {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("There are no rules set for this server.", welcomer.EmbedColourInfo),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				embeds := []discord.Embed{}
				embed := discord.Embed{Title: "Rules", Color: welcomer.EmbedColourInfo}

				for ruleNumber, rule := range guildSettingsRules.Rules {
					ruleWithNumber := fmt.Sprintf("%d. %s\n", ruleNumber+1, rule)

					// If the embed content will go over 4000 characters then create a new embed and continue from that one.
					if len(embed.Description)+len(ruleWithNumber) > 4000 {
						embeds = append(embeds, embed)
						embed = discord.Embed{Color: welcomer.EmbedColourInfo}
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

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "addrule",
		Description: "Add a rule to the server.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeString,
				Name:         "rule",
				Description:  "The rule to add.",

				MaxLength: welcomer.ToPointer(int32(welcomer.MaxRuleLength)),
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				rule := subway.MustGetArgument(ctx, "rule").MustString()

				rules, err := welcomer.Queries.GetRulesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						rules = &database.GuildSettingsRules{
							GuildID:          int64(*interaction.GuildID),
							ToggleEnabled:    welcomer.DefaultRules.ToggleEnabled,
							ToggleDmsEnabled: welcomer.DefaultRules.ToggleDmsEnabled,
							Rules:            welcomer.DefaultRules.Rules,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get rules guild settings")

						return nil, err
					}
				}

				if len(rule) > welcomer.MaxRuleLength {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed(fmt.Sprintf("The rule is too long. Maximum length is %d characters.", welcomer.MaxRuleLength), welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				if len(rules.Rules) >= welcomer.MaxRuleLength {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed(fmt.Sprintf("You have reached the maximum number of rules for this server (%d).", welcomer.MaxRuleLength), welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				rules.Rules = append(rules.Rules, rule)

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.Queries.CreateOrUpdateRulesGuildSettings(ctx, database.CreateOrUpdateRulesGuildSettingsParams{
							GuildID:          int64(*interaction.GuildID),
							ToggleEnabled:    rules.ToggleEnabled,
							ToggleDmsEnabled: rules.ToggleDmsEnabled,
							Rules:            rules.Rules,
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
						Msg("Failed to update rules guild settings")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Your rule has been added. Run `/rules list` to see the list of rules configured.", welcomer.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	ruleGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "removerule",
		Description: "Remove a rule from the server.",

		AutocompleteHandler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) ([]discord.ApplicationCommandOptionChoice, error) {
			autocompleteRule := subway.MustGetArgument(ctx, "rule").MustString()

			rules, err := welcomer.Queries.GetRulesGuildSettings(ctx, int64(*interaction.GuildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					rules = &database.GuildSettingsRules{
						GuildID:          int64(*interaction.GuildID),
						ToggleEnabled:    welcomer.DefaultRules.ToggleEnabled,
						ToggleDmsEnabled: welcomer.DefaultRules.ToggleDmsEnabled,
						Rules:            welcomer.DefaultRules.Rules,
					}
				} else {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get rules guild settings")

					return nil, err
				}
			}

			choices := make([]discord.ApplicationCommandOptionChoice, 0, len(rules.Rules))

			isAutocompleteRuleNumber := false
			if _, err := strconv.Atoi(autocompleteRule); err == nil {
				isAutocompleteRuleNumber = true
			}

			for i, rule := range rules.Rules {
				if autocompleteRule != "" {
					if isAutocompleteRuleNumber {
						// If autocomplete is present and can be converted to a number, check if the rule number is in the list.
						if !strings.Contains(strconv.Itoa(i+1), autocompleteRule) {
							continue
						}
					} else {
						if !welcomer.CompareStrings(rule, autocompleteRule) {
							continue
						}
					}
				}

				choices = append(choices, discord.ApplicationCommandOptionChoice{
					Name:  welcomer.Overflow(fmt.Sprintf("%d. %s", i+1, rule), 100),
					Value: welcomer.StringToJsonLiteral(strconv.Itoa(i + 1)),
				})
			}

			if len(choices) > 25 {
				choices = choices[:25]
			}

			return choices, nil
		},

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeString,
				Name:         "rule",
				Description:  "The rule to remove.",
				Autocomplete: &welcomer.True,
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				ruleString := subway.MustGetArgument(ctx, "rule").MustString()

				rules, err := welcomer.Queries.GetRulesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						rules = &database.GuildSettingsRules{
							GuildID:          int64(*interaction.GuildID),
							ToggleEnabled:    welcomer.DefaultRules.ToggleEnabled,
							ToggleDmsEnabled: welcomer.DefaultRules.ToggleDmsEnabled,
							Rules:            welcomer.DefaultRules.Rules,
						}
					} else {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to get rules guild settings")

						return nil, err
					}
				}

				rule, err := strconv.Atoi(ruleString)
				if err != nil {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Invalid rule number.", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				if rule < 1 || rule > len(rules.Rules) {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed(fmt.Sprintf("Invalid rule number. Must be between 1 and %d.", len(rules.Rules)), welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				rules.Rules = append(rules.Rules[:rule-1], rules.Rules[rule:]...)

				err = welcomer.RetryWithFallback(
					func() error {
						_, err = welcomer.Queries.CreateOrUpdateRulesGuildSettings(ctx, database.CreateOrUpdateRulesGuildSettingsParams{
							GuildID:          int64(*interaction.GuildID),
							ToggleEnabled:    rules.ToggleEnabled,
							ToggleDmsEnabled: rules.ToggleDmsEnabled,
							Rules:            rules.Rules,
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
						Msg("Failed to update rules guild settings")

					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Your rule has been removed. Run `/rules list` to see the list of rules configured.", welcomer.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	r.InteractionCommands.MustAddInteractionCommand(ruleGroup)

	return nil
}
