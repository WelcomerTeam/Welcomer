package plugins

import (
	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
)

type EntitlementsCog struct {
	EventHandlers *sandwich.Handlers
}

// Assert types.
var (
	_ sandwich.Cog           = (*EntitlementsCog)(nil)
	_ sandwich.CogWithEvents = (*EntitlementsCog)(nil)
)

func NewEntitlementsCog() *EntitlementsCog {
	return &EntitlementsCog{
		EventHandlers: sandwich.SetupHandler(nil),
	}
}

func (c *EntitlementsCog) CogInfo() *sandwich.CogInfo {
	return &sandwich.CogInfo{
		Name:        "Entitlements",
		Description: "Handles discord entitlements",
	}
}

func (c *EntitlementsCog) GetEventHandlers() *sandwich.Handlers {
	return c.EventHandlers
}

func (c *EntitlementsCog) RegisterCog(bot *sandwich.Bot) error {
	// Register event when entitlement is created.
	c.EventHandlers.RegisterOnEntitlementCreate(func(eventCtx *sandwich.EventContext, entitlement discord.Entitlement) error {
		queries := welcomer.GetQueriesFromContext(eventCtx.Context)

		return welcomer.OnDiscordEntitlementCreated(eventCtx.Context, eventCtx.Logger, queries, entitlement)
	})

	// Register event when entitlement is updated.
	c.EventHandlers.RegisterOnEntitlementUpdate(func(eventCtx *sandwich.EventContext, entitlement discord.Entitlement) error {
		queries := welcomer.GetQueriesFromContext(eventCtx.Context)

		return welcomer.OnDiscordEntitlementUpdated(eventCtx.Context, eventCtx.Logger, queries, entitlement)
	})

	// Register event when entitlement is deleted.
	c.EventHandlers.RegisterOnEntitlementDelete(func(eventCtx *sandwich.EventContext, entitlement discord.Entitlement) error {
		queries := welcomer.GetQueriesFromContext(eventCtx.Context)

		return welcomer.OnDiscordEntitlementDeleted(eventCtx.Context, eventCtx.Logger, queries, entitlement)
	})

	return nil
}
