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
		if bad(normalize(word)) {
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

func normalize(w []rune) string {
	var b strings.Builder
	var last rune
	for _, r := range w {
		r = unicode.ToLower(r)
		if m, ok := leet[r]; ok {
			r = m
		}
		if r == last { // "fuuuck" -> "fuck"
			continue
		}
		last = r
		b.WriteRune(r)
	}
	return b.String()
}

func bad(w string) bool {
	for _, a := range anywhere {
		if strings.Contains(w, dedupe(a)) {
			return true
		}
	}
	for _, x := range whole {
		for _, e := range endings {
			if w == dedupe(x+e) {
				return true
			}
		}
	}
	return false
}

// dedupe collapses doubled letters the same way normalize does.
func dedupe(s string) string { return normalize([]rune(s)) }
