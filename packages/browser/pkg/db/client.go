package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"path/filepath"

	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/datakit-dev/dtkt-integrations/browser/pkg/db/ent/runtime"
	_ "github.com/ncruces/go-sqlite3/driver"

	"entgo.io/ent/dialect"
	"github.com/datakit-dev/dtkt-integrations/browser/pkg/db/ent"
)

const dbName = "browser.db"

func NewDB(dbDir string) (*sql.DB, error) {
	// Connection string with optimized pragmas for ncruces/go-sqlite3
	// Order matters: busy_timeout should be set first
	dsn := fmt.Sprintf(
		"file:%s?cache=shared&"+
			"_pragma=busy_timeout(30000)&"+ // 30 second timeout for long-running ops
			"_pragma=journal_mode(WAL)&"+ // Write-Ahead Logging for better concurrency
			"_pragma=synchronous(NORMAL)&"+ // Good balance of safety and performance with WAL
			"_pragma=foreign_keys(ON)&"+ // Enforce foreign key constraints
			"_pragma=temp_store(MEMORY)&"+ // Store temp tables in memory
			"_pragma=cache_size(-64000)&"+ // ~64MB cache (negative = KB)
			"_pragma=mmap_size(268435456)", // 256MB memory-mapped I/O
		filepath.Join(dbDir, dbName),
	)

	db, err := sql.Open(dialect.SQLite, dsn)
	if err != nil {
		return nil, err
	}

	// WAL mode allows concurrent readers with one writer
	// Allow multiple connections for better read concurrency
	// Writers still serialize, but readers can proceed in parallel
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(0) // Reuse connections indefinitely
	db.SetConnMaxIdleTime(0) // Don't close idle connections

	return db, nil
}

func NewClient(log *slog.Logger, db *sql.DB) *ent.Client {
	return ent.NewClient(
		ent.Logger(log),
		ent.Driver(
			entsql.OpenDB(dialect.SQLite, db),
		),
	)
}

func GetClient(ctx context.Context, log *slog.Logger, dbDir string) (*ent.Client, error) {
	db, err := NewDB(dbDir)
	if err != nil {
		return nil, err
	}

	client := NewClient(log, db)
	err = client.Schema.Create(ctx)
	if err != nil {
		return nil, err
	}

	return client, nil
}
