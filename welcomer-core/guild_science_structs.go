package welcomer

import (
	"github.com/WelcomerTeam/Discord/discord"
	"github.com/gofrs/uuid"
)

type GuildScienceUserJoin struct {
	UserID discord.Snowflake `json:"user_id"`
}

type GuildScienceUserLeave struct {
	UserID discord.Snowflake `json:"user_id"`
}

type GuildScienceUserWelcomed struct {
	UserID            discord.Snowflake `json:"user_id"`
	HasImage          bool              `json:"has_image"`
	HasMessage        bool              `json:"has_message"`
	HasDM             bool              `json:"has_dm"`
	HasInviteTracking bool              `json:"has_invite_tracking"`
	IsInviteTracked   bool              `json:"is_invite_tracked"`
}

type GuildScienceTimeRoleGiven struct {
	UserID discord.Snowflake `json:"user_id"`
	RoleID discord.Snowflake `json:"role_id"`
}

type GuildScienceBorderwallChallenge struct {
	UserID     discord.Snowflake `json:"user_id"`
	HasMessage bool              `json:"has_message"`
	HasDM      bool              `json:"has_dm"`
}

type GuildScienceBorderwallCompleted struct {
	UserID     discord.Snowflake `json:"user_id"`
	HasMessage bool              `json:"has_message"`
	HasDM      bool              `json:"has_dm"`
}

type GuildScienceTempChannelCreated struct {
	UserID discord.Snowflake `json:"user_id"`
}

type GuildScienceMembershipReceived struct {
	UserID         discord.Snowflake `json:"user_id"`
	MembershipUUID uuid.UUID         `json:"membership_uuid"`
}

type GuildScienceMembershipRemoved struct {
	UserID         discord.Snowflake `json:"user_id"`
	MembershipUUID uuid.UUID         `json:"membership_uuid"`
}
