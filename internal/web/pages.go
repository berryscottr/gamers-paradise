package web

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/berryscottr/gamers-paradise/internal/art"
	"github.com/berryscottr/gamers-paradise/internal/guestbook"
	"github.com/berryscottr/gamers-paradise/internal/store"
	"github.com/berryscottr/gamers-paradise/internal/visitor"
)

// page is everything a template can use.
type page struct {
	Title    string // <title>
	Era      string // body class: each object's page dresses in its own era
	Nav      string // which nav link is current
	Visitors []string
	SiteURL  string

	// home
	Hotspots []art.Hotspot
	Phases   string // "hour:name,..." for site.js
	// arcade and links
	Games    []game
	TopEight []string
	// guestbook
	Entries  []store.Entry
	Form     guestbook.Input
	Problems []guestbook.Problem
	Signed   bool
}

type game struct {
	Name, Blurb, Href string
}

// games on the beanbag's select screen.
var games = []game{
	{"Half-Pipe Hero", "Skate a half-pipe. Pump, launch, chain flips, grabs and spins, and land clean.", "/arcade/skate"},
	{"Fret Night", "Five lanes, an original chiptune, expert or bust.", "/arcade/fret"},
	{"Status Effect", "A monster battle you win with sleep, poison and paralysis instead of grinding.", "/arcade/status"},
	{"Neon Mile", "A top-down street race through a city at night.", "/arcade/neon"},
}

// topEight are the sites on the Top 8 page; empty slots show as open.
var topEight = []string{}

// gamePages are the playable games; they share the arcade look.
var gamePages = map[string]bool{"skate": true, "fret": true, "status": true, "neon": true}

var titles = map[string]string{
	"home": "Gamer's Paradise", "xbox": "The Xbox years", "gameboy": "Game Boy",
	"shelf": "The shelf", "now": "Right now", "box": "Coming soon", "arcade": "Arcade",
	"skate": "Half-Pipe Hero", "fret": "Fret Night", "status": "Status Effect", "neon": "Neon Mile", "links": "Top 8", "guestbook": "Guestbook", "notfound": "404",
}

func (s *Server) base(ctx context.Context, name string) page {
	p := page{Title: titles[name], Era: name, Nav: name, SiteURL: s.opt.SiteURL}
	if name != "home" {
		p.Title += " · Gamer's Paradise"
	}
	if gamePages[name] {
		p.Nav, p.Era = "arcade", "game"
	}
	n, err := s.opt.Store.Visitors(ctx)
	if err != nil {
		s.opt.Logger.Error("visitors", "err", err)
	}
	p.Visitors = digits(n)
	return p
}

// digits zero-pads the counter for the odometer.
func digits(n int64) []string {
	return strings.Split(fmt.Sprintf("%07d", n), "")
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var count int64 = -1
	if !visitor.IsBot(r.UserAgent()) {
		now := s.opt.Now()
		n, err := s.opt.Store.Visit(ctx, now, s.opt.Hasher.ID(r, now))
		if err != nil {
			s.opt.Logger.Error("visit", "err", err)
		} else {
			count = n
		}
	}
	p := s.base(ctx, "home")
	if count >= 0 {
		p.Visitors = digits(count)
	}
	p.Hotspots = art.Hotspots
	var ph []string
	for _, x := range art.Phases {
		ph = append(ph, fmt.Sprintf("%d:%s", x.FromHour, x.Name))
	}
	p.Phases = strings.Join(ph, ",")
	w.Header().Set("Cache-Control", "no-store")
	s.render(w, http.StatusOK, "home", p)
}

func (s *Server) simple(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := s.base(r.Context(), name)
		p.Games = games
		p.TopEight = make([]string, 8)
		copy(p.TopEight, topEight)
		w.Header().Set("Cache-Control", "no-cache")
		s.render(w, http.StatusOK, name, p)
	}
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	s.render(w, http.StatusNotFound, "notfound", s.base(r.Context(), "notfound"))
}

