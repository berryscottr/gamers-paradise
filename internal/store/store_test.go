package store

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// contract runs the same checks against any Store.
func contract(t *testing.T, s Store) {
	ctx := context.Background()
	day := time.Date(2026, 11, 22, 15, 0, 0, 0, time.UTC)

	base, err := s.Visitors(ctx)
	if err != nil {
		t.Fatal(err)
	}
	n, _ := s.Visit(ctx, day, "aaa")
	n2, _ := s.Visit(ctx, day.Add(time.Hour), "aaa") // same visitor, same day
	n3, _ := s.Visit(ctx, day, "bbb")
	n4, _ := s.Visit(ctx, day.AddDate(0, 0, 1), "aaa") // a new day counts again
	if n != base+1 || n2 != base+1 || n3 != base+2 || n4 != base+3 {
		t.Fatalf("counts %d %d %d %d from base %d", n, n2, n3, n4, base)
	}
	if v, _ := s.Visitors(ctx); v != base+3 {
		t.Fatalf("Visitors = %d", v)
	}

	since := time.Now().Add(-time.Hour)
	e, err := s.AddEntry(ctx, Entry{Name: "Ash", Homepage: "https://example.org/", Message: "first!"}, "vvv")
	if err != nil || e.ID == 0 || e.Created.IsZero() {
		t.Fatalf("AddEntry = %+v, %v", e, err)
	}
	e2, _ := s.AddEntry(ctx, Entry{Name: "Misty", Message: "second"}, "vvv")
	if c, _ := s.SignedSince(ctx, "vvv", since); c < 2 {
		t.Fatalf("SignedSince = %d", c)
	}
	if c, _ := s.SignedSince(ctx, "nobody", since); c != 0 {
		t.Fatalf("SignedSince(nobody) = %d", c)
	}
	if dup, _ := s.RecentDuplicate(ctx, "first!", since); !dup {
		t.Fatal("duplicate not found")
	}
	if dup, _ := s.RecentDuplicate(ctx, "never said", since); dup {
		t.Fatal("false duplicate")
	}
	got, _ := s.Entries(ctx, 2)
	if len(got) != 2 || got[0].ID != e2.ID || got[1].Homepage != "https://example.org/" {
		t.Fatalf("Entries = %+v", got)
	}
	if ok, _ := s.DeleteEntry(ctx, e2.ID); !ok {
		t.Fatal("delete failed")
	}
	if ok, _ := s.DeleteEntry(ctx, e2.ID); ok {
		t.Fatal("deleted twice")
	}
	got, _ = s.Entries(ctx, 1)
	if len(got) != 1 || got[0].ID != e.ID {
		t.Fatalf("after delete: %+v", got)
	}
}

func TestMemory(t *testing.T) { contract(t, NewMemory()) }

// TestPostgres runs when TEST_DATABASE_URL points at a throwaway database
// (CI provides one).
func TestPostgres(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	p, err := OpenPostgres(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	// migrations are idempotent
	p2, err := OpenPostgres(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	p2.Close()
	contract(t, p)

	// old visit hashes are forgotten
	var old int
	if err := p.pool.QueryRow(ctx, `SELECT count(*) FROM visits WHERE day < '2026-11-22'`).Scan(&old); err != nil {
		t.Fatal(err)
	}
	if old != 0 {
		t.Fatalf("%d old visit hashes kept", old)
	}
}

func TestOpenPostgresHidesTheURL(t *testing.T) {
	_, err := OpenPostgres(context.Background(), "postgres://u:secretpw@%zz")
	if err == nil {
		t.Fatal("bad URL accepted")
	}
	if strings.Contains(err.Error(), "secretpw") {
		t.Fatalf("error leaks the password: %v", err)
	}
}

func TestDay(t *testing.T) {
	loc := time.FixedZone("x", -5*3600)
	got := Day(time.Date(2026, 11, 22, 21, 30, 0, 0, loc)) // 02:30 UTC next day
	if !got.Equal(time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("Day = %v", got)
	}
}
