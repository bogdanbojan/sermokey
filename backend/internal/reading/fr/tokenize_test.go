package fr

import (
	"testing"

	"github.com/bogdanbojan/sermokey/backend/internal/reading"
)

// words returns just the word tokens' surfaces, which is what the splitting
// rules are actually about.
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

func TestTokenizeElision(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		// Live elision: the clitic and the word it attaches to are
		// separate lexemes and each needs its own lookup.
		{"definite article", "l'homme", []string{"l'", "homme"}},
		{"subject pronoun", "j'ai", []string{"j'", "ai"}},
		{"complementizer", "qu'il", []string{"qu'", "il"}},
		{"preposition de", "d'un", []string{"d'", "un"}},
		{"negation", "n'est", []string{"n'", "est"}},
		{"object pronoun", "s'appelle", []string{"s'", "appelle"}},
		{"chained", "qu'il n'y", []string{"qu'", "il", "n'", "y"}},

		// Frozen apostrophes: one lexeme. Splitting these produces
		// fragments ("aujourd", "hui") that are in no dictionary.
		{"aujourd'hui", "aujourd'hui", []string{"aujourd'hui"}},
		{"quelqu'un", "quelqu'un", []string{"quelqu'un"}},
		{"presqu'île", "presqu'île", []string{"presqu'île"}},
		{"d'abord", "d'abord", []string{"d'abord"}},

		// Capitalisation must not defeat the exception tables.
		{"sentence-initial", "Aujourd'hui", []string{"Aujourd'hui"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := words(t, tc.in)
			assertEqual(t, got, tc.want)
		})
	}
}

func TestTokenizeHyphens(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		// Productive hyphens join independent words.
		{"interrogative", "est-ce", []string{"est", "ce"}},
		// The euphonic t in inversion is a liaison consonant, not a word:
		// it has no dictionary entry, so it must not become a hoverable
		// span.
		{"euphonic t", "parle-t-il", []string{"parle", "il"}},
		{"euphonic t capitalised", "Demande-T-il", []string{"Demande", "il"}},
		// A real word that happens to be one letter between hyphens is
		// not affected, because only t is euphonic.
		{"clitic y between hyphens", "va-y-en", []string{"va", "y", "en"}},
		{"clitic y", "vas-y", []string{"vas", "y"}},
		{"imperative clitic", "donne-moi", []string{"donne", "moi"}},

		// Lexicalised hyphens do not.
		{"peut-être", "peut-être", []string{"peut-être"}},
		{"après-midi", "après-midi", []string{"après-midi"}},
		{"grand-père", "grand-père", []string{"grand-père"}},
		{"rendez-vous", "rendez-vous", []string{"rendez-vous"}},
		{"mixed separators", "c'est-à-dire", []string{"c'est-à-dire"}},

		// An em dash is punctuation, not a word joiner.
		{"em dash", "oui — non", []string{"oui", "non"}},
		{"spaced hyphen", "oui - non", []string{"oui", "non"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertEqual(t, words(t, tc.in), tc.want)
		})
	}
}

// Published French uses U+2019; keyboards produce U+0027. The same word must
// tokenize identically either way, or a word looks up fine in one document
// and fails in the next.
func TestTokenizeApostropheVariants(t *testing.T) {
	for _, tc := range []struct{ name, straight, curly string }{
		{"elided article", "l'homme", "l\u2019homme"},
		{"frozen", "aujourd'hui", "aujourd\u2019hui"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, b := words(t, tc.straight), words(t, tc.curly)
			if len(a) != len(b) {
				t.Fatalf("different token counts: %v vs %v", a, b)
			}
			for i := range a {
				if foldApostrophes(a[i]) != foldApostrophes(b[i]) {
					t.Errorf("token %d: %q vs %q", i, a[i], b[i])
				}
			}
		})
	}
}

