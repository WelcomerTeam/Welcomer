package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsReactionRoles struct {
	ToggleEnabled bool                                 `json:"enabled"`
	ReactionRoles []welcomer.GuildSettingsReactionRole `json:"reaction_roles"`
}

func GuildSettingsReactionRolesSettingsToPartial(
	reactionRoles *database.GuildSettingsReactionRoles,
) *GuildSettingsReactionRoles {
	partial := &GuildSettingsReactionRoles{
		ToggleEnabled: reactionRoles.ToggleEnabled,
		ReactionRoles: welcomer.UnmarshalReactionRolesJSON(welcomer.JSONBToBytes(reactionRoles.ReactionRoles)),
	}

	if len(partial.ReactionRoles) == 0 {
		partial.ReactionRoles = make([]welcomer.GuildSettingsReactionRole, 0)
	}

	return partial
}

func PartialToGuildSettingsReactionRolesSettings(guildID int64, guildSettings *GuildSettingsReactionRoles) *database.GuildSettingsReactionRoles {
	return &database.GuildSettingsReactionRoles{
		GuildID:       guildID,
		ToggleEnabled: guildSettings.ToggleEnabled,
		ReactionRoles: welcomer.BytesToJSONB(welcomer.MarshalReactionRolesJSON(guildSettings.ReactionRoles)),
	}
}
