package plugins

import (
	"errors"
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/jackc/pgx/v4"
)

type RulesCog struct {
	EventHandler *sandwich.Handlers
}

// Assert types.

var (
	_ sandwich.Cog           = (*RulesCog)(nil)
	_ sandwich.CogWithEvents = (*RulesCog)(nil)
)

func NewRulesCog() *RulesCog {
	return &RulesCog{
		EventHandler: sandwich.SetupHandler(nil),
	}
}

func (p *RulesCog) CogInfo() *sandwich.CogInfo {
	return &sandwich.CogInfo{
		Name:        "Rules",
		Description: "Provides the functionality for the 'Rules' feature",
	}
}

func (p *RulesCog) GetEventHandlers() *sandwich.Handlers {
	return p.EventHandler
}

func (p *RulesCog) RegisterCog(bot *sandwich.Bot) error {

	// Trigger OnInvokeRules when ON_GUILD_MEMBER_ADD event is received.
	p.EventHandler.RegisterOnGuildMemberAddEvent(func(eventCtx *sandwich.EventContext, member discord.GuildMember) error {
		return p.OnInvokeRules(eventCtx, member)
	})

	return nil
}

func (p *RulesCog) OnInvokeRules(eventCtx *sandwich.EventContext, member discord.GuildMember) (err error) {
	queries := welcomer.GetQueriesFromContext(eventCtx.Context)

	// Fetch guild settings.

	guildSettingsRules, err := queries.GetRulesGuildSettings(eventCtx.Context, int64(eventCtx.Guild.ID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			guildSettingsRules = &database.GuildSettingsRules{
				GuildID:          int64(eventCtx.Guild.ID),
				ToggleEnabled:    database.DefaultRules.ToggleEnabled,
				ToggleDmsEnabled: database.DefaultRules.ToggleDmsEnabled,
				Rules:            database.DefaultRules.Rules,
			}
		} else {
			eventCtx.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Msg("Failed to get rule settings")

			return err
		}
	}

	// Quit if not enabled or no rules are set.
	if !guildSettingsRules.ToggleEnabled || !guildSettingsRules.ToggleDmsEnabled || len(guildSettingsRules.Rules) == 0 {
		return nil
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

	_, err = member.User.Send(eventCtx.Session, discord.MessageParams{
		Embeds: embeds,
	})
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(member.User.ID)).
			Msg("Failed to send rules to user")

		return err
	}

	return nil
}
