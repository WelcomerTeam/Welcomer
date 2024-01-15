// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1
// source: guild_borderwall_requests_query.sql

package database

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

const CreateBorderwallRequest = `-- name: CreateBorderwallRequest :one
INSERT INTO borderwall_requests (request_uuid, created_at, updated_at, guild_id, user_id, is_verified, verified_at)
    VALUES (uuid_generate_v7(), now(), now(), $1, $2, $3, $4)
RETURNING
    request_uuid, created_at, updated_at, guild_id, user_id, is_verified, verified_at
`

type CreateBorderwallRequestParams struct {
	GuildID    int64     `json:"guild_id"`
	UserID     int64     `json:"user_id"`
	IsVerified bool      `json:"is_verified"`
	VerifiedAt time.Time `json:"verified_at"`
}

func (q *Queries) CreateBorderwallRequest(ctx context.Context, arg *CreateBorderwallRequestParams) (*BorderwallRequests, error) {
	row := q.db.QueryRow(ctx, CreateBorderwallRequest,
		arg.GuildID,
		arg.UserID,
		arg.IsVerified,
		arg.VerifiedAt,
	)
	var i BorderwallRequests
	err := row.Scan(
		&i.RequestUuid,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.GuildID,
		&i.UserID,
		&i.IsVerified,
		&i.VerifiedAt,
	)
	return &i, err
}

const CreateOrUpdateBorderwallRequest = `-- name: CreateOrUpdateBorderwallRequest :one
INSERT INTO borderwall_requests (request_uuid, created_at, updated_at, guild_id, user_id, is_verified, verified_at)
    VALUES (uuid_generate_v7(), now(), now(), $1, $2, $3, $4)
ON CONFLICT(request_uuid) DO UPDATE
    SET updated_at = EXCLUDED.updated_at,
        guild_id = EXCLUDED.guild_id,
        is_verified = EXCLUDED.is_verified,
        verified_at = EXCLUDED.verified_at
RETURNING
    request_uuid, created_at, updated_at, guild_id, user_id, is_verified, verified_at
`

type CreateOrUpdateBorderwallRequestParams struct {
	GuildID    int64     `json:"guild_id"`
	UserID     int64     `json:"user_id"`
	IsVerified bool      `json:"is_verified"`
	VerifiedAt time.Time `json:"verified_at"`
}

func (q *Queries) CreateOrUpdateBorderwallRequest(ctx context.Context, arg *CreateOrUpdateBorderwallRequestParams) (*BorderwallRequests, error) {
	row := q.db.QueryRow(ctx, CreateOrUpdateBorderwallRequest,
		arg.GuildID,
		arg.UserID,
		arg.IsVerified,
		arg.VerifiedAt,
	)
	var i BorderwallRequests
	err := row.Scan(
		&i.RequestUuid,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.GuildID,
		&i.UserID,
		&i.IsVerified,
		&i.VerifiedAt,
	)
	return &i, err
}

const GetBorderwallRequest = `-- name: GetBorderwallRequest :one
SELECT
    request_uuid, created_at, updated_at, guild_id, user_id, is_verified, verified_at
FROM
    borderwall_requests
WHERE
    request_uuid = $1
`

func (q *Queries) GetBorderwallRequest(ctx context.Context, requestUuid uuid.UUID) (*BorderwallRequests, error) {
	row := q.db.QueryRow(ctx, GetBorderwallRequest, requestUuid)
	var i BorderwallRequests
	err := row.Scan(
		&i.RequestUuid,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.GuildID,
		&i.UserID,
		&i.IsVerified,
		&i.VerifiedAt,
	)
	return &i, err
}

const GetBorderwallRequestsByGuildIDUserID = `-- name: GetBorderwallRequestsByGuildIDUserID :many
SELECT
    request_uuid, created_at, updated_at, guild_id, user_id, is_verified, verified_at
FROM
    borderwall_requests
WHERE
    guild_id = $1
    AND user_id = $2
`

type GetBorderwallRequestsByGuildIDUserIDParams struct {
	GuildID int64 `json:"guild_id"`
	UserID  int64 `json:"user_id"`
}

func (q *Queries) GetBorderwallRequestsByGuildIDUserID(ctx context.Context, arg *GetBorderwallRequestsByGuildIDUserIDParams) ([]*BorderwallRequests, error) {
	rows, err := q.db.Query(ctx, GetBorderwallRequestsByGuildIDUserID, arg.GuildID, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*BorderwallRequests{}
	for rows.Next() {
		var i BorderwallRequests
		if err := rows.Scan(
			&i.RequestUuid,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.GuildID,
			&i.UserID,
			&i.IsVerified,
			&i.VerifiedAt,
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

const UpdateBorderwallRequest = `-- name: UpdateBorderwallRequest :execrows
UPDATE
    borderwall_requests
SET
    updated_at = now(),
    guild_id = $2,
    is_verified = $3,
    verified_at = $4
WHERE
    request_uuid = $1
`

type UpdateBorderwallRequestParams struct {
	RequestUuid uuid.UUID `json:"request_uuid"`
	GuildID     int64     `json:"guild_id"`
	IsVerified  bool      `json:"is_verified"`
	VerifiedAt  time.Time `json:"verified_at"`
}

func (q *Queries) UpdateBorderwallRequest(ctx context.Context, arg *UpdateBorderwallRequestParams) (int64, error) {
	result, err := q.db.Exec(ctx, UpdateBorderwallRequest,
		arg.RequestUuid,
		arg.GuildID,
		arg.IsVerified,
		arg.VerifiedAt,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
