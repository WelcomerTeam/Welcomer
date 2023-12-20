package plugins

import (
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
	pb "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	"github.com/WelcomerTeam/Sandwich-Daemon/structs"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	jsoniter "github.com/json-iterator/go"
	"github.com/savsgio/gotils/strconv"
)

type WelcomerCog struct {
	EventHandler *sandwich.Handlers
}

// Assert types.

var (
	_ sandwich.Cog           = (*WelcomerCog)(nil)
	_ sandwich.CogWithEvents = (*WelcomerCog)(nil)
)

func NewWelcomerCog() *WelcomerCog {
	return &WelcomerCog{
		EventHandler: sandwich.SetupHandler(nil),
	}
}

func (p *WelcomerCog) CogInfo() *sandwich.CogInfo {
	return &sandwich.CogInfo{
		Name:        "Welcomer",
		Description: "Provides the functionality for the 'Welcomer' feature",
	}
}

func (p *WelcomerCog) GetEventHandlers() *sandwich.Handlers {
	return p.EventHandler
}

func (p *WelcomerCog) RegisterCog(bot *sandwich.Bot) error {

	// Register CustomEventInvokeWelcomer event.
	bot.RegisterEventHandler(welcomer.CustomEventInvokeWelcomer, func(eventCtx *sandwich.EventContext, payload structs.SandwichPayload) error {
		var invokeWelcomerPayload welcomer.CustomEventInvokeWelcomerStructure
		if err := eventCtx.DecodeContent(payload, &invokeWelcomerPayload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		if invokeWelcomerPayload.Member.GuildID != nil {
			eventCtx.Guild = sandwich.NewGuild(*invokeWelcomerPayload.Member.GuildID)
		}

		eventCtx.EventHandler.EventsMu.RLock()
		defer eventCtx.EventHandler.EventsMu.RUnlock()

		for _, event := range eventCtx.EventHandler.Events {
			if f, ok := event.(welcomer.OnInvokeWelcomerFuncType); ok {
				return eventCtx.Handlers.WrapFuncType(eventCtx, f(eventCtx, invokeWelcomerPayload))
			}
		}

		return nil
	})

	// Trigger CustomEventInvokeWelcomer when ON_GUILD_MEMBER_ADD is triggered.
	bot.RegisterOnGuildMemberAddEvent(func(eventCtx *sandwich.EventContext, member discord.GuildMember) error {
		return p.OnInvokeWelcomerEvent(eventCtx, welcomer.CustomEventInvokeWelcomerStructure{
			Interaction: nil,
			Member:      &member,
		})
	})

	// Call OnInvokeWelcomerEvent when CustomEventInvokeWelcomer is triggered.
	RegisterOnInvokeWelcomerEvent(bot.Handlers, p.OnInvokeWelcomerEvent)

	return nil
}

// RegisterOnInvokeWelcomerEvent adds a new event handler for the WELCOMER_INVOKE_WELCOMER event.
// It does not override a handler and instead will add another handler.
func RegisterOnInvokeWelcomerEvent(h *sandwich.Handlers, event welcomer.OnInvokeWelcomerFuncType) {
	eventName := welcomer.CustomEventInvokeWelcomer

	h.RegisterEvent(eventName, nil, event)
}

// OnInvokeWelcomerEvent is called when CustomEventInvokeWelcomer is triggered.
// This can be from when a user joins or a user uses /welcomer test.
func (p *WelcomerCog) OnInvokeWelcomerEvent(eventCtx *sandwich.EventContext, event welcomer.CustomEventInvokeWelcomerStructure) (err error) {
	defer func() {
		// Send follow-up if present.
		if event.Interaction != nil && event.Interaction.Token != "" {
			var message discord.WebhookMessageParams

			if err == nil {
				message = discord.WebhookMessageParams{
					Content: "✔️ Executed successfully!",
				}
			} else {
				message = discord.WebhookMessageParams{
					Content: fmt.Sprintf("❌ Failed to execute: `%s`", err.Error()),
				}
			}

			_, err = event.Interaction.SendFollowup(eventCtx.Session, message)
			if err != nil {
				eventCtx.Logger.Error().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("application_id", int64(event.Interaction.ApplicationID)).
					Str("token", event.Interaction.Token).
					Msg("Failed to send interaction follow-up")
			}
		}
	}()

	queries := welcomer.GetQueriesFromContext(eventCtx.Context)

	guildSettingsWelcomerText, err := queries.GetWelcomerTextGuildSettings(eventCtx.Context, int64(eventCtx.Guild.ID))
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.Member.User.ID)).
			Msg("Failed to fetch welcomer text guild settings")

		return err
	}

	// Quit if disabled or not channel configured.
	if !guildSettingsWelcomerText.ToggleEnabled || guildSettingsWelcomerText.Channel == 0 {
		return nil
	}

	// Query state cache for guild.
	guilds, err := eventCtx.Sandwich.SandwichClient.FetchGuild(eventCtx, &pb.FetchGuildRequest{
		GuildIDs: []int64{int64(eventCtx.Guild.ID)},
	})
	if err != nil {
		return err
	}

	var guild *discord.Guild
	guildPb, ok := guilds.Guilds[int64(eventCtx.Guild.ID)]
	if ok {
		guild, err = pb.GRPCToGuild(guildPb)
		if err != nil {
			return err
		}
	}

	functions := welcomer.GatherFunctions()
	variables := welcomer.GatherVariables(eventCtx, *event.Member, *guild)

	messageFormat, err := welcomer.FormatString(functions, variables, strconv.B2S(guildSettingsWelcomerText.MessageFormat.Bytes))
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.Member.User.ID)).
			Msg("Failed to format welcomer text payload")

		return err
	}

	var message discord.MessageParams

	// Convert MessageFormat to MessageParams so we can send it.
	err = jsoniter.UnmarshalFromString(messageFormat, &message)
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.Member.User.ID)).
			Msg("Failed to unmarshal embedJSON")

		return err
	}

	channel := discord.Channel{ID: discord.Snowflake(guildSettingsWelcomerText.Channel)}

	_, err = channel.Send(eventCtx.Session, message)
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("channel_id", guildSettingsWelcomerText.Channel).
			Msg("Failed to send message to channel")

		return err
	}

	// Create welcomer image

	// Construct a message for the server
	// Construct a message for DM

	return nil
}
