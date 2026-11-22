// Package visitor turns a request into an anonymous, daily-rotating ID for
// the hit counter and guestbook rate limits. The ID is an HMAC of the day,
// IP and user agent; no IP is ever stored (ADR-006).
package visitor

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"strings"
	"time"
)

// Hasher makes visitor IDs. Its secret must stay private: with it, an ID
// could be checked against guessed IPs.
type Hasher struct {
	secret []byte
	// TrustProxy reads the client IP from X-Real-IP, which Railway's edge
	// sets. Only enable it behind such a proxy, or clients can spoof it.
	TrustProxy bool
}

// NewHasher uses secret, or a random one (IDs then reset on restart).
func NewHasher(secret string, trustProxy bool) *Hasher {
	k := []byte(secret)
	if len(k) == 0 {
		k = make([]byte, 32)
		if _, err := rand.Read(k); err != nil {
			panic(err)
		}
	}
	return &Hasher{secret: k, TrustProxy: trustProxy}
}

// ID returns the visitor's ID for the UTC day of now.
func (h *Hasher) ID(r *http.Request, now time.Time) string {
	m := hmac.New(sha256.New, h.secret)
	m.Write([]byte(now.UTC().Format("2006-01-02")))
	m.Write([]byte{0})
	m.Write([]byte(h.IP(r)))
	m.Write([]byte{0})
	m.Write([]byte(r.UserAgent()))
	return hex.EncodeToString(m.Sum(nil)[:16])
}

// IP is the client address used for hashing and in-memory rate limits.
func (h *Hasher) IP(r *http.Request) string {
	if h.TrustProxy {
		if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

var botMarks = []string{"bot", "crawl", "spider", "slurp", "curl", "wget", "python", "go-http-client", "headless", "preview", "fetch", "monitor", "uptime", "scan"}

// IsBot guesses whether a user agent is automated, so it isn't counted.
func IsBot(ua string) bool {
	if ua == "" {
		return true
	}
	ua = strings.ToLower(ua)
	if !strings.HasPrefix(ua, "mozilla/") && !strings.HasPrefix(ua, "opera/") {
		return true
	}
	for _, m := range botMarks {
		if strings.Contains(ua, m) {
			return true
		}
	}
	return false
}
