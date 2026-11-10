package database

import (
	"context"
	_ "embed"
)

// guildAnalyticsBenchmarksSchema is also the schema SQLC uses for generated queries.
//
//go:embed schema/guild_analytics_benchmarks_schema.sql
var guildAnalyticsBenchmarksSchema string

// EnsureGuildAnalyticsBenchmarkSchema creates or updates the benchmark table.
func (q *Queries) EnsureGuildAnalyticsBenchmarkSchema(ctx context.Context) error {
	if _, err := q.db.Exec(ctx, guildAnalyticsBenchmarksSchema); err != nil {
		return err
	}
	_, err := q.db.Exec(ctx, "ALTER TABLE guild_analytics_benchmarks ADD COLUMN IF NOT EXISTS statistic text NOT NULL DEFAULT 'median'")
	return err
}
