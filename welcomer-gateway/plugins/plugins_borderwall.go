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
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	jsoniter "github.com/json-iterator/go"
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
	p.EventHandler.RegisterEventHandler(welcomer.CustomEventInvokeBorderwallCompletion, func(eventCtx *sandwich.EventContext, payload structs.SandwichPayload) error {
		var invokeBorderwallCompletionPayload welcomer.CustomEventInvokeBorderwallCompletionStructure
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
		return p.OnInvokeBorderwallEvent(eventCtx, welcomer.CustomEventInvokeBorderwallStructure{
			Member: member,
		})
	})

	// Call OnInvokeBorderwallCompletionEvent when CustomEventInvokeBorderwallCompletion event is triggered.
	p.EventHandler.RegisterEvent(welcomer.CustomEventInvokeBorderwallCompletion, nil, (welcomer.OnInvokeBorderwallCompletionFuncType)(p.OnInvokeBorderwallCompletionEvent))

	return nil
}

func (p *BorderwallCog) OnInvokeBorderwallEvent(eventCtx *sandwich.EventContext, event welcomer.CustomEventInvokeBorderwallStructure) (err error) {

	queries := welcomer.GetQueriesFromContext(eventCtx.Context)

	guildSettingsBorderwall, err := queries.GetBorderwallGuildSettings(eventCtx.Context, int64(eventCtx.Guild.ID))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		eventCtx.Logger.Error().Err(err).
			Int64("guildID", int64(eventCtx.Guild.ID)).
			Msg("Failed to get borderwall settings for guild")

		return err
	}

	// Quit if nothing is enabled.
	if !guildSettingsBorderwall.ToggleEnabled && !(guildSettingsBorderwall.ToggleSendDm || guildSettingsBorderwall.Channel != 0) {
		return nil
	}

	assignableRoles, err := welcomer.FilterAssignableRoles(eventCtx.Context, eventCtx.Sandwich.SandwichClient, eventCtx.Logger, int64(eventCtx.Guild.ID), int64(eventCtx.Identifier.ID), guildSettingsBorderwall.RolesOnJoin)
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Msg("Failed to filter assignable roles")

		return err
	}

	if len(assignableRoles) > 0 {
		member := sandwich.NewGuildMember(&eventCtx.Guild.ID, event.Member.User.ID)

		err = member.AddRoles(eventCtx.Session, assignableRoles, welcomer.ToPointer("Automatically assigned with BorderWall"), true)
		if err != nil {
			eventCtx.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("member_id", int64(event.Member.User.ID)).
				Msg("Failed to add roles to member")
		}
	} else {
		eventCtx.Logger.Warn().
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Msg("No roles to assign for borderwall")
	}

	var user *discord.User

	// Query state cache for user if welcomer DMs are enabled.
	// This is for fetching direct message channels for the user.
	if guildSettingsBorderwall.ToggleSendDm {
		// Query state cache for user.
		users, err := eventCtx.Sandwich.SandwichClient.FetchUsers(eventCtx, &pb.FetchUsersRequest{
			UserIDs:         []int64{int64(event.Member.User.ID)},
			CreateDMChannel: true,
		})
		if err != nil {
			return err
		}

		userPb, ok := users.Users[int64(event.Member.User.ID)]
		if ok {
			user, err = pb.GRPCToUser(userPb)
			if err != nil {
				return err
			}
		} else {
			eventCtx.Logger.Warn().
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to get user from state cache, falling back to event.Member.User")

			user = event.Member.User
		}
	} else {
		user = event.Member.User
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

	var existingRequestUuid uuid.UUID

	borderwallRequests, err := queries.GetBorderwallRequestsByGuildIDUserID(eventCtx.Context, database.GetBorderwallRequestsByGuildIDUserIDParams{
		GuildID: int64(eventCtx.Guild.ID),
		UserID:  int64(event.Member.User.ID),
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.Member.User.ID)).
			Msg("Failed to check existing borderwall request")
	}

	for _, request := range borderwallRequests {
		if !request.IsVerified {
			existingRequestUuid = request.RequestUuid

			break
		}
	}

	if existingRequestUuid.IsNil() {
		borderwallRequest, err := queries.CreateBorderwallRequest(eventCtx.Context, database.CreateBorderwallRequestParams{
			GuildID: int64(eventCtx.Guild.ID),
			UserID:  int64(event.Member.User.ID),
		})
		if err != nil {
			eventCtx.Logger.Error().Err(err).
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
	variables := welcomer.GatherVariables(eventCtx, event.Member, *guild, map[string]interface{}{
		"Borderwall": BorderwallVariables{
			Link: borderwallLink,
		},
	})

	var serverMessage *discord.MessageParams

	var directMessage *discord.MessageParams

	// Fallback to default message if the message is empty.
	guildSettingsBorderwall.MessageVerify = welcomer.SetupJSONB(guildSettingsBorderwall.MessageVerify)


	if guildSettingsBorderwall.Channel != 0 || guildSettingsBorderwall.ToggleSendDm {
		messageFormat, err := welcomer.FormatString(functions, variables, strconv.B2S(guildSettingsBorderwall.MessageVerify.Bytes))
		if err != nil {
			eventCtx.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to format borderwall text payload")

			return err
		}

		if guildSettingsBorderwall.Channel != 0 {
			// Convert MessageFormat to MessageParams so we can send it.
			err = jsoniter.UnmarshalFromString(messageFormat, &serverMessage)
			if err != nil {
				eventCtx.Logger.Error().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("user_id", int64(event.Member.User.ID)).
					Msg("Failed to unmarshal borderwall verify messageFormat")

				return err
			}

			// Fallback to default message if the message is empty.
			if welcomer.IsMessageParamsEmpty(*serverMessage) {
				serverMessage.Content, _ = welcomer.FormatString(functions, variables, FallbackBorderwallMessage)
			}
		}

		if guildSettingsBorderwall.ToggleSendDm {
			// Convert MessageFormat to MessageParams so we can send it.
			err = jsoniter.UnmarshalFromString(messageFormat, &directMessage)
			if err != nil {
				eventCtx.Logger.Error().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("user_id", int64(event.Member.User.ID)).
					Msg("Failed to unmarshal borderwall verify messageFormat")

				return err
			}

			// Fallback to default message if the message is empty.
			if welcomer.IsMessageParamsEmpty(*directMessage) {
				directMessage.Content, _ = welcomer.FormatString(functions, variables, FallbackBorderwallMessage)
			}
		}
	}

	// Send server message if it's not empty.
	if serverMessage != nil {
		channel := discord.Channel{ID: discord.Snowflake(guildSettingsBorderwall.Channel)}

		serverMessage = welcomer.IncludeBorderwallVerifyButton(serverMessage, borderwallLink)
		serverMessage = welcomer.IncludeScamsButton(serverMessage)

		_, err = channel.Send(eventCtx.Session, *serverMessage)
		if err != nil {
			eventCtx.Logger.Warn().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("channel_id", guildSettingsBorderwall.Channel).
				Msg("Failed to send borderwall verify message to channel")
		}
	}

	// Send direct message if it's not empty.
	if directMessage != nil {
		directMessage = welcomer.IncludeBorderwallVerifyButton(directMessage, borderwallLink)
		directMessage = welcomer.IncludeSentByButton(directMessage, guild.Name)
		directMessage = welcomer.IncludeScamsButton(directMessage)

		_, err = user.Send(eventCtx.Session, *directMessage)
		if err != nil {
			eventCtx.Logger.Warn().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to send message to user")
		}
	}

	return nil
}

func (p *BorderwallCog) OnInvokeBorderwallCompletionEvent(eventCtx *sandwich.EventContext, event welcomer.CustomEventInvokeBorderwallCompletionStructure) (err error) {

	queries := welcomer.GetQueriesFromContext(eventCtx.Context)

	guildSettingsBorderwall, err := queries.GetBorderwallGuildSettings(eventCtx.Context, int64(eventCtx.Guild.ID))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		eventCtx.Logger.Error().Err(err).
			Int64("guildID", int64(eventCtx.Guild.ID)).
			Msg("Failed to get borderwall settings for guild")

		return err
	}

	// Quit if nothing is enabled.
	if !guildSettingsBorderwall.ToggleEnabled && !(guildSettingsBorderwall.ToggleSendDm || guildSettingsBorderwall.Channel != 0) {
		return nil
	}

	var member *discord.GuildMember

	// Fetch guild member.
	guildMembers, err := eventCtx.Sandwich.SandwichClient.FetchGuildMembers(eventCtx, &pb.FetchGuildMembersRequest{
		GuildID:    int64(eventCtx.Guild.ID),
		UserIDs:    []int64{int64(event.Member.User.ID)},
		ChunkGuild: true,
	})
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Msg("Failed to fetch guild members")

		return err
	}

	memberPb, ok := guildMembers.GuildMembers[int64(event.Member.User.ID)]
	if ok {
		member, err = pb.GRPCToGuildMember(memberPb)
		if err != nil {
			return err
		}
	} else {
		eventCtx.Logger.Warn().
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.Member.User.ID)).
			Msg("Failed to get member from state cache, falling back to event.Member")

		member, err = discord.GetGuildMember(eventCtx.Session, eventCtx.Guild.ID, event.Member.User.ID)
		if err != nil {
			eventCtx.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to get member from discord")

			return err
		}
	}

	member.GuildID = &eventCtx.Guild.ID

	assignableJoinRoles, err := welcomer.FilterAssignableRoles(eventCtx.Context, eventCtx.Sandwich.SandwichClient, eventCtx.Logger, int64(eventCtx.Guild.ID), int64(eventCtx.Identifier.ID), guildSettingsBorderwall.RolesOnJoin)
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Msg("Failed to filter assignable roles")

		return err
	}

	assignableVerifyRoles, err := welcomer.FilterAssignableRoles(eventCtx.Context, eventCtx.Sandwich.SandwichClient, eventCtx.Logger, int64(eventCtx.Guild.ID), int64(eventCtx.Identifier.ID), guildSettingsBorderwall.RolesOnVerify)
	if err != nil {
		eventCtx.Logger.Error().Err(err).
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
					rolePresent := false

					for _, role := range member.Roles {
						if role == assignableJoinRoles[i] {
							rolePresent = true

							break
						}
					}

					if rolePresent {
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
		err = member.RemoveRoles(eventCtx.Session, assignableJoinRoles, welcomer.ToPointer("Automatically removed with BorderWall"), true)
		if err != nil {
			eventCtx.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("member_id", int64(event.Member.User.ID)).
				Msg("Failed to remove roles from member for borderwall completion")
		}
	} else {
		eventCtx.Logger.Warn().
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Msg("No roles to assign for borderwall completion")
	}

	if len(assignableVerifyRoles) > 0 {
		err = member.AddRoles(eventCtx.Session, assignableVerifyRoles, welcomer.ToPointer("Automatically assigned with BorderWall"), true)
		if err != nil {
			eventCtx.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("member_id", int64(event.Member.User.ID)).
				Msg("Failed to add roles to member for borderwall completion")
		}
	} else {
		eventCtx.Logger.Warn().
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Msg("No roles to assign for borderwall completion")
	}

	var user *discord.User

	// Query state cache for user if welcomer DMs are enabled.
	// This is for fetching direct message channels for the user.
	if guildSettingsBorderwall.ToggleSendDm {
		// Query state cache for user.
		users, err := eventCtx.Sandwich.SandwichClient.FetchUsers(eventCtx, &pb.FetchUsersRequest{
			UserIDs:         []int64{int64(event.Member.User.ID)},
			CreateDMChannel: true,
		})
		if err != nil {
			return err
		}

		userPb, ok := users.Users[int64(event.Member.User.ID)]
		if ok {
			user, err = pb.GRPCToUser(userPb)
			if err != nil {
				return err
			}
		} else {
			eventCtx.Logger.Warn().
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to get user from state cache, falling back to event.Member.User")

			user = event.Member.User
		}
	} else {
		user = event.Member.User
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
	variables := welcomer.GatherVariables(eventCtx, event.Member, *guild, nil)

	var serverMessage *discord.MessageParams

	var directMessage *discord.MessageParams

	if !welcomer.IsJSONBEmpty(guildSettingsBorderwall.MessageVerified.Bytes) {
		messageFormat, err := welcomer.FormatString(functions, variables, strconv.B2S(guildSettingsBorderwall.MessageVerified.Bytes))
		if err != nil {
			eventCtx.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to format borderwall text payload")

			return err
		}

		if guildSettingsBorderwall.Channel != 0 {
			// Convert MessageFormat to MessageParams so we can send it.
			err = jsoniter.UnmarshalFromString(messageFormat, &serverMessage)
			if err != nil {
				eventCtx.Logger.Error().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("user_id", int64(event.Member.User.ID)).
					Msg("Failed to unmarshal borderwall verified messageFormat")

				return err
			}
		}

		if guildSettingsBorderwall.ToggleSendDm {
			// Convert MessageFormat to MessageParams so we can send it.
			err = jsoniter.UnmarshalFromString(messageFormat, &directMessage)
			if err != nil {
				eventCtx.Logger.Error().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("user_id", int64(event.Member.User.ID)).
					Msg("Failed to unmarshal borderwall verified messageFormat")

				return err
			}
		}
	}

	// Send server message if it's not empty.
	if serverMessage != nil && !welcomer.IsMessageParamsEmpty(*serverMessage) {
		channel := discord.Channel{ID: discord.Snowflake(guildSettingsBorderwall.Channel)}

		_, err = channel.Send(eventCtx.Session, *serverMessage)
		if err != nil {
			eventCtx.Logger.Warn().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("channel_id", guildSettingsBorderwall.Channel).
				Msg("Failed to send borderwall verified message to channel")
		}
	}

	// Send direct message if it's not empty.
	if directMessage != nil {
		directMessage = welcomer.IncludeSentByButton(directMessage, guild.Name)
		directMessage = welcomer.IncludeScamsButton(directMessage)

		_, err = user.Send(eventCtx.Session, *directMessage)
		if err != nil {
			eventCtx.Logger.Warn().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to send message to user")
		}
	}

	return nil
}

type BorderwallVariables struct {
	Link string
}
