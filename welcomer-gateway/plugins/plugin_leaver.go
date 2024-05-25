package plugins

import (
	"errors"
	"fmt"
	"net/http"

	"encoding/json"

	"github.com/WelcomerTeam/Discord/discord"
	pb "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	"github.com/WelcomerTeam/Sandwich-Daemon/structs"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/jackc/pgx/v4"
	"github.com/savsgio/gotils/strconv"
)

type LeaverCog struct {
	EventHandler *sandwich.Handlers
	Client       http.Client
}

// Assert types.

var (
	_ sandwich.Cog           = (*LeaverCog)(nil)
	_ sandwich.CogWithEvents = (*LeaverCog)(nil)
)

func NewLeaverCog() *LeaverCog {
	return &LeaverCog{
		EventHandler: sandwich.SetupHandler(nil),
		Client:       http.Client{},
	}
}

func (p *LeaverCog) CogInfo() *sandwich.CogInfo {
	return &sandwich.CogInfo{
		Name:        "Leaver",
		Description: "Provides the functionality for the 'Leaver' feature",
	}
}

func (p *LeaverCog) GetEventHandlers() *sandwich.Handlers {
	return p.EventHandler
}

func (p *LeaverCog) RegisterCog(bot *sandwich.Bot) error {

	// Register CustomEventInvokeLeaver event.
	p.EventHandler.RegisterEventHandler(core.CustomEventInvokeLeaver, func(eventCtx *sandwich.EventContext, payload structs.SandwichPayload) error {
		var invokeLeaverPayload core.CustomEventInvokeLeaverStructure
		if err := eventCtx.DecodeContent(payload, &invokeLeaverPayload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		eventCtx.Guild = sandwich.NewGuild(invokeLeaverPayload.GuildID)

		eventCtx.EventHandler.EventsMu.RLock()
		defer eventCtx.EventHandler.EventsMu.RUnlock()

		for _, event := range eventCtx.EventHandler.Events {
			if f, ok := event.(welcomer.OnInvokeLeaverFuncType); ok {
				return eventCtx.Handlers.WrapFuncType(eventCtx, f(eventCtx, invokeLeaverPayload))
			}
		}

		return nil
	})

	// Trigger CustomEventInvokeLeaver when ON_GUILD_MEMBER_REMOVE event is received.
	p.EventHandler.RegisterOnGuildMemberRemoveEvent(func(eventCtx *sandwich.EventContext, user discord.User) error {
		return p.OnInvokeLeaverEvent(eventCtx, core.CustomEventInvokeLeaverStructure{
			Interaction: nil,
			User:        user,
		})
	})

	// Call OnInvokeLeaverEvent when CustomEventInvokeLeaver is triggered.
	p.EventHandler.RegisterEvent(core.CustomEventInvokeLeaver, nil, (welcomer.OnInvokeLeaverFuncType)(p.OnInvokeLeaverEvent))

	return nil
}

// OnInvokeLeaverEvent is called when CustomEventInvokeLeaver is triggered.
// This can be from when a user joins or a user uses /welcomer test.
func (p *LeaverCog) OnInvokeLeaverEvent(eventCtx *sandwich.EventContext, event core.CustomEventInvokeLeaverStructure) (err error) {
	defer func() {
		// Send follow-up if present.
		if event.Interaction != nil && event.Interaction.Token != "" {
			var message discord.WebhookMessageParams

			if err == nil {
				message = discord.WebhookMessageParams{
					Embeds: utils.NewEmbed("Executed successfully", utils.EmbedColourSuccess),
				}
			} else {
				message = discord.WebhookMessageParams{
					Embeds: utils.NewEmbed(fmt.Sprintf("Failed to execute: `%s`", err.Error()), utils.EmbedColourError),
				}
			}

			_, err = event.Interaction.SendFollowup(eventCtx.Session, message)
			if err != nil {
				eventCtx.Logger.Warn().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("application_id", int64(event.Interaction.ApplicationID)).
					Str("token", event.Interaction.Token).
					Msg("Failed to send interaction follow-up")
			}
		}
	}()

	queries := welcomer.GetQueriesFromContext(eventCtx.Context)

	// Fetch guild settings.

	guildSettingsLeaver, err := queries.GetLeaverGuildSettings(eventCtx.Context, int64(eventCtx.Guild.ID))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.User.ID)).
			Msg("Failed to get leaver guild settings")

		return err
	}

	// Quit if leaver is not enabled or configured.
	if !guildSettingsLeaver.ToggleEnabled || guildSettingsLeaver.Channel == 0 || utils.IsJSONBEmpty(guildSettingsLeaver.MessageFormat.Bytes) {
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
	variables := welcomer.GatherVariables(eventCtx, discord.GuildMember{
		GuildID: &event.GuildID,
		User:    &event.User,
	}, *guild, nil)

	messageFormat, err := welcomer.FormatString(functions, variables, strconv.B2S(guildSettingsLeaver.MessageFormat.Bytes))
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.User.ID)).
			Msg("Failed to format leaver text payload")

		return err
	}

	var serverMessage discord.MessageParams

	// Convert MessageFormat to MessageParams so we can send it.
	err = json.Unmarshal(strconv.S2B(messageFormat), &serverMessage)
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.User.ID)).
			Msg("Failed to unmarshal leaver messageFormat")

		return err
	}

	// Send the message if it's not empty.
	if !utils.IsMessageParamsEmpty(serverMessage) {
		channel := discord.Channel{ID: discord.Snowflake(guildSettingsLeaver.Channel)}

		_, err = channel.Send(eventCtx.Session, serverMessage)
		if err != nil {
			eventCtx.Logger.Warn().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("channel_id", guildSettingsLeaver.Channel).
				Msg("Failed to send leaver message to channel")
		}
	}

	return nil
}
