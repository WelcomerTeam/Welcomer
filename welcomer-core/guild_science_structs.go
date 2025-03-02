package welcomer

import (
	"github.com/WelcomerTeam/Discord/discord"
	"github.com/gofrs/uuid"
)

type GuildScienceUserWelcomed struct {
	HasImage          bool `json:"has_image"`
	HasMessage        bool `json:"has_message"`
	HasDM             bool `json:"has_dm"`
	HasInviteTracking bool `json:"has_invite_tracking"`
	IsInviteTracked   bool `json:"is_invite_tracked"`
}

type GuildScienceTimeRoleGiven struct {
	RoleID discord.Snowflake `json:"role_id"`
}

type GuildScienceBorderwallChallenge struct {
	HasMessage bool `json:"has_message"`
	HasDM      bool `json:"has_dm"`
}

type GuildScienceBorderwallCompleted struct {
	HasMessage bool `json:"has_message"`
	HasDM      bool `json:"has_dm"`
}

type GuildScienceMembershipReceived struct {
	MembershipUUID uuid.UUID `json:"membership_uuid"`
}

type GuildScienceMembershipRemoved struct {
	MembershipUUID uuid.UUID `json:"membership_uuid"`
}
