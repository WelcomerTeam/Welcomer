// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: guild_settings_rules_query.sql

package database

import (
	"context"
	"database/sql"
)

const CreateRulesGuildSettings = `-- name: CreateRulesGuildSettings :one
INSERT INTO guild_settings_rules (guild_id)
    VALUES ($1)
RETURNING
    guild_id, toggle_enabled, toggle_dms_enabled, rules
`

func (q *Queries) CreateRulesGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsRules, error) {
	row := q.db.QueryRow(ctx, CreateRulesGuildSettings, guildID)
	var i GuildSettingsRules
	err := row.Scan(
		&i.GuildID,
		&i.ToggleEnabled,
		&i.ToggleDmsEnabled,
		&i.Rules,
	)
	return &i, err
}

const GetRulesGuildSettings = `-- name: GetRulesGuildSettings :one
SELECT
    guild_id, toggle_enabled, toggle_dms_enabled, rules
FROM
    guild_settings_rules
WHERE
    guild_id = $1
`

func (q *Queries) GetRulesGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsRules, error) {
	row := q.db.QueryRow(ctx, GetRulesGuildSettings, guildID)
	var i GuildSettingsRules
	err := row.Scan(
		&i.GuildID,
		&i.ToggleEnabled,
		&i.ToggleDmsEnabled,
		&i.Rules,
	)
	return &i, err
}

const UpdateRuleGuildSettings = `-- name: UpdateRuleGuildSettings :execrows
UPDATE
    guild_settings_rules
SET
    toggle_enabled = $2,
    toggle_dms_enabled = $3,
    rules = $4
WHERE
    guild_id = $1
`

type UpdateRuleGuildSettingsParams struct {
	GuildID          int64        `json:"guild_id"`
	ToggleEnabled    sql.NullBool `json:"toggle_enabled"`
	ToggleDmsEnabled sql.NullBool `json:"toggle_dms_enabled"`
	Rules            []string     `json:"rules"`
}

func (q *Queries) UpdateRuleGuildSettings(ctx context.Context, arg *UpdateRuleGuildSettingsParams) (int64, error) {
	result, err := q.db.Exec(ctx, UpdateRuleGuildSettings,
		arg.GuildID,
		arg.ToggleEnabled,
		arg.ToggleDmsEnabled,
		arg.Rules,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