func (s *Server) render(w http.ResponseWriter, status int, name string, p page) {
	var buf bytes.Buffer
	if err := s.pages[name].Execute(&buf, p); err != nil {
		s.opt.Logger.Error("render", "page", name, "err", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

// Guestbook limits.
const (
	signaturesPerDay = 3
	entriesShown     = 100
)

func (s *Server) guestbook(w http.ResponseWriter, r *http.Request) {
	s.showGuestbook(w, r, http.StatusOK, guestbook.Input{}, nil)
}

func (s *Server) showGuestbook(w http.ResponseWriter, r *http.Request, status int, in guestbook.Input, problems []guestbook.Problem) {
	p := s.base(r.Context(), "guestbook")
	entries, err := s.opt.Store.Entries(r.Context(), entriesShown)
	if err != nil {
		s.opt.Logger.Error("entries", "err", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	p.Entries, p.Form, p.Problems = entries, in, problems
	p.Signed = r.Method == http.MethodGet && r.URL.Query().Get("signed") == "1"
	w.Header().Set("Cache-Control", "no-store")
	s.render(w, status, "guestbook", p)
}

// sign handles the guestbook form. Signatures go live immediately; the
// guestbook package does the moderation (ADR-007).
func (s *Server) sign(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if !s.posts.allow(s.opt.Hasher.IP(r)) {
		http.Error(w, "Slow down a little and try again in a few minutes.", http.StatusTooManyRequests)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	in := guestbook.Input{Name: r.PostForm.Get("name"), Homepage: r.PostForm.Get("homepage"), Message: r.PostForm.Get("message")}
	// Honeypot: a field people never see. Bots that fill it get a fake success.
	if r.PostForm.Get("email") != "" {
		http.Redirect(w, r, "/guestbook?signed=1", http.StatusSeeOther)
		return
	}
	clean, problems := guestbook.Check(in)
	ctx := r.Context()
	now := s.opt.Now()
	id := s.opt.Hasher.ID(r, now)
	if len(problems) == 0 {
		n, err := s.opt.Store.SignedSince(ctx, id, now.Add(-24*time.Hour))
		if err != nil {
			s.fail(w, err)
			return
		}
		if n >= signaturesPerDay {
			problems = append(problems, "You've signed three times today. Come back tomorrow.")
		}
	}
	if len(problems) == 0 {
		dup, err := s.opt.Store.RecentDuplicate(ctx, clean.Message, now.Add(-24*time.Hour))
		if err != nil {
			s.fail(w, err)
			return
		}
		if dup {
			problems = append(problems, "Someone already signed with that exact message today.")
		}
	}
	if len(problems) > 0 {
		s.showGuestbook(w, r, http.StatusUnprocessableEntity, in, problems)
		return
	}
	if _, err := s.opt.Store.AddEntry(ctx, store.Entry{Name: clean.Name, Homepage: clean.Homepage, Message: clean.Message}, id); err != nil {
		s.fail(w, err)
		return
	}
	http.Redirect(w, r, "/guestbook?signed=1#entries", http.StatusSeeOther)
}

func (s *Server) fail(w http.ResponseWriter, err error) {
	s.opt.Logger.Error("guestbook", "err", err)
	http.Error(w, "something went wrong", http.StatusInternalServerError)
}

// sameOrigin is the CSRF guard for POSTs: browsers send Sec-Fetch-Site;
// older ones send Origin, which must match the host.
func sameOrigin(r *http.Request) bool {
	switch r.Header.Get("Sec-Fetch-Site") {
	case "same-origin":
		return true
	case "":
		o := r.Header.Get("Origin")
		return o != "" && (o == "https://"+r.Host || o == "http://"+r.Host)
	default:
		return false
	}
}

// limiter is a small in-memory sliding-window rate limiter keyed by IP.
// It is never persisted.
type limiter struct {
	mu     sync.Mutex
	max    int
	window time.Duration
	hits   map[string][]time.Time
	now    func() time.Time
}

func newLimiter(max int, window time.Duration) *limiter {
	return &limiter{max: max, window: window, hits: map[string][]time.Time{}, now: time.Now}
}

func (l *limiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	cut := now.Add(-l.window)
	kept := l.hits[key][:0]
	for _, t := range l.hits[key] {
		if t.After(cut) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= l.max {
		l.hits[key] = kept
		return false
	}
	l.hits[key] = append(kept, now)
	if len(l.hits) > 10000 { // bound memory
		l.hits = map[string][]time.Time{key: l.hits[key]}
	}
	return true
}
