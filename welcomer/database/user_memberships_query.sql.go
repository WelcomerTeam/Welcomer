// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.14.0
// source: user_memberships_query.sql

package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
)

const CreateNewMembership = `-- name: CreateNewMembership :one
INSERT INTO user_memberships (membership_uuid, origin_membership_id, created_at, updated_at, started_at, expires_at, status, membership_type, transaction_uuid, user_id, guild_id)
    VALUES (uuid_generate_v4(), $1, now(), now(), $2, $3, $4, $5, $6, $7, $8)
RETURNING
    membership_uuid, origin_membership_id, created_at, updated_at, started_at, expires_at, status, membership_type, transaction_uuid, user_id, guild_id
`

type CreateNewMembershipParams struct {
	OriginMembershipID uuid.NullUUID `json:"origin_membership_id"`
	StartedAt          time.Time     `json:"started_at"`
	ExpiresAt          time.Time     `json:"expires_at"`
	Status             int32         `json:"status"`
	MembershipType     int32         `json:"membership_type"`
	TransactionUuid    uuid.NullUUID `json:"transaction_uuid"`
	UserID             int64         `json:"user_id"`
	GuildID            sql.NullInt64 `json:"guild_id"`
}

func (q *Queries) CreateNewMembership(ctx context.Context, arg *CreateNewMembershipParams) (*UserMemberships, error) {
	row := q.db.QueryRow(ctx, CreateNewMembership,
		arg.OriginMembershipID,
		arg.StartedAt,
		arg.ExpiresAt,
		arg.Status,
		arg.MembershipType,
		arg.TransactionUuid,
		arg.UserID,
		arg.GuildID,
	)
	var i UserMemberships
	err := row.Scan(
		&i.MembershipUuid,
		&i.OriginMembershipID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.StartedAt,
		&i.ExpiresAt,
		&i.Status,
		&i.MembershipType,
		&i.TransactionUuid,
		&i.UserID,
		&i.GuildID,
	)
	return &i, err
}

const GetUserMembership = `-- name: GetUserMembership :one
SELECT
    membership_uuid, origin_membership_id, created_at, updated_at, started_at, expires_at, status, membership_type, transaction_uuid, user_id, guild_id
FROM
    user_memberships
WHERE
    membership_uuid = $1
`

func (q *Queries) GetUserMembership(ctx context.Context, membershipUuid uuid.UUID) (*UserMemberships, error) {
	row := q.db.QueryRow(ctx, GetUserMembership, membershipUuid)
	var i UserMemberships
	err := row.Scan(
		&i.MembershipUuid,
		&i.OriginMembershipID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.StartedAt,
		&i.ExpiresAt,
		&i.Status,
		&i.MembershipType,
		&i.TransactionUuid,
		&i.UserID,
		&i.GuildID,
	)
	return &i, err
}

const GetUserMembershipsByGuildID = `-- name: GetUserMembershipsByGuildID :many
SELECT
    membership_uuid, origin_membership_id, created_at, updated_at, started_at, expires_at, status, membership_type, transaction_uuid, user_id, guild_id
FROM
    user_memberships
WHERE
    guild_id = $1
`

func (q *Queries) GetUserMembershipsByGuildID(ctx context.Context, guildID sql.NullInt64) ([]*UserMemberships, error) {
	rows, err := q.db.Query(ctx, GetUserMembershipsByGuildID, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*UserMemberships{}
	for rows.Next() {
		var i UserMemberships
		if err := rows.Scan(
			&i.MembershipUuid,
			&i.OriginMembershipID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.StartedAt,
			&i.ExpiresAt,
			&i.Status,
			&i.MembershipType,
			&i.TransactionUuid,
			&i.UserID,
			&i.GuildID,
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

const GetUserMembershipsByUserID = `-- name: GetUserMembershipsByUserID :many
SELECT
    membership_uuid, origin_membership_id, created_at, updated_at, started_at, expires_at, status, membership_type, transaction_uuid, user_id, guild_id
FROM
    user_memberships
WHERE
    user_id = $1
`

func (q *Queries) GetUserMembershipsByUserID(ctx context.Context, userID int64) ([]*UserMemberships, error) {
	rows, err := q.db.Query(ctx, GetUserMembershipsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*UserMemberships{}
	for rows.Next() {
		var i UserMemberships
		if err := rows.Scan(
			&i.MembershipUuid,
			&i.OriginMembershipID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.StartedAt,
			&i.ExpiresAt,
			&i.Status,
			&i.MembershipType,
			&i.TransactionUuid,
			&i.UserID,
			&i.GuildID,
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

const UpdateUserMembership = `-- name: UpdateUserMembership :execrows
UPDATE
    user_memberships
SET
    origin_membership_id = $2,
    started_at = $3,
    expires_at = $4,
    status = $5,
    transaction_uuid = $6,
    user_id = $7,
    guild_id = $8,
    updated_at = now()
WHERE
    membership_uuid = $1
`

type UpdateUserMembershipParams struct {
	MembershipUuid     uuid.UUID     `json:"membership_uuid"`
	OriginMembershipID uuid.NullUUID `json:"origin_membership_id"`
	StartedAt          time.Time     `json:"started_at"`
	ExpiresAt          time.Time     `json:"expires_at"`
	Status             int32         `json:"status"`
	TransactionUuid    uuid.NullUUID `json:"transaction_uuid"`
	UserID             int64         `json:"user_id"`
	GuildID            sql.NullInt64 `json:"guild_id"`
}

func (q *Queries) UpdateUserMembership(ctx context.Context, arg *UpdateUserMembershipParams) (int64, error) {
	result, err := q.db.Exec(ctx, UpdateUserMembership,
		arg.MembershipUuid,
		arg.OriginMembershipID,
		arg.StartedAt,
		arg.ExpiresAt,
		arg.Status,
		arg.TransactionUuid,
		arg.UserID,
		arg.GuildID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
