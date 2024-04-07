// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: guilds_query.sql

package database

import (
	"context"
)

const CreateGuild = `-- name: CreateGuild :one
INSERT INTO guilds (guild_id, embed_colour, site_splash_url, site_staff_visible, site_guild_visible, site_allow_invites)
    VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    guild_id, embed_colour, site_splash_url, site_staff_visible, site_guild_visible, site_allow_invites
`

type CreateGuildParams struct {
	GuildID          int64  `json:"guild_id"`
	EmbedColour      int32  `json:"embed_colour"`
	SiteSplashUrl    string `json:"site_splash_url"`
	SiteStaffVisible bool   `json:"site_staff_visible"`
	SiteGuildVisible bool   `json:"site_guild_visible"`
	SiteAllowInvites bool   `json:"site_allow_invites"`
}

func (q *Queries) CreateGuild(ctx context.Context, arg *CreateGuildParams) (*Guilds, error) {
	row := q.db.QueryRow(ctx, CreateGuild,
		arg.GuildID,
		arg.EmbedColour,
		arg.SiteSplashUrl,
		arg.SiteStaffVisible,
		arg.SiteGuildVisible,
		arg.SiteAllowInvites,
	)
	var i Guilds
	err := row.Scan(
		&i.GuildID,
		&i.EmbedColour,
		&i.SiteSplashUrl,
		&i.SiteStaffVisible,
		&i.SiteGuildVisible,
		&i.SiteAllowInvites,
	)
	return &i, err
}

const CreateOrUpdateGuild = `-- name: CreateOrUpdateGuild :one
INSERT INTO guilds (guild_id, embed_colour, site_splash_url, site_staff_visible, site_guild_visible, site_allow_invites)
    VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT(guild_id) DO UPDATE
    SET embed_colour = EXCLUDED.embed_colour,
        site_splash_url = EXCLUDED.site_splash_url,
        site_staff_visible = EXCLUDED.site_staff_visible,
        site_guild_visible = EXCLUDED.site_guild_visible,
        site_allow_invites = EXCLUDED.site_allow_invites
RETURNING
    guild_id, embed_colour, site_splash_url, site_staff_visible, site_guild_visible, site_allow_invites
`

type CreateOrUpdateGuildParams struct {
	GuildID          int64  `json:"guild_id"`
	EmbedColour      int32  `json:"embed_colour"`
	SiteSplashUrl    string `json:"site_splash_url"`
	SiteStaffVisible bool   `json:"site_staff_visible"`
	SiteGuildVisible bool   `json:"site_guild_visible"`
	SiteAllowInvites bool   `json:"site_allow_invites"`
}

func (q *Queries) CreateOrUpdateGuild(ctx context.Context, arg *CreateOrUpdateGuildParams) (*Guilds, error) {
	row := q.db.QueryRow(ctx, CreateOrUpdateGuild,
		arg.GuildID,
		arg.EmbedColour,
		arg.SiteSplashUrl,
		arg.SiteStaffVisible,
		arg.SiteGuildVisible,
		arg.SiteAllowInvites,
	)
	var i Guilds
	err := row.Scan(
		&i.GuildID,
		&i.EmbedColour,
		&i.SiteSplashUrl,
		&i.SiteStaffVisible,
		&i.SiteGuildVisible,
		&i.SiteAllowInvites,
	)
	return &i, err
}

const GetGuild = `-- name: GetGuild :one
SELECT
    guild_id, embed_colour, site_splash_url, site_staff_visible, site_guild_visible, site_allow_invites
FROM
    guilds
WHERE
    guild_id = $1
`

func (q *Queries) GetGuild(ctx context.Context, guildID int64) (*Guilds, error) {
	row := q.db.QueryRow(ctx, GetGuild, guildID)
	var i Guilds
	err := row.Scan(
		&i.GuildID,
		&i.EmbedColour,
		&i.SiteSplashUrl,
		&i.SiteStaffVisible,
		&i.SiteGuildVisible,
		&i.SiteAllowInvites,
	)
	return &i, err
}

const UpdateGuild = `-- name: UpdateGuild :execrows
UPDATE
    guilds
SET
    embed_colour = $2,
    site_splash_url = $3,
    site_staff_visible = $4,
    site_guild_visible = $5,
    site_allow_invites = $6
WHERE
    guild_id = $1
`

type UpdateGuildParams struct {
	GuildID          int64  `json:"guild_id"`
	EmbedColour      int32  `json:"embed_colour"`
	SiteSplashUrl    string `json:"site_splash_url"`
	SiteStaffVisible bool   `json:"site_staff_visible"`
	SiteGuildVisible bool   `json:"site_guild_visible"`
	SiteAllowInvites bool   `json:"site_allow_invites"`
}

func (q *Queries) UpdateGuild(ctx context.Context, arg *UpdateGuildParams) (int64, error) {
	result, err := q.db.Exec(ctx, UpdateGuild,
		arg.GuildID,
		arg.EmbedColour,
		arg.SiteSplashUrl,
		arg.SiteStaffVisible,
		arg.SiteGuildVisible,
		arg.SiteAllowInvites,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
