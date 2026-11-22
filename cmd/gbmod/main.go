// Command gbmod moderates the guestbook from your own machine. There is no
// admin page on the site at all (no logins, nothing to break into).
//
//	DATABASE_URL=... go run ./cmd/gbmod list        # newest 50 signatures
//	DATABASE_URL=... go run ./cmd/gbmod delete 17   # hide signature #17
//
// Get DATABASE_URL from Railway without printing it:
//
//	railway run --service gamers-paradise go run ./cmd/gbmod list
//
// (railway run injects the service's variables; the app's DATABASE_URL uses
// Railway's private network, so for local runs use the Postgres service's
// DATABASE_PUBLIC_URL instead. See docs/moderation.md.)
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/berryscottr/gamers-paradise/internal/store"
)

func main() {
	if err := run(os.Args[1:], os.Getenv("DATABASE_URL"), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "gbmod:", err)
		os.Exit(1)
	}
}

func run(args []string, url string, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: gbmod list | gbmod delete <id>")
	}
	if url == "" {
		return fmt.Errorf("set DATABASE_URL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	st, err := store.OpenPostgres(ctx, url)
	if err != nil {
		return err
	}
	defer st.Close()
	return command(ctx, st, args, out)
}

func command(ctx context.Context, st store.Store, args []string, out io.Writer) error {
	switch args[0] {
	case "list":
		entries, err := st.Entries(ctx, 50)
		if err != nil {
			return err
		}
		for _, e := range entries {
			home := ""
			if e.Homepage != "" {
				home = " <" + e.Homepage + ">"
			}
			fmt.Fprintf(out, "#%d  %s  %s%s\n    %s\n", e.ID, e.Created.Format("2006-01-02 15:04"), e.Name, home, strings.ReplaceAll(e.Message, "\n", "\n    "))
		}
		if len(entries) == 0 {
			fmt.Fprintln(out, "no signatures")
		}
		return nil
	case "delete":
		if len(args) != 2 {
			return fmt.Errorf("usage: gbmod delete <id>")
		}
		id, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil || id <= 0 {
			return fmt.Errorf("id must be a positive number")
		}
		ok, err := st.DeleteEntry(ctx, id)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("no visible signature #%d", id)
		}
		fmt.Fprintf(out, "hid signature #%d\n", id)
		return nil
	}
	return fmt.Errorf("unknown command %q (use list or delete)", args[0])
}
