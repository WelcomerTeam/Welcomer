package welcomer

import "github.com/WelcomerTeam/Discord/discord"

// Contains the structure for the "values" attribute in sandwich configuration files.

type ApplicationValues struct {
	IsCustomBot bool              `json:"custom_bot"`
	GuildID     discord.Snowflake `json:"guild_id"`
}
