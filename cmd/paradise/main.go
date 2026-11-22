// Command paradise serves Gamer's Paradise.
//
//	go run ./cmd/paradise            # http://localhost:8080, in-memory store
//
// See docs/development.md for the environment variables.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/berryscottr/gamers-paradise/internal/store"
	"github.com/berryscottr/gamers-paradise/internal/visitor"
	"github.com/berryscottr/gamers-paradise/internal/web"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(log, os.Getenv); err != nil {
		log.Error("exit", "err", err)
		os.Exit(1)
	}
}

// config is read from the environment.
type config struct {
	Port        string
	Production  bool
	DatabaseURL string
	Secret      string
	SiteURL     string
	TrustProxy  bool
}

func loadConfig(getenv func(string) string) (config, error) {
	c := config{
		Port:        getenv("PORT"),
		Production:  getenv("APP_ENV") == "production",
		DatabaseURL: getenv("DATABASE_URL"),
		Secret:      getenv("VISITOR_SECRET"),
		SiteURL:     strings.TrimRight(getenv("SITE_URL"), "/"),
	}
	if c.Port == "" {
		c.Port = "8080"
	}
	// Railway's edge sets X-Real-IP; trust it only in production by default.
	c.TrustProxy = c.Production
	if v := getenv("TRUST_PROXY"); v != "" {
		c.TrustProxy = v == "true"
	}
	if c.SiteURL == "" {
		c.SiteURL = "http://localhost:" + c.Port
	}
	if c.Production {
		switch {
		case c.DatabaseURL == "":
			return c, errors.New("production needs DATABASE_URL")
		case len(c.Secret) < 32:
			return c, errors.New("production needs VISITOR_SECRET (32+ characters)")
		case !strings.HasPrefix(c.SiteURL, "https://"):
			return c, errors.New("production needs an https SITE_URL")
		}
	}
	return c, nil
}

func run(log *slog.Logger, getenv func(string) string) error {
	cfg, err := loadConfig(getenv)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var st store.Store
	if cfg.DatabaseURL != "" {
		pg, err := store.OpenPostgres(ctx, cfg.DatabaseURL)
		if err != nil {
			return err
		}
		st = pg
	} else {
		log.Warn("no DATABASE_URL: using an in-memory store (resets on restart)")
		st = store.NewMemory()
	}
	defer st.Close()

	h, err := web.New(web.Options{
		Store:      st,
		Hasher:     visitor.NewHasher(cfg.Secret, cfg.TrustProxy),
		Production: cfg.Production,
		SiteURL:    cfg.SiteURL,
		Logger:     log,
	})
	if err != nil {
		return err
	}
	srv := web.NewHTTPServer(":"+cfg.Port, h)
	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	log.Info("listening", "addr", fmt.Sprintf("http://localhost:%s", cfg.Port), "production", cfg.Production)

	select {
	case err := <-errc:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
	}
	shut, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shut)
}
