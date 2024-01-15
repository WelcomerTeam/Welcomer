// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1
// source: user_transactions_query.sql

package database

import (
	"context"

	"github.com/gofrs/uuid"
)

const CreateOrUpdateUserTransaction = `-- name: CreateOrUpdateUserTransaction :one
INSERT INTO user_transactions (transaction_uuid, created_at, updated_at, user_id, platform_type, transaction_id, transaction_status, currency_code, amount)
    VALUES (uuid_generate_v7(), now(), now(), $1, $2, $3, $4, $5, $6)
ON CONFLICT(transaction_uuid) DO UPDATE
    SET user_id = EXCLUDED.user_id,
        platform_type = EXCLUDED.platform_type,
        transaction_id = EXCLUDED.transaction_id,
        transaction_status = EXCLUDED.transaction_status,
        currency_code = EXCLUDED.currency_code,
        amount = EXCLUDED.amount,
        updated_at = EXCLUDED.updated_at
RETURNING
    transaction_uuid, created_at, updated_at, user_id, platform_type, transaction_id, transaction_status, currency_code, amount
`

type CreateOrUpdateUserTransactionParams struct {
	UserID            int64  `json:"user_id"`
	PlatformType      int32  `json:"platform_type"`
	TransactionID     string `json:"transaction_id"`
	TransactionStatus int32  `json:"transaction_status"`
	CurrencyCode      string `json:"currency_code"`
	Amount            int32  `json:"amount"`
}

func (q *Queries) CreateOrUpdateUserTransaction(ctx context.Context, arg *CreateOrUpdateUserTransactionParams) (*UserTransactions, error) {
	row := q.db.QueryRow(ctx, CreateOrUpdateUserTransaction,
		arg.UserID,
		arg.PlatformType,
		arg.TransactionID,
		arg.TransactionStatus,
		arg.CurrencyCode,
		arg.Amount,
	)
	var i UserTransactions
	err := row.Scan(
		&i.TransactionUuid,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
		&i.PlatformType,
		&i.TransactionID,
		&i.TransactionStatus,
		&i.CurrencyCode,
		&i.Amount,
	)
	return &i, err
}

const CreateUserTransaction = `-- name: CreateUserTransaction :one
INSERT INTO user_transactions (transaction_uuid, created_at, updated_at, user_id, platform_type, transaction_id, transaction_status, currency_code, amount)
    VALUES (uuid_generate_v7(), now(), now(), $1, $2, $3, $4, $5, $6)
RETURNING
    transaction_uuid, created_at, updated_at, user_id, platform_type, transaction_id, transaction_status, currency_code, amount
`

type CreateUserTransactionParams struct {
	UserID            int64  `json:"user_id"`
	PlatformType      int32  `json:"platform_type"`
	TransactionID     string `json:"transaction_id"`
	TransactionStatus int32  `json:"transaction_status"`
	CurrencyCode      string `json:"currency_code"`
	Amount            int32  `json:"amount"`
}

func (q *Queries) CreateUserTransaction(ctx context.Context, arg *CreateUserTransactionParams) (*UserTransactions, error) {
	row := q.db.QueryRow(ctx, CreateUserTransaction,
		arg.UserID,
		arg.PlatformType,
		arg.TransactionID,
		arg.TransactionStatus,
		arg.CurrencyCode,
		arg.Amount,
	)
	var i UserTransactions
	err := row.Scan(
		&i.TransactionUuid,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
		&i.PlatformType,
		&i.TransactionID,
		&i.TransactionStatus,
		&i.CurrencyCode,
		&i.Amount,
	)
	return &i, err
}

const GetUserTransaction = `-- name: GetUserTransaction :one
SELECT
    transaction_uuid, created_at, updated_at, user_id, platform_type, transaction_id, transaction_status, currency_code, amount
FROM
    user_transactions
WHERE
    transaction_uuid = $1
`

func (q *Queries) GetUserTransaction(ctx context.Context, transactionUuid uuid.UUID) (*UserTransactions, error) {
	row := q.db.QueryRow(ctx, GetUserTransaction, transactionUuid)
	var i UserTransactions
	err := row.Scan(
		&i.TransactionUuid,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
		&i.PlatformType,
		&i.TransactionID,
		&i.TransactionStatus,
		&i.CurrencyCode,
		&i.Amount,
	)
	return &i, err
}

const GetUserTransactionsByUserID = `-- name: GetUserTransactionsByUserID :many
SELECT
    transaction_uuid, created_at, updated_at, user_id, platform_type, transaction_id, transaction_status, currency_code, amount
FROM
    user_transactions
WHERE
    user_id = $1
`

func (q *Queries) GetUserTransactionsByUserID(ctx context.Context, userID int64) ([]*UserTransactions, error) {
	rows, err := q.db.Query(ctx, GetUserTransactionsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*UserTransactions{}
	for rows.Next() {
		var i UserTransactions
		if err := rows.Scan(
			&i.TransactionUuid,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.UserID,
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

const UpdateUserTransaction = `-- name: UpdateUserTransaction :execrows
UPDATE
    user_transactions
SET
    user_id = $2,
    platform_type = $3,
    transaction_id = $4,
    transaction_status = $5,
    currency_code = $6,
    amount = $7,
    updated_at = now()
WHERE
    transaction_id = $1
`

type UpdateUserTransactionParams struct {
	TransactionID     string `json:"transaction_id"`
	UserID            int64  `json:"user_id"`
	PlatformType      int32  `json:"platform_type"`
	TransactionID_2   string `json:"transaction_id_2"`
	TransactionStatus int32  `json:"transaction_status"`
	CurrencyCode      string `json:"currency_code"`
	Amount            int32  `json:"amount"`
}

func (q *Queries) UpdateUserTransaction(ctx context.Context, arg *UpdateUserTransactionParams) (int64, error) {
	result, err := q.db.Exec(ctx, UpdateUserTransaction,
		arg.TransactionID,
		arg.UserID,
		arg.PlatformType,
		arg.TransactionID_2,
		arg.TransactionStatus,
		arg.CurrencyCode,
		arg.Amount,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
