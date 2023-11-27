package plugins

import subway "github.com/WelcomerTeam/Subway/subway"

type WelcomerCog struct {
	InteractionCommands *subway.InteractionCommandable
}

// Assert types.

var (
	_ subway.Cog                        = (*GeneralCog)(nil)
	_ subway.CogWithInteractionCommands = (*GeneralCog)(nil)
)

func (p *WelcomerCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{}
}

func (p *WelcomerCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return p.InteractionCommands
}

func (p *WelcomerCog) RegisterCog(sub *subway.Subway) error {
	return nil
}
