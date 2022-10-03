// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: user_memberships_query.sql

package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
)

const CreateNewMembership = `-- name: CreateNewMembership :one
INSERT INTO user_memberships (membership_uuid, created_at, updated_at, started_at, expires_at, status, membership_type, transaction_uuid, user_id, guild_id)
    VALUES (uuid_generate_v4(), now(), now(), $1, $2, $3, $4, $5, $6, $7)
RETURNING
    membership_uuid, created_at, updated_at, started_at, expires_at, status, membership_type, transaction_uuid, user_id, guild_id
`

type CreateNewMembershipParams struct {
	StartedAt       time.Time `json:"started_at"`
	ExpiresAt       time.Time `json:"expires_at"`
	Status          int32     `json:"status"`
	MembershipType  int32     `json:"membership_type"`
	TransactionUuid uuid.UUID `json:"transaction_uuid"`
	UserID          int64     `json:"user_id"`
	GuildID         int64     `json:"guild_id"`
}

func (q *Queries) CreateNewMembership(ctx context.Context, arg *CreateNewMembershipParams) (*UserMemberships, error) {
	row := q.db.QueryRow(ctx, CreateNewMembership,
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

const CreateOrUpdateNewMembership = `-- name: CreateOrUpdateNewMembership :one
INSERT INTO user_memberships (membership_uuid, created_at, updated_at, started_at, expires_at, status, membership_type, transaction_uuid, user_id, guild_id)
    VALUES (uuid_generate_v4(), now(), now(), $1, $2, $3, $4, $5, $6, $7)
ON CONFLICT(membership_uuid) DO UPDATE
    SET started_at = EXCLUDED.started_at,
        expires_at = EXCLUDED.expires_at,
        status = EXCLUDED.status,
        transaction_uuid = EXCLUDED.transaction_uuid,
        user_id = EXCLUDED.user_id,
        guild_id = EXCLUDED.guild_id,
        updated_at = EXCLUDED.updated_at
RETURNING
    membership_uuid, created_at, updated_at, started_at, expires_at, status, membership_type, transaction_uuid, user_id, guild_id
`

type CreateOrUpdateNewMembershipParams struct {
	StartedAt       time.Time `json:"started_at"`
	ExpiresAt       time.Time `json:"expires_at"`
	Status          int32     `json:"status"`
	MembershipType  int32     `json:"membership_type"`
	TransactionUuid uuid.UUID `json:"transaction_uuid"`
	UserID          int64     `json:"user_id"`
	GuildID         int64     `json:"guild_id"`
}

func (q *Queries) CreateOrUpdateNewMembership(ctx context.Context, arg *CreateOrUpdateNewMembershipParams) (*UserMemberships, error) {
	row := q.db.QueryRow(ctx, CreateOrUpdateNewMembership,
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
    membership_uuid, user_memberships.created_at, user_memberships.updated_at, started_at, expires_at, status, membership_type, user_memberships.transaction_uuid, user_memberships.user_id, guild_id, user_transactions.transaction_uuid, user_transactions.created_at, user_transactions.updated_at, user_transactions.user_id, platform_type, transaction_id, transaction_status, currency_code, amount
FROM
    user_memberships
    LEFT JOIN user_transactions ON (user_memberships.transaction_uuid = user_transactions.transaction_uuid)
WHERE
    user_memberships.membership_uuid = $1
`

type GetUserMembershipRow struct {
	MembershipUuid    uuid.UUID      `json:"membership_uuid"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	StartedAt         time.Time      `json:"started_at"`
	ExpiresAt         time.Time      `json:"expires_at"`
	Status            int32          `json:"status"`
	MembershipType    int32          `json:"membership_type"`
	TransactionUuid   uuid.UUID      `json:"transaction_uuid"`
	UserID            int64          `json:"user_id"`
	GuildID           int64          `json:"guild_id"`
	TransactionUuid_2 uuid.NullUUID  `json:"transaction_uuid_2"`
	CreatedAt_2       sql.NullTime   `json:"created_at_2"`
	UpdatedAt_2       sql.NullTime   `json:"updated_at_2"`
	UserID_2          sql.NullInt64  `json:"user_id_2"`
	PlatformType      sql.NullInt32  `json:"platform_type"`
	TransactionID     sql.NullString `json:"transaction_id"`
	TransactionStatus sql.NullInt32  `json:"transaction_status"`
	CurrencyCode      sql.NullString `json:"currency_code"`
	Amount            sql.NullInt32  `json:"amount"`
}

func (q *Queries) GetUserMembership(ctx context.Context, membershipUuid uuid.UUID) (*GetUserMembershipRow, error) {
	row := q.db.QueryRow(ctx, GetUserMembership, membershipUuid)
	var i GetUserMembershipRow
	err := row.Scan(
		&i.MembershipUuid,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.StartedAt,
		&i.ExpiresAt,
		&i.Status,
		&i.MembershipType,
		&i.TransactionUuid,
		&i.UserID,
		&i.GuildID,
		&i.TransactionUuid_2,
		&i.CreatedAt_2,
		&i.UpdatedAt_2,
		&i.UserID_2,
		&i.PlatformType,
		&i.TransactionID,
		&i.TransactionStatus,
		&i.CurrencyCode,
		&i.Amount,
	)
	return &i, err
}

const GetUserMembershipsByGuildID = `-- name: GetUserMembershipsByGuildID :many
SELECT
    membership_uuid, user_memberships.created_at, user_memberships.updated_at, started_at, expires_at, status, membership_type, user_memberships.transaction_uuid, user_memberships.user_id, guild_id, user_transactions.transaction_uuid, user_transactions.created_at, user_transactions.updated_at, user_transactions.user_id, platform_type, transaction_id, transaction_status, currency_code, amount
FROM
    user_memberships
    LEFT JOIN user_transactions ON (user_memberships.transaction_uuid = user_transactions.transaction_uuid)
WHERE
    user_memberships.guild_id = $1
`

type GetUserMembershipsByGuildIDRow struct {
	MembershipUuid    uuid.UUID      `json:"membership_uuid"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	StartedAt         time.Time      `json:"started_at"`
	ExpiresAt         time.Time      `json:"expires_at"`
	Status            int32          `json:"status"`
	MembershipType    int32          `json:"membership_type"`
	TransactionUuid   uuid.UUID      `json:"transaction_uuid"`
	UserID            int64          `json:"user_id"`
	GuildID           int64          `json:"guild_id"`
	TransactionUuid_2 uuid.NullUUID  `json:"transaction_uuid_2"`
	CreatedAt_2       sql.NullTime   `json:"created_at_2"`
	UpdatedAt_2       sql.NullTime   `json:"updated_at_2"`
	UserID_2          sql.NullInt64  `json:"user_id_2"`
	PlatformType      sql.NullInt32  `json:"platform_type"`
	TransactionID     sql.NullString `json:"transaction_id"`
	TransactionStatus sql.NullInt32  `json:"transaction_status"`
	CurrencyCode      sql.NullString `json:"currency_code"`
	Amount            sql.NullInt32  `json:"amount"`
}

func (q *Queries) GetUserMembershipsByGuildID(ctx context.Context, guildID int64) ([]*GetUserMembershipsByGuildIDRow, error) {
	rows, err := q.db.Query(ctx, GetUserMembershipsByGuildID, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*GetUserMembershipsByGuildIDRow{}
	for rows.Next() {
		var i GetUserMembershipsByGuildIDRow
		if err := rows.Scan(
			&i.MembershipUuid,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.StartedAt,
			&i.ExpiresAt,
			&i.Status,
			&i.MembershipType,
			&i.TransactionUuid,
			&i.UserID,
			&i.GuildID,
			&i.TransactionUuid_2,
			&i.CreatedAt_2,
			&i.UpdatedAt_2,
			&i.UserID_2,
			&i.PlatformType,
			&i.TransactionID,
			&i.TransactionStatus,
			&i.CurrencyCode,
			&i.Amount,
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
    membership_uuid, user_memberships.created_at, user_memberships.updated_at, started_at, expires_at, status, membership_type, user_memberships.transaction_uuid, user_memberships.user_id, guild_id, user_transactions.transaction_uuid, user_transactions.created_at, user_transactions.updated_at, user_transactions.user_id, platform_type, transaction_id, transaction_status, currency_code, amount
FROM
    user_memberships
    LEFT JOIN user_transactions ON (user_memberships.transaction_uuid = user_transactions.transaction_uuid)
WHERE
    user_memberships.user_id = $1
`

type GetUserMembershipsByUserIDRow struct {
	MembershipUuid    uuid.UUID      `json:"membership_uuid"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	StartedAt         time.Time      `json:"started_at"`
	ExpiresAt         time.Time      `json:"expires_at"`
	Status            int32          `json:"status"`
	MembershipType    int32          `json:"membership_type"`
	TransactionUuid   uuid.UUID      `json:"transaction_uuid"`
	UserID            int64          `json:"user_id"`
	GuildID           int64          `json:"guild_id"`
	TransactionUuid_2 uuid.NullUUID  `json:"transaction_uuid_2"`
	CreatedAt_2       sql.NullTime   `json:"created_at_2"`
	UpdatedAt_2       sql.NullTime   `json:"updated_at_2"`
	UserID_2          sql.NullInt64  `json:"user_id_2"`
	PlatformType      sql.NullInt32  `json:"platform_type"`
	TransactionID     sql.NullString `json:"transaction_id"`
	TransactionStatus sql.NullInt32  `json:"transaction_status"`
	CurrencyCode      sql.NullString `json:"currency_code"`
	Amount            sql.NullInt32  `json:"amount"`
}

func (q *Queries) GetUserMembershipsByUserID(ctx context.Context, userID int64) ([]*GetUserMembershipsByUserIDRow, error) {
	rows, err := q.db.Query(ctx, GetUserMembershipsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*GetUserMembershipsByUserIDRow{}
	for rows.Next() {
		var i GetUserMembershipsByUserIDRow
		if err := rows.Scan(
			&i.MembershipUuid,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.StartedAt,
			&i.ExpiresAt,
			&i.Status,
			&i.MembershipType,
			&i.TransactionUuid,
			&i.UserID,
			&i.GuildID,
			&i.TransactionUuid_2,
			&i.CreatedAt_2,
			&i.UpdatedAt_2,
			&i.UserID_2,
			&i.PlatformType,
			&i.TransactionID,
			&i.TransactionStatus,
			&i.CurrencyCode,
			&i.Amount,
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
    started_at = $2,
    expires_at = $3,
    status = $4,
    transaction_uuid = $5,
    user_id = $6,
    guild_id = $7,
    updated_at = now()
WHERE
    membership_uuid = $1
`

type UpdateUserMembershipParams struct {
	MembershipUuid  uuid.UUID `json:"membership_uuid"`
	StartedAt       time.Time `json:"started_at"`
	ExpiresAt       time.Time `json:"expires_at"`
	Status          int32     `json:"status"`
	TransactionUuid uuid.UUID `json:"transaction_uuid"`
	UserID          int64     `json:"user_id"`
	GuildID         int64     `json:"guild_id"`
}

func (q *Queries) UpdateUserMembership(ctx context.Context, arg *UpdateUserMembershipParams) (int64, error) {
	result, err := q.db.Exec(ctx, UpdateUserMembership,
		arg.MembershipUuid,
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
