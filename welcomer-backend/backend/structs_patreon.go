package backend

import "github.com/WelcomerTeam/Discord/discord"

type PatreonUser struct {
	Data PatreonUser_Data `json:"data"`
}

type PatreonUser_Data struct {
	Attributes PatreonUser_Attributes `json:"attributes"`
	ID         discord.Snowflake      `json:"id"`
	Type       string                 `json:"type"`
}

type PatreonUser_Attributes struct {
	Email             string                        `json:"email"`
	FullName          string                        `json:"full_name"`
	SocialConnections PatreonUser_SocialConnections `json:"social_connections"`
	ThumbURL          string                        `json:"thumb_url"`
}

type PatreonUser_SocialConnections struct {
	Discord PatreonUser_Discord `json:"discord"`
}

type PatreonUser_Discord struct {
	UserID discord.Snowflake `json:"user_id"`
}
