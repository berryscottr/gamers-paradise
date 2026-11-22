package web

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/berryscottr/gamers-paradise/internal/art"
	"github.com/berryscottr/gamers-paradise/internal/store"
	"github.com/berryscottr/gamers-paradise/internal/visitor"
)

const browser = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/140.0 Safari/537.36"

func newServer(t *testing.T, production bool) (*Server, *store.Memory) {
	t.Helper()
	st := store.NewMemory()
	s, err := New(Options{
		Store:      st,
		Hasher:     visitor.NewHasher("test-secret-test-secret-test-secret", false),
		Production: production,
		SiteURL:    "https://gamers-paradise.example",
		Logger:     nil,
		Now:        func() time.Time { return time.Date(2026, 11, 22, 18, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatal(err)
	}
	return s, st
}

func get(t *testing.T, h http.Handler, path string, mod ...func(*http.Request)) (*http.Response, string) {
	t.Helper()
	r := httptest.NewRequest("GET", path, nil)
	r.Header.Set("User-Agent", browser)
	for _, m := range mod {
		m(r)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	body, _ := io.ReadAll(w.Result().Body)
	return w.Result(), string(body)
}

var pagePaths = []string{"/", "/xbox", "/gameboy", "/shelf", "/now", "/box", "/arcade", "/arcade/skate", "/arcade/fret", "/arcade/status", "/arcade/neon", "/links", "/guestbook"}

func TestEveryPageRenders(t *testing.T) {
	s, _ := newServer(t, false)
	for _, p := range pagePaths {
		res, body := get(t, s, p)
		if res.StatusCode != 200 {
			t.Errorf("%s: %d", p, res.StatusCode)
		}
		if !strings.Contains(body, "<title>") || !strings.Contains(body, `id="footer"`) {
			t.Errorf("%s: missing layout", p)
		}
	}
	res, body := get(t, s, "/missing")
	if res.StatusCode != 404 || !strings.Contains(body, "cartridge") {
		t.Errorf("404 page: %d", res.StatusCode)
	}
}

// The CSP forbids inline script and style, so none may exist in any page.
func TestNoInlineScriptOrStyle(t *testing.T) {
	s, _ := newServer(t, false)
	inlineScript := regexp.MustCompile(`<script(?:\s[^>]*)?>\s*[^<\s]`)
	for _, p := range append(pagePaths, "/missing") {
		_, body := get(t, s, p)
		if inlineScript.MatchString(body) {
			t.Errorf("%s has an inline script", p)
		}
		for _, bad := range []string{"<style", " style=", " onclick=", " onload=", " onerror=", "javascript:"} {
			if strings.Contains(body, bad) {
				t.Errorf("%s contains %q", p, bad)
			}
		}
	}
}

func TestSecurityHeaders(t *testing.T) {
	s, _ := newServer(t, true)
	res, _ := get(t, s, "/")
	h := res.Header
	csp := h.Get("Content-Security-Policy")
	for _, want := range []string{"default-src 'none'", "script-src 'self'", "frame-ancestors 'none'", "require-trusted-types-for 'script'", "upgrade-insecure-requests", "form-action 'self'"} {
		if !strings.Contains(csp, want) {
			t.Errorf("CSP missing %q: %s", want, csp)
		}
	}
	if strings.Contains(csp, "unsafe") {
		t.Errorf("CSP allows unsafe sources: %s", csp)
	}
	for k, v := range map[string]string{
		"X-Content-Type-Options":       "nosniff",
		"X-Frame-Options":              "DENY",
		"Cross-Origin-Opener-Policy":   "same-origin",
		"Cross-Origin-Resource-Policy": "same-origin",
		"Strict-Transport-Security":    "max-age=31536000; includeSubDomains",
	} {
		if got := h.Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
	if h.Get("X-Robots-Tag") != "" {
		t.Error("production must be indexable")
	}
	// The 88x31 button may be shown on other sites.
	res, _ = get(t, s, "/static/art/button.png")
	if res.Header.Get("Cross-Origin-Resource-Policy") != "cross-origin" {
		t.Error("button.png not embeddable")
	}
}

func TestDevSiteIsNotIndexed(t *testing.T) {
	s, _ := newServer(t, false)
	res, _ := get(t, s, "/")
	if res.Header.Get("X-Robots-Tag") == "" || res.Header.Get("Strict-Transport-Security") != "" {
		t.Error("dev headers wrong")
	}
	_, body := get(t, s, "/robots.txt")
	if !strings.Contains(body, "Disallow: /") {
		t.Error("dev robots.txt allows crawling")
	}
}

func TestMethodsAndBodyLimit(t *testing.T) {
	s, _ := newServer(t, false)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest("DELETE", "/", nil))
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("DELETE: %d", w.Code)
	}
	big := strings.NewReader("message=" + strings.Repeat("a", 64<<10))
	r := httptest.NewRequest("POST", "/guestbook", big)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Sec-Fetch-Site", "same-origin")
	w = httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("oversized POST: %d", w.Code)
	}
}

func TestStaticAssetsAndETags(t *testing.T) {
	s, _ := newServer(t, false)
	for _, p := range []string{"/static/style.css", "/static/site.js", "/static/skate.js", "/static/skate-core.js", "/static/sound.js", "/static/phase.js", "/static/arcade.js", "/static/rng.js", "/static/fret.js", "/static/fret-core.js", "/static/battle.js", "/static/battle-core.js", "/static/neon.js", "/static/neon-core.js", "/static/art/font.js", "/static/art/hotspots.css", "/static/art/room-night.png"} {
		res, _ := get(t, s, p)
		if res.StatusCode != 200 || res.Header.Get("ETag") == "" {
			t.Errorf("%s: %d", p, res.StatusCode)
			continue
		}
		res2, _ := get(t, s, p, func(r *http.Request) { r.Header.Set("If-None-Match", res.Header.Get("ETag")) })
		if res2.StatusCode != http.StatusNotModified {
			t.Errorf("%s revalidation: %d", p, res2.StatusCode)
		}
	}
	if res, _ := get(t, s, "/static/nope.js"); res.StatusCode != 404 {
		t.Errorf("missing asset: %d", res.StatusCode)
	}
}

func TestEveryRoomLinkWorks(t *testing.T) {
	s, _ := newServer(t, false)
	for _, h := range art.Hotspots {
		if h.Href == "" {
			continue
		}
		u, _ := url.Parse(h.Href)
		res, body := get(t, s, u.Path)
		if res.StatusCode != 200 {
			t.Errorf("%s -> %s: %d", h.ID, h.Href, res.StatusCode)
		}
		if u.Fragment != "" && !strings.Contains(body, `id="`+u.Fragment+`"`) {
			t.Errorf("%s -> %s: no #%s on the page", h.ID, h.Href, u.Fragment)
		}
	}
}

func TestCounterCountsPeopleOncePerDay(t *testing.T) {
	s, st := newServer(t, false)
	get(t, s, "/")
	get(t, s, "/")
	get(t, s, "/xbox") // only the bedroom counts
	get(t, s, "/", func(r *http.Request) { r.Header.Set("User-Agent", "Googlebot/2.1") })
	get(t, s, "/", func(r *http.Request) { r.RemoteAddr = "198.51.100.9:1234" })
	n, _ := st.Visitors(context.Background())
	if n != 2 {
		t.Fatalf("visitors = %d, want 2", n)
	}
	_, body := get(t, s, "/xbox")
	if !strings.Contains(body, "<span>0</span><span>0</span><span>0</span><span>0</span><span>0</span><span>0</span><span>2</span>") {
		t.Error("odometer doesn't show 0000002")
	}
}

func TestPromptsShowOnlyOutsideProduction(t *testing.T) {
	dev, _ := newServer(t, false)
	prod, _ := newServer(t, true)
	if _, body := get(t, dev, "/gameboy"); !strings.Contains(body, `class="prompt"`) {
		t.Error("dev hides the writing prompts")
	}
	if _, body := get(t, prod, "/gameboy"); strings.Contains(body, `class="prompt"`) {
		t.Error("production shows writing prompts")
	}
}

func post(t *testing.T, h http.Handler, form url.Values, mod ...func(*http.Request)) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest("POST", "/guestbook", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("User-Agent", browser)
	r.Header.Set("Sec-Fetch-Site", "same-origin")
	for _, m := range mod {
		m(r)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestGuestbookSignAndShow(t *testing.T) {
	s, _ := newServer(t, false)
	w := post(t, s, url.Values{"name": {"<b>Red</b>"}, "homepage": {"https://red.example/"}, "message": {"cool room <script>alert(1)</script> check www.spam.com"}})
	if w.Code != http.StatusSeeOther || !strings.HasPrefix(w.Header().Get("Location"), "/guestbook?signed=1") {
		t.Fatalf("sign: %d %s", w.Code, w.Header().Get("Location"))
	}
	_, body := get(t, s, "/guestbook?signed=1")
	if !strings.Contains(body, "Thanks for signing") {
		t.Error("no thanks")
	}
	if strings.Contains(body, "<script>alert") || strings.Contains(body, "<b>Red</b>") {
		t.Error("signature not escaped")
	}
	if !strings.Contains(body, "&lt;b&gt;Red&lt;/b&gt;") || !strings.Contains(body, "[link removed]") {
		t.Error("signature missing or link kept")
	}
	if !strings.Contains(body, `href="https://red.example/" rel="ugc nofollow noopener"`) {
		t.Error("homepage link not marked as user content")
	}
}

func TestGuestbookRefusals(t *testing.T) {
	s, st := newServer(t, false)
	s.posts = newLimiter(100, time.Minute) // this test is about the other limits
	ok := url.Values{"name": {"Blue"}, "message": {"hello"}}

	if w := post(t, s, ok, func(r *http.Request) { r.Header.Set("Sec-Fetch-Site", "cross-site") }); w.Code != http.StatusForbidden {
		t.Errorf("cross-site: %d", w.Code)
	}
	if w := post(t, s, url.Values{"name": {""}, "message": {"x"}}); w.Code != http.StatusUnprocessableEntity || !strings.Contains(w.Body.String(), "Write a name") {
		t.Errorf("missing name: %d", w.Code)
	}
	// honeypot: looks like success, stores nothing
	if w := post(t, s, url.Values{"name": {"bot"}, "message": {"buy"}, "email": {"x@y.z"}}); w.Code != http.StatusSeeOther {
		t.Errorf("honeypot: %d", w.Code)
	}
	if e, _ := st.Entries(context.Background(), 10); len(e) != 0 {
		t.Fatalf("honeypot stored %d entries", len(e))
	}
	// duplicate message
	post(t, s, ok)
	if w := post(t, s, url.Values{"name": {"Other"}, "message": {"hello"}}, func(r *http.Request) { r.RemoteAddr = "198.51.100.2:1" }); w.Code != http.StatusUnprocessableEntity {
		t.Errorf("duplicate: %d", w.Code)
	}
	// three a day per visitor
	post(t, s, url.Values{"name": {"Blue"}, "message": {"two"}})
	post(t, s, url.Values{"name": {"Blue"}, "message": {"three"}})
	if w := post(t, s, url.Values{"name": {"Blue"}, "message": {"four"}}); w.Code != http.StatusUnprocessableEntity || !strings.Contains(w.Body.String(), "three times") {
		t.Errorf("daily limit: %d", w.Code)
	}
}

func TestGuestbookBurstLimit(t *testing.T) {
	s, _ := newServer(t, false)
	var last int
	for i := 0; i < 7; i++ {
		last = post(t, s, url.Values{"name": {"x"}, "message": {""}}).Code
	}
	if last != http.StatusTooManyRequests {
		t.Fatalf("7th post in a burst: %d", last)
	}
}

func TestSameOrigin(t *testing.T) {
	cases := []struct {
		site, origin string
		want         bool
	}{
		{"same-origin", "", true},
		{"cross-site", "", false},
		{"same-site", "", false},
		{"", "https://example.com", true},
		{"", "https://evil.com", false},
		{"", "", false},
	}
	for _, c := range cases {
		r := httptest.NewRequest("POST", "https://example.com/guestbook", nil)
		r.Header.Set("Sec-Fetch-Site", c.site)
		r.Header.Set("Origin", c.origin)
		if got := sameOrigin(r); got != c.want {
			t.Errorf("site=%q origin=%q: %v", c.site, c.origin, got)
		}
	}
}

func TestLinksPageHasButtonCode(t *testing.T) {
	s, _ := newServer(t, false)
	_, body := get(t, s, "/links")
	if !strings.Contains(body, "https://gamers-paradise.example/static/art/button.png") {
		t.Error("link-back code missing the site URL")
	}
	if strings.Count(body, `class="open"`) != 8 {
		t.Error("Top 8 should show 8 open slots")
	}
}

func TestEveryArcadeGameIsPlayable(t *testing.T) {
	s, _ := newServer(t, false)
	_, body := get(t, s, "/arcade")
	for _, g := range games {
		if g.Href == "" {
			t.Errorf("%s has no page", g.Name)
			continue
		}
		if !strings.Contains(body, `href="`+g.Href+`"`) {
			t.Errorf("arcade doesn't link %s", g.Href)
		}
		res, page := get(t, s, g.Href)
		if res.StatusCode != 200 || !strings.Contains(page, `<canvas id=`) || !strings.Contains(page, `class="era-game"`) {
			t.Errorf("%s: %d, missing canvas or game look", g.Href, res.StatusCode)
		}
	}
	if strings.Contains(body, "COMING SOON") {
		t.Error("a game still says coming soon")
	}
}
