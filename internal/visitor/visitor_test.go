package visitor

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func req(ip, realIP, ua string) *http.Request {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = ip + ":5555"
	if realIP != "" {
		r.Header.Set("X-Real-IP", realIP)
	}
	r.Header.Set("User-Agent", ua)
	return r
}

func TestIDRotatesDailyAndIgnoresSpoofedHeaders(t *testing.T) {
	h := NewHasher("a-test-secret-that-is-long-enough-123", false)
	day := time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC)
	a := h.ID(req("1.2.3.4", "", "Mozilla/5.0"), day)
	if a != h.ID(req("1.2.3.4", "", "Mozilla/5.0"), day.Add(5*time.Hour)) {
		t.Fatal("same visitor, same day, different ID")
	}
	if a == h.ID(req("1.2.3.4", "", "Mozilla/5.0"), day.AddDate(0, 0, 1)) {
		t.Fatal("ID did not rotate the next day")
	}
	if a == h.ID(req("5.6.7.8", "", "Mozilla/5.0"), day) {
		t.Fatal("different IPs share an ID")
	}
	// Without TrustProxy, X-Real-IP is ignored (it could be spoofed).
	if a != h.ID(req("1.2.3.4", "9.9.9.9", "Mozilla/5.0"), day) {
		t.Fatal("untrusted X-Real-IP changed the ID")
	}
	if len(a) != 32 {
		t.Fatalf("ID %q", a)
	}
}

func TestTrustProxyUsesRealIP(t *testing.T) {
	h := NewHasher("s", true)
	if got := h.IP(req("10.0.0.1", "203.0.113.7", "x")); got != "203.0.113.7" {
		t.Fatalf("IP = %q", got)
	}
	if got := h.IP(req("10.0.0.1", "", "x")); got != "10.0.0.1" {
		t.Fatalf("fallback IP = %q", got)
	}
}

func TestDifferentSecretsDifferentIDs(t *testing.T) {
	day := time.Now()
	r := req("1.2.3.4", "", "Mozilla/5.0")
	if NewHasher("one", false).ID(r, day) == NewHasher("two", false).ID(r, day) {
		t.Fatal("secret does not affect the ID")
	}
	r1, r2 := NewHasher("", false), NewHasher("", false)
	if r1.ID(r, day) == r2.ID(r, day) {
		t.Fatal("random secrets collided")
	}
}

func TestIsBot(t *testing.T) {
	bots := []string{"", "curl/8.0", "Googlebot/2.1", "Mozilla/5.0 (compatible; bingbot/2.0)", "python-requests/2.31", "Mozilla/5.0 HeadlessChrome/120", "Go-http-client/1.1"}
	people := []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0 Safari/537.36", "Mozilla/5.0 (Macintosh; Intel Mac OS X 15_0) Gecko/20100101 Firefox/150.0"}
	for _, ua := range bots {
		if !IsBot(ua) {
			t.Errorf("IsBot(%q) = false", ua)
		}
	}
	for _, ua := range people {
		if IsBot(ua) {
			t.Errorf("IsBot(%q) = true", ua)
		}
	}
}
