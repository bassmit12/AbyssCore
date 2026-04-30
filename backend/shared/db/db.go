package db

import (
	"context"
	"fmt"
	"os"

	"encore.dev/storage/sqldb"
)

// Each service declares its own database via encore's sqldb package.
// The DB name maps to the schema in Postgres.
// This file documents the pattern - each service has its own db.go declaring:
//
//   var DB = sqldb.NewDatabase("service_name", sqldb.DatabaseConfig{
//       Migrations: "./migrations",
//   })
//
// Connection string comes from ENCORE_DB_URL env or Encore's built-in provisioning.

func DSN(schema string) string {
	base := os.Getenv("DATABASE_URL")
	if base == "" {
		base = "postgres://abysscore:abysscore@localhost:5432/abysscore"
	}
	return fmt.Sprintf("%s?search_path=%s&sslmode=disable", base, schema)
}

// QueryContext is a helper interface both *sql.DB and *sql.Tx satisfy.
type QueryContext interface {
	QueryContext(ctx context.Context, query string, args ...any) (sqldb.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sqldb.Row
	ExecContext(ctx context.Context, query string, args ...any) (sqldb.ExecResult, error)
}
