package plugins

import (
	"errors"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	pb "github.com/WelcomerTeam/Sandwich-Daemon/proto"
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
		if message.Author.Bot || message.GuildID == nil {
			return nil
		}

		return p.OnInvokeTimeRoles(eventCtx, *message.GuildID, &message.Author.ID)
	})

	return nil
}

func (p *TimeRolesCog) OnInvokeTimeRoles(eventCtx *sandwich.EventContext, guildID discord.Snowflake, memberID *discord.Snowflake) (err error) {
	// Fetch guild settings.

	guildSettingsTimeRoles, timeRoles, err := p.FetchGuildInformation(eventCtx, guildID)
	if err != nil {
		return err
	}

	if !guildSettingsTimeRoles.ToggleEnabled || len(timeRoles) == 0 {
		return nil
	}

	// Chunk users
	_, err = welcomer.SandwichClient.RequestGuildChunk(eventCtx.Context, &pb.RequestGuildChunkRequest{
		GuildId: int64(guildID),
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("Failed to request guild chunk")
	}

	guildMembers, err := welcomer.SandwichClient.FetchGuildMember(eventCtx.Context, &pb.FetchGuildMemberRequest{
		GuildId: int64(guildID),
		UserIds: welcomer.If(memberID != nil, []int64{int64(*memberID)}, nil),
	})
	if err != nil || guildMembers == nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("Failed to fetch guild members")

		return err
	}

	allRolesToAssign := make(map[discord.Snowflake][]discord.Snowflake)
	rolesToAssign := make([]discord.Snowflake, 0, len(timeRoles))

	for _, guildMember := range guildMembers.GuildMembers {
		clear(rolesToAssign)

		joinedAtTime, err := time.Parse(time.RFC3339, guildMember.JoinedAt)
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Str("joined_at", guildMember.JoinedAt).
				Msg("Failed to parse joined_at time for user")

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

	welcomer.Logger.Info().
		Int("guild_id", int(guildID)).
		Int("members", len(allRolesToAssign)).
		Msg("Assigning time roles for members")

	for memberID, roles := range allRolesToAssign {
		var member *discord.GuildMember

		memberPb, ok := guildMembers.GuildMembers[int64(memberID)]
		if ok {
			member = pb.PBToGuildMember(memberPb)
		} else {
			welcomer.Logger.Warn().
				Int64("guild_id", int64(guildID)).
				Int64("user_id", int64(memberID)).
				Msg("Failed to get member from state cache, falling back to default member")

			member = &discord.GuildMember{
				User: &discord.User{
					ID: discord.Snowflake(memberID),
				},
			}
		}

		member.GuildID = &guildID

		err = member.AddRoles(eventCtx.Context, eventCtx.Session, roles, welcomer.ToPointer("Automatically assigned with TimeRoles"), true)
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(guildID)).
				Int64("member_id", int64(memberID)).
				Interface("roles", roles).
				Msg("Failed to add roles to member for timeroles")
		}

		for _, role := range roles {
			welcomer.PushGuildScience.Push(
				eventCtx.Context,
				eventCtx.Guild.ID,
				member.User.ID,
				database.ScienceGuildEventTypeTimeRoleGiven,
				welcomer.GuildScienceTimeRoleGiven{
					RoleID: role,
				})
		}
	}

	return nil
}

func (p *TimeRolesCog) FetchGuildInformation(eventCtx *sandwich.EventContext, guildID discord.Snowflake) (guildSettingsTimeRoles *database.GuildSettingsTimeroles, assignableTimeRoles []welcomer.GuildSettingsTimeRolesRole, err error) {
	// TODO: Add caching

	guildSettingsTimeRoles, err = welcomer.Queries.GetTimeRolesGuildSettings(eventCtx.Context, int64(guildID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			guildSettingsTimeRoles = &database.GuildSettingsTimeroles{
				GuildID:       int64(guildID),
				ToggleEnabled: welcomer.DefaultTimeRoles.ToggleEnabled,
				Timeroles:     welcomer.DefaultTimeRoles.Timeroles,
			}
		} else {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(guildID)).
				Msg("Failed to get time role settings")

			return nil, nil, err
		}
	}

	timeRoles := welcomer.UnmarshalTimeRolesJSON(guildSettingsTimeRoles.Timeroles.Bytes)
	if len(timeRoles) == 0 {
		return guildSettingsTimeRoles, nil, nil
	}

	assignableTimeRoles, err = welcomer.FilterAssignableTimeRoles(eventCtx.Context, welcomer.SandwichClient, int64(guildID), int64(eventCtx.Identifier.UserId), timeRoles)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("Failed to filter assignable timeroles")

		return nil, nil, err
	}

	return guildSettingsTimeRoles, assignableTimeRoles, nil
}
