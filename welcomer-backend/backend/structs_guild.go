package backend

import (
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type Guild struct {
	CreatedAt            time.Time               `json:"created_at"`
	UpdatedAt            time.Time               `json:"updated_at"`
	Guild                *PartialGuild           `json:"guild,omitempty"`
	SplashURL            string                  `json:"splash_url"`
	EmbedColour          int                     `json:"embed_colour"`
	HasWelcomerPro       bool                    `json:"has_welcomer_pro"`
	HasCustomBackgrounds bool                    `json:"has_custom_backgrounds"`
	Features             []welcomer.GuildFeature `json:"features"`
	StaffVisible         bool                    `json:"staff_visible"`
	GuildVisible         bool                    `json:"guild_visible"`
	AllowInvites         bool                    `json:"allow_invites"`
}

type PartialGuild struct {
	*MinimalGuild
	Channels []*MinimalChannel          `json:"channels"`
	Roles    []*welcomer.AssignableRole `json:"roles"`
	Emojis   []*MinimalEmoji            `json:"emojis"`

	MemberCount   int32  `json:"member_count"`
	MembersJoined int32  `json:"members_joined"`
	NumberLocale  string `json:"number_locale"`
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

type MinimalEmoji struct {
	Name      string            `json:"name"`
	ID        discord.Snowflake `json:"id"`
	Managed   bool              `json:"managed"`
	Animated  bool              `json:"animated"`
	Available bool              `json:"available"`
}

func GuildToPartial(guild *discord.Guild, guildConfig *database.Guilds) *PartialGuild {
	return &PartialGuild{
		MinimalGuild:  GuildToMinimal(guild),
		Channels:      ChannelsToMinimal(guild.Channels),
		Roles:         RolesToMinimal(guild.Roles),
		Emojis:        EmojisToMinimal(guild.Emojis),
		MemberCount:   guild.MemberCount,
		NumberLocale:  database.NumberLocale(guildConfig.NumberLocale.Int32).String(),
		MembersJoined: guildConfig.MemberCount,
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

func ChannelsToMinimal(channels []discord.Channel) []*MinimalChannel {
	minimalChannels := make([]*MinimalChannel, len(channels))

	for i, channel := range channels {
		minimalChannels[i] = &MinimalChannel{
			ID:       channel.ID,
			Type:     channel.Type,
			Position: channel.Position,
			Name:     channel.Name,
		}
	}

	return minimalChannels
}

func RolesToMinimal(roles []discord.Role) []*welcomer.AssignableRole {
	minimalRoles := make([]*welcomer.AssignableRole, len(roles))

	for i, role := range roles {
		minimalRoles[i] = &welcomer.AssignableRole{
			Role:         &role,
			IsAssignable: false,
			IsElevated:   false,
		}
	}

	return minimalRoles
}

func EmojisToMinimal(emojis []discord.Emoji) []*MinimalEmoji {
	minimalEmojis := make([]*MinimalEmoji, len(emojis))

	for i, emoji := range emojis {
		minimalEmojis[i] = &MinimalEmoji{
			ID:        emoji.ID,
			Name:      emoji.Name,
			Managed:   emoji.Managed,
			Animated:  emoji.Animated,
			Available: emoji.Available,
		}
	}

	return minimalEmojis
}
