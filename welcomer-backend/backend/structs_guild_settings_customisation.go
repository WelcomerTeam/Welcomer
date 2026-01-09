package backend

import "github.com/WelcomerTeam/Discord/discord"

type GuildSettingsCustomisation struct {
	Nickname *string           `json:"nickname,omitempty"`
	Avatar   *string           `json:"avatar,omitempty"`
	Banner   *string           `json:"banner,omitempty"`
	Bio      *string           `json:"bio,omitempty"`
	UserID   discord.Snowflake `json:"user_id"`
}
