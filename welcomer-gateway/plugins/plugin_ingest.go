package plugins

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4"
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

		// Ignore non-default message types (e.g., system messages).
		if message.Type != discord.MessageTypeDefault {
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

		if before.ChannelID == after.ChannelID {
			return nil
		}

		if !before.ChannelID.IsNil() && after.ChannelID.IsNil() ||
			!before.ChannelID.IsNil() && !after.ChannelID.IsNil() {
			// Left or moved voice channel.
			err := closeSession(eventCtx.Context, guildID, before.ChannelID, member.User.ID)
			if err != nil {
				welcomer.Logger.Error().Err(err).Msg("Failed to close voice channel session")

				return err
			}
		}

		if before.ChannelID.IsNil() && !after.ChannelID.IsNil() ||
			!before.ChannelID.IsNil() && !after.ChannelID.IsNil() {
			// Joined or moved voice channel.
			err := createSession(eventCtx.Context, guildID, after.ChannelID, member.User.ID)
			if err != nil {
				welcomer.Logger.Error().Err(err).Msg("Failed to create voice channel session")

				return err
			}
		}

		return nil
	})

	return nil
}

func createSession(ctx context.Context, guildID, channelID, userID discord.Snowflake) error {
	err := welcomer.Queries.CreateGuildVoiceChannelOpenSession(ctx, database.CreateGuildVoiceChannelOpenSessionParams{
		GuildID:    int64(guildID),
		UserID:     int64(userID),
		ChannelID:  int64(channelID),
		StartTs:    time.Now(),
		LastSeenTs: time.Now(),
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to create voice channel open session")

		return fmt.Errorf("failed to create voice channel open session: %w", err)
	}

	return nil
}

func closeSession(ctx context.Context, guildID, channelID, userID discord.Snowflake) error {
	session, err := welcomer.Queries.DeleteAndGetGuildVoiceChannelOpenSession(ctx, database.DeleteAndGetGuildVoiceChannelOpenSessionParams{
		GuildID: int64(guildID),
		UserID:  int64(userID),
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		welcomer.Logger.Error().Err(err).Msg("Failed to get voice channel open session")

		return fmt.Errorf("failed to get voice channel open session: %w", err)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		welcomer.Logger.Warn().
			Int64("guild_id", int64(guildID)).
			Int64("user_id", int64(userID)).
			Msg("No open voice channel session found on voice channel leave")

		return nil
	}

	totalTime := time.Since(session.StartTs)

	err = welcomer.Queries.CreateVoiceChannelStat(ctx, database.CreateVoiceChannelStatParams{
		GuildID:     int64(guildID),
		ChannelID:   int64(channelID),
		UserID:      int64(userID),
		StartTs:     session.StartTs,
		EndTs:       time.Now(),
		TotalTimeMs: totalTime.Milliseconds(),
		Inferred:    false,
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Int64("channel_id", int64(channelID)).
			Int64("user_id", int64(userID)).
			Dur("duration", totalTime).
			Msg("Failed to create voice channel stat")

		return fmt.Errorf("failed to create voice channel stat: %w", err)
	}

	return nil
}

// TODO: handle lingering open sessions
