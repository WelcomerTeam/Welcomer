package plugins

import (
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Sandwich-Daemon/structs"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
)

func NewGeneralCog() *GeneralCog {
	return &GeneralCog{
		EventHandler: sandwich.SetupHandler(nil),
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
	// Register custom events.
	bot.RegisterEventHandler(core.CustomEventInvokeWelcomer, func(eventCtx *sandwich.EventContext, payload structs.SandwichPayload) error {
		bot.Logger.Trace().Msg("Called " + core.CustomEventInvokeWelcomer + " handler")

		var guildMemberAddPayload discord.GuildMemberAdd
		if err := eventCtx.DecodeContent(payload, &guildMemberAddPayload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		eventCtx.Guild = sandwich.NewGuild(*guildMemberAddPayload.GuildID)

		eventCtx.EventHandler.EventsMu.RLock()
		defer eventCtx.EventHandler.EventsMu.RUnlock()

		for _, event := range eventCtx.EventHandler.Events {
			if f, ok := event.(core.OnInvokeWelcomerFuncType); ok {
				return eventCtx.Handlers.WrapFuncType(eventCtx, f(eventCtx, *guildMemberAddPayload))
			}
		}

		return nil
	})

	// Links ON_GUILD_MEMBER_ADD to CustomEventInvokeWelcomer event.
	bot.RegisterOnGuildMemberAddEvent(func(eventCtx *sandwich.EventContext, member discord.GuildMember) error {
		bot.Logger.Trace().Msg("Called " + eventCtx.Payload.Type + " in " + discord.DiscordEventGuildMemberAdd + " handler")

		err := bot.DispatchType(eventCtx, core.CustomEventInvokeWelcomer, *eventCtx.Payload)
		if err != nil {
			return fmt.Errorf("failed to dispatch %s: %w", core.CustomEventInvokeWelcomer, err)
		}

		return nil
	})

	// Register events.
	RegisterOnInvokeWelcomerEvent(bot.Handlers, func(eventCtx *sandwich.EventContext, member discord.GuildMember) error {
		bot.Logger.Trace().Msg("Called " + eventCtx.Payload.Type + " in " + core.CustomEventInvokeWelcomer + " event")

		return nil
	})

	return nil
}

// RegisterOnGuildMemberAddEvent adds a new event handler for the GUILD_MEMBER_ADD event.
// It does not override a handler and instead will add another handler.
func RegisterOnInvokeWelcomerEvent(h *sandwich.Handlers, event core.OnInvokeWelcomerFuncType) {
	eventName := core.CustomEventInvokeWelcomer

	h.RegisterEvent(eventName, nil, event)
}

func (p *GeneralCog) GetEventHandlers() *sandwich.Handlers {
	return p.EventHandler
}
