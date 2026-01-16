package welcomer

import (
	"fmt"
	"log/slog"

	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	plugins "github.com/WelcomerTeam/Welcomer/welcomer-gateway/plugins"
)

type Welcomer struct {
	Bot *sandwich.Bot
}

func NewWelcomer(identifierName string, sandwichClient *sandwich.Sandwich) (welcomer *Welcomer) {
	welcomer = &Welcomer{}

	// Register bot (cogs, events)
	err := welcomer.Register()
	if err != nil {
		panic(fmt.Sprintf("welcomer.Register(): %v", err.Error()))
	}

	return welcomer
}

func (w *Welcomer) Register() error {
	bot := sandwich.NewBot(slog.Default())

	// Register cogs
	bot.MustRegisterCog(plugins.NewWelcomerCog())
	bot.MustRegisterCog(plugins.NewRulesCog())
	bot.MustRegisterCog(plugins.NewAutoRolesCog())
	bot.MustRegisterCog(plugins.NewLeaverCog())
	bot.MustRegisterCog(plugins.NewTimeRolesCog())
	bot.MustRegisterCog(plugins.NewTempChannelsCog())
	bot.MustRegisterCog(plugins.NewBorderwallCog())
	bot.MustRegisterCog(plugins.NewEventsCog())
	bot.MustRegisterCog(plugins.NewOnboardingCog())
	bot.MustRegisterCog(plugins.NewEntitlementsCog())
	bot.MustRegisterCog(plugins.NewIngestCog())

	w.Bot = bot

	return nil
}
