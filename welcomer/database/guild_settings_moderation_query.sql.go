// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: guild_settings_moderation_query.sql

package database

import (
	"context"
	"database/sql"
)

const CreateModerationGuildSettings = `-- name: CreateModerationGuildSettings :one
INSERT INTO guild_settings_moderation (guild_id)
    VALUES ($1)
RETURNING
    guild_id, toggle_reasons_required
`

func (q *Queries) CreateModerationGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsModeration, error) {
	row := q.db.QueryRow(ctx, CreateModerationGuildSettings, guildID)
	var i GuildSettingsModeration
	err := row.Scan(&i.GuildID, &i.ToggleReasonsRequired)
	return &i, err
}

const GetModerationGuildSettings = `-- name: GetModerationGuildSettings :one
SELECT
    guild_id, toggle_reasons_required
FROM
    guild_settings_moderation
WHERE
    guild_id = $1
`

func (q *Queries) GetModerationGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsModeration, error) {
	row := q.db.QueryRow(ctx, GetModerationGuildSettings, guildID)
	var i GuildSettingsModeration
	err := row.Scan(&i.GuildID, &i.ToggleReasonsRequired)
	return &i, err
}

const UpdateModerationGuildSettings = `-- name: UpdateModerationGuildSettings :execrows
UPDATE
    guild_settings_moderation
SET
    toggle_reasons_required = $2
WHERE
    guild_id = $1
`

type UpdateModerationGuildSettingsParams struct {
	GuildID               int64        `json:"guild_id"`
	ToggleReasonsRequired sql.NullBool `json:"toggle_reasons_required"`
}

func (q *Queries) UpdateModerationGuildSettings(ctx context.Context, arg *UpdateModerationGuildSettingsParams) (int64, error) {
	result, err := q.db.Exec(ctx, UpdateModerationGuildSettings, arg.GuildID, arg.ToggleReasonsRequired)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
