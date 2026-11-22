package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/berryscottr/gamers-paradise/internal/store"
)

func TestListAndDelete(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	e, _ := st.AddEntry(ctx, store.Entry{Name: "Gary", Homepage: "https://g.example/", Message: "smell ya\nlater"}, "v")

	var out bytes.Buffer
	if err := command(ctx, st, []string{"list"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "#1") || !strings.Contains(out.String(), "<https://g.example/>") || !strings.Contains(out.String(), "    later") {
		t.Fatalf("list:\n%s", out.String())
	}
	out.Reset()
	if err := command(ctx, st, []string{"delete", "1"}, &out); err != nil || !strings.Contains(out.String(), "hid signature #1") {
		t.Fatalf("delete: %v %s", err, out.String())
	}
	if err := command(ctx, st, []string{"delete", "1"}, &out); err == nil {
		t.Fatal("deleted twice")
	}
	out.Reset()
	_ = command(ctx, st, []string{"list"}, &out)
	if !strings.Contains(out.String(), "no signatures") {
		t.Fatalf("after delete:\n%s", out.String())
	}
	_ = e
}

func TestBadUsage(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	for _, args := range [][]string{{"delete"}, {"delete", "abc"}, {"delete", "-3"}, {"nuke"}} {
		if err := command(ctx, st, args, &bytes.Buffer{}); err == nil {
			t.Errorf("%v accepted", args)
		}
	}
	if err := run(nil, "postgres://x", &bytes.Buffer{}); err == nil {
		t.Error("no args accepted")
	}
	if err := run([]string{"list"}, "", &bytes.Buffer{}); err == nil {
		t.Error("no DATABASE_URL accepted")
	}
}
