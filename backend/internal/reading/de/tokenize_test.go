package de

import (
	"testing"

	"github.com/bogdanbojan/sermokey/backend/internal/reading"
)

func words(t *testing.T, text string) []string {
	t.Helper()
	var out []string
	for _, tok := range (Tokenizer{}).Tokenize(text) {
		if tok.Word {
			out = append(out, tok.Surface(text))
		}
	}
	return out
}

func TestTokenizeGerman(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"umlauts and eszett", "Die Straße öffnet früh", []string{"Die", "Straße", "öffnet", "früh"}},
		// Solid compounds are one word; splitting them is a research
		// problem of its own and out of scope.
		{"solid compound", "der Samstagsmarkt", []string{"der", "Samstagsmarkt"}},
		// Hyphenated compounds are kept whole: the hyphenated form is what
		// the dictionary lists, and it is what the reader perceives as one
		// item.
		{"hyphenated compound", "Kaffee-Ersatz", []string{"Kaffee-Ersatz"}},
		{"em dash is punctuation", "ja — nein", []string{"ja", "nein"}},
		{"digits", "Im Jahr 1920 kam er", []string{"Im", "Jahr", "1920", "kam", "er"}},
		{"guillemets", "« Guten Tag »", []string{"Guten", "Tag"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := words(t, tc.in)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("token %d: got %q, want %q (full %v)", i, got[i], tc.want[i], got)
				}
			}
		})
	}
}

func TestTokenizeGermanRoundTrip(t *testing.T) {
	for _, in := range []string{
		"",
		"Es war einmal ein kleines süßes Mädchen, das hatte jedermann lieb.",
		"« Guten Tag, Rotkäppchen », sprach er. — « Schönen Dank, Wolf. »",
		"Der Kaffee-Ersatz kostete 4,50 €.",
	} {
		if err := reading.AssertRoundTrip(Provider.Tokenizer, in); err != nil {
			t.Errorf("round trip failed for %q: %v", in, err)
		}
	}
}

// The vocabulary list stores nouns with their article ("der Hund") because
// the gender is half of what has to be learned. Running text has the bare
// noun. If the alias were missing, every German noun would miss the
// curated layer and be saved under a different card id than the one the
// vocabulary tab schedules, silently splitting its review history.
func TestCuratedNounsResolveFromTheirBareForm(t *testing.T) {
	for _, tc := range []struct{ surface, lemma string }{
		{"Haus", "das Haus"},
		{"Fenster", "das Fenster"},
		{"Markt", "der Markt"},
		{"Kinder", "das Kind"}, // inflected, via the form index
		{"Tische", "der Tisch"},
	} {
		as := Analyzer.Analyze(tc.surface)
		if len(as) == 0 {
			t.Errorf("Analyze(%q) returned nothing", tc.surface)
			continue
		}
		if as[0].Lemma != tc.lemma {
			t.Errorf("Analyze(%q) -> %q, want %q", tc.surface, as[0].Lemma, tc.lemma)
			continue
		}
		if as[0].Source != reading.SourceCurated || as[0].WordID == "" {
			t.Errorf("Analyze(%q): source %v, id %q — not linked to the deck",
				tc.surface, as[0].Source, as[0].WordID)
		}
	}
}

// German capitalises every noun, so an initial capital says nothing about
// whether a word is a name. Acting on it would promote half the words on
// a page to proper nouns and drop them from the vocabulary count.
func TestCapitalisationIsNotAProperNounSignal(t *testing.T) {
	for _, s := range []string{"Haus", "Markt", "Fenster", "Brot"} {
		as := Analyzer.Analyze(s)
		if len(as) == 0 {
			t.Fatalf("Analyze(%q) returned nothing", s)
		}
		if as[0].Proper {
			t.Errorf("Analyze(%q) was promoted to a proper noun", s)
		}
	}
}

func TestGermanFunctionWords(t *testing.T) {
	for _, s := range []string{"der", "die", "das", "den", "dem", "ein", "eine", "ich", "sie", "nicht", "und", "zu"} {
		as := Analyzer.Analyze(s)
		if len(as) == 0 || !as[0].Function {
			t.Errorf("%q is not marked as a function word", s)
		}
	}
	for _, s := range []string{"Fenster", "kaufen", "Markt"} {
		if as := Analyzer.Analyze(s); len(as) > 0 && as[0].Function {
			t.Errorf("%q is wrongly marked as a function word", s)
		}
	}
}

// "Hause" is both the dative of Haus and a form of the verb hausen. With
// only one analysis kept per surface the verb won, and "nach Hause" was
// glossed "to dwell".
func TestAmbiguousFormsPreferTheCuratedReading(t *testing.T) {
	as := Analyzer.Analyze("Hause")
	if len(as) == 0 {
		t.Fatal("Analyze(\"Hause\") returned nothing")
	}
	if as[0].Lemma != "das Haus" {
		t.Errorf("Analyze(\"Hause\") -> %q, want the noun das Haus", as[0].Lemma)
	}
}

// The vocabulary list writes nouns as "das Fenster" so the flashcard
// teaches the gender, but on the page those are two words: a determiner
// and a noun, each of which a reader may want to look up. Treating them as
// multi-word expressions merged them into one token and made both
// unhoverable.
func TestArticleNounPairsAreNotMergedIntoOneToken(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want []string
	}{
		{"das Fenster", []string{"das", "Fenster"}},
		{"der Markt ist offen", []string{"der", "Markt", "ist", "offen"}},
		{"die Küche", []string{"die", "Küche"}},
	} {
		var got []string
		for _, tok := range Provider.Tokenizer.Tokenize(tc.in) {
			if tok.Word {
				got = append(got, tok.Surface(tc.in))
			}
		}
		if len(got) != len(tc.want) {
			t.Errorf("Tokenize(%q) = %v, want %v", tc.in, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("Tokenize(%q) = %v, want %v", tc.in, got, tc.want)
				break
			}
		}
	}

	// The bare noun must still reach its curated card.
	if as := Analyzer.Analyze("Fenster"); len(as) == 0 || as[0].WordID != "das Fenster|haus" {
		t.Errorf("Analyze(\"Fenster\") lost its deck link: %+v", as)
	}
}
