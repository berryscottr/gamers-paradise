package store

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Postgres implements Store. All SQL is parameterized.
type Postgres struct {
	pool *pgxpool.Pool
}

// OpenPostgres connects and applies pending migrations.
func OpenPostgres(ctx context.Context, url string) (*Postgres, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, errors.New("parse DATABASE_URL: invalid URL") // never echo the URL: it holds a password
	}
	cfg.MaxConns = 6
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	p := &Postgres{pool: pool}
	if err := p.migrate(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return p, nil
}

func (p *Postgres) migrate(ctx context.Context) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(7700112233)`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (name TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
		return err
	}
	names, err := fs.Glob(migrations, "migrations/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(names)
	for _, name := range names {
		var done bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE name = $1)`, name).Scan(&done); err != nil {
			return err
		}
		if done {
			continue
		}
		body, err := migrations.ReadFile(name)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(body)); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (name) VALUES ($1)`, name); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (p *Postgres) Visit(ctx context.Context, day time.Time, visitor string) (int64, error) {
	day = Day(day)
	var n int64
	err := p.pool.QueryRow(ctx, `
		WITH ins AS (
			INSERT INTO visits (day, visitor) VALUES ($1, $2)
			ON CONFLICT DO NOTHING RETURNING 1
		)
		UPDATE counters SET n = n + (SELECT count(*) FROM ins)
		WHERE name = 'visitors' RETURNING n`, day, visitor).Scan(&n)
	if err != nil {
		return 0, err
	}
	// Forget hashes once their day is over (privacy, ADR-006).
	if _, err := p.pool.Exec(ctx, `DELETE FROM visits WHERE day < $1`, day.AddDate(0, 0, -1)); err != nil {
		return 0, err
	}
	if _, err := p.pool.Exec(ctx, `UPDATE guestbook SET visitor = NULL WHERE visitor IS NOT NULL AND created_at < $1`, day.AddDate(0, 0, -1)); err != nil {
		return 0, err
	}
	return n, nil
}

func (p *Postgres) Visitors(ctx context.Context) (int64, error) {
	var n int64
	err := p.pool.QueryRow(ctx, `SELECT n FROM counters WHERE name = 'visitors'`).Scan(&n)
	return n, err
}

func (p *Postgres) AddEntry(ctx context.Context, e Entry, visitor string) (Entry, error) {
	err := p.pool.QueryRow(ctx, `
		INSERT INTO guestbook (name, homepage, message, visitor) VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`, e.Name, e.Homepage, e.Message, visitor).Scan(&e.ID, &e.Created)
	return e, err
}

func (p *Postgres) Entries(ctx context.Context, limit int) ([]Entry, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT id, name, homepage, message, created_at FROM guestbook
		WHERE deleted_at IS NULL ORDER BY created_at DESC, id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Entry, error) {
		var e Entry
		err := r.Scan(&e.ID, &e.Name, &e.Homepage, &e.Message, &e.Created)
		return e, err
	})
}

func (p *Postgres) SignedSince(ctx context.Context, visitor string, t time.Time) (int, error) {
	var n int
	err := p.pool.QueryRow(ctx, `SELECT count(*) FROM guestbook WHERE visitor = $1 AND created_at >= $2`, visitor, t).Scan(&n)
	return n, err
}

func (p *Postgres) RecentDuplicate(ctx context.Context, message string, t time.Time) (bool, error) {
	var ok bool
	err := p.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM guestbook WHERE message = $1 AND created_at >= $2)`, message, t).Scan(&ok)
	return ok, err
}

func (p *Postgres) DeleteEntry(ctx context.Context, id int64) (bool, error) {
	tag, err := p.pool.Exec(ctx, `UPDATE guestbook SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

func (p *Postgres) Close() { p.pool.Close() }
