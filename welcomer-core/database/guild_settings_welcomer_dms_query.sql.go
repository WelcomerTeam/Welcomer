// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1
// source: guild_settings_welcomer_dms_query.sql

package database

import (
	"context"

	"github.com/jackc/pgtype"
)

const CreateOrUpdateWelcomerDMsGuildSettings = `-- name: CreateOrUpdateWelcomerDMsGuildSettings :one
INSERT INTO guild_settings_welcomer_dms (guild_id, toggle_enabled, toggle_use_text_format, toggle_include_image, message_format)
    VALUES ($1, $2, $3, $4, $5)
ON CONFLICT(guild_id) DO UPDATE
    SET toggle_enabled = EXCLUDED.toggle_enabled, 
        toggle_use_text_format = EXCLUDED.toggle_use_text_format, 
        toggle_include_image = EXCLUDED.toggle_include_image, 
        message_format = EXCLUDED.message_format
RETURNING
    guild_id, toggle_enabled, toggle_use_text_format, toggle_include_image, message_format
`

type CreateOrUpdateWelcomerDMsGuildSettingsParams struct {
	GuildID             int64        `json:"guild_id"`
	ToggleEnabled       bool         `json:"toggle_enabled"`
	ToggleUseTextFormat bool         `json:"toggle_use_text_format"`
	ToggleIncludeImage  bool         `json:"toggle_include_image"`
	MessageFormat       pgtype.JSONB `json:"message_format"`
}

func (q *Queries) CreateOrUpdateWelcomerDMsGuildSettings(ctx context.Context, arg *CreateOrUpdateWelcomerDMsGuildSettingsParams) (*GuildSettingsWelcomerDms, error) {
	row := q.db.QueryRow(ctx, CreateOrUpdateWelcomerDMsGuildSettings,
		arg.GuildID,
		arg.ToggleEnabled,
		arg.ToggleUseTextFormat,
		arg.ToggleIncludeImage,
		arg.MessageFormat,
	)
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

const CreateWelcomerDMsGuildSettings = `-- name: CreateWelcomerDMsGuildSettings :one
INSERT INTO guild_settings_welcomer_dms (guild_id, toggle_enabled, toggle_use_text_format, toggle_include_image, message_format)
    VALUES ($1, $2, $3, $4, $5)
RETURNING
    guild_id, toggle_enabled, toggle_use_text_format, toggle_include_image, message_format
`

type CreateWelcomerDMsGuildSettingsParams struct {
	GuildID             int64        `json:"guild_id"`
	ToggleEnabled       bool         `json:"toggle_enabled"`
	ToggleUseTextFormat bool         `json:"toggle_use_text_format"`
	ToggleIncludeImage  bool         `json:"toggle_include_image"`
	MessageFormat       pgtype.JSONB `json:"message_format"`
}

func (q *Queries) CreateWelcomerDMsGuildSettings(ctx context.Context, arg *CreateWelcomerDMsGuildSettingsParams) (*GuildSettingsWelcomerDms, error) {
	row := q.db.QueryRow(ctx, CreateWelcomerDMsGuildSettings,
		arg.GuildID,
		arg.ToggleEnabled,
		arg.ToggleUseTextFormat,
		arg.ToggleIncludeImage,
		arg.MessageFormat,
	)
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
	ToggleEnabled       bool         `json:"toggle_enabled"`
	ToggleUseTextFormat bool         `json:"toggle_use_text_format"`
	ToggleIncludeImage  bool         `json:"toggle_include_image"`
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
