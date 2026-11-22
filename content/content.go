// Package content loads the site's writing from words.txt (see the notes at
// the top of that file).
package content

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed words.txt
var wordsTxt string

// Section is one piece of writing: its prompt, and the owner's paragraphs.
type Section struct {
	ID         string
	Prompt     string
	Paragraphs []string
}

// Words maps section IDs to their writing.
type Words map[string]Section

// Load parses the embedded words.txt.
func Load() (Words, error) { return Parse(wordsTxt) }

// Parse reads the words.txt format.
func Parse(src string) (Words, error) {
	w := Words{}
	var cur *Section
	var para []string
	flush := func() {
		if cur != nil && len(para) > 0 {
			cur.Paragraphs = append(cur.Paragraphs, strings.Join(para, " "))
		}
		para = nil
	}
	save := func() {
		flush()
		if cur != nil {
			w[cur.ID] = *cur
		}
	}
	for n, line := range strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n") {
		t := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(t, "#"):
		case strings.HasPrefix(t, "==") && strings.HasSuffix(t, "=="):
			save()
			id := strings.TrimSpace(strings.Trim(t, "="))
			if id == "" {
				return nil, fmt.Errorf("words.txt line %d: empty section id", n+1)
			}
			if _, dup := w[id]; dup {
				return nil, fmt.Errorf("words.txt line %d: section %q appears twice", n+1, id)
			}
			cur = &Section{ID: id}
		case cur == nil && t != "":
			return nil, fmt.Errorf("words.txt line %d: text before the first section", n+1)
		case strings.HasPrefix(t, "?") && cur != nil && cur.Prompt == "" && len(cur.Paragraphs) == 0 && len(para) == 0:
			cur.Prompt = strings.TrimSpace(strings.TrimPrefix(t, "?"))
		case t == "":
			flush()
		default:
			para = append(para, t)
		}
	}
	save()
	return w, nil
}
