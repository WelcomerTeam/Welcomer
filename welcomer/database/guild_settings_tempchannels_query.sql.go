// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: guild_settings_tempchannels_query.sql

package database

import (
	"context"
	"database/sql"
)

const CreateTempChannelsGuildSettings = `-- name: CreateTempChannelsGuildSettings :one
INSERT INTO guild_settings_tempchannels (guild_id)
    VALUES ($1)
RETURNING
    guild_id, toggle_enabled, toggle_autopurge, channel_lobby, channel_category, default_user_count
`

func (q *Queries) CreateTempChannelsGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsTempchannels, error) {
	row := q.db.QueryRow(ctx, CreateTempChannelsGuildSettings, guildID)
	var i GuildSettingsTempchannels
	err := row.Scan(
		&i.GuildID,
		&i.ToggleEnabled,
		&i.ToggleAutopurge,
		&i.ChannelLobby,
		&i.ChannelCategory,
		&i.DefaultUserCount,
	)
	return &i, err
}

const GetTempChannelsGuildSettings = `-- name: GetTempChannelsGuildSettings :one
SELECT
    guild_id, toggle_enabled, toggle_autopurge, channel_lobby, channel_category, default_user_count
FROM
    guild_settings_tempchannels
WHERE
    guild_id = $1
`

func (q *Queries) GetTempChannelsGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsTempchannels, error) {
	row := q.db.QueryRow(ctx, GetTempChannelsGuildSettings, guildID)
	var i GuildSettingsTempchannels
	err := row.Scan(
		&i.GuildID,
		&i.ToggleEnabled,
		&i.ToggleAutopurge,
		&i.ChannelLobby,
		&i.ChannelCategory,
		&i.DefaultUserCount,
	)
	return &i, err
}

const UpdateTempChannelsGuildSettings = `-- name: UpdateTempChannelsGuildSettings :execrows
UPDATE
    guild_settings_tempchannels
SET
    toggle_enabled = $2,
    toggle_autopurge = $3,
    channel_lobby = $4,
    channel_category = $5,
    default_user_count = $6
WHERE
    guild_id = $1
`

type UpdateTempChannelsGuildSettingsParams struct {
	GuildID          int64         `json:"guild_id"`
	ToggleEnabled    sql.NullBool  `json:"toggle_enabled"`
	ToggleAutopurge  sql.NullBool  `json:"toggle_autopurge"`
	ChannelLobby     sql.NullInt64 `json:"channel_lobby"`
	ChannelCategory  sql.NullInt64 `json:"channel_category"`
	DefaultUserCount sql.NullInt32 `json:"default_user_count"`
}

func (q *Queries) UpdateTempChannelsGuildSettings(ctx context.Context, arg *UpdateTempChannelsGuildSettingsParams) (int64, error) {
	result, err := q.db.Exec(ctx, UpdateTempChannelsGuildSettings,
		arg.GuildID,
		arg.ToggleEnabled,
		arg.ToggleAutopurge,
		arg.ChannelLobby,
		arg.ChannelCategory,
		arg.DefaultUserCount,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
