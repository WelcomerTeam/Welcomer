package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Database struct {
	ScienceCommandUsages *ScienceCommandUsages
	ScienceCommandErrors *ScienceCommandErrors
}

func NewDatabase(pool *pgxpool.Pool) (database *Database) {
	return &Database{
		ScienceCommandUsages: newScienceCommandUsages(pool),
		ScienceCommandErrors: newScienceCommandErrors(pool),
	}
}

func (d *Database) CreateTables(ctx context.Context, pool *pgxpool.Pool) {
	// Import uuid functions to postgres.
	_, err := pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if err != nil {
		panic(fmt.Sprintf("pool.Exec('CREATE EXTENSIoN uuid-ossp'): %v", err.Error()))
	}

	// Setup tables.
	mustCreate(ctx, pool,
		d.ScienceCommandUsages,
		d.ScienceCommandErrors,
	)
}

func mustCreate(ctx context.Context, pool *pgxpool.Pool, tables ...Table) {
	for _, table := range tables {
		_, err := pool.Exec(ctx, table.Schema())
		if err != nil {
			panic(fmt.Sprintf("mustCreate(): %v", err.Error()))
		}
	}
}
