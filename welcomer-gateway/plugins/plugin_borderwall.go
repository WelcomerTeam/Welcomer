package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich_daemon "github.com/WelcomerTeam/Sandwich-Daemon"
	sandwich_protobuf "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/savsgio/gotils/strconv"
)

type BorderwallCog struct {
	EventHandler *sandwich.Handlers
}

// Assert types.

var (
	_ sandwich.Cog           = (*BorderwallCog)(nil)
	_ sandwich.CogWithEvents = (*BorderwallCog)(nil)
)

var FallbackBorderwallMessage = "Welcome to {{Guild.Name}}, {{User.Mention}}. This server server is protected by Borderwall, please verify at {{Borderwall.Link}}"

func NewBorderwallCog() *BorderwallCog {
	return &BorderwallCog{
		EventHandler: sandwich.SetupHandler(nil),
	}
}

func (p *BorderwallCog) CogInfo() *sandwich.CogInfo {
	return &sandwich.CogInfo{
		Name:        "Borderwall",
		Description: "Provides the functionality for the 'Borderwall' feature",
	}
}

func (p *BorderwallCog) GetEventHandlers() *sandwich.Handlers {
	return p.EventHandler
}

func (p *BorderwallCog) RegisterCog(bot *sandwich.Bot) error {
	// Register CustomEventInvokeBorderwallCompletion event.
	p.EventHandler.RegisterEventHandler(core.CustomEventInvokeBorderwallCompletion, func(eventCtx *sandwich.EventContext, payload sandwich_daemon.ProducedPayload) error {
		var invokeBorderwallCompletionPayload core.CustomEventInvokeBorderwallCompletionStructure
		if err := eventCtx.DecodeContent(payload, &invokeBorderwallCompletionPayload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		eventCtx.Guild = sandwich.NewGuild(*invokeBorderwallCompletionPayload.Member.GuildID)

		eventCtx.EventHandler.EventsMu.RLock()
		defer eventCtx.EventHandler.EventsMu.RUnlock()

		for _, event := range eventCtx.EventHandler.Events {
			if f, ok := event.(welcomer.OnInvokeBorderwallCompletionFuncType); ok {
				return eventCtx.Handlers.WrapFuncType(eventCtx, f(eventCtx, invokeBorderwallCompletionPayload))
			}
		}

		return nil
	})

	// Trigger OnInvokeBorderwallEvent when ON_GUILD_MEMBER_ADD event is triggered.
	p.EventHandler.RegisterOnGuildMemberAddEvent(func(eventCtx *sandwich.EventContext, member discord.GuildMember) error {
		return p.OnInvokeBorderwallEvent(eventCtx, core.CustomEventInvokeBorderwallStructure{
			Member: member,
		})
	})

	// Call OnInvokeBorderwallCompletionEvent when CustomEventInvokeBorderwallCompletion event is triggered.
	p.EventHandler.RegisterEvent(core.CustomEventInvokeBorderwallCompletion, nil, (welcomer.OnInvokeBorderwallCompletionFuncType)(p.OnInvokeBorderwallCompletionEvent))

	return nil
}

func (p *BorderwallCog) OnInvokeBorderwallEvent(eventCtx *sandwich.EventContext, event core.CustomEventInvokeBorderwallStructure) (err error) {
	guildSettingsBorderwall, err := welcomer.Queries.GetBorderwallGuildSettings(eventCtx.Context, int64(eventCtx.Guild.ID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			guildSettingsBorderwall = &database.GuildSettingsBorderwall{
				GuildID:         int64(eventCtx.Guild.ID),
				ToggleEnabled:   welcomer.DefaultBorderwall.ToggleEnabled,
				ToggleSendDm:    welcomer.DefaultBorderwall.ToggleSendDm,
				Channel:         welcomer.DefaultBorderwall.Channel,
				MessageVerify:   welcomer.DefaultBorderwall.MessageVerify,
				MessageVerified: welcomer.DefaultBorderwall.MessageVerified,
				RolesOnJoin:     welcomer.DefaultBorderwall.RolesOnJoin,
				RolesOnVerify:   welcomer.DefaultBorderwall.RolesOnVerify,
			}
		} else {
			welcomer.Logger.Error().Err(err).
				Int64("guildID", int64(eventCtx.Guild.ID)).
				Msg("Failed to get borderwall settings for guild")

			return err
		}
	}

	// Quit if nothing is enabled.
	if !guildSettingsBorderwall.ToggleEnabled || !(guildSettingsBorderwall.ToggleSendDm || guildSettingsBorderwall.Channel != 0) {
		return nil
	}

	assignableRoles, err := welcomer.FilterAssignableRolesAsSnowflakes(eventCtx.Context, welcomer.SandwichClient, int64(eventCtx.Guild.ID), int64(eventCtx.Identifier.UserId), guildSettingsBorderwall.RolesOnJoin)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Msg("Failed to filter assignable roles")

		return err
	}

	if len(assignableRoles) > 0 {
		member := sandwich.NewGuildMember(&eventCtx.Guild.ID, event.Member.User.ID)

		err = member.AddRoles(eventCtx.Context, eventCtx.Session, assignableRoles, welcomer.ToPointer("Automatically assigned with BorderWall"), true)
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("member_id", int64(event.Member.User.ID)).
				Msg("Failed to add roles to member")
		}
	} else {
		welcomer.Logger.Warn().
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Msg("No roles to assign for borderwall")
	}

	var user *discord.User

	// Query state cache for user if welcomer DMs are enabled.
	// This is for fetching direct message channels for the user.
	if guildSettingsBorderwall.ToggleSendDm {
		// Query state cache for user.
		users, err := welcomer.SandwichClient.FetchUser(eventCtx, &sandwich_protobuf.FetchUserRequest{
			UserIds: []int64{int64(event.Member.User.ID)},
		})
		if err != nil {
			return err
		}

		userPb, ok := users.Users[int64(event.Member.User.ID)]
		if ok {
			user = sandwich_protobuf.PBToUser(userPb)
		} else {
			welcomer.Logger.Warn().
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to get user from state cache, falling back to event.Member.User")

			user = event.Member.User
		}
	}

	// Query state cache for guild.
	guild, err := welcomer.FetchGuild(eventCtx.Context, eventCtx.Guild.ID)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Msg("Failed to fetch guild from state cache")

		return err
	}

	var existingRequestUuid uuid.UUID

	borderwallRequests, err := welcomer.Queries.GetBorderwallRequestsByGuildIDUserID(eventCtx.Context, database.GetBorderwallRequestsByGuildIDUserIDParams{
		GuildID: int64(eventCtx.Guild.ID),
		UserID:  int64(event.Member.User.ID),
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.Member.User.ID)).
			Msg("Failed to check existing borderwall request")

		return err
	}

	for _, request := range borderwallRequests {
		if !request.IsVerified {
			existingRequestUuid = request.RequestUuid

			break
		}
	}

	if existingRequestUuid.IsNil() {
		borderwallRequest, err := welcomer.Queries.CreateBorderwallRequest(eventCtx.Context, database.CreateBorderwallRequestParams{
			GuildID: int64(eventCtx.Guild.ID),
			UserID:  int64(event.Member.User.ID),
		})
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to create borderwall request")

			return err
		}

		existingRequestUuid = borderwallRequest.RequestUuid
	}

	borderwallLink := fmt.Sprintf(
		"%s/borderwall/%s",
		welcomer.WebsiteURL,
		existingRequestUuid.String(),
	)

	functions := welcomer.GatherFunctions()
	variables := welcomer.GatherVariables(eventCtx, &event.Member, guild, nil, map[string]any{
		"Borderwall": BorderwallVariables{
			Link: borderwallLink,
		},
	})

	var serverMessage discord.MessageParams
	var directMessage discord.MessageParams

	// Fallback to default message if the message is empty.
	guildSettingsBorderwall.MessageVerify = welcomer.SetupJSONB(guildSettingsBorderwall.MessageVerify)

	if guildSettingsBorderwall.Channel != 0 || guildSettingsBorderwall.ToggleSendDm {
		messageFormat, err := welcomer.FormatString(functions, variables, strconv.B2S(guildSettingsBorderwall.MessageVerify.Bytes))
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to format borderwall text payload")

			return err
		}

		if guildSettingsBorderwall.Channel != 0 {
			// Convert MessageFormat to MessageParams so we can send it.
			err = json.Unmarshal(strconv.S2B(messageFormat), &serverMessage)
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("user_id", int64(event.Member.User.ID)).
					Msg("Failed to unmarshal borderwall verify messageFormat")

				return err
			}

			// Fallback to default message if the message is empty.
			if welcomer.IsMessageParamsEmpty(serverMessage) {
				serverMessage.Content, _ = welcomer.FormatString(functions, variables, FallbackBorderwallMessage)
			}
		}

		if guildSettingsBorderwall.ToggleSendDm {
			// Convert MessageFormat to MessageParams so we can send it.
			err = json.Unmarshal(strconv.S2B(messageFormat), &directMessage)
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("user_id", int64(event.Member.User.ID)).
					Msg("Failed to unmarshal borderwall verify messageFormat")

				return err
			}

			// Fallback to default message if the message is empty.
			if welcomer.IsMessageParamsEmpty(directMessage) {
				directMessage.Content, _ = welcomer.FormatString(functions, variables, FallbackBorderwallMessage)
			}
		}
	}

	// Send server message if it's not empty.
	if !welcomer.IsMessageParamsEmpty(serverMessage) {
		validGuild, err := core.CheckChannelGuild(eventCtx.Context, welcomer.SandwichClient, eventCtx.Guild.ID, discord.Snowflake(guildSettingsBorderwall.Channel))
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("channel_id", guildSettingsBorderwall.Channel).
				Msg("Failed to check channel guild")
		} else if !validGuild {
			welcomer.Logger.Warn().
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("channel_id", guildSettingsBorderwall.Channel).
				Msg("Channel does not belong to guild")
		} else {
			channel := discord.Channel{ID: discord.Snowflake(guildSettingsBorderwall.Channel)}

			serverMessage = welcomer.IncludeBorderwallVerifyButton(serverMessage, borderwallLink)
			serverMessage = welcomer.IncludeScamsButton(serverMessage)

			_, err = channel.Send(eventCtx.Context, eventCtx.Session, serverMessage)

			welcomer.Logger.Debug().
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("channel_id", guildSettingsBorderwall.Channel).
				Msg("Sent borderwall verify message to channel")

			if err != nil {
				welcomer.Logger.Warn().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("channel_id", guildSettingsBorderwall.Channel).
					Msg("Failed to send borderwall verify message to channel")
			}
		}
	}

	// Send direct message if it's not empty.
	if user != nil && !welcomer.IsMessageParamsEmpty(directMessage) {
		directMessage = welcomer.IncludeBorderwallVerifyButton(directMessage, borderwallLink)
		directMessage = welcomer.IncludeSentByButton(directMessage, guild.Name)
		directMessage = welcomer.IncludeScamsButton(directMessage)

		_, err = user.Send(eventCtx.Context, eventCtx.Session, directMessage)

		welcomer.Logger.Debug().
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.Member.User.ID)).
			Msg("Sent borderwall verify message to user")

		if err != nil {
			welcomer.Logger.Warn().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to send message to user")
		}
	}

	welcomer.PushGuildScience.Push(
		eventCtx.Context,
		eventCtx.Guild.ID,
		event.Member.User.ID,
		database.ScienceGuildEventTypeBorderwallChallenge,
		welcomer.GuildScienceBorderwallChallenge{
			HasMessage: !welcomer.IsMessageParamsEmpty(serverMessage),
			HasDM:      !welcomer.IsMessageParamsEmpty(directMessage),
		})

	return nil
}

