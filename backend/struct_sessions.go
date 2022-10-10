package backend

import (
	"time"

	"github.com/WelcomerTeam/Discord/discord"
)

const (
	UserKey         = "user"
	GuildKey        = "guild"
	GuildIDKey      = "guildID"
	KeyKey          = "key"
	TokenKey        = "token"
	StateKey        = "state"
	PreviousPathKey = "previous_path"
)

// SessionUser stores the user in a session.
type SessionUser struct {
	ID                         discord.Snowflake                   `json:"id"`
	Username                   string                              `json:"username"`
	Discriminator              string                              `json:"discriminator"`
	Avatar                     string                              `json:"avatar"`
	Guilds                     map[discord.Snowflake]*SessionGuild `json:"guilds"`
	GuildsLastRequestedAt      time.Time                           `json:"-"`
	Memberships                []*Membership                       `json:"memberships"`
	MembershipsLastRequestedAt time.Time                           `json:"-"`
}

// SessionGuild represents a guild passed through /api/users/guilds and is stored in the session.
type SessionGuild struct {
	ID                   discord.Snowflake `json:"id"`
	Name                 string            `json:"name"`
	Icon                 string            `json:"icon"`
	HasWelcomer          bool              `json:"has_welcomer"`
	HasWelcomerPro       bool              `json:"has_welcomer_pro"`
	HasCustomBackgrounds bool              `json:"has_custom_backgrounds"`
	HasElevation         bool              `json:"has_elevation"`
	IsOwner              bool              `json:"is_owner"`
}
