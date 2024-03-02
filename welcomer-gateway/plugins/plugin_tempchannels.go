package plugins

import (
	"errors"
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
	pb "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	"github.com/WelcomerTeam/Sandwich-Daemon/structs"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
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
	p.EventHandler.RegisterEventHandler(welcomer.CustomEventInvokeTempChannels, func(eventCtx *sandwich.EventContext, payload structs.SandwichPayload) error {
		var invokeTempChannelsPayload welcomer.CustomEventInvokeTempChannelsStructure
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
	p.EventHandler.RegisterEventHandler(welcomer.CustomEventInvokeTempChannelsRemove, func(eventCtx *sandwich.EventContext, payload structs.SandwichPayload) error {
		var invokeTempChannelsRemovePayload welcomer.CustomEventInvokeTempChannelsRemoveStructure
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
		return p.OnInvokeVoiceStateUpdate(eventCtx, member, before, after)
	})

	// Call OnInvokeTempChannelsEvent when CustomEventInvokeTempChannels is triggered.
	p.EventHandler.RegisterEvent(welcomer.CustomEventInvokeTempChannels, nil, (welcomer.OnInvokeTempChannelsFuncType)(p.OnInvokeTempChannelsEvent))

	// Call OnInvokeTempChannelsRemoveEvent when CustomEventInvokeTempChannelsRemove is triggered.
	p.EventHandler.RegisterEvent(welcomer.CustomEventInvokeTempChannelsRemove, nil, (welcomer.OnInvokeTempChannelsRemoveFuncType)(p.OnInvokeTempChannelsRemoveEvent))

	return nil
}

func (p *TempChannelsCog) OnInvokeVoiceStateUpdate(eventCtx *sandwich.EventContext, member discord.GuildMember, before, after discord.VoiceState) error {
	queries := welcomer.GetQueriesFromContext(eventCtx.Context)

	// Fetch guild settings.

	var guildID discord.Snowflake

	if after.GuildID != nil {
		guildID = *after.GuildID
	} else if member.GuildID != nil {
		guildID = *member.GuildID
	} else {
		return nil
	}

	guildSettingsTimeroles, err := p.FetchGuildInformation(eventCtx, queries, guildID)
	if err != nil {
		return err
	}

	if !guildSettingsTimeroles.ToggleEnabled {

		return nil
	}

	// If user is moving to lobby, create channel and move user.
	if after.ChannelID == discord.Snowflake(guildSettingsTimeroles.ChannelLobby) {
		return p.CreateChannelAndMove(eventCtx, guildID, welcomer.ToPointer(discord.Snowflake(guildSettingsTimeroles.ChannelCategory)), member)
	}

	// If user is leaving or moving to a different channel, delete channel if empty.
	if guildSettingsTimeroles.ToggleAutopurge {
		if (before.GuildID == nil || after.GuildID == nil) || *before.GuildID != *after.GuildID || before.ChannelID != after.ChannelID {
			_, err = p.DeleteChannelIfEmpty(eventCtx, guildID, discord.Snowflake(guildSettingsTimeroles.ChannelCategory), discord.Snowflake(guildSettingsTimeroles.ChannelLobby), before.ChannelID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *TempChannelsCog) FindChannelForUser(eventCtx *sandwich.EventContext, guildID discord.Snowflake, category *discord.Snowflake, member discord.GuildMember) (channel *discord.Channel, err error) {
	channels, err := eventCtx.Sandwich.SandwichClient.FetchGuildChannels(eventCtx.Context, &pb.FetchGuildChannelsRequest{
		GuildID: int64(guildID),
		Query:   "[" + member.User.ID.String() + "]",
	})
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Str("guild_id", guildID.String()).
			Str("user_id", member.User.ID.String()).
			Msg("failed to fetch guild channels")
	}

	for _, guildChannel := range channels.GuildChannels {
		if category == nil || guildChannel.ParentID == int64(*category) {
			channel, err = pb.GRPCToChannel(guildChannel)
			if err != nil {
				eventCtx.Logger.Error().Err(err).
					Str("guild_id", guildID.String()).
					Int64("channel_id", guildChannel.ID).
					Msg("failed to convert channel")
			}

			return channel, nil
		}
	}

	return nil, nil
}

func (p *TempChannelsCog) CreateChannelAndMove(eventCtx *sandwich.EventContext, guildID discord.Snowflake, category *discord.Snowflake, member discord.GuildMember) (err error) {
	channel, err := p.FindChannelForUser(eventCtx, guildID, category, member)
	if err != nil {
		return err
	}

	if channel == nil {
		if category == nil {
			return nil
		}

		guild := sandwich.NewGuild(guildID)
		channel, err = guild.CreateChannel(eventCtx.Session, discord.ChannelParams{
			Name:     p.FormatChannelName(member),
			Type:     discord.ChannelTypeGuildVoice,
			ParentID: category,
		}, welcomer.ToPointer("Automatically created by TempChannels"))
		if err != nil {
			eventCtx.Logger.Error().Err(err).
				Str("guild_id", guildID.String()).
				Msg("failed to create channel")

			return err
		}
	}

	err = member.MoveTo(eventCtx.Session, &channel.ID, welcomer.ToPointer("Automatically move by TempChannels"))
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Str("guild_id", guildID.String()).
			Str("member_id", member.User.ID.String()).
			Str("channel_id", channel.ID.String()).
			Msg("failed to move member to channel")

		return err
	}

	return nil
}

func (p *TempChannelsCog) FormatChannelName(member discord.GuildMember) string {
	return fmt.Sprintf("🔊 %s's Channel [%d]", welcomer.GetGuildMemberDisplayName(member), member.User.ID)
}

func (p *TempChannelsCog) DeleteChannelIfEmpty(eventCtx *sandwich.EventContext, guildID discord.Snowflake, category discord.Snowflake, lobby discord.Snowflake, channelID discord.Snowflake) (ok bool, err error) {
	channel := sandwich.NewChannel(&guildID, channelID)
	channel, err = sandwich.FetchChannel(eventCtx.ToGRPCContext(), channel)
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Str("guild_id", guildID.String()).
			Str("channel_id", channelID.String()).
			Msg("failed to fetch channel")

		return false, err
	}

	if (channel.ParentID != nil && *channel.ParentID == category) && channel.MemberCount == 0 && channel.ID != lobby {
		err = channel.Delete(eventCtx.Session, welcomer.ToPointer("Automatically deleted by TempChannels"))
		if err != nil {
			eventCtx.Logger.Error().Err(err).
				Str("guild_id", guildID.String()).
				Str("channel_id", channelID.String()).
				Msg("failed to delete channel")

			return false, err
		}

		return true, nil
	}

	return false, nil
}

func (p *TempChannelsCog) OnInvokeTempChannelsEvent(eventCtx *sandwich.EventContext, payload welcomer.CustomEventInvokeTempChannelsStructure) error {
	queries := welcomer.GetQueriesFromContext(eventCtx.Context)

	// Fetch guild settings.

	guildSettingsTimeroles, err := p.FetchGuildInformation(eventCtx, queries, *payload.Member.GuildID)
	if err != nil {
		return err
	}

	return p.CreateChannelAndMove(eventCtx, *payload.Member.GuildID, welcomer.ToPointer(discord.Snowflake(guildSettingsTimeroles.ChannelCategory)), payload.Member)
}

func (p *TempChannelsCog) OnInvokeTempChannelsRemoveEvent(eventCtx *sandwich.EventContext, payload welcomer.CustomEventInvokeTempChannelsRemoveStructure) error {
	queries := welcomer.GetQueriesFromContext(eventCtx.Context)

	// Fetch guild settings.

	guildSettingsTimeroles, err := p.FetchGuildInformation(eventCtx, queries, *payload.Member.GuildID)
	if err != nil {
		return err
	}

	channel, err := p.FindChannelForUser(eventCtx, *payload.Interaction.GuildID, welcomer.ToPointer(discord.Snowflake(guildSettingsTimeroles.ChannelCategory)), payload.Member)
	if err != nil {
		return err
	}

	if channel != nil && channel.ID != discord.Snowflake(guildSettingsTimeroles.ChannelLobby) {
		err = channel.Delete(eventCtx.Session, welcomer.ToPointer("Automatically deleted by TempChannels"))
		if err != nil {
			eventCtx.Logger.Error().Err(err).
				Str("guild_id", payload.Member.GuildID.String()).
				Str("channel_id", channel.ID.String()).
				Msg("failed to delete channel")
		}
	}

	return nil
}

func (p *TempChannelsCog) FetchGuildInformation(eventCtx *sandwich.EventContext, queries *database.Queries, guildID discord.Snowflake) (guildSettingsTempChannels *database.GuildSettingsTempchannels, err error) {
	guildSettingsTempChannels, err = queries.GetTempChannelsGuildSettings(eventCtx.Context, int64(guildID))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("failed to get temp channel settings")

		return nil, err
	}

	return guildSettingsTempChannels, nil
}
