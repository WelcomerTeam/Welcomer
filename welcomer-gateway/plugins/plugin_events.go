package plugins

import (
	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type EventsCog struct {
	EventHandler *sandwich.Handlers
}

// Assert types.

var (
	_ sandwich.Cog           = (*EventsCog)(nil)
	_ sandwich.CogWithEvents = (*EventsCog)(nil)
)

func NewEventsCog() *EventsCog {
	return &EventsCog{
		EventHandler: sandwich.SetupHandler(nil),
	}
}

func (c *EventsCog) CogInfo() *sandwich.CogInfo {
	return &sandwich.CogInfo{
		Name:        "Invites",
		Description: "Handles misc discord events",
	}
}

func (c *EventsCog) GetEventHandlers() *sandwich.Handlers {
	return c.EventHandler
}

func (c *EventsCog) RegisterCog(bot *sandwich.Bot) error {
	// Register event for when a guild is joined.
	c.EventHandler.RegisterOnGuildJoinEvent(func(eventCtx *sandwich.EventContext, guild discord.Guild) error {
		queries := welcomer.GetQueriesFromContext(eventCtx.Context)

		return welcomer.EnsureGuild(eventCtx.Context, queries, guild.ID)
	})

	// Register event for when an invite is created.
	c.EventHandler.RegisterOnInviteCreateEvent(func(eventCtx *sandwich.EventContext, invite discord.Invite) error {
		queries := welcomer.GetQueriesFromContext(eventCtx.Context)

		var guildID int64

		if invite.GuildID != nil {
			guildID = int64(*invite.GuildID)
		} else if invite.Guild != nil {
			guildID = int64(invite.Guild.ID)
		} else {
			welcomer.Logger.Warn().Msg("Invite does not have a guild ID")

			return nil
		}

		var inviter discord.Snowflake
		if invite.Inviter != nil {
			inviter = invite.Inviter.ID
		}

		err := welcomer.RetryWithFallback(
			func() error {
				_, err := queries.CreateOrUpdateGuildInvites(eventCtx.Context, database.CreateOrUpdateGuildInvitesParams{
					InviteCode: invite.Code,
					GuildID:    guildID,
					CreatedBy:  int64(inviter),
					CreatedAt:  invite.CreatedAt,
					Uses:       int64(invite.Uses),
				})
				return err
			},
			func() error {
				return welcomer.EnsureGuild(eventCtx.Context, queries, discord.Snowflake(guildID))
			},
			nil,
		)
		if err != nil {
			welcomer.Logger.Error().Err(err).
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
			welcomer.Logger.Warn().Msg("Invite does not have a guild ID")

			return nil
		}

		_, err := queries.DeleteGuildInvites(eventCtx.Context, database.DeleteGuildInvitesParams{
			InviteCode: invite.Code,
			GuildID:    guildID,
		})
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guildID", guildID).
				Msg("Failed to create or update invite")
		}

		return nil
	})

	return nil
}
