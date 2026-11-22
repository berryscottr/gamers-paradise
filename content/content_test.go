package content

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	w, err := Parse(`# note
== a ==
? What's up?
First line
continues here.

Second paragraph.
# a note inside, ignored

== b ==
? Empty one
`)
	if err != nil {
		t.Fatal(err)
	}
	if got := w["a"]; got.Prompt != "What's up?" || !reflect.DeepEqual(got.Paragraphs, []string{"First line continues here.", "Second paragraph."}) {
		t.Fatalf("a = %+v", got)
	}
	if got := w["b"]; got.Prompt != "Empty one" || len(got.Paragraphs) != 0 {
		t.Fatalf("b = %+v", got)
	}
}

func TestParseQuestionMarkInsideWritingIsText(t *testing.T) {
	w, err := Parse("== a ==\n? prompt\nWhy not?\n? still me talking\n")
	if err != nil {
		t.Fatal(err)
	}
	if got := w["a"].Paragraphs; !reflect.DeepEqual(got, []string{"Why not? ? still me talking"}) {
		t.Fatalf("got %q", got)
	}
}

func TestParseErrors(t *testing.T) {
	for _, src := range []string{"stray text\n== a ==", "== a ==\n== a ==", "== ==\n"} {
		if _, err := Parse(src); err == nil {
			t.Errorf("Parse(%q) accepted", src)
		}
	}
}

func TestEmbeddedWordsLoad(t *testing.T) {
	w, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for id, s := range w {
		if s.Prompt == "" {
			t.Errorf("section %q has no prompt", id)
		}
	}
}
