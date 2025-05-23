// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: guild_settings_welcomer_images_query.sql

package database

import (
	"context"
)

const CreateOrUpdateWelcomerImagesGuildSettings = `-- name: CreateOrUpdateWelcomerImagesGuildSettings :one
INSERT INTO guild_settings_welcomer_images (guild_id, toggle_enabled, toggle_image_border, toggle_show_avatar, background_name, colour_text, colour_text_border, colour_image_border, colour_profile_border, image_alignment, image_theme, image_message, image_profile_border_type)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
ON CONFLICT(guild_id) DO UPDATE
    SET toggle_enabled = EXCLUDED.toggle_enabled,
        toggle_image_border = EXCLUDED.toggle_image_border,
        toggle_show_avatar = EXCLUDED.toggle_show_avatar,
        background_name = EXCLUDED.background_name,
        colour_text = EXCLUDED.colour_text,
        colour_text_border = EXCLUDED.colour_text_border,
        colour_image_border = EXCLUDED.colour_image_border,
        colour_profile_border = EXCLUDED.colour_profile_border,
        image_alignment = EXCLUDED.image_alignment,
        image_theme = EXCLUDED.image_theme,
        image_message = EXCLUDED.image_message,
        image_profile_border_type = EXCLUDED.image_profile_border_type
RETURNING
    guild_id, toggle_enabled, toggle_image_border, toggle_show_avatar, background_name, colour_text, colour_text_border, colour_image_border, colour_profile_border, image_alignment, image_theme, image_message, image_profile_border_type
`

type CreateOrUpdateWelcomerImagesGuildSettingsParams struct {
	GuildID                int64  `json:"guild_id"`
	ToggleEnabled          bool   `json:"toggle_enabled"`
	ToggleImageBorder      bool   `json:"toggle_image_border"`
	ToggleShowAvatar       bool   `json:"toggle_show_avatar"`
	BackgroundName         string `json:"background_name"`
	ColourText             string `json:"colour_text"`
	ColourTextBorder       string `json:"colour_text_border"`
	ColourImageBorder      string `json:"colour_image_border"`
	ColourProfileBorder    string `json:"colour_profile_border"`
	ImageAlignment         int32  `json:"image_alignment"`
	ImageTheme             int32  `json:"image_theme"`
	ImageMessage           string `json:"image_message"`
	ImageProfileBorderType int32  `json:"image_profile_border_type"`
}

func (q *Queries) CreateOrUpdateWelcomerImagesGuildSettings(ctx context.Context, arg CreateOrUpdateWelcomerImagesGuildSettingsParams) (*GuildSettingsWelcomerImages, error) {
	row := q.db.QueryRow(ctx, CreateOrUpdateWelcomerImagesGuildSettings,
		arg.GuildID,
		arg.ToggleEnabled,
		arg.ToggleImageBorder,
		arg.ToggleShowAvatar,
		arg.BackgroundName,
		arg.ColourText,
		arg.ColourTextBorder,
		arg.ColourImageBorder,
		arg.ColourProfileBorder,
		arg.ImageAlignment,
		arg.ImageTheme,
		arg.ImageMessage,
		arg.ImageProfileBorderType,
	)
	var i GuildSettingsWelcomerImages
	err := row.Scan(
		&i.GuildID,
		&i.ToggleEnabled,
		&i.ToggleImageBorder,
		&i.ToggleShowAvatar,
		&i.BackgroundName,
		&i.ColourText,
		&i.ColourTextBorder,
		&i.ColourImageBorder,
		&i.ColourProfileBorder,
		&i.ImageAlignment,
		&i.ImageTheme,
		&i.ImageMessage,
		&i.ImageProfileBorderType,
	)
	return &i, err
}

const CreateWelcomerImagesGuildSettings = `-- name: CreateWelcomerImagesGuildSettings :one
INSERT INTO guild_settings_welcomer_images (guild_id, toggle_enabled, toggle_image_border, toggle_show_avatar, background_name, colour_text, colour_text_border, colour_image_border, colour_profile_border, image_alignment, image_theme, image_message, image_profile_border_type)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING
    guild_id, toggle_enabled, toggle_image_border, toggle_show_avatar, background_name, colour_text, colour_text_border, colour_image_border, colour_profile_border, image_alignment, image_theme, image_message, image_profile_border_type
`

type CreateWelcomerImagesGuildSettingsParams struct {
	GuildID                int64  `json:"guild_id"`
	ToggleEnabled          bool   `json:"toggle_enabled"`
	ToggleImageBorder      bool   `json:"toggle_image_border"`
	ToggleShowAvatar       bool   `json:"toggle_show_avatar"`
	BackgroundName         string `json:"background_name"`
	ColourText             string `json:"colour_text"`
	ColourTextBorder       string `json:"colour_text_border"`
	ColourImageBorder      string `json:"colour_image_border"`
	ColourProfileBorder    string `json:"colour_profile_border"`
	ImageAlignment         int32  `json:"image_alignment"`
	ImageTheme             int32  `json:"image_theme"`
	ImageMessage           string `json:"image_message"`
	ImageProfileBorderType int32  `json:"image_profile_border_type"`
}