func (p *BorderwallCog) OnInvokeBorderwallCompletionEvent(eventCtx *sandwich.EventContext, event core.CustomEventInvokeBorderwallCompletionStructure) (err error) {
	guildSettingsBorderwall, err := welcomer.Queries.GetBorderwallGuildSettings(eventCtx.Context, int64(eventCtx.Guild.ID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			guildSettingsBorderwall = &database.GuildSettingsBorderwall{
				GuildID:         int64(eventCtx.Guild.ID),
				ToggleEnabled:   welcomer.DefaultBorderwall.ToggleEnabled,
				ToggleSendDm:    welcomer.DefaultBorderwall.ToggleSendDm,
				Channel:         welcomer.DefaultBorderwall.Channel,
				MessageVerify:   welcomer.DefaultBorderwall.MessageVerify,
				MessageVerified: welcomer.DefaultBorderwall.MessageVerified,
				RolesOnJoin:     welcomer.DefaultBorderwall.RolesOnJoin,
				RolesOnVerify:   welcomer.DefaultBorderwall.RolesOnVerify,
			}
		} else {
			welcomer.Logger.Error().Err(err).
				Int64("guildID", int64(eventCtx.Guild.ID)).
				Msg("Failed to get borderwall settings for guild")

			return err
		}
	}

	// Quit if nothing is enabled.
	if !guildSettingsBorderwall.ToggleEnabled || (!guildSettingsBorderwall.ToggleSendDm && guildSettingsBorderwall.Channel == 0) {
		return nil
	}

	var member discord.GuildMember

	// Chunk guild
	_, err = welcomer.SandwichClient.RequestGuildChunk(eventCtx, &sandwich_protobuf.RequestGuildChunkRequest{
		GuildId:     int64(eventCtx.Guild.ID),
		AlwaysChunk: false,
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Msg("Failed to chunk guild")

		return err
	}

	// Fetch guild member.
	guildMembers, err := welcomer.SandwichClient.FetchGuildMember(eventCtx, &sandwich_protobuf.FetchGuildMemberRequest{
		GuildId: int64(eventCtx.Guild.ID),
		UserIds: []int64{int64(event.Member.User.ID)},
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Msg("Failed to fetch guild members")

		return err
	}

	memberPb, ok := guildMembers.GuildMembers[int64(event.Member.User.ID)]
	if ok {
		member = *sandwich_protobuf.PBToGuildMember(memberPb)
	} else {
		welcomer.Logger.Warn().
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.Member.User.ID)).
			Msg("Failed to get member from state cache, falling back to event.Member")

		discordMember, err := discord.GetGuildMember(eventCtx.Context, eventCtx.Session, eventCtx.Guild.ID, event.Member.User.ID)
		if err != nil || discordMember == nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to get member from discord")

			return err
		}

		member = *discordMember
	}

	member.GuildID = &eventCtx.Guild.ID

	assignableJoinRoles, err := welcomer.FilterAssignableRolesAsSnowflakes(eventCtx.Context, welcomer.SandwichClient, int64(eventCtx.Guild.ID), int64(eventCtx.Identifier.UserId), guildSettingsBorderwall.RolesOnJoin)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Msg("Failed to filter assignable roles")

		return err
	}

	assignableVerifyRoles, err := welcomer.FilterAssignableRolesAsSnowflakes(eventCtx.Context, welcomer.SandwichClient, int64(eventCtx.Guild.ID), int64(eventCtx.Identifier.UserId), guildSettingsBorderwall.RolesOnVerify)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Msg("Failed to filter assignable roles")

		return err
	}

	// Do not remove roles if we are about to re-assign them, and the user already has them.
	// This is to prevent the bot from removing and re-adding the same roles.
	for i := 0; i < len(assignableJoinRoles); i++ {
		for j := 0; j < len(assignableVerifyRoles); j++ {
			if assignableJoinRoles[i] == assignableVerifyRoles[j] {
				// If the role is present in both the add and remove list, remove
				// it from both lists if the user has the role already.
				if len(member.Roles) > 0 {
					if slices.Contains(member.Roles, assignableJoinRoles[i]) {
						assignableJoinRoles = append(assignableJoinRoles[:i], assignableJoinRoles[i+1:]...)
						assignableVerifyRoles = append(assignableVerifyRoles[:j], assignableVerifyRoles[j+1:]...)
						i--

						break
					}
				}
			}
		}
	}

	if len(assignableJoinRoles) > 0 {
		err = member.RemoveRoles(eventCtx.Context, eventCtx.Session, assignableJoinRoles, welcomer.ToPointer("Automatically removed with BorderWall"), true)
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("member_id", int64(event.Member.User.ID)).
				Msg("Failed to remove roles from member for borderwall completion")
		}
	} else {
		welcomer.Logger.Warn().
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Msg("No roles to assign for borderwall completion")
	}

	if len(assignableVerifyRoles) > 0 {
		err = member.AddRoles(eventCtx.Context, eventCtx.Session, assignableVerifyRoles, welcomer.ToPointer("Automatically assigned with BorderWall"), true)
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("member_id", int64(event.Member.User.ID)).
				Msg("Failed to add roles to member for borderwall completion")
		}
	} else {
		welcomer.Logger.Warn().
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Msg("No roles to assign for borderwall completion")
	}

	var user *discord.User

	// Query state cache for user if welcomer DMs are enabled.
	// This is for fetching direct message channels for the user.
	if guildSettingsBorderwall.ToggleSendDm {
		// Query state cache for user.
		users, err := welcomer.SandwichClient.FetchUser(eventCtx, &sandwich_protobuf.FetchUserRequest{
			UserIds: []int64{int64(event.Member.User.ID)},
		})
		if err != nil {
			return err
		}

		userPb, ok := users.Users[int64(event.Member.User.ID)]
		if ok {
			user = sandwich_protobuf.PBToUser(userPb)
		} else {
			welcomer.Logger.Warn().
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to get user from state cache, falling back to event.Member.User")

			user = event.Member.User
		}
	} else {
		user = event.Member.User
	}

	// Query state cache for guild.
	guild, err := welcomer.FetchGuild(eventCtx.Context, eventCtx.Guild.ID)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Msg("Failed to fetch guild from state cache")

		guild = eventCtx.Guild
	}

	functions := welcomer.GatherFunctions()
	variables := welcomer.GatherVariables(eventCtx, &event.Member, guild, nil, nil)

	var serverMessage discord.MessageParams
	var directMessage discord.MessageParams

	if !welcomer.IsJSONBEmpty(guildSettingsBorderwall.MessageVerified.Bytes) {
		messageFormat, err := welcomer.FormatString(functions, variables, strconv.B2S(guildSettingsBorderwall.MessageVerified.Bytes))
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to format borderwall text payload")

			return err
		}

		if guildSettingsBorderwall.Channel != 0 {
			// Convert MessageFormat to MessageParams so we can send it.
			err = json.Unmarshal(strconv.S2B(messageFormat), &serverMessage)
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("user_id", int64(event.Member.User.ID)).
					Msg("Failed to unmarshal borderwall verified messageFormat")

				return err
			}
		}

		if guildSettingsBorderwall.ToggleSendDm {
			// Convert MessageFormat to MessageParams so we can send it.
			err = json.Unmarshal(strconv.S2B(messageFormat), &directMessage)
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("user_id", int64(event.Member.User.ID)).
					Msg("Failed to unmarshal borderwall verified messageFormat")

				return err
			}
		}
	}

	// Send server message if it's not empty.
	if !welcomer.IsMessageParamsEmpty(serverMessage) {
		validGuild, err := core.CheckChannelGuild(eventCtx.Context, welcomer.SandwichClient, eventCtx.Guild.ID, discord.Snowflake(guildSettingsBorderwall.Channel))
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("channel_id", guildSettingsBorderwall.Channel).
				Msg("Failed to check channel guild")
		} else if !validGuild {
			welcomer.Logger.Warn().
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("channel_id", guildSettingsBorderwall.Channel).
				Msg("Channel does not belong to guild")
		} else {
			channel := discord.Channel{ID: discord.Snowflake(guildSettingsBorderwall.Channel)}

			_, err = channel.Send(eventCtx.Context, eventCtx.Session, serverMessage)

			welcomer.Logger.Debug().
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("channel_id", guildSettingsBorderwall.Channel).
				Msg("Sent borderwall verified message to channel")

			if err != nil {
				welcomer.Logger.Warn().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("channel_id", guildSettingsBorderwall.Channel).
					Msg("Failed to send borderwall verified message to channel")
			}
		}
	}

	// Send direct message if it's not empty.
	if !welcomer.IsMessageParamsEmpty(directMessage) {
		directMessage = welcomer.IncludeSentByButton(directMessage, guild.Name)
		directMessage = welcomer.IncludeScamsButton(directMessage)

		_, err = user.Send(eventCtx.Context, eventCtx.Session, directMessage)

		welcomer.Logger.Debug().
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.Member.User.ID)).
			Msg("Sent borderwall verified message to user")

		if err != nil {
			welcomer.Logger.Warn().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to send message to user")
		}
	}

	welcomer.PushGuildScience.Push(
		eventCtx.Context,
		eventCtx.Guild.ID,
		event.Member.User.ID,
		database.ScienceGuildEventTypeBorderwallCompleted,
		welcomer.GuildScienceBorderwallCompleted{
			HasMessage: !welcomer.IsMessageParamsEmpty(serverMessage),
			HasDM:      !welcomer.IsMessageParamsEmpty(directMessage),
		})

	return nil
}

type BorderwallVariables struct {
	Link string
}
