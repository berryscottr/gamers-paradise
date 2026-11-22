// Command artgen writes every generated image (and the room's hotspot CSS)
// into internal/web/static/art. Output is deterministic; CI regenerates it
// and fails if anything differs from what is committed.
//
//	go run ./cmd/artgen
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/berryscottr/gamers-paradise/internal/art"
)

func main() {
	out := flag.String("out", "internal/web/static/art", "output directory")
	flag.Parse()
	if err := run(*out); err != nil {
		fmt.Fprintln(os.Stderr, "artgen:", err)
		os.Exit(1)
	}
}

func run(dir string) error {
	files, err := art.Render()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), body, 0o644); err != nil {
			return err
		}
	}
	fmt.Printf("artgen: wrote %d files to %s\n", len(files), dir)
	return nil
}
