package plugins

import (
	"errors"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	pb "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4"
)

type TimeRolesCog struct {
	EventHandler *sandwich.Handlers
}

// Assert Types

var (
	_ sandwich.Cog           = (*TimeRolesCog)(nil)
	_ sandwich.CogWithEvents = (*TimeRolesCog)(nil)
)

func NewTimeRolesCog() *TimeRolesCog {
	return &TimeRolesCog{
		EventHandler: sandwich.SetupHandler(nil),
	}
}

func (p *TimeRolesCog) CogInfo() *sandwich.CogInfo {
	return &sandwich.CogInfo{
		Name:        "TimeRoles",
		Description: "Provides the functionality for the 'TimeRoles' feature",
	}
}

func (p *TimeRolesCog) GetEventHandlers() *sandwich.Handlers {
	return p.EventHandler
}

func (p *TimeRolesCog) RegisterCog(bot *sandwich.Bot) error {

	// Trigger OnInvokeTimeRoles when ON_MESSAGE_CREATE event is received.
	p.EventHandler.RegisterOnMessageCreateEvent(func(eventCtx *sandwich.EventContext, message discord.Message) error {
		if message.Author.Bot {
			return nil
		}

		return p.OnInvokeTimeRoles(eventCtx, *message.GuildID, &message.Author.ID)
	})

	return nil
}

func (p *TimeRolesCog) OnInvokeTimeRoles(eventCtx *sandwich.EventContext, guildID discord.Snowflake, memberID *discord.Snowflake) (err error) {
	queries := welcomer.GetQueriesFromContext(eventCtx.Context)

	// Fetch guild settings.

	guildSettingsTimeRoles, timeRoles, err := p.FetchGuildInformation(eventCtx, queries, guildID)
	if err != nil {
		return err
	}

	if !guildSettingsTimeRoles.ToggleEnabled || len(timeRoles) == 0 {
		return nil
	}

	// Chunk users
	_, err = eventCtx.Sandwich.SandwichClient.RequestGuildChunk(eventCtx.Context, &pb.RequestGuildChunkRequest{
		GuildId: int64(guildID),
	})
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("failed to request guild chunk")
	}

	guildMembers, err := eventCtx.Sandwich.SandwichClient.FetchGuildMembers(eventCtx.Context, &pb.FetchGuildMembersRequest{
		GuildID: int64(guildID),
		UserIDs: welcomer.If(memberID != nil, []int64{int64(*memberID)}, nil),
	})
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("failed to fetch guild members")
	}

	allRolesToAssign := make(map[discord.Snowflake][]discord.Snowflake)
	rolesToAssign := make([]discord.Snowflake, 0, len(timeRoles))

	for _, guildMember := range guildMembers.GuildMembers {
		clear(rolesToAssign)

		joinedAtTime, err := time.Parse(time.RFC3339, guildMember.JoinedAt)
		if err != nil {
			eventCtx.Logger.Error().Err(err).
				Str("joined_at", guildMember.JoinedAt).
				Msg("failed to parse joined_at time for user")

			continue
		}

		secondsSinceJoining := int(time.Since(joinedAtTime).Seconds())

		for _, timeRole := range timeRoles {
			if timeRole.Seconds <= secondsSinceJoining {
				hasRole := false

				for _, role := range guildMember.Roles {
					if role == int64(timeRole.Role) {
						hasRole = true

						break
					}
				}

				if !hasRole {
					rolesToAssign = append(rolesToAssign, timeRole.Role)
				}
			}
		}

		if len(rolesToAssign) > 0 {
			for _, role := range guildMember.Roles {
				rolesToAssign = append(rolesToAssign, discord.Snowflake(role))
			}

			allRolesToAssign[discord.Snowflake(guildMember.User.ID)] = rolesToAssign
		}
	}

	eventCtx.Logger.Info().
		Int("guild_id", int(guildID)).
		Int("members", len(allRolesToAssign)).
		Msg("Assigning time roles for members")

	for memberID, roles := range allRolesToAssign {
		var member *discord.GuildMember

		memberPb, ok := guildMembers.GuildMembers[int64(memberID)]
		if ok {
			member, err = pb.GRPCToGuildMember(memberPb)
			if err != nil {
				return err
			}
		} else {
			eventCtx.Logger.Warn().
				Int64("guild_id", int64(guildID)).
				Int64("user_id", int64(memberID)).
				Msg("failed to get member from state cache, falling back to default member")

			member = &discord.GuildMember{
				User: &discord.User{
					ID: discord.Snowflake(memberID),
				},
			}
		}

		member.GuildID = &guildID

		err = member.AddRoles(eventCtx.Session, roles, welcomer.ToPointer("Automatically assigned with TimeRoles"), true)
		if err != nil {
			eventCtx.Logger.Error().Err(err).
				Int64("guild_id", int64(guildID)).
				Int64("member_id", int64(memberID)).
				Interface("roles", roles).
				Msg("failed to add roles to member")
		}
	}

	return nil
}

func (p *TimeRolesCog) FetchGuildInformation(eventCtx *sandwich.EventContext, queries *database.Queries, guildID discord.Snowflake) (guildSettingsTimeRoles *database.GuildSettingsTimeroles, assignableTimeRoles []welcomer.GuildSettingsTimeRolesRole, err error) {
	// TODO: Add caching

	guildSettingsTimeRoles, err = queries.GetTimeRolesGuildSettings(eventCtx.Context, int64(guildID))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("failed to get time role settings")

		return nil, nil, err
	}

	timeRoles := welcomer.UnmarshalTimeRolesJSON(guildSettingsTimeRoles.Timeroles.Bytes)
	if len(timeRoles) == 0 {
		return guildSettingsTimeRoles, nil, nil
	}

	assignableTimeRoles, err = welcomer.FilterAssignableTimeRoles(eventCtx.Context, eventCtx.Sandwich.SandwichClient, eventCtx.Logger, int64(guildID), int64(eventCtx.Identifier.ID), timeRoles)
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("failed to filter assignable time roles")

		return nil, nil, err
	}

	return guildSettingsTimeRoles, assignableTimeRoles, nil
}
