package welcomer

import (
	"github.com/WelcomerTeam/Discord/discord"
	"github.com/gofrs/uuid"
)

type GuildScienceUserJoined struct {
	HasInviteTracking bool   `json:"has_invite_tracking,omitempty"`
	IsInviteTracked   bool   `json:"is_invite_tracked,omitempty"`
	InviteCode        string `json:"invite_code,omitempty"`
	MemberCount       int32  `json:"member_count,omitempty"`
	IsPending         bool   `json:"is_pending,omitempty"`
}

type GuildScienceUserWelcomed struct {
	HasImage bool `json:"has_image,omitempty"`

	HasMessage       bool              `json:"has_message,omitempty"`
	MessageID        discord.Snowflake `json:"message_id,omitempty"`
	MessageChannelID discord.Snowflake `json:"channel_id,omitempty"`

	HasDM             bool   `json:"has_dm,omitempty"`
	HasInviteTracking bool   `json:"has_invite_tracking,omitempty"`
	IsInviteTracked   bool   `json:"is_invite_tracked,omitempty"`
	InviteCode        string `json:"invite_code,omitempty"`
}

type GuildScienceUserLeftMessage struct {
	HasMessage       bool              `json:"has_message,omitempty"`
	MessageID        discord.Snowflake `json:"message_id,omitempty"`
	MessageChannelID discord.Snowflake `json:"channel_id,omitempty"`
}

type GuildScienceTimeRoleGiven struct {
	RoleID discord.Snowflake `json:"role_id"`
}

type GuildScienceBorderwallChallenge struct {
	HasMessage bool `json:"has_message,omitempty"`
	HasDM      bool `json:"has_dm,omitempty"`
}

type GuildScienceBorderwallCompleted struct {
	HasMessage bool `json:"has_message,omitempty"`
	HasDM      bool `json:"has_dm,omitempty"`
}

type GuildScienceMembershipReceived struct {
	MembershipUUID uuid.UUID `json:"membership_uuid"`
}

type GuildScienceMembershipRemoved struct {
	MembershipUUID uuid.UUID `json:"membership_uuid"`
}

type GuildScienceWelcomeMessageRemoved struct {
	HasMessage       bool              `json:"has_message,omitempty"`
	Successful       bool              `json:"successful,omitempty"`
	MessageID        discord.Snowflake `json:"message_id,omitempty"`
	MessageChannelID discord.Snowflake `json:"channel_id,omitempty"`
}

type GuildScienceLeaverMessageRemoved struct {
	HasMessage       bool              `json:"has_message,omitempty"`
	Successful       bool              `json:"successful,omitempty"`
	MessageID        discord.Snowflake `json:"message_id,omitempty"`
	MessageChannelID discord.Snowflake `json:"channel_id,omitempty"`
}
