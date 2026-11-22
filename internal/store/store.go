// Package store keeps the only state the site has: the visitor counter and
// the guestbook. Postgres in deployed environments; memory for local dev
// and tests.
package store

import (
	"context"
	"time"
)

// Entry is one guestbook signature.
type Entry struct {
	ID       int64
	Name     string
	Homepage string // empty or an http(s) URL
	Message  string
	Created  time.Time
}

// Store is implemented by Memory and Postgres. Days are UTC dates.
type Store interface {
	// Visit records one unique visitor for a day (repeat calls are no-ops)
	// and returns the all-time count of unique visitor-days.
	Visit(ctx context.Context, day time.Time, visitor string) (int64, error)
	// Visitors returns the count without recording anything.
	Visitors(ctx context.Context) (int64, error)

	// AddEntry stores a signature; visitor is the daily hash, kept only
	// briefly for rate limits.
	AddEntry(ctx context.Context, e Entry, visitor string) (Entry, error)
	// Entries returns the newest visible signatures first.
	Entries(ctx context.Context, limit int) ([]Entry, error)
	// SignedSince counts signatures by a visitor since t.
	SignedSince(ctx context.Context, visitor string, t time.Time) (int, error)
	// RecentDuplicate reports whether the same message was signed since t.
	RecentDuplicate(ctx context.Context, message string, t time.Time) (bool, error)
	// DeleteEntry hides a signature. It reports whether one was hidden.
	DeleteEntry(ctx context.Context, id int64) (bool, error)

	Close()
}

// Day truncates t to its UTC date.
func Day(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
