// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: guild_settings_welcomer_dms_query.sql

package database

import (
	"context"
	"database/sql"

	"github.com/jackc/pgtype"
)

const CreateWelcomerDMsGuildSettings = `-- name: CreateWelcomerDMsGuildSettings :one
INSERT INTO guild_settings_welcomer_dms (guild_id)
    VALUES ($1)
RETURNING
    guild_id, toggle_enabled, toggle_use_text_format, toggle_include_image, message_format
`

func (q *Queries) CreateWelcomerDMsGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsWelcomerDms, error) {
	row := q.db.QueryRow(ctx, CreateWelcomerDMsGuildSettings, guildID)
	var i GuildSettingsWelcomerDms
	err := row.Scan(
		&i.GuildID,
		&i.ToggleEnabled,
		&i.ToggleUseTextFormat,
		&i.ToggleIncludeImage,
		&i.MessageFormat,
	)
	return &i, err
}

const GetWelcomerDMsGuildSettings = `-- name: GetWelcomerDMsGuildSettings :one
SELECT
    guild_id, toggle_enabled, toggle_use_text_format, toggle_include_image, message_format
FROM
    guild_settings_welcomer_dms
WHERE
    guild_id = $1
`

func (q *Queries) GetWelcomerDMsGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsWelcomerDms, error) {
	row := q.db.QueryRow(ctx, GetWelcomerDMsGuildSettings, guildID)
	var i GuildSettingsWelcomerDms
	err := row.Scan(
		&i.GuildID,
		&i.ToggleEnabled,
		&i.ToggleUseTextFormat,
		&i.ToggleIncludeImage,
		&i.MessageFormat,
	)
	return &i, err
}

const UpdateWelcomerDMsGuildSettings = `-- name: UpdateWelcomerDMsGuildSettings :execrows
UPDATE
    guild_settings_welcomer_dms
SET
    toggle_enabled = $2,
    toggle_use_text_format = $3,
    toggle_include_image = $4,
    message_format = $5
WHERE
    guild_id = $1
`

type UpdateWelcomerDMsGuildSettingsParams struct {
	GuildID             int64        `json:"guild_id"`
	ToggleEnabled       sql.NullBool `json:"toggle_enabled"`
	ToggleUseTextFormat sql.NullBool `json:"toggle_use_text_format"`
	ToggleIncludeImage  sql.NullBool `json:"toggle_include_image"`
	MessageFormat       pgtype.JSONB `json:"message_format"`
}

func (q *Queries) UpdateWelcomerDMsGuildSettings(ctx context.Context, arg *UpdateWelcomerDMsGuildSettingsParams) (int64, error) {
	result, err := q.db.Exec(ctx, UpdateWelcomerDMsGuildSettings,
		arg.GuildID,
		arg.ToggleEnabled,
		arg.ToggleUseTextFormat,
		arg.ToggleIncludeImage,
		arg.MessageFormat,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