func (q *Queries) CreateWelcomerImagesGuildSettings(ctx context.Context, arg CreateWelcomerImagesGuildSettingsParams) (*GuildSettingsWelcomerImages, error) {
	row := q.db.QueryRow(ctx, CreateWelcomerImagesGuildSettings,
		arg.GuildID,
		arg.ToggleEnabled,
		arg.ToggleImageBorder,
		arg.ToggleShowAvatar,
		arg.BackgroundName,
		arg.ColourText,
		arg.ColourTextBorder,
		arg.ColourImageBorder,
		arg.ColourProfileBorder,
		arg.ImageAlignment,
		arg.ImageTheme,
		arg.ImageMessage,
		arg.ImageProfileBorderType,
	)
	var i GuildSettingsWelcomerImages
	err := row.Scan(
		&i.GuildID,
		&i.ToggleEnabled,
		&i.ToggleImageBorder,
		&i.ToggleShowAvatar,
		&i.BackgroundName,
		&i.ColourText,
		&i.ColourTextBorder,
		&i.ColourImageBorder,
		&i.ColourProfileBorder,
		&i.ImageAlignment,
		&i.ImageTheme,
		&i.ImageMessage,
		&i.ImageProfileBorderType,
	)
	return &i, err
}

const GetWelcomerImagesGuildSettings = `-- name: GetWelcomerImagesGuildSettings :one
SELECT
    guild_id, toggle_enabled, toggle_image_border, toggle_show_avatar, background_name, colour_text, colour_text_border, colour_image_border, colour_profile_border, image_alignment, image_theme, image_message, image_profile_border_type
FROM
    guild_settings_welcomer_images
WHERE
    guild_id = $1
`

func (q *Queries) GetWelcomerImagesGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsWelcomerImages, error) {
	row := q.db.QueryRow(ctx, GetWelcomerImagesGuildSettings, guildID)
	var i GuildSettingsWelcomerImages
	err := row.Scan(
		&i.GuildID,
		&i.ToggleEnabled,
		&i.ToggleImageBorder,
		&i.ToggleShowAvatar,
		&i.BackgroundName,
		&i.ColourText,
		&i.ColourTextBorder,
		&i.ColourImageBorder,
		&i.ColourProfileBorder,
		&i.ImageAlignment,
		&i.ImageTheme,
		&i.ImageMessage,
		&i.ImageProfileBorderType,
	)
	return &i, err
}

const UpdateWelcomerImagesGuildSettings = `-- name: UpdateWelcomerImagesGuildSettings :execrows
UPDATE
    guild_settings_welcomer_images
SET
    toggle_enabled = $2,
    toggle_image_border = $3,
    toggle_show_avatar = $4,
    background_name = $5,
    colour_text = $6,
    colour_text_border = $7,
    colour_image_border = $8,
    colour_profile_border = $9,
    image_alignment = $10,
    image_theme = $11,
    image_message = $12,
    image_profile_border_type = $13
WHERE
    guild_id = $1
`

type UpdateWelcomerImagesGuildSettingsParams struct {
	GuildID                int64  `json:"guild_id"`
	ToggleEnabled          bool   `json:"toggle_enabled"`
	ToggleImageBorder      bool   `json:"toggle_image_border"`
	ToggleShowAvatar       bool   `json:"toggle_show_avatar"`
	BackgroundName         string `json:"background_name"`
	ColourText             string `json:"colour_text"`
	ColourTextBorder       string `json:"colour_text_border"`
	ColourImageBorder      string `json:"colour_image_border"`
	ColourProfileBorder    string `json:"colour_profile_border"`
	ImageAlignment         int32  `json:"image_alignment"`
	ImageTheme             int32  `json:"image_theme"`
	ImageMessage           string `json:"image_message"`
	ImageProfileBorderType int32  `json:"image_profile_border_type"`
}

func (q *Queries) UpdateWelcomerImagesGuildSettings(ctx context.Context, arg UpdateWelcomerImagesGuildSettingsParams) (int64, error) {
	result, err := q.db.Exec(ctx, UpdateWelcomerImagesGuildSettings,
		arg.GuildID,
		arg.ToggleEnabled,
		arg.ToggleImageBorder,
		arg.ToggleShowAvatar,
		arg.BackgroundName,
		arg.ColourText,
		arg.ColourTextBorder,
		arg.ColourImageBorder,
		arg.ColourProfileBorder,
		arg.ImageAlignment,
		arg.ImageTheme,
		arg.ImageMessage,
		arg.ImageProfileBorderType,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
