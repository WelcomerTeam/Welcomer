// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.14.0
// source: guild_settings_welcomer_text_query.sql

package database

import (
	"context"
	"database/sql"

	"github.com/jackc/pgtype"
)

const CreateWelcomerTextGuildSettings = `-- name: CreateWelcomerTextGuildSettings :one
INSERT INTO guild_settings_welcomer_text (guild_id)
    VALUES ($1)
RETURNING
    guild_id, toggle_enabled, channel, message_format
`

func (q *Queries) CreateWelcomerTextGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsWelcomerText, error) {
	row := q.db.QueryRow(ctx, CreateWelcomerTextGuildSettings, guildID)
	var i GuildSettingsWelcomerText
	err := row.Scan(
		&i.GuildID,
		&i.ToggleEnabled,
		&i.Channel,
		&i.MessageFormat,
	)
	return &i, err
}

const GetWelcomerTextGuildSettings = `-- name: GetWelcomerTextGuildSettings :one
SELECT
    guild_id, toggle_enabled, channel, message_format
FROM
    guild_settings_welcomer_text
WHERE
    guild_id = $1
`

func (q *Queries) GetWelcomerTextGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsWelcomerText, error) {
	row := q.db.QueryRow(ctx, GetWelcomerTextGuildSettings, guildID)
	var i GuildSettingsWelcomerText
	err := row.Scan(
		&i.GuildID,
		&i.ToggleEnabled,
		&i.Channel,
		&i.MessageFormat,
	)
	return &i, err
}

const UpdateWelcomerTextGuildSettings = `-- name: UpdateWelcomerTextGuildSettings :execrows
UPDATE
    guild_settings_welcomer_text
SET
    toggle_enabled = $2,
    channel = $3,
    message_format = $4
WHERE
    guild_id = $1
`

type UpdateWelcomerTextGuildSettingsParams struct {
	GuildID       int64         `json:"guild_id"`
	ToggleEnabled sql.NullBool  `json:"toggle_enabled"`
	Channel       sql.NullInt64 `json:"channel"`
	MessageFormat pgtype.JSONB  `json:"message_format"`
}

func (q *Queries) UpdateWelcomerTextGuildSettings(ctx context.Context, arg *UpdateWelcomerTextGuildSettingsParams) (int64, error) {
	result, err := q.db.Exec(ctx, UpdateWelcomerTextGuildSettings,
		arg.GuildID,
		arg.ToggleEnabled,
		arg.Channel,
		arg.MessageFormat,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
