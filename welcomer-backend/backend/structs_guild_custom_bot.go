package backend

import discord "github.com/WelcomerTeam/Discord/discord"

type GuildCustomBotResponse struct {
	Limit int              `json:"limit"`
	Bots  []GuildCustomBot `json:"bots"`
}

type GuildCustomBot struct {
	UUID              string            `json:"id"`
	IsActive          bool              `json:"is_active"`
	ApplicationID     discord.Snowflake `json:"application_id"`
	ApplicationName   string            `json:"application_name"`
	ApplicationAvatar string            `json:"application_avatar"`

	Shards []GetStatusResponseShard `json:"shards,omitempty"`
}

type GuildCustomBotPayload struct {
	Token string `json:"token"`
}
