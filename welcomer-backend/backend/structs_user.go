package backend

import (
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/gofrs/uuid"
)

type MinimalUser struct {
	Username      string            `json:"username"`
	Discriminator string            `json:"discriminator"`
	GlobalName    string            `json:"global_name"`
	Avatar        string            `json:"avatar"`
	Memberships   []*Membership     `json:"memberships,omitempty"`
	ID            discord.Snowflake `json:"id"`
}

// Membership represents a membership to a server.
type Membership struct {
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	StartedAt      time.Time         `json:"started_at"`
	ExpiresAt      time.Time         `json:"expires_at"`
	Guild          *MinimalGuild     `json:"guild"`
	GuildID        discord.Snowflake `json:"guild_id"`
	Status         int32             `json:"status"`
	MembershipType int32             `json:"membership_type"`
	MembershipUuid uuid.UUID         `json:"membership_uuid"`
}

func SessionUserToMinimal(sessionUser *SessionUser) *MinimalUser {
	return &MinimalUser{
		ID:            sessionUser.ID,
		Username:      sessionUser.Username,
		Discriminator: sessionUser.Discriminator,
		GlobalName:    sessionUser.GlobalName,
		Avatar:        sessionUser.Avatar,
		Memberships:   sessionUser.Memberships,
	}
}
