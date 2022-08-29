// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: guild_settings_automoderation_query.sql

package database

import (
	"context"
	"database/sql"
)

const CreateAutoModerationGuildSettings = `-- name: CreateAutoModerationGuildSettings :one
INSERT INTO guild_settings_automoderation (guild_id)
    VALUES ($1)
RETURNING
    guild_id, toggle_masscaps, toggle_massmentions, toggle_urls, toggle_invites, toggle_regex, toggle_phishing, toggle_massemojis, threshold_mass_caps_percentage, threshold_mass_mentions, threshold_mass_emojis, blacklisted_urls, whitelisted_urls, regex_patterns, toggle_smartmod
`

func (q *Queries) CreateAutoModerationGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsAutomoderation, error) {
	row := q.db.QueryRow(ctx, CreateAutoModerationGuildSettings, guildID)
	var i GuildSettingsAutomoderation
	err := row.Scan(
		&i.GuildID,
		&i.ToggleMasscaps,
		&i.ToggleMassmentions,
		&i.ToggleUrls,
		&i.ToggleInvites,
		&i.ToggleRegex,
		&i.TogglePhishing,
		&i.ToggleMassemojis,
		&i.ThresholdMassCapsPercentage,
		&i.ThresholdMassMentions,
		&i.ThresholdMassEmojis,
		&i.BlacklistedUrls,
		&i.WhitelistedUrls,
		&i.RegexPatterns,
		&i.ToggleSmartmod,
	)
	return &i, err
}

const GetAutoModerationGuildSettings = `-- name: GetAutoModerationGuildSettings :one
SELECT
    guild_id, toggle_masscaps, toggle_massmentions, toggle_urls, toggle_invites, toggle_regex, toggle_phishing, toggle_massemojis, threshold_mass_caps_percentage, threshold_mass_mentions, threshold_mass_emojis, blacklisted_urls, whitelisted_urls, regex_patterns, toggle_smartmod
FROM
    guild_settings_automoderation
WHERE
    guild_id = $1
`

func (q *Queries) GetAutoModerationGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsAutomoderation, error) {
	row := q.db.QueryRow(ctx, GetAutoModerationGuildSettings, guildID)
	var i GuildSettingsAutomoderation
	err := row.Scan(
		&i.GuildID,
		&i.ToggleMasscaps,
		&i.ToggleMassmentions,
		&i.ToggleUrls,
		&i.ToggleInvites,
		&i.ToggleRegex,
		&i.TogglePhishing,
		&i.ToggleMassemojis,
		&i.ThresholdMassCapsPercentage,
		&i.ThresholdMassMentions,
		&i.ThresholdMassEmojis,
		&i.BlacklistedUrls,
		&i.WhitelistedUrls,
		&i.RegexPatterns,
		&i.ToggleSmartmod,
	)
	return &i, err
}

const UpdateAutoModerationGuildSettings = `-- name: UpdateAutoModerationGuildSettings :execrows
UPDATE
    guild_settings_automoderation
SET
    toggle_masscaps = $2,
    toggle_massmentions = $3,
    toggle_urls = $4,
    toggle_invites = $5,
    toggle_regex = $6,
    toggle_phishing = $7,
    toggle_massemojis = $8,
    threshold_mass_caps_percentage = $9,
    threshold_mass_mentions = $10,
    threshold_mass_emojis = $11,
    blacklisted_urls = $12,
    whitelisted_urls = $13,
    regex_patterns = $14,
    toggle_smartmod = $15
WHERE
    guild_id = $1
`

type UpdateAutoModerationGuildSettingsParams struct {
	GuildID                     int64        `json:"guild_id"`
	ToggleMasscaps              sql.NullBool `json:"toggle_masscaps"`
	ToggleMassmentions          sql.NullBool `json:"toggle_massmentions"`
	ToggleUrls                  sql.NullBool `json:"toggle_urls"`
	ToggleInvites               sql.NullBool `json:"toggle_invites"`
	ToggleRegex                 sql.NullBool `json:"toggle_regex"`
	TogglePhishing              sql.NullBool `json:"toggle_phishing"`
	ToggleMassemojis            sql.NullBool `json:"toggle_massemojis"`
	ThresholdMassCapsPercentage int32        `json:"threshold_mass_caps_percentage"`
	ThresholdMassMentions       int32        `json:"threshold_mass_mentions"`
	ThresholdMassEmojis         int32        `json:"threshold_mass_emojis"`
	BlacklistedUrls             []string     `json:"blacklisted_urls"`
	WhitelistedUrls             []string     `json:"whitelisted_urls"`
	RegexPatterns               []string     `json:"regex_patterns"`
	ToggleSmartmod              bool         `json:"toggle_smartmod"`
}

func (q *Queries) UpdateAutoModerationGuildSettings(ctx context.Context, arg *UpdateAutoModerationGuildSettingsParams) (int64, error) {
	result, err := q.db.Exec(ctx, UpdateAutoModerationGuildSettings,
		arg.GuildID,
		arg.ToggleMasscaps,
		arg.ToggleMassmentions,
		arg.ToggleUrls,
		arg.ToggleInvites,
		arg.ToggleRegex,
		arg.TogglePhishing,
		arg.ToggleMassemojis,
		arg.ThresholdMassCapsPercentage,
		arg.ThresholdMassMentions,
		arg.ThresholdMassEmojis,
		arg.BlacklistedUrls,
		arg.WhitelistedUrls,
		arg.RegexPatterns,
		arg.ToggleSmartmod,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
