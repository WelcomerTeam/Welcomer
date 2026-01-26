package welcomer

import (
	"encoding/json"

	"github.com/WelcomerTeam/Discord/discord"
)

//go:generate go-enum -f=$GOFILE --marshal

// ENUM(emoji, buttons, dropdown)
type ReactionRoleType int32

type GuildSettingsReactionRoles []GuildSettingsReactionRole

type GuildSettingsReactionRole struct {
	Enabled   bool              `json:"enabled"`
	ChannelID discord.Snowflake `json:"channel_id"`
	MessageID discord.Snowflake `json:"message_id"`
	// Indicates if it is a message sent by Welcomer. If false, the user cannot
	// change the message embed through Welcomer.
	IsSystemMessage bool                 `json:"is_system_message"`
	MessageEmbed    *discord.Embed       `json:"message,omitempty"`
	Type            ReactionRoleType     `json:"type"`
	Roles           []ReactionRoleOption `json:"roles"`
}

type ReactionRoleOption struct {
	Style       discord.InteractionComponentStyle `json:"style,omitempty"`
	RoleID      discord.Snowflake                 `json:"role_id"`
	Emoji       string                            `json:"emoji"`
	Name        string                            `json:"name,omitempty"`
	Description string                            `json:"description,omitempty"`
}

func UnmarshalReactionRolesJSON(reactionRolesJSON []byte) (reactionRoles []GuildSettingsReactionRole) {
	_ = json.Unmarshal(reactionRolesJSON, &reactionRoles)

	return
}

func MarshalReactionRolesJSON(reactionRoles []GuildSettingsReactionRole) (reactionRolesJSON []byte) {
	reactionRolesJSON, _ = json.Marshal(reactionRoles)

	return
}
