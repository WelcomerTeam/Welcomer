package plugins

import (
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
)

type IngestCog struct {
	EventHandler *sandwich.Handlers
}

// Assert types.

var (
	_ sandwich.Cog           = (*IngestCog)(nil)
	_ sandwich.CogWithEvents = (*IngestCog)(nil)
)

func NewIngestCog() *IngestCog {
	return &IngestCog{
		EventHandler: sandwich.SetupHandler(nil),
	}
}

func (c *IngestCog) CogInfo() *sandwich.CogInfo {
	return &sandwich.CogInfo{
		Name:        "Ingest",
		Description: "Handles ingesting of various Discord events",
	}
}

func (c *IngestCog) GetEventHandlers() *sandwich.Handlers {
	return c.EventHandler
}

func (c *IngestCog) RegisterCog(bot *sandwich.Bot) error {
	// Register event for message create.
	c.EventHandler.RegisterOnMessageCreateEvent(func(eventCtx *sandwich.EventContext, message discord.Message) error {
		if message.GuildID == nil {
			return nil
		}

		welcomer.PusherIngestMessageEvents.Push(eventCtx.Context,
			message.ID, *message.GuildID, message.ChannelID, message.Author.ID,
			welcomer.IngestMessageEventTypeCreate, message.Timestamp)

		return nil
	})

	// Register event for message update.
	c.EventHandler.RegisterOnMessageUpdateEvent(func(eventCtx *sandwich.EventContext, _, newMessage discord.Message) error {
		if newMessage.GuildID == nil {
			return nil
		}

		welcomer.PusherIngestMessageEvents.Push(eventCtx.Context,
			newMessage.ID, *newMessage.GuildID, newMessage.ChannelID, newMessage.Author.ID,
			welcomer.IngestMessageEventTypeEdit, time.Now())

		return nil
	})

	// Register event for voice state update.
	c.EventHandler.RegisterOnVoiceStateUpdateEvent(func(eventCtx *sandwich.EventContext, member discord.GuildMember, before, after discord.VoiceState) error {
		var guildID discord.Snowflake

		if after.GuildID != nil {
			guildID = *after.GuildID
		} else if before.GuildID != nil {
			guildID = *before.GuildID
		} else {
			return nil
		}

		if before.ChannelID != after.ChannelID {
			if before.ChannelID.IsNil() && !after.ChannelID.IsNil() {
				// Joined a voice channel.
				welcomer.PusherIngestVoiceChannelEvents.Push(eventCtx.Context,
					guildID, after.ChannelID, member.User.ID,
					welcomer.IngestVoiceChannelEventTypeJoin, time.Now())
			} else if !before.ChannelID.IsNil() && after.ChannelID.IsNil() {
				// Left a voice channel.
				welcomer.PusherIngestVoiceChannelEvents.Push(eventCtx.Context,
					guildID, before.ChannelID, member.User.ID,
					welcomer.IngestVoiceChannelEventTypeLeave, time.Now())
			} else if !before.ChannelID.IsNil() && !after.ChannelID.IsNil() {
				// Moved voice channels.
				welcomer.PusherIngestVoiceChannelEvents.Push(eventCtx.Context,
					guildID, before.ChannelID, member.User.ID,
					welcomer.IngestVoiceChannelEventTypeLeave, time.Now())

				welcomer.PusherIngestVoiceChannelEvents.Push(eventCtx.Context,
					guildID, after.ChannelID, member.User.ID,
					welcomer.IngestVoiceChannelEventTypeJoin, time.Now())
			}
		}

		return nil
	})

	return nil
}
