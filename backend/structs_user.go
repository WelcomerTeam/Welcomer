package backend

import (
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/gofrs/uuid"
)

type MinimalUser struct {
	ID            discord.Snowflake `json:"id"`
	Username      string            `json:"username"`
	Discriminator string            `json:"discriminator"`
	Avatar        string            `json:"avatar"`

	Memberships []*Membership `json:"memberships,omitempty"`
}

// Membership represents a membership to a server.
type Membership struct {
	MembershipUuid uuid.UUID         `json:"membership_uuid"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	StartedAt      time.Time         `json:"started_at"`
	ExpiresAt      time.Time         `json:"expires_at"`
	Status         int32             `json:"status"`
	MembershipType int32             `json:"membership_type"`
	GuildID        discord.Snowflake `json:"guild_id"`

	Guild *MinimalGuild `json:"guild"`
}

func SessionUserToMinimal(sessionUser *SessionUser) *MinimalUser {
	return &MinimalUser{
		ID:            sessionUser.ID,
		Username:      sessionUser.Username,
		Discriminator: sessionUser.Discriminator,
		Avatar:        sessionUser.Avatar,
		Memberships:   sessionUser.Memberships,
	}
}
