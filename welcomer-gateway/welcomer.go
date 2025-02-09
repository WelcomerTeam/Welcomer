package welcomer

import (
	"fmt"

	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	plugins "github.com/WelcomerTeam/Welcomer/welcomer-gateway/plugins"
	"github.com/rs/zerolog"
)

type Welcomer struct {
	Logger zerolog.Logger
	Bot    *sandwich.Bot
}

func NewWelcomer(identifierName string, sandwichClient *sandwich.Sandwich) (welcomer *Welcomer) {
	welcomer = &Welcomer{
		Logger: sandwichClient.Logger,
	}

	// Register bot (cogs, events)
	err := welcomer.Register()
	if err != nil {
		panic(fmt.Sprintf("welcomer.Register(): %v", err.Error()))
	}

	return welcomer
}

func (w *Welcomer) Register() error {
	bot := sandwich.NewBot(w.Logger)

	// Register cogs
	bot.MustRegisterCog(plugins.NewWelcomerCog())
	bot.MustRegisterCog(plugins.NewRulesCog())
	bot.MustRegisterCog(plugins.NewAutoRolesCog())
	bot.MustRegisterCog(plugins.NewLeaverCog())
	bot.MustRegisterCog(plugins.NewTimeRolesCog())
	bot.MustRegisterCog(plugins.NewTempChannelsCog())
	bot.MustRegisterCog(plugins.NewBorderwallCog())
	bot.MustRegisterCog(plugins.NewEventsCog())

	w.Bot = bot

	return nil
}