func TestTokenizeTypography(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"guillemets", "« Bonjour »", []string{"Bonjour"}},
		// French puts a narrow no-break space before ? ! ; : — it is
		// whitespace and must not glue onto the preceding word.
		{"narrow nbsp", "Quoi\u202f?", []string{"Quoi"}},
		{"nbsp", "Quoi\u00a0!", []string{"Quoi"}},
		{"ligature oe", "cœur", []string{"cœur"}},
		{"ligature ae", "curriculum vitæ", []string{"curriculum", "vitæ"}},
		{"ellipsis", "Alors…", []string{"Alors"}},
		{"digits", "En 1920, il", []string{"En", "1920", "il"}},
		{"decimal comma", "3,5 kilos", []string{"3", "5", "kilos"}},
		// A decomposed accent (e + combining acute) is one word, not two.
		{"decomposed accent", "e\u0301te\u0301", []string{"e\u0301te\u0301"}},
		{"quoted word", "'bonjour'", []string{"bonjour"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertEqual(t, words(t, tc.in), tc.want)
		})
	}
}

// The round-trip property is the one invariant that must never break: a
// reader that drops or reorders a byte corrupts the text the learner is
// trying to read.
func TestTokenizeRoundTrip(t *testing.T) {
	inputs := []string{
		"",
		" ",
		"Bonjour.",
		"L'homme qu'il a vu aujourd'hui n'était pas là.",
		"« Est-ce que vous parlez français ? » demanda-t-elle.",
		"Il y a 3,5 kilos de pommes — presque 8 livres — dans le sac.",
		"Peut-être… mais c'est-à-dire que je n'en sais rien !",
		"\u202f\u00a0\t\n",
		"---",
		"'''",
		"Ça coûte 12 € ; c'est cher.",
	}
	for _, in := range inputs {
		if err := reading.AssertRoundTrip(Tokenizer{}, in); err != nil {
			t.Errorf("round trip failed for %q: %v", in, err)
		}
	}
}

func TestNormalizeSurface(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Bonjour", "bonjour"},
		{"L\u2019", "l"},
		{"qu'", "qu"},
		{"  Été  ", "été"},
		// Accents are lexical in French; folding them would merge
		// genuinely different words.
		{"où", "où"},
		{"sûr", "sûr"},
	}
	for _, tc := range tests {
		if got := NormalizeSurface(tc.in); got != tc.want {
			t.Errorf("NormalizeSurface(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func assertEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d tokens %v, want %d %v", len(got), got, len(want), want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("token %d: got %q, want %q (full: %v)", i, got[i], want[i], got)
		}
	}
}

// Multi-word items must survive tokenization as single tokens, in the
// inflected shapes running text actually uses. Split apart they are
// glossed in the most misleading way available: "terre" as earth in the
// middle of a potato, "parce" as nothing at all.
func TestTokenizeMergesCuratedPhrases(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"citation form", "une pomme de terre", []string{"une", "pomme de terre"}},
		{"inflected plural", "les pommes de terre", []string{"les", "pommes de terre"}},
		{"conjugated reflexive", "elle se lève tôt", []string{"elle", "se lève", "tôt"}},
		{"connector", "parce que c'est tard", []string{"parce que", "c'", "est", "tard"}},
		{"longest match wins", "la salle de bains", []string{"la", "salle de bains"}},
		// A sequence that merely starts like a phrase must not be merged.
		{"no false merge", "la salle est grande", []string{"la", "salle", "est", "grande"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got []string
			for _, tok := range phraseTokenizer().Tokenize(tc.in) {
				if tok.Word {
					got = append(got, tok.Surface(tc.in))
				}
			}
			assertEqual(t, got, tc.want)
		})
	}
}

// Merging must not break the one invariant the reader depends on.
func TestPhraseTokenizerRoundTrips(t *testing.T) {
	for _, in := range []string{
		"Elle se lève et achète des pommes de terre, parce que c'est bon marché.",
		"La salle de bains, la salle à manger et le centre commercial.",
		"",
		"se",
	} {
		if err := reading.AssertRoundTrip(phraseTokenizer(), in); err != nil {
			t.Errorf("round trip failed for %q: %v", in, err)
		}
	}
}

// phraseTokenizer returns the tokenizer the corpus actually uses: the
// French orthographic tokenizer plus multi-word merging.
func phraseTokenizer() reading.Tokenizer { return Provider.Tokenizer }
