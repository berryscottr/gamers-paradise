package guestbook

import (
	"strings"
	"testing"
)

func TestCheckAcceptsANormalSignature(t *testing.T) {
	c, problems := Check(Input{Name: "  Player   One ", Homepage: "https://example.neocities.org/", Message: "cool room!\r\n\r\n\r\nlove the CRT"})
	if len(problems) != 0 {
		t.Fatalf("problems: %v", problems)
	}
	if c.Name != "Player One" {
		t.Errorf("name %q", c.Name)
	}
	if c.Message != "cool room!\n\nlove the CRT" {
		t.Errorf("message %q", c.Message)
	}
	if c.Homepage != "https://example.neocities.org/" {
		t.Errorf("homepage %q", c.Homepage)
	}
}

func TestCheckProblems(t *testing.T) {
	cases := []struct {
		name string
		in   Input
		want string
	}{
		{"empty name", Input{Name: "  ", Message: "hi"}, "Write a name"},
		{"long name", Input{Name: strings.Repeat("a", 41), Message: "hi"}, "under 40"},
		{"link in name", Input{Name: "buy-pills.com", Message: "hi"}, "homepage box"},
		{"empty message", Input{Name: "a", Message: "\n\n"}, "Write a message"},
		{"only a link", Input{Name: "a", Message: "http://spam.example"}, "Write a message"},
		{"long message", Input{Name: "a", Message: strings.Repeat("b", 501)}, "under 500"},
		{"link spam", Input{Name: "a", Message: "x a.com b.com c.com d.com"}, "spam"},
		{"javascript homepage", Input{Name: "a", Message: "hi", Homepage: "javascript:alert(1)"}, "http://"},
		{"no host", Input{Name: "a", Message: "hi", Homepage: "https://"}, "http://"},
		{"userinfo", Input{Name: "a", Message: "hi", Homepage: "https://user:pw@example.com"}, "http://"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, problems := Check(tc.in)
			for _, p := range problems {
				if strings.Contains(string(p), tc.want) {
					return
				}
			}
			t.Fatalf("want a problem containing %q, got %v", tc.want, problems)
		})
	}
}

func TestStripLinks(t *testing.T) {
	for in, want := range map[string]string{
		"see https://x.io/a?b=1 now": "see [link removed] now",
		"www.spam.net":               "[link removed]",
		"go to cheap-stuff.shop/buy": "go to [link removed]",
		"ftp://files.example/x":      "[link removed]",
		"version 1.2 is out":         "version 1.2 is out",
		"e.g. this is fine":          "e.g. this is fine",
	} {
		if got := StripLinks(in); got != want {
			t.Errorf("StripLinks(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMask(t *testing.T) {
	for in, want := range map[string]string{
		"what the fuck":          "what the f***",
		"FUUUCK yeah":            "F***** yeah",
		"sh1t happens":           "s*** happens",
		"you bastards":           "you b*******",
		"Dickens and cocktails":  "Dickens and cocktails",
		"classic assassin moves": "classic assassin moves",
		"nice room":              "nice room",
		"motherfucker":           "m***********",
		"sick as heck":           "sick as heck",
		"pass the glass":         "pass the glass",
		"what an a$$":            "what an a**",
		"you asss":               "you a***",
	} {
		if got := Mask(in); got != want {
			t.Errorf("Mask(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMultiLineCapsLines(t *testing.T) {
	got := multiLine(strings.Repeat("line\n", 20))
	if n := strings.Count(got, "\n") + 1; n != MaxLines {
		t.Fatalf("got %d lines, want %d", n, MaxLines)
	}
}

func TestCleanDropsInvisibleCharacters(t *testing.T) {
	if got := singleLine("a" + string(rune(0x200b)) + "b" + string(rune(0)) + "c" + string(rune(0x202e))); got != "abc" {
		t.Fatalf("got %q", got)
	}
}
