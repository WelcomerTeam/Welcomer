package backend

import (
	"time"

	"github.com/WelcomerTeam/Discord/discord"
)

type Guild struct {
	Guild *PartialGuild `json:"guild,omitempty"`

	HasWelcomerPro       bool `json:"has_welcomer_pro"`
	HasCustomBackgrounds bool `json:"has_custom_backgrounds"`

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

	Channels []*MinimalChannel `json:"channels"`
	Roles    []*MinimalRole    `json:"roles"`
	Emojis   []*MinimalEmoji   `json:"emojis"`
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

type MinimalChannel struct {
	ID       discord.Snowflake   `json:"id"`
	Type     discord.ChannelType `json:"type"`
	Position int32               `json:"position,omitempty"`
	Name     string              `json:"name,omitempty"`
}

type MinimalRole struct {
	ID       discord.Snowflake `json:"id"`
	Name     string            `json:"name"`
	Color    int32             `json:"color"`
	Position int32             `json:"position"`

	IsAssignable bool `json:"is_assignable"`
	IsElevated   bool `json:"is_elevated"`

	// Attributes used for IsAssignable and IsElevated calculations
	permissions discord.Int64
	managed     bool
	tags        *discord.RoleTag
}

type MinimalEmoji struct {
	ID        discord.Snowflake `json:"id"`
	Name      string            `json:"name"`
	Managed   bool              `json:"managed"`
	Animated  bool              `json:"animated"`
	Available bool              `json:"available"`
}

func GuildToPartial(guild *discord.Guild) *PartialGuild {
	return &PartialGuild{
		MinimalGuild: GuildToMinimal(guild),
		MemberCount:  guild.MemberCount,
		Channels:     ChannelsToMinimal(guild.Channels),
		Roles:        RolesToMinimal(guild.Roles),
		Emojis:       EmojisToMinimal(guild.Emojis),
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

func ChannelsToMinimal(channels []*discord.Channel) []*MinimalChannel {
	minimalChannels := make([]*MinimalChannel, 0, len(channels))

	for _, channel := range channels {
		minimalChannels = append(minimalChannels, &MinimalChannel{
			ID:       channel.ID,
			Type:     channel.Type,
			Position: channel.Position,
			Name:     channel.Name,
		})
	}

	return minimalChannels
}

func RolesToMinimal(roles []*discord.Role) []*MinimalRole {
	minimalRoles := make([]*MinimalRole, 0, len(roles))

	for _, role := range roles {
		minimalRoles = append(minimalRoles, &MinimalRole{
			ID:       role.ID,
			Name:     role.Name,
			Color:    role.Color,
			Position: role.Position,

			permissions: role.Permissions,
			managed:     role.Managed,
			tags:        role.Tags,
		})
	}

	return minimalRoles
}

func EmojisToMinimal(emojis []*discord.Emoji) []*MinimalEmoji {
	minimalEmojis := make([]*MinimalEmoji, 0, len(emojis))

	for _, emoji := range emojis {
		minimalEmojis = append(minimalEmojis, &MinimalEmoji{
			ID:        emoji.ID,
			Name:      emoji.Name,
			Managed:   emoji.Managed,
			Animated:  emoji.Animated,
			Available: emoji.Available,
		})
	}

	return minimalEmojis
}

func MinimalRolesToMap(roles []*MinimalRole) map[discord.Snowflake]*MinimalRole {
	roleMap := map[discord.Snowflake]*MinimalRole{}

	for _, role := range roles {
		roleMap[role.ID] = role
	}

	return roleMap
}
