package backend

import (
	"time"

	"github.com/WelcomerTeam/Discord/discord"
)

const (
	UserKey         = "user"
	TokenKey        = "token"
	StateKey        = "state"
	PreviousPathKey = "previous_path"
)

// SessionUser stores the user in a session.
type SessionUser struct {
	ID                    discord.Snowflake `json:"id"`
	Username              string            `json:"username"`
	Discriminator         string            `json:"discriminator"`
	Avatar                string            `json:"avatar"`
	Guilds                []*SessionGuild   `json:"guilds"`
	GuildsLastRequestedAt time.Time         `json:"-"`
}

// SessionGuild represents a guild passed through /api/users/guilds and is stored in the session.
type SessionGuild struct {
	ID            discord.Snowflake `json:"id"`
	Name          string            `json:"name"`
	Icon          string            `json:"icon"`
	HasWelcomer   bool              `json:"has_welcomer"`
	HasMembership bool              `json:"has_membership"`
}
