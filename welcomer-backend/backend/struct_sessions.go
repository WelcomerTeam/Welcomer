package backend

import (
	"time"

	"github.com/WelcomerTeam/Discord/discord"
)

const (
	UserKey         = "user"
	GuildKey        = "guild"
	GuildIDKey      = "guildID"
	PatreonIDKey    = "patreonID"
	MembershipIDKey = "membershipID"
	KeyKey          = "key"
	TokenKey        = "token"
	StateKey        = "state"
	PreviousPathKey = "previous_path"
)

// SessionUser stores the user in a session.
type SessionUser struct {
	GuildsLastRequestedAt      time.Time                           `json:"-"`
	MembershipsLastRequestedAt time.Time                           `json:"-"`
	Guilds                     map[discord.Snowflake]*SessionGuild `json:"guilds"`
	Username                   string                              `json:"username"`
	Discriminator              string                              `json:"discriminator"`
	GlobalName                 string                              `json:"global_name"`
	Avatar                     string                              `json:"avatar"`
	Memberships                []*Membership                       `json:"memberships"`
	ID                         discord.Snowflake                   `json:"id"`
}

// SessionGuild represents a guild passed through /api/users/guilds and is stored in the session.
type SessionGuild struct {
	Name                 string            `json:"name"`
	Icon                 string            `json:"icon"`
	ID                   discord.Snowflake `json:"id"`
	HasWelcomer          bool              `json:"has_welcomer"`
	HasWelcomerPro       bool              `json:"has_welcomer_pro"`
	HasCustomBackgrounds bool              `json:"has_custom_backgrounds"`
	HasElevation         bool              `json:"has_elevation"`
	IsOwner              bool              `json:"is_owner"`
}
