// Code generated by sqlc. DO NOT EDIT.
// source: guild_settings_borderwall_query.sql

package database

import (
	"context"
	"database/sql"

	"github.com/jackc/pgtype"
)

const CreateBorderwallGuildSettings = `-- name: CreateBorderwallGuildSettings :one
INSERT INTO guild_settings_borderwall (guild_id)
    VALUES ($1)
RETURNING
    guild_id, toggle_enabled, message_verify, message_verified, roles_on_join, roles_on_verify
`

func (q *Queries) CreateBorderwallGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsBorderwall, error) {
	row := q.db.QueryRow(ctx, CreateBorderwallGuildSettings, guildID)
	var i GuildSettingsBorderwall
	err := row.Scan(
		&i.GuildID,
		&i.ToggleEnabled,
		&i.MessageVerify,
		&i.MessageVerified,
		&i.RolesOnJoin,
		&i.RolesOnVerify,
	)
	return &i, err
}

const GetBorderwallGuildSettings = `-- name: GetBorderwallGuildSettings :one
SELECT
    guild_id, toggle_enabled, message_verify, message_verified, roles_on_join, roles_on_verify
FROM
    guild_settings_borderwall
WHERE
    guild_id = $1
`

func (q *Queries) GetBorderwallGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsBorderwall, error) {
	row := q.db.QueryRow(ctx, GetBorderwallGuildSettings, guildID)
	var i GuildSettingsBorderwall
	err := row.Scan(
		&i.GuildID,
		&i.ToggleEnabled,
		&i.MessageVerify,
		&i.MessageVerified,
		&i.RolesOnJoin,
		&i.RolesOnVerify,
	)
	return &i, err
}

const UpdateBorderwallGuildSettings = `-- name: UpdateBorderwallGuildSettings :execrows
UPDATE
    guild_settings_borderwall
SET
    toggle_enabled = $2,
    message_verify = $3,
    message_verified = $4,
    roles_on_join = $5,
    roles_on_verify = $6
WHERE
    guild_id = $1
`

type UpdateBorderwallGuildSettingsParams struct {
	GuildID         int64        `json:"guild_id"`
	ToggleEnabled   sql.NullBool `json:"toggle_enabled"`
	MessageVerify   pgtype.JSONB `json:"message_verify"`
	MessageVerified pgtype.JSONB `json:"message_verified"`
	RolesOnJoin     []int64      `json:"roles_on_join"`
	RolesOnVerify   []int64      `json:"roles_on_verify"`
}

func (q *Queries) UpdateBorderwallGuildSettings(ctx context.Context, arg *UpdateBorderwallGuildSettingsParams) (int64, error) {
	result, err := q.db.Exec(ctx, UpdateBorderwallGuildSettings,
		arg.GuildID,
		arg.ToggleEnabled,
		arg.MessageVerify,
		arg.MessageVerified,
		arg.RolesOnJoin,
		arg.RolesOnVerify,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
