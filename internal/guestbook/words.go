package guestbook

import (
	"strings"
	"unicode"
)

// Word lists for the guestbook's auto-moderation. "anywhere" words are
// masked even inside longer words; "whole" words only when they stand
// alone (with common endings), so "Dickens" and "cocktail" survive.
var (
	anywhere = []string{"fuck", "shit", "cunt", "nigger", "nigga", "faggot", "motherfucker"}
	whole    = []string{"ass", "asshole", "bitch", "dick", "cock", "pussy", "bastard", "slut", "whore", "fag", "retard", "twat", "wanker", "kike", "spic", "chink", "dyke", "tranny"}
	endings  = []string{"", "s", "es", "ed", "er", "ers", "ing", "y"}
)

// leet maps look-alike characters to letters before matching.
var leet = map[rune]rune{'0': 'o', '1': 'i', '!': 'i', '3': 'e', '4': 'a', '@': 'a', '5': 's', '$': 's', '7': 't', '+': 't'}

// Mask replaces each bad word with its first letter and asterisks.
func Mask(s string) string {
	runes := []rune(s)
	var out []rune
	for i := 0; i < len(runes); {
		if !isWordRune(runes[i]) {
			out = append(out, runes[i])
			i++
			continue
		}
		j := i
		for j < len(runes) && isWordRune(runes[j]) {
			j++
		}
		word := runes[i:j]
		if bad(word) {
			out = append(out, word[0])
			for range word[1:] {
				out = append(out, '*')
			}
		} else {
			out = append(out, word...)
		}
		i = j
	}
	return string(out)
}

func isWordRune(r rune) bool {
	_, l := leet[r]
	return unicode.IsLetter(r) || unicode.IsDigit(r) || l
}

// plain lowercases and undoes look-alike characters.
func plain(w []rune) string {
	var b strings.Builder
	for _, r := range w {
		r = unicode.ToLower(r)
		if m, ok := leet[r]; ok {
			r = m
		}
		b.WriteRune(r)
	}
	return b.String()
}

// collapse also squeezes repeated letters: "fuuuck" -> "fuck".
func collapse(s string) string {
	var b strings.Builder
	var last rune
	for _, r := range s {
		if r != last {
			b.WriteRune(r)
		}
		last = r
	}
	return b.String()
}

func bad(word []rune) bool {
	p := plain(word)
	c := collapse(p)
	for _, a := range anywhere {
		if strings.Contains(c, collapse(a)) {
			return true
		}
	}
	for _, x := range whole {
		for _, e := range endings {
			full := x + e
			// Collapsed matches only count when the word is at least as long
			// as the bad one, so "as" isn't mistaken for "ass".
			if p == full || (c == collapse(full) && len(p) >= len(full)) {
				return true
			}
		}
	}
	return false
}
