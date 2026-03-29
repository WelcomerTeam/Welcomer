package plugins

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich_daemon "github.com/WelcomerTeam/Sandwich-Daemon"
	pb "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GiveawayCog struct {
	EventHandler *sandwich.Handlers
}

// Assert types.

var (
	_ sandwich.Cog           = (*GiveawayCog)(nil)
	_ sandwich.CogWithEvents = (*GiveawayCog)(nil)
)

func NewGiveawayCog() *GiveawayCog {
	return &GiveawayCog{
		EventHandler: sandwich.SetupHandler(nil),
	}
}

func (g *GiveawayCog) CogInfo() *sandwich.CogInfo {
	return &sandwich.CogInfo{
		Name:        "Giveaways",
		Description: "Provides the functionality for the 'Giveaways' feature",
	}
}

func (g *GiveawayCog) GetEventHandlers() *sandwich.Handlers {
	return g.EventHandler
}

func (g *GiveawayCog) RegisterCog(bot *sandwich.Bot) error {
	// Register giveaway end handler.

	g.EventHandler.RegisterEventHandler(core.CustomEventInvokeEndGiveaway, func(eventCtx *sandwich.EventContext, payload sandwich_daemon.ProducedPayload) error {
		var invokeGiveawayEndPayload core.CustomEventEndGiveawayStructure
		if err := eventCtx.DecodeContent(payload, &invokeGiveawayEndPayload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		eventCtx.Guild = sandwich.NewGuild(invokeGiveawayEndPayload.GuildID)

		eventCtx.EventHandler.EventsMu.RLock()
		defer eventCtx.EventHandler.EventsMu.RUnlock()

		for _, event := range eventCtx.EventHandler.Events {
			if f, ok := event.(welcomer.OnInvokeEndGiveawayFuncType); ok {
				return eventCtx.Handlers.WrapFuncType(eventCtx, f(eventCtx, invokeGiveawayEndPayload))
			}
		}

		return nil
	})

	// Call OnInvokeEndGiveaway when CustomEventInvokeEndGiveaway is triggered.
	g.EventHandler.RegisterEvent(core.CustomEventInvokeEndGiveaway, nil, (welcomer.OnInvokeEndGiveawayFuncType)(g.OnInvokeEndGiveaway))

	return nil
}

func (g *GiveawayCog) OnInvokeEndGiveaway(eventCtx *sandwich.EventContext, event core.CustomEventEndGiveawayStructure) error {
	giveaway, err := welcomer.Queries.GetGiveaway(eventCtx.Context, database.GetGiveawayParams{
		GiveawayUuid: event.GiveawayUUID,
		GuildID:      int64(event.GuildID),
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", event.GiveawayUUID.String()).
			Msg("Failed to get giveaway for giveaway end event")

		return err
	}

	if giveaway.HasEnded {
		welcomer.Logger.Error().
			Str("giveaway_uuid", event.GiveawayUUID.String()).
			Msg("Giveaway has already ended, skipping giveaway end event")

		return nil
	}

	if err := g.EndGiveaway(eventCtx.Context, giveaway); err != nil {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", event.GiveawayUUID.String()).
			Msg("Failed to end giveaway for giveaway end event")

		return err
	}

	return nil
}

func secureRandomInt(n int64) (int64, error) {
	max := big.NewInt(n)

	r, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}

	return r.Int64(), nil
}

func (g *GiveawayCog) EndGiveaway(ctx context.Context, giveaway *database.GuildGiveaways) error {
	if giveaway.HasEnded {
		welcomer.Logger.Error().
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Giveaway has already ended, skipping giveaway end")

		return nil
	}

	session, err := welcomer.AcquireSession(ctx, welcomer.GetManagerNameFromContext(ctx))
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to acquire session for giveaway end")

		return err
	}

	prizes := welcomer.UnmarshalGiveawayPrizeJSON(giveaway.GiveawayPrizes.Bytes)

	_, err = welcomer.SandwichClient.RequestGuildChunk(ctx, &pb.RequestGuildChunkRequest{
		GuildId: giveaway.GuildID,
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to request guild chunk for giveaway end")

		return err
	}

	entries, err := welcomer.Queries.GetGiveawayEntryUsers(ctx, giveaway.GiveawayUuid)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to get giveaway entries for giveaway end")

		return err
	}

	members, err := welcomer.SandwichClient.FetchGuildMember(ctx, &pb.FetchGuildMemberRequest{
		GuildId: giveaway.GuildID,
		UserIds: entries,
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to fetch guild members for giveaway end")

		return err
	}

	clear(entries)

	for _, member := range members.GetGuildMembers() {
		entries = append(entries, member.GetUser().GetID())
	}

	// Assign winners

	for _, prize := range prizes {
		for i := 0; i < prize.Count; i++ {
			if len(entries) == 0 {
				welcomer.Logger.Info().
					Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
					Msg("Not enough entries to assign all winners for giveaway")

				break
			}

			randomIndex, err := secureRandomInt(int64(len(entries)))
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
					Msg("Failed to generate secure random index for giveaway winner selection")

				continue
			}

			winnerID := entries[randomIndex]

			if _, err = welcomer.Queries.CreateGiveawayWinner(ctx, database.CreateGiveawayWinnerParams{
				GiveawayUuid: giveaway.GiveawayUuid,
				UserID:       winnerID,
				Prize:        prize.Title,
				MessageID:    0,
			}); err != nil {
				welcomer.Logger.Error().Err(err).
					Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
					Int64("winner_id", winnerID).
					Msg("Failed to add giveaway winner to database")

				continue
			}

			// Remove the selected winner from the entries slice
			entries = append(entries[:randomIndex], entries[randomIndex+1:]...)
		}
	}

	// Announce winners in giveaway channel

	winners, err := welcomer.Queries.GetGiveawayWinners(ctx, giveaway.GiveawayUuid)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to get giveaway winners")

		return err
	}

	channel := discord.Channel{ID: discord.Snowflake(giveaway.ChannelID)}

	for _, winner := range winners {
		welcomer.Logger.Info().
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Int64("user_id", winner.UserID).
			Str("prize", winner.Prize).
			Msg("Giveaway winner selected")

		message, err := channel.Send(ctx, session, discord.MessageParams{
			Content: fmt.Sprintf("Congratulations <@%d>, you won **%s**", winner.UserID, winner.Prize),
		})
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
				Int64("user_id", winner.UserID).
				Str("prize", winner.Prize).
				Msg("Failed to send giveaway winner message")
		} else {
			_, err = welcomer.Queries.UpdateGiveawayWinnerMessageID(ctx, database.UpdateGiveawayWinnerMessageIDParams{
				GiveawayWinnerUuid: winner.GiveawayWinnerUuid,
				MessageID:          int64(message.ID),
			})
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
					Int64("user_id", winner.UserID).
					Str("prize", winner.Prize).
					Msg("Failed to update giveaway winner message ID")
			}
		}
	}

	// Set giveaway as ended

	giveaway, err = welcomer.Queries.SetGiveawayEnded(ctx, database.SetGiveawayEndedParams{
		GiveawayUuid: giveaway.GiveawayUuid,
		HasEnded:     true,
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to set giveaway as ended")

		return err
	}

	return nil
}
