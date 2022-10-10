// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: guild_settings_welcomer_backgrounds_query.sql

package database

import (
	"context"

	"github.com/gofrs/uuid"
)

const CreateWelcomerBackground = `-- name: CreateWelcomerBackground :one
INSERT INTO guild_settings_welcomer_backgrounds (image_uuid, created_at, guild_id, filename, filesize, filetype)
    VALUES (uuid_generate_v4(), now(), $1, $2, $3, $4)
RETURNING
    image_uuid, created_at, guild_id, filename, filesize, filetype
`

type CreateWelcomerBackgroundParams struct {
	GuildID  int64  `json:"guild_id"`
	Filename string `json:"filename"`
	Filesize int32  `json:"filesize"`
	Filetype string `json:"filetype"`
}

func (q *Queries) CreateWelcomerBackground(ctx context.Context, arg *CreateWelcomerBackgroundParams) (*GuildSettingsWelcomerBackgrounds, error) {
	row := q.db.QueryRow(ctx, CreateWelcomerBackground,
		arg.GuildID,
		arg.Filename,
		arg.Filesize,
		arg.Filetype,
	)
	var i GuildSettingsWelcomerBackgrounds
	err := row.Scan(
		&i.ImageUuid,
		&i.CreatedAt,
		&i.GuildID,
		&i.Filename,
		&i.Filesize,
		&i.Filetype,
	)
	return &i, err
}

const DeleteWelcomerBackground = `-- name: DeleteWelcomerBackground :execrows
DELETE FROM guild_settings_welcomer_backgrounds
WHERE image_uuid = $1
`

func (q *Queries) DeleteWelcomerBackground(ctx context.Context, imageUuid uuid.UUID) (int64, error) {
	result, err := q.db.Exec(ctx, DeleteWelcomerBackground, imageUuid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const GetWelcomerBackground = `-- name: GetWelcomerBackground :one
SELECT
    image_uuid, created_at, guild_id, filename, filesize, filetype
FROM
    guild_settings_welcomer_backgrounds
WHERE
    image_uuid = $1
`

func (q *Queries) GetWelcomerBackground(ctx context.Context, imageUuid uuid.UUID) (*GuildSettingsWelcomerBackgrounds, error) {
	row := q.db.QueryRow(ctx, GetWelcomerBackground, imageUuid)
	var i GuildSettingsWelcomerBackgrounds
	err := row.Scan(
		&i.ImageUuid,
		&i.CreatedAt,
		&i.GuildID,
		&i.Filename,
		&i.Filesize,
		&i.Filetype,
	)
	return &i, err
}

const GetWelcomerBackgroundByGuildID = `-- name: GetWelcomerBackgroundByGuildID :many
SELECT
    image_uuid, created_at, guild_id, filename, filesize, filetype
FROM
    guild_settings_welcomer_backgrounds
WHERE
    guild_id = $1
`

func (q *Queries) GetWelcomerBackgroundByGuildID(ctx context.Context, guildID int64) ([]*GuildSettingsWelcomerBackgrounds, error) {
	rows, err := q.db.Query(ctx, GetWelcomerBackgroundByGuildID, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*GuildSettingsWelcomerBackgrounds{}
	for rows.Next() {
		var i GuildSettingsWelcomerBackgrounds
		if err := rows.Scan(
			&i.ImageUuid,
			&i.CreatedAt,
			&i.GuildID,
			&i.Filename,
			&i.Filesize,
			&i.Filetype,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
