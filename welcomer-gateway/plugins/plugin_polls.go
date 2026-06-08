package plugins

import (
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich_daemon "github.com/WelcomerTeam/Sandwich-Daemon"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type PollCog struct {
	EventHandler *sandwich.Handlers
}

// Assert types.

var (
	_ sandwich.Cog           = (*PollCog)(nil)
	_ sandwich.CogWithEvents = (*PollCog)(nil)
)

func NewPollCog() *PollCog {
	return &PollCog{
		EventHandler: sandwich.SetupHandler(nil),
	}
}

func (g *PollCog) CogInfo() *sandwich.CogInfo {
	return &sandwich.CogInfo{
		Name:        "Polls",
		Description: "Provides the functionality for the 'Polls' feature",
	}
}

func (g *PollCog) GetEventHandlers() *sandwich.Handlers {
	return g.EventHandler
}

func (g *PollCog) RegisterCog(bot *sandwich.Bot) error {
	// Register poll end handler.

	g.EventHandler.RegisterEventHandler(core.CustomEventInvokeEndPoll, func(eventCtx *sandwich.EventContext, payload sandwich_daemon.ProducedPayload) error {
		var invokePollEndPayload core.CustomEventInvokeEndPollStructure
		if err := eventCtx.DecodeContent(payload, &invokePollEndPayload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		eventCtx.Guild = sandwich.NewGuild(invokePollEndPayload.GuildID)

		eventCtx.EventHandler.EventsMu.RLock()
		defer eventCtx.EventHandler.EventsMu.RUnlock()

		for _, event := range eventCtx.EventHandler.Events {
			if f, ok := event.(welcomer.OnInvokeEndPollFuncType); ok {
				return eventCtx.Handlers.WrapFuncType(eventCtx, f(eventCtx, invokePollEndPayload))
			}
		}

		return nil
	})

	// Call OnInvokeEndPoll when CustomEventInvokeEndPoll is triggered.
	g.EventHandler.RegisterEvent(core.CustomEventInvokeEndPoll, nil, (welcomer.OnInvokeEndPollFuncType)(g.OnInvokeEndPoll))

	return nil
}

func (g *PollCog) OnInvokeEndPoll(eventCtx *sandwich.EventContext, event core.CustomEventInvokeEndPollStructure) error {
	welcomer.Logger.Info().
		Str("poll_uuid", event.PollUUID.String()).
		Msg("Received poll end event, processing poll end")

	poll, err := welcomer.Queries.GetPoll(eventCtx.Context, database.GetPollParams{
		PollUuid: event.PollUUID,
		GuildID:  int64(event.GuildID),
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("poll_uuid", event.PollUUID.String()).
			Msg("Failed to get poll for poll end event")

		return err
	}

	if poll.HasEnded {
		welcomer.Logger.Error().
			Str("poll_uuid", event.PollUUID.String()).
			Msg("Poll has already ended, skipping poll end event")

		return nil
	}

	if err := g.EndPoll(eventCtx, poll); err != nil {
		welcomer.Logger.Error().Err(err).
			Str("poll_uuid", event.PollUUID.String()).
			Msg("Failed to end poll for poll end event")

		return err
	}

	return nil
}

func (g *PollCog) EndPoll(eventCtx *sandwich.EventContext, poll *database.GuildPolls) error {
	if poll.HasEnded {
		welcomer.Logger.Error().
			Str("poll_uuid", poll.PollUuid.String()).
			Msg("Poll has already ended, skipping poll end")

		return nil
	}

	// Set poll as ended

	var err error

	poll, err = welcomer.Queries.SetPollEnded(eventCtx.Context, database.SetPollEndedParams{
		PollUuid: poll.PollUuid,
		HasEnded: true,
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("poll_uuid", poll.PollUuid.String()).
			Msg("Failed to set poll as ended")

		return err
	}

	defer func() {
		if err != nil {
			// Unset poll as ended on error

			welcomer.Logger.Error().Err(err).
				Str("poll_uuid", poll.PollUuid.String()).
				Msg("An error occurred during poll end processing, unsetting poll as ended")

			poll, err = welcomer.Queries.SetPollEnded(eventCtx.Context, database.SetPollEndedParams{
				PollUuid: poll.PollUuid,
				HasEnded: false,
			})
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Str("poll_uuid", poll.PollUuid.String()).
					Msg("Failed to unset poll as ended")
			}

		}
	}()

	msg, err := discord.GetChannelMessage(eventCtx.Context, eventCtx.Session, discord.Snowflake(poll.ChannelID), discord.Snowflake(poll.MessageID))
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("poll_uuid", poll.PollUuid.String()).
			Msg("Failed to fetch poll message for poll end")

		return err
	}

	for i := range msg.Components {
		for j := range msg.Components[i].Components {
			if msg.Components[i].Components[j].Type == discord.InteractionComponentTypeButton {
				msg.Components[i].Components[j].Label = "This poll has ended"
				msg.Components[i].Components[j].Disabled = true
			}
		}
	}

	_, err = msg.Edit(eventCtx.Context, eventCtx.Session, discord.MessageParams{
		Components: msg.Components,
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("poll_uuid", poll.PollUuid.String()).
			Msg("Failed to edit poll message to disable buttons for poll end")

		return err
	}

	welcomer.PusherGuildScience.Push(
		eventCtx.Context,
		discord.Snowflake(poll.GuildID),
		0,
		database.ScienceGuildEventTypePollEnded,
		&welcomer.GuildSciencePollEvents{
			PollUUID: poll.PollUuid,
		},
	)

	return nil
}
