package storage

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"

	nekodb "novelclaw/db"
	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/paths"
)

type Storage struct {
	db *sql.DB
	q  *sqlc.Queries
}

func (s *Storage) DB() *sql.DB {
	return s.db
}

func (s *Storage) Queries() *sqlc.Queries {
	return s.q
}

func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// WithTx executes a set of queries within a transaction, rolling back on error or panic.
func (s *Storage) WithTx(ctx context.Context, fn func(q *sqlc.Queries) error) (err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	qtx := s.q.WithTx(tx)
	if err = fn(qtx); err != nil {
		return err
	}
	return tx.Commit()
}

// OpenStorage opens a SQLite database, configures WAL mode, connection pool and applies migrations.
func OpenStorage(dbPath string) (*Storage, error) {
	if dbPath == "" {
		dbPath = paths.DB()
	}

	// One-time migration: if the legacy "neko.db" exists and the new path
	// does not, copy the old database over so existing projects survive the
	// rebrand instead of being silently replaced by a fresh empty database.
	if dbPath != ":memory:" && dbPath != "./data/neko.db" {
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			if _, legacyErr := os.Stat("./data/neko.db"); legacyErr == nil {
				if mkErr := os.MkdirAll(filepath.Dir(dbPath), 0755); mkErr == nil {
					if copyErr := copyFile(dbPath, "./data/neko.db"); copyErr != nil {
						return nil, fmt.Errorf("migrate legacy neko.db: %w", copyErr)
					}
				}
			}
		}
	}

	if dbPath != ":memory:" {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	dsn := buildDSN(dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)

	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	if err := ApplySchema(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	queries := sqlc.New(db)
	return &Storage{
		db: db,
		q:  queries,
	}, nil
}

// ApplySchema executes embedded SQL schema migration files in alphanumeric order.
func ApplySchema(db *sql.DB) error {
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY);`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := fs.ReadDir(nekodb.SchemaFS, "schema")
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		var count int
		err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, entry.Name()).Scan(&count)
		if err != nil {
			return fmt.Errorf("check migration version %s: %w", entry.Name(), err)
		}

		if count > 0 {
			continue
		}

		content, err := fs.ReadFile(nekodb.SchemaFS, "schema/"+entry.Name())
		if err != nil {
			return fmt.Errorf("read migration content %s: %w", entry.Name(), err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration tx %s: %w", entry.Name(), err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			if strings.Contains(err.Error(), "duplicate column name") {
				_ = tx.Rollback()
				_, _ = db.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES (?)`, entry.Name())
				continue
			}
			_ = tx.Rollback()
			return fmt.Errorf("execute migration %s: %w", entry.Name(), err)
		}

		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES (?)`, entry.Name()); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", entry.Name(), err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// buildDSN constructs the SQLite connection string with WAL and performance pragma options.
func buildDSN(dbPath string) string {
	if dbPath == ":memory:" {
		return ":memory:?_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)"
	}

	values := url.Values{}
	values.Add("_txlock", "immediate")
	values.Add("_pragma", "busy_timeout=10000")
	values.Add("_pragma", "foreign_keys(ON)")
	values.Add("_pragma", "trusted_schema(OFF)")
	values.Add("_pragma", "journal_mode(WAL)")
	values.Add("_pragma", "synchronous(NORMAL)")
	values.Add("_pragma", "temp_store(MEMORY)")
	values.Add("_pragma", "cache_size(-65536)")
	values.Add("_pragma", "mmap_size(268435456)")

	return "file:" + dbPath + "?" + values.Encode()
}

// copyFile copies src to dst using a bounded buffer. Used to migrate the
// legacy "neko.db" into the rebranded "novelclaw.db" path exactly once.
func copyFile(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	buf := make([]byte, 1<<20)
	for {
		n, rerr := in.Read(buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				return werr
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return rerr
		}
	}
	return out.Sync()
}
