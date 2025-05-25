package plugins

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich_daemon "github.com/WelcomerTeam/Sandwich-Daemon"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4"
)

type TempChannelsCog struct {
	EventHandler *sandwich.Handlers
}

// Assert Types

var (
	_ sandwich.Cog           = (*TempChannelsCog)(nil)
	_ sandwich.CogWithEvents = (*TempChannelsCog)(nil)
)

func NewTempChannelsCog() *TempChannelsCog {
	return &TempChannelsCog{
		EventHandler: sandwich.SetupHandler(nil),
	}
}

func (p *TempChannelsCog) CogInfo() *sandwich.CogInfo {
	return &sandwich.CogInfo{
		Name:        "TempChannels",
		Description: "Provides the functionality for the 'TempChannels' feature",
	}
}

func (p *TempChannelsCog) GetEventHandlers() *sandwich.Handlers {
	return p.EventHandler
}

func (p *TempChannelsCog) RegisterCog(bot *sandwich.Bot) error {
	// Register CustomEventInvokeTempChannels event.
	p.EventHandler.RegisterEventHandler(core.CustomEventInvokeTempChannels, func(eventCtx *sandwich.EventContext, payload sandwich_daemon.ProducedPayload) error {
		var invokeTempChannelsPayload core.CustomEventInvokeTempChannelsStructure
		if err := eventCtx.DecodeContent(payload, &invokeTempChannelsPayload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		if invokeTempChannelsPayload.Member.GuildID == nil {
			eventCtx.Guild = sandwich.NewGuild(*invokeTempChannelsPayload.Member.GuildID)
		}

		eventCtx.EventHandler.EventsMu.RLock()
		defer eventCtx.EventHandler.EventsMu.RUnlock()

		for _, event := range eventCtx.EventHandler.Events {
			if f, ok := event.(welcomer.OnInvokeTempChannelsFuncType); ok {
				return eventCtx.Handlers.WrapFuncType(eventCtx, f(eventCtx, invokeTempChannelsPayload))
			}
		}

		return nil
	})

	// Register CustomEventInvokeTempChannelsRemove event.
	p.EventHandler.RegisterEventHandler(core.CustomEventInvokeTempChannelsRemove, func(eventCtx *sandwich.EventContext, payload sandwich_daemon.ProducedPayload) error {
		var invokeTempChannelsRemovePayload core.CustomEventInvokeTempChannelsRemoveStructure
		if err := eventCtx.DecodeContent(payload, &invokeTempChannelsRemovePayload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		if invokeTempChannelsRemovePayload.Member.GuildID == nil {
			eventCtx.Guild = sandwich.NewGuild(*invokeTempChannelsRemovePayload.Member.GuildID)
		}

		eventCtx.EventHandler.EventsMu.RLock()
		defer eventCtx.EventHandler.EventsMu.RUnlock()

		for _, event := range eventCtx.EventHandler.Events {
			if f, ok := event.(welcomer.OnInvokeTempChannelsRemoveFuncType); ok {
				return eventCtx.Handlers.WrapFuncType(eventCtx, f(eventCtx, invokeTempChannelsRemovePayload))
			}
		}

		return nil
	})

	// Trigger OnInvokeVoiceStateUpdate when VoiceStateUpdate event is triggered.
	p.EventHandler.RegisterOnVoiceStateUpdateEvent(func(eventCtx *sandwich.EventContext, member discord.GuildMember, before, after discord.VoiceState) error {
		if before.ChannelID != after.ChannelID {
			return p.OnInvokeVoiceStateUpdate(eventCtx, member, before, after)
		}

		return nil
	})

	// Call OnInvokeTempChannelsEvent when CustomEventInvokeTempChannels is triggered.
	p.EventHandler.RegisterEvent(core.CustomEventInvokeTempChannels, nil, (welcomer.OnInvokeTempChannelsFuncType)(p.OnInvokeTempChannelsEvent))

	// Call OnInvokeTempChannelsRemoveEvent when CustomEventInvokeTempChannelsRemove is triggered.
	p.EventHandler.RegisterEvent(core.CustomEventInvokeTempChannelsRemove, nil, (welcomer.OnInvokeTempChannelsRemoveFuncType)(p.OnInvokeTempChannelsRemoveEvent))

	return nil
}

func (p *TempChannelsCog) OnInvokeVoiceStateUpdate(eventCtx *sandwich.EventContext, member discord.GuildMember, before, after discord.VoiceState) error {
	// Fetch guild settings.

	var guildID discord.Snowflake

	if after.GuildID != nil {
		guildID = *after.GuildID
	} else if member.GuildID != nil {
		guildID = *member.GuildID
	} else {
		welcomer.Logger.Error().
			Str("member_id", member.User.ID.String()).
			Msg("Failed to get guild ID for temp channels")

		return nil
	}

	guildSettingsTimeroles, err := p.FetchGuildInformation(eventCtx, guildID)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("guild_id", guildID.String()).
			Str("member_id", member.User.ID.String()).
			Msg("Failed to get temp channel settings")

		return err
	}

	if !guildSettingsTimeroles.ToggleEnabled {
		return nil
	}

	// If user is moving to lobby, create channel and move user.
	if guildSettingsTimeroles.ChannelLobby != 0 && after.ChannelID == discord.Snowflake(guildSettingsTimeroles.ChannelLobby) {
		return p.createChannelAndMove(eventCtx, guildID, welcomer.ToPointer(discord.Snowflake(guildSettingsTimeroles.ChannelCategory)), &member)
	}

	var beforeGuildID discord.Snowflake
	var afterGuildID discord.Snowflake

	if before.GuildID != nil {
		beforeGuildID = *before.GuildID
	}

	if after.GuildID != nil {
		afterGuildID = *after.GuildID
	}

	// If user is leaving or moving to a different channel, delete channel if empty.
	if guildSettingsTimeroles.ToggleAutopurge {
		if (after.GuildID == nil || beforeGuildID != afterGuildID || before.ChannelID != after.ChannelID) && !before.ChannelID.IsNil() {
			_, err = p.deleteChannelIfEmpty(eventCtx, guildID, discord.Snowflake(guildSettingsTimeroles.ChannelCategory), discord.Snowflake(guildSettingsTimeroles.ChannelLobby), before.ChannelID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *TempChannelsCog) findChannelForUser(eventCtx *sandwich.EventContext, guildID discord.Snowflake, category *discord.Snowflake, member *discord.GuildMember) (channel *discord.Channel, err error) {
	channels, err := welcomer.FetchGuildChannels(eventCtx.Context, guildID)
	if err != nil {
		return nil, err
	}

	for _, guildChannel := range channels {
		if category == nil || (guildChannel.ParentID != nil && *guildChannel.ParentID == *category) {
			return guildChannel, nil
		}
	}

	return nil, nil
}

func (p *TempChannelsCog) createChannelAndMove(eventCtx *sandwich.EventContext, guildID discord.Snowflake, category *discord.Snowflake, member *discord.GuildMember) (err error) {
	channel, err := p.findChannelForUser(eventCtx, guildID, category, member)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("guild_id", guildID.String()).
			Str("member_id", member.User.ID.String()).
			Msg("Failed to find channel for tempchannels")

		return err
	}

	if channel == nil {
		if category == nil {
			welcomer.Logger.Error().
				Msg("Failed to get category for temp channels")

			return nil
		}

		guild := sandwich.NewGuild(guildID)
		channel, err = guild.CreateChannel(eventCtx.Context, eventCtx.Session, discord.ChannelParams{
			Name:     p.formatChannelName(member),
			Type:     discord.ChannelTypeGuildVoice,
			ParentID: category,
		}, welcomer.ToPointer("Automatically created by TempChannels"))
		if err != nil || channel == nil {
			welcomer.Logger.Error().Err(err).
				Str("guild_id", guildID.String()).
				Msg("Failed to create channel for tempchannels")

			return err
		}
	}

	err = member.MoveTo(eventCtx.Context, eventCtx.Session, &channel.ID, welcomer.ToPointer("Automatically move by TempChannels"))
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("guild_id", guildID.String()).
			Str("member_id", member.User.ID.String()).
			Str("channel_id", channel.ID.String()).
			Msg("Failed to move member to channel for tempchannels")

		return err
	}

	welcomer.PushGuildScience.Push(
		eventCtx.Context,
		eventCtx.Guild.ID,
		member.User.ID,
		database.ScienceGuildEventTypeTempChannelCreated,
		nil,
	)

	return nil
}

func (p *TempChannelsCog) formatChannelName(member *discord.GuildMember) string {
	return fmt.Sprintf("🔊 %s's Channel [%d]", welcomer.GetGuildMemberDisplayName(member), member.User.ID)
}

var tempChannelRegex = regexp.MustCompile(`🔊 (.+)'s Channel \[(\d+)\]`)

func (p *TempChannelsCog) isTempChannel(name string) bool {
	return tempChannelRegex.MatchString(name)
}

func (p *TempChannelsCog) deleteChannelIfEmpty(eventCtx *sandwich.EventContext, guildID, category, lobby, channelID discord.Snowflake) (ok bool, err error) {
	channel := sandwich.NewChannel(&guildID, channelID)

	channel, err = sandwich.FetchChannel(eventCtx.ToGRPCContext(), channel)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("guild_id", guildID.String()).
			Str("channel_id", channelID.String()).
			Msg("Failed to fetch channel for tempchannels")

		return false, err
	}

	if !p.isTempChannel(channel.Name) {
		return false, welcomer.ErrInvalidTempChannel
	}
	if (channel.ParentID != nil && *channel.ParentID == category) && channel.MemberCount == 0 && channel.ID != lobby {
		err = channel.Delete(eventCtx.Context, eventCtx.Session, welcomer.ToPointer("Automatically deleted by TempChannels"))
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Str("guild_id", guildID.String()).
				Str("channel_id", channelID.String()).
				Msg("Failed to delete channel for tempchannels")

			return false, err
		}

		return true, nil
	}

	return false, nil
}

