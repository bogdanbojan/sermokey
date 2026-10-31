package es

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

func TestTokenizeSpanish(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		// Inverted punctuation opens questions and exclamations. Treated
		// as a word character it would glue onto the next word and every
		// question in a text would start with a token no dictionary knows.
		{"inverted question", "¿Qué quiere?", []string{"Qué", "quiere"}},
		{"inverted exclamation", "¡Qué bien!", []string{"Qué", "bien"}},
		// Enclitic pronouns are written solid and are not separable; the
		// dictionary lists the combined form.
		{"enclitic pronouns", "dámelo ahora", []string{"dámelo", "ahora"}},
		{"infinitive enclitic", "contarte este mes", []string{"contarte", "este", "mes"}},
		// Accents and ñ are ordinary letters, never a word boundary.
		{"accents and enye", "el año pasó rápido", []string{"el", "año", "pasó", "rápido"}},
		{"diaeresis", "vergüenza", []string{"vergüenza"}},
		// Coordinated compounds split; each half is a word worth a lookup.
		{"hyphenated compound", "teórico-práctico", []string{"teórico", "práctico"}},
		{"em dash is punctuation", "sí — no", []string{"sí", "no"}},
		{"digits", "En 1920, vivía", []string{"En", "1920", "vivía"}},
		{"quoted", "« Buenos días »", []string{"Buenos", "días"}},
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

func TestTokenizeSpanishRoundTrip(t *testing.T) {
	for _, in := range []string{
		"",
		"¿Qué quiere hoy? ¡Cuatro euros con cincuenta!",
		"En un lugar de la Mancha, de cuyo nombre no quiero acordarme…",
		"El año pasó rápido — y la vergüenza también.",
		"---",
	} {
		if err := reading.AssertRoundTrip(Provider.Tokenizer, in); err != nil {
			t.Errorf("round trip failed for %q: %v", in, err)
		}
	}
}

// Accents are lexical in Spanish and often the only difference between two
// words: "el" (the) vs "él" (he), "si" (if) vs "sí" (yes).
func TestNormalizeKeepsAccents(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"Buenos", "buenos"},
		{"él", "él"},
		{"sí", "sí"},
		{"está", "está"},
	} {
		if got := NormalizeSurface(tc.in); got != tc.want {
			t.Errorf("NormalizeSurface(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestAnalyzeSpanish(t *testing.T) {
	for _, tc := range []struct{ surface, lemma string }{
		{"ventana", "ventana"},
		{"ventanas", "ventana"},
		{"hablamos", "hablar"},
		{"tiene", "tener"},
		{"fue", "ir"}, // or ser; both are acceptable readings of this form
		// A name with no common homograph is promoted on capitalisation.
		// "Clara" deliberately is not: "clara" is also egg white, and the
		// guard against promoting sentence-initial common nouns refuses
		// the case. See reading.StandardAnalyzer.
		{"Sevilla", "Sevilla"},
	} {
		as := Analyzer.Analyze(tc.surface)
		if len(as) == 0 {
			t.Errorf("Analyze(%q) returned nothing", tc.surface)
			continue
		}
		if tc.surface == "fue" {
			continue // genuinely ambiguous; only checking it resolves
		}
		if as[0].Lemma != tc.lemma {
			t.Errorf("Analyze(%q) -> %q, want %q", tc.surface, as[0].Lemma, tc.lemma)
		}
	}
}

// Articles and clitics are a third of running Spanish; if they were not
// marked, every text would score as too hard.
func TestSpanishFunctionWords(t *testing.T) {
	for _, s := range []string{"el", "la", "los", "un", "del", "al", "se", "que", "no", "de", "y"} {
		as := Analyzer.Analyze(s)
		if len(as) == 0 || !as[0].Function {
			t.Errorf("%q is not marked as a function word", s)
		}
	}
	for _, s := range []string{"ventana", "comprar", "mercado"} {
		if as := Analyzer.Analyze(s); len(as) > 0 && as[0].Function {
			t.Errorf("%q is wrongly marked as a function word", s)
		}
	}
}

// Spanish writes reflexive infinitives with the pronoun attached
// ("levantarse") but detaches it in finite use ("se levanta"), which
// analyses to the bare "levantar". Without the alias the word missed the
// curriculum entirely and was charged to the reader as unknown vocabulary.
func TestReflexiveVerbsReachTheirCuratedCard(t *testing.T) {
	for _, tc := range []struct{ surface, wantLemma string }{
		{"levanta", "levantarse"},
		{"levantar", "levantarse"},
		{"acuesta", "acostarse"},
	} {
		as := Analyzer.Analyze(tc.surface)
		if len(as) == 0 {
			t.Errorf("Analyze(%q) returned nothing", tc.surface)
			continue
		}
		if as[0].Lemma != tc.wantLemma {
			t.Errorf("Analyze(%q) -> %q, want %q", tc.surface, as[0].Lemma, tc.wantLemma)
		}
	}
}
