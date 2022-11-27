package plugins

import (
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
)

func NewGeneralCog() *GeneralCog {
	return &GeneralCog{
		EventHandler: &sandwich.Handlers{},
	}
}

type GeneralCog struct {
	EventHandler *sandwich.Handlers
}

// Assert types.

var (
	_ sandwich.Cog           = (*GeneralCog)(nil)
	_ sandwich.CogWithEvents = (*GeneralCog)(nil)
)

func (p *GeneralCog) CogInfo() *sandwich.CogInfo {
	return &sandwich.CogInfo{
		Name:        "GeneralCog",
		Description: "General commands",
	}
}

func (p *GeneralCog) RegisterCog(bot *sandwich.Bot) error {
	return nil
}

func (p *GeneralCog) GetEventHandlers() *sandwich.Handlers {
	return p.EventHandler
}
