package backend

import (
	discord "github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsReactionRoles struct {
	ReactionRoles []welcomer.GuildSettingsReactionRole `json:"reaction_roles"`
}

func GuildSettingsReactionRolesSettingsToPartial(reactionRoles []*database.GuildSettingsReactionRoles) *GuildSettingsReactionRoles {
	partial := &GuildSettingsReactionRoles{
		ReactionRoles: make([]welcomer.GuildSettingsReactionRole, len(reactionRoles)),
	}

	for i, reactionRole := range reactionRoles {
		partial.ReactionRoles[i] = GuildSettingsReactionRoleToPartial(reactionRole)
	}

	return partial
}

func PartialToGuildSettingsReactionRolesSettings(guildID int64, guildSettings *GuildSettingsReactionRoles) []database.CreateOrUpdateReactionRoleSettingParams {
	reactionRoles := make([]database.CreateOrUpdateReactionRoleSettingParams, len(guildSettings.ReactionRoles))

	for i, reactionRole := range guildSettings.ReactionRoles {
		reactionRoles[i] = database.CreateOrUpdateReactionRoleSettingParams(PartialToGuildSettingsReactionRole(guildID, &reactionRole))
	}

	return reactionRoles
}

func GuildSettingsReactionRoleToPartial(reactionRole *database.GuildSettingsReactionRoles) welcomer.GuildSettingsReactionRole {
	return welcomer.GuildSettingsReactionRole{
		ReactionRoleID:  reactionRole.ReactionRoleID,
		Enabled:         reactionRole.ToggleEnabled,
		ChannelID:       discord.Snowflake(reactionRole.ChannelID),
		MessageID:       discord.Snowflake(reactionRole.MessageID),
		IsSystemMessage: reactionRole.IsSystemMessage,
		Message:         welcomer.JSONBToString(reactionRole.SystemMessageFormat),
		Type:            welcomer.ReactionRoleType(reactionRole.ReactionRoleType),
		Roles:           welcomer.UnmarshalReactionRolesJSON(welcomer.JSONBToBytes(reactionRole.Roles)),
	}
}

func PartialToGuildSettingsReactionRole(guildID int64, reactionRole *welcomer.GuildSettingsReactionRole) database.GuildSettingsReactionRoles {
	return database.GuildSettingsReactionRoles{
		GuildID:             guildID,
		ReactionRoleID:      reactionRole.ReactionRoleID,
		ToggleEnabled:       reactionRole.Enabled,
		ChannelID:           int64(reactionRole.ChannelID),
		MessageID:           int64(reactionRole.MessageID),
		IsSystemMessage:     reactionRole.IsSystemMessage,
		SystemMessageFormat: welcomer.StringToJSONB(reactionRole.Message),
		ReactionRoleType:    int32(reactionRole.Type),
		Roles:               welcomer.BytesToJSONB(welcomer.MarshalReactionRolesJSON(reactionRole.Roles)),
	}
}
