package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsRules struct {
	ToggleEnabled    bool     `json:"enabled"`
	ToggleDmsEnabled bool     `json:"dms_enabled"`
	Rules            []string `json:"rules"`
}

func GuildSettingsRulesSettingsToPartial(
	rules *database.GuildSettingsRules,
) *GuildSettingsRules {
	partial := &GuildSettingsRules{
		ToggleEnabled:    rules.ToggleEnabled,
		ToggleDmsEnabled: rules.ToggleDmsEnabled,
		Rules:            rules.Rules,
	}

	if len(partial.Rules) == 0 {
		partial.Rules = make([]string, 0)
	}

	return partial
}

func PartialToGuildSettingsRulesSettings(guildID int64, guildSettings *GuildSettingsRules) *database.GuildSettingsRules {
	return &database.GuildSettingsRules{
		GuildID:          guildID,
		ToggleEnabled:    guildSettings.ToggleEnabled,
		ToggleDmsEnabled: guildSettings.ToggleDmsEnabled,
		Rules:            guildSettings.Rules,
	}
}
