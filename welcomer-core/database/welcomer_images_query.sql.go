// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1
// source: welcomer_images_query.sql

package database

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

const CreateWelcomerImages = `-- name: CreateWelcomerImages :one
INSERT INTO welcomer_images (welcomer_image_uuid, guild_id, created_at, image_type, data)
    VALUES ($1, $2, $3, $4, $5)
RETURNING
    welcomer_image_uuid, guild_id, created_at, image_type, data
`

type CreateWelcomerImagesParams struct {
	WelcomerImageUuid uuid.UUID `json:"welcomer_image_uuid"`
	GuildID           int64     `json:"guild_id"`
	CreatedAt         time.Time `json:"created_at"`
	ImageType         string    `json:"image_type"`
	Data              []byte    `json:"data"`
}

func (q *Queries) CreateWelcomerImages(ctx context.Context, arg *CreateWelcomerImagesParams) (*WelcomerImages, error) {
	row := q.db.QueryRow(ctx, CreateWelcomerImages,
		arg.WelcomerImageUuid,
		arg.GuildID,
		arg.CreatedAt,
		arg.ImageType,
		arg.Data,
	)
	var i WelcomerImages
	err := row.Scan(
		&i.WelcomerImageUuid,
		&i.GuildID,
		&i.CreatedAt,
		&i.ImageType,
		&i.Data,
	)
	return &i, err
}

const DeleteWelcomerImage = `-- name: DeleteWelcomerImage :execrows
DELETE FROM welcomer_images
WHERE welcomer_image_uuid = $1
`

func (q *Queries) DeleteWelcomerImage(ctx context.Context, welcomerImageUuid uuid.UUID) (int64, error) {
	result, err := q.db.Exec(ctx, DeleteWelcomerImage, welcomerImageUuid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const GetWelcomerImages = `-- name: GetWelcomerImages :one
SELECT
    welcomer_image_uuid, guild_id, created_at, image_type, data
FROM
    welcomer_images
WHERE
    welcomer_image_uuid = $1
`

func (q *Queries) GetWelcomerImages(ctx context.Context, welcomerImageUuid uuid.UUID) (*WelcomerImages, error) {
	row := q.db.QueryRow(ctx, GetWelcomerImages, welcomerImageUuid)
	var i WelcomerImages
	err := row.Scan(
		&i.WelcomerImageUuid,
		&i.GuildID,
		&i.CreatedAt,
		&i.ImageType,
		&i.Data,
	)
	return &i, err
}

const GetWelcomerImagesByGuildId = `-- name: GetWelcomerImagesByGuildId :many
SELECT
    welcomer_image_uuid, guild_id, created_at, image_type, data
FROM
    welcomer_images
WHERE
    guild_id = $1
`

func (q *Queries) GetWelcomerImagesByGuildId(ctx context.Context, guildID int64) ([]*WelcomerImages, error) {
	rows, err := q.db.Query(ctx, GetWelcomerImagesByGuildId, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*WelcomerImages{}
	for rows.Next() {
		var i WelcomerImages
		if err := rows.Scan(
			&i.WelcomerImageUuid,
			&i.GuildID,
			&i.CreatedAt,
			&i.ImageType,
			&i.Data,
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