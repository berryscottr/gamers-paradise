// Package guestbook cleans and checks signatures before they are stored.
// Signatures go live at once (no approval queue), so this is the moderation:
// normalize, strip links from messages, mask bad words, reject spam shapes.
// Templates escape everything on output as well.
package guestbook

import (
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Limits on each field, in characters.
const (
	MaxName     = 40
	MaxHomepage = 200
	MaxMessage  = 500
	MaxLines    = 8
)

// Input is what the form sends.
type Input struct {
	Name     string
	Homepage string
	Message  string
}

// Clean is a signature ready to store.
type Clean struct {
	Name     string
	Homepage string
	Message  string
}

// Problem is a user-facing reason a signature was refused.
type Problem string

// Check normalizes and moderates a signature. It returns problems for the
// visitor to fix; an empty slice means c can be stored.
func Check(in Input) (c Clean, problems []Problem) {
	c.Name = Mask(singleLine(in.Name))
	c.Message = Mask(StripLinks(multiLine(in.Message)))
	c.Homepage = strings.TrimSpace(in.Homepage)

	switch n := utf8.RuneCountInString(c.Name); {
	case n == 0:
		problems = append(problems, "Write a name (a nickname is fine).")
	case n > MaxName:
		problems = append(problems, "Keep your name under 40 characters.")
	}
	if StripLinks(c.Name) != c.Name {
		problems = append(problems, "Put your website in the homepage box, not your name.")
	}
	switch n := utf8.RuneCountInString(c.Message); {
	case strings.TrimSpace(strings.ReplaceAll(c.Message, linkRemoved, "")) == "":
		problems = append(problems, "Write a message.")
	case n > MaxMessage:
		problems = append(problems, "Keep your message under 500 characters.")
	}
	if strings.Count(c.Message, linkRemoved) > 2 {
		problems = append(problems, "That looks like spam. Links go in the homepage box.")
	}
	if c.Homepage != "" {
		h, ok := Homepage(c.Homepage)
		if !ok {
			problems = append(problems, "Your homepage needs to be a full http:// or https:// address.")
		}
		c.Homepage = h
	}
	return c, problems
}

// Homepage validates and normalizes a visitor's homepage link.
func Homepage(raw string) (string, bool) {
	if len(raw) > MaxHomepage {
		return "", false
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil {
		return "", false
	}
	if strings.ContainsAny(u.Host, " <>\"'") {
		return "", false
	}
	s := u.String()
	if len(s) > MaxHomepage {
		return "", false
	}
	return s, true
}

const linkRemoved = "[link removed]"

var linkRE = regexp.MustCompile(`(?i)(?:\b[a-z][a-z0-9+.-]*://\S+|\bwww\.\S+|\b[a-z0-9-]+(?:\.[a-z0-9-]+)*\.(?:com|net|org|io|ru|cn|xyz|info|biz|co|me|app|dev|link|site|top|shop|online|club|live|tk|ly|gg|tv)\b(?:/\S*)?)`)

// StripLinks replaces anything that looks like a link.
func StripLinks(s string) string {
	return linkRE.ReplaceAllString(s, linkRemoved)
}

func singleLine(s string) string {
	return strings.Join(strings.Fields(clean(s)), " ")
}

// multiLine keeps line breaks (up to MaxLines) but collapses other space.
func multiLine(s string) string {
	s = strings.ReplaceAll(clean(s), "\r\n", "\n")
	var lines []string
	blank := 0
	for _, l := range strings.Split(s, "\n") {
		l = strings.Join(strings.Fields(l), " ")
		if l == "" {
			blank++
			if blank > 1 || len(lines) == 0 {
				continue
			}
		} else {
			blank = 0
		}
		lines = append(lines, l)
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) > MaxLines {
		lines = append(lines[:MaxLines-1], strings.Join(lines[MaxLines-1:], " "))
	}
	return strings.Join(lines, "\n")
}

// clean drops control and invisible formatting characters (keeping \n).
func clean(s string) string {
	s = strings.ToValidUTF8(s, "")
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' {
			return r
		}
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, s)
}
