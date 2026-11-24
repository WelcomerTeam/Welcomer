package database

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type benchmarkSchemaTestDB struct {
	statements []string
	failAt     int
	err        error
}

func (db *benchmarkSchemaTestDB) Exec(_ context.Context, sql string, _ ...interface{}) (pgconn.CommandTag, error) {
	db.statements = append(db.statements, sql)
	if len(db.statements) == db.failAt {
		return nil, db.err
	}
	return nil, nil
}

func (db *benchmarkSchemaTestDB) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	panic("unexpected Query")
}

func (db *benchmarkSchemaTestDB) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	panic("unexpected QueryRow")
}

func (db *benchmarkSchemaTestDB) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	panic("unexpected CopyFrom")
}

func TestEnsureGuildAnalyticsBenchmarkSchema(t *testing.T) {
	db := &benchmarkSchemaTestDB{}
	if err := New(db).EnsureGuildAnalyticsBenchmarkSchema(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(db.statements) != 2 ||
		!strings.Contains(db.statements[0], "statistic text NOT NULL DEFAULT 'median'") ||
		!strings.Contains(db.statements[1], "ADD COLUMN IF NOT EXISTS statistic text NOT NULL DEFAULT 'median'") {
		t.Fatalf("unexpected schema statements: %v", db.statements)
	}

	wantErr := errors.New("schema failure")
	for _, failAt := range []int{1, 2} {
		db := &benchmarkSchemaTestDB{failAt: failAt, err: wantErr}
		err := New(db).EnsureGuildAnalyticsBenchmarkSchema(context.Background())
		if !errors.Is(err, wantErr) || len(db.statements) != failAt {
			t.Errorf("failure at statement %d: got error %v and %d statements", failAt, err, len(db.statements))
		}
	}
}

func TestGuildAnalyticsBenchmarkQueriesUseP90(t *testing.T) {
	if !strings.Contains(GetGuildAnalyticsBenchmarkBucket, "AND statistic = 'p90'") {
		t.Fatal("benchmark query must exclude legacy median rows")
	}
	if !strings.Contains(UpsertGuildAnalyticsBenchmark, "'p90'") ||
		!strings.Contains(UpsertGuildAnalyticsBenchmark, "statistic = EXCLUDED.statistic") {
		t.Fatal("benchmark upsert must mark new and updated rows as p90")
	}
}
