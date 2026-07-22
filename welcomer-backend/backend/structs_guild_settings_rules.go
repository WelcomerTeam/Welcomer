package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsRules struct {
	Rules                   []string                `json:"rules"`
	ToggleEnabled           bool                    `json:"enabled"`
	ToggleDmsEnabled        bool                    `json:"dms_enabled"`
	ModerationCheckupStatus ModerationCheckupStatus `json:"moderation_checkup_status"`
}

func GuildSettingsRulesSettingsToPartial(
	rules *database.GetRulesGuildSettingsRow,
) *GuildSettingsRules {
	partial := &GuildSettingsRules{
		ToggleEnabled:    rules.ToggleEnabled,
		ToggleDmsEnabled: rules.ToggleDmsEnabled,
		Rules:            rules.Rules,
		ModerationCheckupStatus: welcomer.If(
			rules.ModerationCheckupUuid.UUID.IsNil(),
			ModerationCheckupStatusUnknown,
			welcomer.If(
				rules.CompletedAt.Time.IsZero(),
				ModerationCheckupStatusPending,
				welcomer.If(
					rules.IsBlocked.Bool,
					ModerationCheckupStatusRejected,
					ModerationCheckupStatusApproved,
				),
			),
		),
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
