package plugins

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"slices"

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
		var invokeGiveawayEndPayload core.CustomEventInvokeEndGiveawayStructure
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

func (g *GiveawayCog) OnInvokeEndGiveaway(eventCtx *sandwich.EventContext, event core.CustomEventInvokeEndGiveawayStructure) error {
	welcomer.Logger.Info().
		Str("giveaway_uuid", event.GiveawayUUID.String()).
		Msg("Received giveaway end event, processing giveaway end")

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

	if err := g.EndGiveaway(eventCtx, giveaway); err != nil {
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

func (g *GiveawayCog) EndGiveaway(eventCtx *sandwich.EventContext, giveaway *database.GuildGiveaways) error {
	if giveaway.HasEnded {
		welcomer.Logger.Error().
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Giveaway has already ended, skipping giveaway end")

		return nil
	}

	// Set giveaway as ended

	var err error

	giveaway, err = welcomer.Queries.SetGiveawayEnded(eventCtx.Context, database.SetGiveawayEndedParams{
		GiveawayUuid: giveaway.GiveawayUuid,
		HasEnded:     true,
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to set giveaway as ended")

		return err
	}

	defer func() {
		if err != nil {
			// Unset giveaway as ended on error

			giveaway, err = welcomer.Queries.SetGiveawayEnded(eventCtx.Context, database.SetGiveawayEndedParams{
				GiveawayUuid: giveaway.GiveawayUuid,
				HasEnded:     false,
			})
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
					Msg("Failed to unset giveaway as ended")
			}

		}
	}()

	prizes := welcomer.UnmarshalGiveawayPrizeJSON(giveaway.GiveawayPrizes.Bytes)

	_, err = welcomer.SandwichClient.RequestGuildChunk(eventCtx.Context, &pb.RequestGuildChunkRequest{
		GuildId: giveaway.GuildID,
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to request guild chunk for giveaway end")

		return err
	}

	entries, err := welcomer.Queries.GetGiveawayEntryUsers(eventCtx.Context, giveaway.GiveawayUuid)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to get giveaway entries for giveaway end")

		return err
	}

	// Chunk guild
	_, err = welcomer.SandwichClient.RequestGuildChunk(eventCtx, &pb.RequestGuildChunkRequest{
		GuildId:     int64(eventCtx.Guild.ID),
		AlwaysChunk: false,
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Msg("Failed to chunk guild")

		return err
	}

	members, err := welcomer.SandwichClient.FetchGuildMember(eventCtx.Context, &pb.FetchGuildMemberRequest{
		GuildId: giveaway.GuildID,
		UserIds: entries,
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to fetch guild members for giveaway end")

		return err
	}

	entries = make([]int64, 0)

	existingWinners, err := welcomer.Queries.GetGiveawayWinners(eventCtx.Context, giveaway.GiveawayUuid)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to get existing giveaway winners for giveaway end")
	}

	// Add entries for users still in server and is not an existing winner.
	for _, member := range members.GetGuildMembers() {
		if !slices.ContainsFunc(existingWinners, func(winner *database.GuildGiveawaysWinners) bool {
			return winner.UserID == int64(member.GetUser().GetID())
		}) {
			entries = append(entries, member.GetUser().GetID())
		}
	}

	wonPrizes := make(map[string]int)

	// Remove prize count for existing winners.
	for _, winner := range existingWinners {
		_, ok := wonPrizes[winner.Prize]
		if !ok {
			wonPrizes[winner.Prize] = 0
		}

		wonPrizes[winner.Prize]++
	}

	for _, prize := range prizes {
		count, ok := wonPrizes[prize.Title]
		if ok {
			prize.Count -= count
		}
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

				return err
			}

			winnerID := entries[randomIndex]

			if _, err = welcomer.Queries.CreateGiveawayWinner(eventCtx.Context, database.CreateGiveawayWinnerParams{
				GiveawayUuid: giveaway.GiveawayUuid,
				UserID:       winnerID,
				Prize:        prize.Title,
				MessageID:    0,
			}); err != nil {
				welcomer.Logger.Error().Err(err).
					Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
					Int64("winner_id", winnerID).
					Msg("Failed to add giveaway winner to database")

				return err
			}

			// Remove the selected winner from the entries slice
			entries = append(entries[:randomIndex], entries[randomIndex+1:]...)
		}
	}

	// Announce winners in giveaway channel

	winners, err := welcomer.Queries.GetGiveawayWinners(eventCtx.Context, giveaway.GiveawayUuid)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to get giveaway winners")

		return err
	}

	channel := discord.Channel{ID: discord.Snowflake(giveaway.ChannelID)}

	if giveaway.AnnounceWinners {
		for _, winner := range winners {
			if slices.ContainsFunc(existingWinners, func(value *database.GuildGiveawaysWinners) bool {
				return value.UserID == winner.UserID && value.Prize == winner.Prize && value.MessageID != 0
			}) {
				// Winner has already been announced, skip announcing again.
				continue
			}

			welcomer.Logger.Info().
				Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
				Int64("user_id", winner.UserID).
				Str("prize", winner.Prize).
				Msg("Giveaway winner selected")

			message, err := channel.Send(eventCtx.Context, eventCtx.Session, discord.MessageParams{
				Content: fmt.Sprintf("Congratulations <@%d>, you won **%s**", winner.UserID, winner.Prize),
				MessageReference: &discord.MessageReference{
					ID:              new(discord.Snowflake(giveaway.MessageID)),
					ChannelID:       new(discord.Snowflake(giveaway.ChannelID)),
					FailIfNotExists: false,
				},
			})
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
					Int64("user_id", winner.UserID).
					Str("prize", winner.Prize).
					Msg("Failed to send giveaway winner message")
			} else {
				_, err = welcomer.Queries.UpdateGiveawayWinnerMessageID(eventCtx.Context, database.UpdateGiveawayWinnerMessageIDParams{
					GiveawayWinnerUuid: winner.GiveawayWinnerUuid,
					MessageID:          int64(message.ID),
				})
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
						Int64("user_id", winner.UserID).
						Str("prize", winner.Prize).
						Msg("Failed to update giveaway winner message ID")

					return err
				}
			}
		}
	}

	msg, err := discord.GetChannelMessage(eventCtx.Context, eventCtx.Session, discord.Snowflake(giveaway.ChannelID), discord.Snowflake(giveaway.MessageID))
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to fetch giveaway message for giveaway end")

		return err
	}

	for i := range msg.Components {
		for j := range msg.Components[i].Components {
			if msg.Components[i].Components[j].Type == discord.InteractionComponentTypeButton {
				msg.Components[i].Components[j].Label = "This giveaway has ended"
				msg.Components[i].Components[j].Disabled = true
			}
		}
	}

	_, err = msg.Edit(eventCtx.Context, eventCtx.Session, discord.MessageParams{
		Components: msg.Components,
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to edit giveaway message to disable buttons for giveaway end")

		return err
	}

	return nil
}
