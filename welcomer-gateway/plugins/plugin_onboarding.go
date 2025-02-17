package plugins

import (
	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type OnboardingCog struct {
	EventHandler *sandwich.Handlers
}

// Assert types.

var (
	_ sandwich.Cog           = (*OnboardingCog)(nil)
	_ sandwich.CogWithEvents = (*OnboardingCog)(nil)
)

func NewOnboardingCog() *OnboardingCog {
	return &OnboardingCog{
		EventHandler: sandwich.SetupHandler(nil),
	}
}

func (p *OnboardingCog) CogInfo() *sandwich.CogInfo {
	return &sandwich.CogInfo{
		Name:        "Onboarding",
		Description: "Provides the functionality for the 'Onboarding' feature",
	}
}

func (p *OnboardingCog) GetEventHandlers() *sandwich.Handlers {
	return p.EventHandler
}

func (p *OnboardingCog) RegisterCog(bot *sandwich.Bot) error {
	// Register
	p.EventHandler.RegisterOnGuildJoinEvent(func(eventCtx *sandwich.EventContext, guild discord.Guild) error {
		welcomer.GetPushGuildScienceFromContext(eventCtx.Context).Push(
			eventCtx.Context,
			eventCtx.Guild.ID,
			database.ScienceGuildEventTypeGuildJoin,
			nil,
		)

		return nil
	})

	p.EventHandler.RegisterOnGuildLeaveEvent(func(eventCtx *sandwich.EventContext, unavailableGuild discord.UnavailableGuild) error {
		welcomer.GetPushGuildScienceFromContext(eventCtx.Context).Push(
			eventCtx.Context,
			eventCtx.Guild.ID,
			database.ScienceGuildEventTypeGuildLeave,
			nil,
		)

		return nil
	})

	return nil
}
