package plugins

import (
	"context"
	"errors"
	"fmt"

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
		"Provide rules for the server.",
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

				Choices: []*discord.ApplicationCommandOptionChoice{
					{Name: RuleModuleRules, Value: welcomer.StringToJsonLiteral(RuleModuleRules)},
					{Name: RuleModuleDMs, Value: welcomer.StringToJsonLiteral(RuleModuleDMs)},
				},
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.IntToInt64Pointer(discord.PermissionElevated),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				module := subway.MustGetArgument(ctx, "module").MustString()

				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsRules, err := queries.GetRulesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get rules guild settings.")
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

				_, err = queries.UpdateRuleGuildSettings(ctx, &database.UpdateRuleGuildSettingsParams{
					GuildID:          int64(*interaction.GuildID),
					ToggleEnabled:    guildSettingsRules.ToggleEnabled,
					ToggleDmsEnabled: guildSettingsRules.ToggleDmsEnabled,
					Rules:            guildSettingsRules.Rules,
				})
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update rules guild settings.")

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

				Choices: []*discord.ApplicationCommandOptionChoice{
					{Name: RuleModuleRules, Value: welcomer.StringToJsonLiteral(RuleModuleRules)},
					{Name: RuleModuleDMs, Value: welcomer.StringToJsonLiteral(RuleModuleDMs)},
				},
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.IntToInt64Pointer(discord.PermissionElevated),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				module := subway.MustGetArgument(ctx, "module").MustString()

				queries := welcomer.GetQueriesFromContext(ctx)

				guildSettingsRules, err := queries.GetRulesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get rules guild settings.")
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

				_, err = queries.UpdateRuleGuildSettings(ctx, &database.UpdateRuleGuildSettingsParams{
					GuildID:          int64(*interaction.GuildID),
					ToggleEnabled:    guildSettingsRules.ToggleEnabled,
					ToggleDmsEnabled: guildSettingsRules.ToggleDmsEnabled,
					Rules:            guildSettingsRules.Rules,
				})
				if err != nil {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to update rules guild settings.")

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
				queries := welcomer.GetQueriesFromContext(ctx)
				guildSettingsRules, err := queries.GetRulesGuildSettings(ctx, int64(*interaction.GuildID))
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					sub.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to get rules guild settings.")

					return nil, err
				}

				if len(guildSettingsRules.Rules) == 0 || !guildSettingsRules.ToggleEnabled {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("There are no rules set for this server.", welcomer.EmbedColourInfo),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				embeds := []*discord.Embed{}
				embed := &discord.Embed{Title: "Rules", Color: welcomer.EmbedColourInfo}

				for ruleNumber, rule := range guildSettingsRules.Rules {
					ruleWithNumber := fmt.Sprintf("%d. %s\n", ruleNumber, rule)

					// If the embed content will go over 4000 characters then create a new embed and continue from that one.
					if len(embed.Description)+len(ruleWithNumber) > 4000 {
						embeds = append(embeds, embed)
						embed = &discord.Embed{Color: welcomer.EmbedColourInfo}
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
