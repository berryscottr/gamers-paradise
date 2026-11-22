// Package web serves Gamer's Paradise: server-rendered pages from embedded
// templates, generated art, a hit counter and a guestbook. There are no
// logins, no third-party requests and no inline script or style (ADR-005).
package web

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"html"
	"html/template"
	"image/png"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/berryscottr/gamers-paradise/content"
	"github.com/berryscottr/gamers-paradise/internal/store"
	"github.com/berryscottr/gamers-paradise/internal/visitor"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

// Options configures the server.
type Options struct {
	Store      store.Store
	Hasher     *visitor.Hasher
	Production bool
	// SiteURL is the public address, used in the "link to me" code.
	SiteURL string
	Logger  *slog.Logger
	Now     func() time.Time
}

// Server is the HTTP handler for the whole site.
type Server struct {
	opt    Options
	pages  map[string]*template.Template
	assets map[string]asset
	sizes  map[string][2]int // art name -> width, height
	words  content.Words
	mux    *http.ServeMux
	posts  *limiter
}

type asset struct {
	body  []byte
	etag  string
	ctype string
}

// pageNames are the templates in templates/, each rendered inside layout.html.
var pageNames = []string{"home", "xbox", "gameboy", "shelf", "now", "box", "arcade", "skate", "fret", "status", "neon", "links", "guestbook", "notfound"}

// New builds the server, parsing every template and loading every asset up
// front so mistakes fail at startup (and in tests), not on a visitor.
func New(opt Options) (*Server, error) {
	if opt.Now == nil {
		opt.Now = time.Now
	}
	if opt.Logger == nil {
		opt.Logger = slog.Default()
	}
	if opt.Hasher == nil {
		opt.Hasher = visitor.NewHasher("", false)
	}
	words, err := content.Load()
	if err != nil {
		return nil, err
	}
	s := &Server{opt: opt, pages: map[string]*template.Template{}, assets: map[string]asset{}, sizes: map[string][2]int{}, words: words, posts: newLimiter(5, 10*time.Minute)}
	if err := s.loadAssets(); err != nil {
		return nil, err
	}
	funcs := template.FuncMap{"say": s.say, "art": s.art}
	for _, name := range pageNames {
		t, err := template.New("layout.html").Funcs(funcs).ParseFS(templateFS, "templates/layout.html", "templates/"+name+".html")
		if err != nil {
			return nil, fmt.Errorf("template %s: %w", name, err)
		}
		s.pages[name] = t
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.home)
	for _, p := range []string{"xbox", "gameboy", "shelf", "now", "box", "arcade", "links"} {
		mux.HandleFunc("GET /"+p, s.simple(p))
	}
	for _, g := range []string{"skate", "fret", "status", "neon"} {
		mux.HandleFunc("GET /arcade/"+g, s.simple(g))
	}
	mux.HandleFunc("GET /guestbook", s.guestbook)
	mux.HandleFunc("POST /guestbook", s.sign)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /robots.txt", s.robots)
	mux.HandleFunc("GET /static/", s.static)
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/art/favicon.png", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/", s.notFound)
	s.mux = mux
	return s, nil
}

func (s *Server) loadAssets() error {
	return fs.WalkDir(staticFS, "static", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		body, err := staticFS.ReadFile(p)
		if err != nil {
			return err
		}
		ctype := map[string]string{
			".js":  "text/javascript; charset=utf-8",
			".css": "text/css; charset=utf-8",
			".png": "image/png",
		}[path.Ext(p)]
		if ctype == "" {
			return fmt.Errorf("static file %s: unsupported type", p) // keep the site to known types
		}
		if path.Ext(p) == ".png" {
			cfg, err := png.DecodeConfig(bytes.NewReader(body))
			if err != nil {
				return fmt.Errorf("%s: %w", p, err)
			}
			s.sizes[strings.TrimSuffix(path.Base(p), ".png")] = [2]int{cfg.Width, cfg.Height}
		}
		sum := sha256.Sum256(body)
		s.assets["/"+p] = asset{body: body, etag: `"` + hex.EncodeToString(sum[:8]) + `"`, ctype: ctype}
		return nil
	})
}

