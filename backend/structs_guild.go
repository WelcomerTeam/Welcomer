package backend

import (
	"time"

	"github.com/WelcomerTeam/Discord/discord"
)

type Guild struct {
	Guild *PartialGuild `json:"guild,omitempty"`

	HasMembership bool `json:"has_membership"`

	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	EmbedColour  int       `json:"embed_colour"`
	SplashURL    string    `json:"splash_url"`
	StaffVisible bool      `json:"staff_visible"`
	GuildVisible bool      `json:"guild_visible"`
	AllowInvites bool      `json:"allow_invites"`
}

type PartialGuild struct {
	*MinimalGuild

	MemberCount int32 `json:"member_count"`

	Channels []*discord.Channel `json:"channels"`
	Roles    []*discord.Role    `json:"roles"`
	Emojis   []*discord.Emoji   `json:"emojis"`
}

type MinimalGuild struct {
	ID              discord.Snowflake `json:"id"`
	Name            string            `json:"name"`
	Icon            string            `json:"icon"`
	IconHash        string            `json:"icon_hash"`
	Splash          string            `json:"splash,omitempty"`
	DiscoverySplash string            `json:"discovery_splash,omitempty"`
	Description     string            `json:"description,omitempty"`
	Banner          string            `json:"banner,omitempty"`
}

func GuildToPartial(guild *discord.Guild) *PartialGuild {
	return &PartialGuild{
		MinimalGuild: GuildToMinimal(guild),
		MemberCount:  guild.MemberCount,
		Channels:     guild.Channels,
		Roles:        guild.Roles,
		Emojis:       guild.Emojis,
	}
}

func GuildToMinimal(guild *discord.Guild) *MinimalGuild {
	return &MinimalGuild{
		ID:              guild.ID,
		Name:            guild.Name,
		Icon:            guild.Icon,
		IconHash:        guild.IconHash,
		Splash:          guild.Splash,
		DiscoverySplash: guild.DiscoverySplash,
		Description:     guild.Description,
		Banner:          guild.Banner,
	}
}
