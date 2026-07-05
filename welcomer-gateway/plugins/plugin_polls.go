package plugins

import (
	"fmt"
	"strings"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich_daemon "github.com/WelcomerTeam/Sandwich-Daemon"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	welcomer_interactions "github.com/WelcomerTeam/Welcomer/welcomer-interactions/plugins"
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
	welcomer_interactions.SetupSectionEmojiIDs()

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
		if strings.Contains(err.Error(), "404 Not Found") {
			welcomer.Logger.Warn().
				Str("poll_uuid", poll.PollUuid.String()).
				Msg("Poll message not found, skipping disabling buttons for poll end")

			return nil
		}

		welcomer.Logger.Error().Err(err).
			Str("poll_uuid", poll.PollUuid.String()).
			Msg("Failed to fetch poll message for poll end")

		return err
	}

	entriesCounts, err := welcomer.Queries.GetPollEntriesCounts(eventCtx.Context, poll.PollUuid)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("poll_uuid", poll.PollUuid.String()).
			Msg("Failed to get poll entries counts")

		return err
	}

	answers := welcomer.UnmarshalAnswersListJSON(poll.PollOptions.Bytes)
	results := make([]int, len(answers))

	for _, entryCount := range entriesCounts {
		results[entryCount.OptionIndex] = int(entryCount.EntryCount)
	}

	newMessage := welcomer_interactions.PollView(poll, results, false, true)

	for i := range newMessage.Components {
		for j := range newMessage.Components[i].Components {
			if newMessage.Components[i].Components[j].Type == discord.InteractionComponentTypeButton {
				newMessage.Components[i].Components[j].Label = "This poll has ended"
				newMessage.Components[i].Components[j].Disabled = true
			}
		}
	}

	_, err = msg.Edit(eventCtx.Context, eventCtx.Session, welcomer.WebhookMessageParamsToMessageParams(newMessage))
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("poll_uuid", poll.PollUuid.String()).
			Msg("Failed to edit poll message to disable buttons for poll end")

		return err
	}

	topResult := 0
	totalVotes := 0

	for _, count := range results {
		if count > topResult {
			topResult = count
		}

		totalVotes += count
	}

	wonAnswers := make([]string, 0)

	for i, answer := range answers {
		if results[i] == topResult {
			wonAnswers = append(wonAnswers, answer)
		}
	}

	resultPercentage := float64(float64(topResult) / float64(totalVotes) * 100)

	if topResult > 0 {
		if len(wonAnswers) == 1 {
			_, err = msg.Reply(eventCtx.Context, eventCtx.Session, *discord.NewMessage(fmt.Sprintf("The **%s** poll has ended! The winner is **%s** (**%s%%**)", poll.Title, wonAnswers[0], formatDecimal(resultPercentage))))
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Str("poll_uuid", poll.PollUuid.String()).
					Msg("Failed to send poll end message")
			}
		} else {
			wonAnswersStr := ""

			for i, answer := range wonAnswers {
				if i == len(wonAnswers)-1 {
					wonAnswersStr += fmt.Sprintf("and **%s**", answer)
				} else {
					wonAnswersStr += fmt.Sprintf("**%s**, ", answer)
				}
			}

			_, err = msg.Reply(eventCtx.Context, eventCtx.Session, *discord.NewMessage(fmt.Sprintf("The **%s** poll has ended! The winners are %s (**%s%%**)", poll.Title, wonAnswersStr, formatDecimal(resultPercentage))))
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Str("poll_uuid", poll.PollUuid.String()).
					Msg("Failed to send poll end message")
			}
		}
	} else {
		_, err = msg.Reply(eventCtx.Context, eventCtx.Session, *discord.NewMessage(fmt.Sprintf("The **%s** poll has ended! There were no votes.", poll.Title)))
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Str("poll_uuid", poll.PollUuid.String()).
				Msg("Failed to send poll end message")
		}
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

func formatDecimal(v float64) string {
	s := fmt.Sprintf("%.1f", v)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")

	return s
}