// ServeHTTP applies security headers to every response, allows only the
// methods the site uses, caps request bodies, and routes.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := w.Header()
	h.Set("Content-Security-Policy", s.csp())
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Cross-Origin-Opener-Policy", "same-origin")
	// The 88x31 button is meant to be shown on other people's sites.
	if r.URL.Path == "/static/art/button.png" {
		h.Set("Cross-Origin-Resource-Policy", "cross-origin")
	} else {
		h.Set("Cross-Origin-Resource-Policy", "same-origin")
	}
	h.Set("Permissions-Policy", "accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=(), browsing-topics=()")
	if s.opt.Production {
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	} else {
		h.Set("X-Robots-Tag", "noindex, nofollow")
	}
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodPost:
	default:
		h.Set("Allow", "GET, HEAD, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	s.mux.ServeHTTP(w, r)
}

// csp is strict: same-origin scripts, styles and images only; no inline
// code, no fetches, no frames, and Trusted Types so the DOM can't be fed
// HTML strings (ADR-005).
func (s *Server) csp() string {
	p := "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self'; connect-src 'none'; " +
		"form-action 'self'; base-uri 'none'; frame-ancestors 'none'; " +
		"require-trusted-types-for 'script'; trusted-types 'none'"
	if s.opt.Production {
		p += "; upgrade-insecure-requests"
	}
	return p
}

// say renders a words.txt section. Empty sections show their prompt on
// non-production sites and nothing in production.
func (s *Server) say(id string) (template.HTML, error) {
	sec, ok := s.words[id]
	if !ok {
		return "", fmt.Errorf("words.txt has no section %q", id)
	}
	var b strings.Builder
	for _, p := range sec.Paragraphs {
		b.WriteString("<p>" + html.EscapeString(p) + "</p>")
	}
	if len(sec.Paragraphs) == 0 && !s.opt.Production {
		fmt.Fprintf(&b, `<p class="prompt"><b>Your words go here</b> (content/words.txt, section <code>%s</code>): %s</p>`,
			html.EscapeString(id), html.EscapeString(sec.Prompt))
	}
	return template.HTML(b.String()), nil // #nosec G203 -- every piece is escaped above
}

// art renders an <img> for a generated image with its real size.
func (s *Server) art(name, alt string) (template.HTML, error) {
	sz, ok := s.sizes[name]
	if !ok {
		return "", fmt.Errorf("no art named %q", name)
	}
	return template.HTML(fmt.Sprintf(`<img src="/static/art/%s.png" width="%d" height="%d" alt="%s">`, // #nosec G203 -- name is a known asset, alt is escaped
		name, sz[0], sz[1], html.EscapeString(alt))), nil
}

func (s *Server) robots(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if !s.opt.Production {
		_, _ = w.Write([]byte("User-agent: *\nDisallow: /\n"))
		return
	}
	_, _ = w.Write([]byte("User-agent: *\nAllow: /\n"))
}

// static serves embedded assets with ETags; browsers revalidate each load
// (a cheap 304) so a deploy never leaves stale code behind.
func (s *Server) static(w http.ResponseWriter, r *http.Request) {
	a, ok := s.assets[r.URL.Path]
	if !ok {
		s.notFound(w, r)
		return
	}
	h := w.Header()
	h.Set("ETag", a.etag)
	h.Set("Cache-Control", "no-cache")
	h.Set("Content-Type", a.ctype)
	if match := r.Header.Get("If-None-Match"); match != "" && strings.Contains(match, a.etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	_, _ = w.Write(a.body)
}

// NewHTTPServer wraps a handler with timeouts suitable for the internet.
func NewHTTPServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
}
