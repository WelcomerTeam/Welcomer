package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich_daemon "github.com/WelcomerTeam/Sandwich-Daemon"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
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
	p.EventHandler.RegisterEventHandler(core.CustomEventInvokeLeaver, func(eventCtx *sandwich.EventContext, payload sandwich_daemon.ProducedPayload) error {
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
		welcomer.PusherGuildScience.Push(
			eventCtx.Context,
			eventCtx.Guild.ID,
			user.ID,
			database.ScienceGuildEventTypeUserLeave,
			nil,
		)

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
					Embeds: welcomer.NewEmbed("Executed successfully", welcomer.EmbedColourSuccess),
				}
			} else {
				message = discord.WebhookMessageParams{
					Embeds: welcomer.NewEmbed(fmt.Sprintf("Failed to execute: `%s`", err.Error()), welcomer.EmbedColourError),
				}
			}

			_, err = event.Interaction.SendFollowup(eventCtx.Context, eventCtx.Session, message)
			if err != nil {
				welcomer.Logger.Warn().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("application_id", int64(event.Interaction.ApplicationID)).
					Str("token", event.Interaction.Token).
					Msg("Failed to send interaction follow-up")
			}
		}
	}()

	// Fetch guild settings.

	guildSettingsLeaver, err := welcomer.Queries.GetLeaverGuildSettings(eventCtx.Context, int64(eventCtx.Guild.ID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			guildSettingsLeaver = &database.GuildSettingsLeaver{
				GuildID:       int64(eventCtx.Guild.ID),
				ToggleEnabled: welcomer.DefaultLeaver.ToggleEnabled,
				Channel:       welcomer.DefaultLeaver.Channel,
				MessageFormat: welcomer.DefaultLeaver.MessageFormat,
			}
		} else {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.User.ID)).
				Msg("Failed to get leaver guild settings")

			return err
		}
	}

	// Quit if leaver is not enabled or configured.
	if !guildSettingsLeaver.ToggleEnabled || guildSettingsLeaver.Channel == 0 || welcomer.IsJSONBEmpty(guildSettingsLeaver.MessageFormat.Bytes) {
		return nil
	}

	guild, err := welcomer.FetchGuild(eventCtx.Context, eventCtx.Guild.ID)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.User.ID)).
			Msg("Failed to fetch guild from state cache")

		guild = eventCtx.Guild
	}

	guildSettings, err := welcomer.Queries.GetGuild(eventCtx.Context, int64(eventCtx.Guild.ID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			guildSettings = &welcomer.DefaultGuild
		} else {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Msg("Failed to get guild settings")
		}
	}

	functions := welcomer.GatherFunctions(database.NumberLocale(guildSettings.NumberLocale.Int32))
	variables := welcomer.GatherVariables(eventCtx, &discord.GuildMember{
		GuildID: &event.GuildID,
		User:    &event.User,
	}, core.GuildVariables{
		Guild:         guild,
		MembersJoined: guildSettings.MemberCount,
		NumberLocale:  database.NumberLocale(guildSettings.NumberLocale.Int32),
	}, nil, nil)

	messageFormat, err := welcomer.FormatString(functions, variables, strconv.B2S(guildSettingsLeaver.MessageFormat.Bytes))
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.User.ID)).
			Str("message_format", messageFormat).
			Msg("Failed to format leaver text payload")

		return err
	}

	var serverMessage discord.MessageParams

	// Convert MessageFormat to MessageParams so we can send it.
	err = json.Unmarshal(strconv.S2B(messageFormat), &serverMessage)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.User.ID)).
			Str("message_format", messageFormat).
			Msg("Failed to unmarshal leaver messageFormat")

		return err
	}

	var messageID discord.Snowflake
	var channelID discord.Snowflake

	// Send the message if it's not empty.
	if !welcomer.IsMessageParamsEmpty(serverMessage) {
		validGuild, err := core.CheckChannelGuild(eventCtx.Context, welcomer.SandwichClient, eventCtx.Guild.ID, discord.Snowflake(guildSettingsLeaver.Channel))
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("channel_id", guildSettingsLeaver.Channel).
				Msg("Failed to check channel guild")
		} else if !validGuild {
			welcomer.Logger.Warn().
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("channel_id", guildSettingsLeaver.Channel).
				Msg("Channel does not belong to guild")
		} else {
			channel := discord.Channel{ID: discord.Snowflake(guildSettingsLeaver.Channel)}

			message, err := channel.Send(eventCtx.Context, eventCtx.Session, serverMessage)

			welcomer.Logger.Info().
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("channel_id", guildSettingsLeaver.Channel).
				Msg("Sent leaver message to channel")

			if err != nil {
				welcomer.Logger.Warn().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("channel_id", guildSettingsLeaver.Channel).
					Msg("Failed to send leaver message to channel")
			} else {
				messageID = message.ID
				channelID = channel.ID
			}
		}
	}

	welcomer.PusherGuildScience.Push(
		eventCtx.Context,
		eventCtx.Guild.ID,
		event.User.ID,
		database.ScienceGuildEventTypeUserLeftMessage,
		core.GuildScienceUserLeftMessage{
			HasMessage:       messageID != 0,
			MessageID:        messageID,
			MessageChannelID: channelID,
		},
	)

	return nil
}
