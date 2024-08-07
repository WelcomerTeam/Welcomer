package plugins

import (
	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type InviteCog struct {
	EventHandler *sandwich.Handlers
}

// Assert types.

var (
	_ sandwich.Cog           = (*InviteCog)(nil)
	_ sandwich.CogWithEvents = (*InviteCog)(nil)
)

func NewInviteCog() *InviteCog {
	return &InviteCog{
		EventHandler: sandwich.SetupHandler(nil),
	}
}

func (c *InviteCog) CogInfo() *sandwich.CogInfo {
	return &sandwich.CogInfo{
		Name:        "Invites",
		Description: "Handles storing the creation and deletion of invites",
	}
}

func (c *InviteCog) GetEventHandlers() *sandwich.Handlers {
	return c.EventHandler
}

func (c *InviteCog) RegisterCog(bot *sandwich.Bot) error {

	// Register event for when an invite is created.
	c.EventHandler.RegisterOnInviteCreateEvent(func(eventCtx *sandwich.EventContext, invite discord.Invite) error {
		queries := welcomer.GetQueriesFromContext(eventCtx.Context)

		var guildID int64

		if invite.GuildID != nil {
			guildID = int64(*invite.GuildID)
		} else if invite.Guild != nil {
			guildID = int64(invite.Guild.ID)
		} else {
			eventCtx.Logger.Warn().Msg("Invite does not have a guild ID")

			return nil
		}

		_, err := queries.CreateOrUpdateGuildInvites(eventCtx.Context, database.CreateOrUpdateGuildInvitesParams{
			InviteCode: invite.Code,
			GuildID:    guildID,
			CreatedBy:  int64(invite.Inviter.ID),
			CreatedAt:  invite.CreatedAt,
			Uses:       int64(invite.Uses),
		})
		if err != nil {
			eventCtx.Logger.Error().Err(err).
				Int64("guildID", guildID).
				Msg("Failed to create or update invite")
		}

		return nil
	})

	// Register event for when an invite is deleted.
	c.EventHandler.RegisterOnInviteDeleteEvent(func(eventCtx *sandwich.EventContext, invite discord.Invite) error {
		queries := welcomer.GetQueriesFromContext(eventCtx.Context)

		var guildID int64

		if invite.GuildID != nil {
			guildID = int64(*invite.GuildID)
		} else if invite.Guild != nil {
			guildID = int64(invite.Guild.ID)
		} else {
			eventCtx.Logger.Warn().Msg("Invite does not have a guild ID")

			return nil
		}

		_, err := queries.DeleteGuildInvites(eventCtx.Context, database.DeleteGuildInvitesParams{
			InviteCode: invite.Code,
			GuildID:    guildID,
		})
		if err != nil {
			eventCtx.Logger.Error().Err(err).
				Int64("guildID", guildID).
				Msg("Failed to create or update invite")
		}

		return nil
	})

	return nil
}