func (p *TempChannelsCog) OnInvokeTempChannelsEvent(eventCtx *sandwich.EventContext, payload core.CustomEventInvokeTempChannelsStructure) error {
	// Fetch guild settings.

	guildSettingsTimeroles, err := p.FetchGuildInformation(eventCtx, *payload.Member.GuildID)
	if err != nil {
		return err
	}

	if !guildSettingsTimeroles.ToggleEnabled {
		return nil
	}

	return p.createChannelAndMove(eventCtx, *payload.Member.GuildID, welcomer.ToPointer(discord.Snowflake(guildSettingsTimeroles.ChannelCategory)), &payload.Member)
}

func (p *TempChannelsCog) OnInvokeTempChannelsRemoveEvent(eventCtx *sandwich.EventContext, payload core.CustomEventInvokeTempChannelsRemoveStructure) error {
	// Fetch guild settings.

	guildSettingsTimeroles, err := p.FetchGuildInformation(eventCtx, *payload.Member.GuildID)
	if err != nil {
		return err
	}

	if !guildSettingsTimeroles.ToggleEnabled {
		return nil
	}

	channel, err := p.findChannelForUser(eventCtx, *payload.Interaction.GuildID, welcomer.ToPointer(discord.Snowflake(guildSettingsTimeroles.ChannelCategory)), &payload.Member)
	if err != nil {
		return err
	}

	if channel != nil && channel.ID != discord.Snowflake(guildSettingsTimeroles.ChannelLobby) {
		err = channel.Delete(eventCtx.Context, eventCtx.Session, welcomer.ToPointer("Automatically deleted by TempChannels"))
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Str("guild_id", payload.Member.GuildID.String()).
				Str("channel_id", channel.ID.String()).
				Msg("Failed to delete channel for tempchannels")
		}
	}

	return nil
}

func (p *TempChannelsCog) FetchGuildInformation(eventCtx *sandwich.EventContext, guildID discord.Snowflake) (guildSettingsTempChannels *database.GuildSettingsTempchannels, err error) {
	guildSettingsTempChannels, err = welcomer.Queries.GetTempChannelsGuildSettings(eventCtx.Context, int64(guildID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			guildSettingsTempChannels = &database.GuildSettingsTempchannels{
				GuildID:          int64(guildID),
				ToggleEnabled:    welcomer.DefaultTempChannels.ToggleEnabled,
				ToggleAutopurge:  welcomer.DefaultTempChannels.ToggleAutopurge,
				ChannelLobby:     welcomer.DefaultTempChannels.ChannelLobby,
				ChannelCategory:  welcomer.DefaultTempChannels.ChannelCategory,
				DefaultUserCount: welcomer.DefaultTempChannels.DefaultUserCount,
			}
		} else {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(guildID)).
				Msg("Failed to get temp channel settings")

			return nil, err
		}
	}

	return guildSettingsTempChannels, nil
}
