package backend

import (
	"github.com/WelcomerTeam/Discord/discord"
	"time"
)

type Guild struct {
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`
	Guild                *PartialGuild `json:"guild,omitempty"`
	SplashURL            string        `json:"splash_url"`
	EmbedColour          int           `json:"embed_colour"`
	HasWelcomerPro       bool          `json:"has_welcomer_pro"`
	HasCustomBackgrounds bool          `json:"has_custom_backgrounds"`
	StaffVisible         bool          `json:"staff_visible"`
	GuildVisible         bool          `json:"guild_visible"`
	AllowInvites         bool          `json:"allow_invites"`
}

type PartialGuild struct {
	MinimalGuild
	Channels    []MinimalChannel `json:"channels"`
	Roles       []MinimalRole    `json:"roles"`
	Emojis      []MinimalEmoji   `json:"emojis"`
	MemberCount int32            `json:"member_count"`
}

type MinimalGuild struct {
	Name            string            `json:"name"`
	Icon            string            `json:"icon"`
	IconHash        string            `json:"icon_hash"`
	Splash          string            `json:"splash,omitempty"`
	DiscoverySplash string            `json:"discovery_splash,omitempty"`
	Description     string            `json:"description,omitempty"`
	Banner          string            `json:"banner,omitempty"`
	ID              discord.Snowflake `json:"id"`
}

type MinimalChannel struct {
	Name     string              `json:"name,omitempty"`
	ID       discord.Snowflake   `json:"id"`
	Position int32               `json:"position,omitempty"`
	Type     discord.ChannelType `json:"type"`
}

type MinimalRole struct {
	tags         *discord.RoleTag
	Name         string            `json:"name"`
	ID           discord.Snowflake `json:"id"`
	permissions  discord.Int64
	Color        int32 `json:"color"`
	Position     int32 `json:"position"`
	IsAssignable bool  `json:"is_assignable"`
	IsElevated   bool  `json:"is_elevated"`
	managed      bool
}

type MinimalEmoji struct {
	Name      string            `json:"name"`
	ID        discord.Snowflake `json:"id"`
	Managed   bool              `json:"managed"`
	Animated  bool              `json:"animated"`
	Available bool              `json:"available"`
}

func GuildToPartial(guild discord.Guild) PartialGuild {
	return PartialGuild{
		MinimalGuild: GuildToMinimal(guild),
		MemberCount:  guild.MemberCount,
		Channels:     ChannelsToMinimal(guild.Channels),
		Roles:        RolesToMinimal(guild.Roles),
		Emojis:       EmojisToMinimal(guild.Emojis),
	}
}

func GuildToMinimal(guild discord.Guild) MinimalGuild {
	return MinimalGuild{
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

func ChannelsToMinimal(channels []discord.Channel) []MinimalChannel {
	minimalChannels := make([]MinimalChannel, len(channels))

	for i, channel := range channels {
		minimalChannels[i] = MinimalChannel{
			ID:       channel.ID,
			Type:     channel.Type,
			Position: channel.Position,
			Name:     channel.Name,
		}
	}

	return minimalChannels
}

func RolesToMinimal(roles []discord.Role) []MinimalRole {
	minimalRoles := make([]MinimalRole, len(roles))

	for i, role := range roles {
		minimalRoles[i] = MinimalRole{
			ID:       role.ID,
			Name:     role.Name,
			Color:    role.Color,
			Position: role.Position,

			permissions: role.Permissions,
			managed:     role.Managed,
			tags:        role.Tags,
		}
	}

	return minimalRoles
}

func EmojisToMinimal(emojis []discord.Emoji) []MinimalEmoji {
	minimalEmojis := make([]MinimalEmoji, len(emojis))

	for i, emoji := range emojis {
		minimalEmojis[i] = MinimalEmoji{
			ID:        emoji.ID,
			Name:      emoji.Name,
			Managed:   emoji.Managed,
			Animated:  emoji.Animated,
			Available: emoji.Available,
		}
	}

	return minimalEmojis
}

func MinimalRolesToMap(roles []MinimalRole) map[discord.Snowflake]MinimalRole {
	roleMap := map[discord.Snowflake]MinimalRole{}

	for _, role := range roles {
		roleMap[role.ID] = role
	}

	return roleMap
}
