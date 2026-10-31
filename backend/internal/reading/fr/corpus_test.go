package fr

import (
	"strings"
	"testing"

	vocab "github.com/bogdanbojan/sermokey/backend/internal/grammar/fr"
	"github.com/bogdanbojan/sermokey/backend/internal/reading"
)

func curatedWordIDs() map[string]bool {
	ids := make(map[string]bool, len(vocab.Vocabulary))
	for _, w := range vocab.Vocabulary {
		ids[w.ID] = true
	}
	return ids
}

// The corpus check is the reading equivalent of grammar.ValidateVocab: it
// runs the real tokenizer and analyzer over every bundled text and reports
// every problem at once, so authoring a new text cannot silently ship a
// paragraph the reader renders wrongly.
//
// The coverage floor is what keeps the lexicon and the corpus honest in
// both directions: a text too obscure for the shipped dictionary fails, and
// so does a regression that breaks deinflection.
func TestCorpusIsWellFormed(t *testing.T) {
	errs := reading.ValidateCorpus(reading.CorpusCheck{
		Provider:       Provider,
		MinCoverage:    0.95,
		CuratedWordIDs: curatedWordIDs(),
	})
	for _, err := range errs {
		t.Error(err)
	}
}

func TestCorpusHasThreeTextsAcrossLevels(t *testing.T) {
	metas := Catalog.List()
	if len(metas) != 3 {
		t.Fatalf("got %d texts, want 3", len(metas))
	}
	levels := map[reading.Level]bool{}
	for _, m := range metas {
		levels[m.Level] = true
		if m.TokenCount < 150 {
			t.Errorf("%s: only %d word tokens, too short to read", m.ID, m.TokenCount)
		}
		if len(m.WordFrequency) == 0 {
			t.Errorf("%s: empty word frequency table, so coverage cannot be computed", m.ID)
		}
	}
	for _, want := range []reading.Level{reading.LevelA2, reading.LevelB1, reading.LevelB2} {
		if !levels[want] {
			t.Errorf("no text at level %s", want)
		}
	}
}

// WordFrequency drives the per-reader coverage figure on every text card.
// If it counted types rather than tokens, or drifted from the actual token
// count, the 95%/98% comprehension thresholds it is compared against would
// be meaningless.
func TestWordFrequencyIsTokenWeighted(t *testing.T) {
	for _, m := range Catalog.List() {
		var total int
		for _, n := range m.WordFrequency {
			total += n
		}
		if total == 0 || total > m.TokenCount {
			t.Errorf("%s: word frequencies sum to %d, but the text has %d word tokens",
				m.ID, total, m.TokenCount)
		}
		if len(m.WordFrequency) >= total {
			t.Errorf("%s: %d distinct words for %d counted tokens — not token-weighted",
				m.ID, len(m.WordFrequency), total)
		}
	}
}

// The frequency table must be keyed by SRS card id, not by lemma: that is
// what lets the client intersect it with the deck without guessing at a
// curated word's category.
func TestWordFrequencyIsKeyedByCardID(t *testing.T) {
	for _, m := range Catalog.List() {
		for id := range m.WordFrequency {
			if !strings.Contains(id, "|") {
				t.Errorf("%s: frequency key %q is not a card id", m.ID, id)
			}
		}
	}
}

// The graded texts exist to be comprehensible against the curated deck, and
// the literary one exists to prove the lexicon survives authentic French.
// Asserting the gradient keeps the corpus from silently drifting into three
// texts of the same difficulty.
//
// Note what is deliberately NOT asserted: a 95% floor. The curated
// vocabulary is a study list and omits the closed-class core — articles,
// clitic pronouns and possessives ("le", "la", "il", "ne", "se") — which
// alone are roughly a quarter of running tokens, so curated coverage tops
// out near 75% however simple the prose. The reader-facing coverage figure
// is computed against the reader's own deck plus the words they have marked
// known, which does include those, and that is the number the 95%/98%
// comprehension thresholds (Hu & Nation 2000) apply to.
func TestGradedTextsAreEasierThanTheLiteraryOne(t *testing.T) {
	curated := curatedWordIDs()
	coverage := func(m reading.DocumentMeta) float64 {
		var known, total int
		for id, n := range m.WordFrequency {
			total += n
			if curated[id] {
				known += n
			}
		}
		if total == 0 {
			return 0
		}
		return float64(known) / float64(total)
	}

	byID := map[string]float64{}
	for _, m := range Catalog.List() {
		byID[m.ID] = coverage(m)
		t.Logf("%-28s %s  curated coverage %.1f%%", m.ID, m.Level, byID[m.ID]*100)
	}

	a2, b1, b2 := byID["le-marche-du-samedi"], byID["une-lettre-de-marseille"], byID["la-parure"]
	if !(a2 > b1 && b1 > b2) {
		t.Errorf("curated coverage is not monotonically decreasing with level: A2=%.1f%% B1=%.1f%% B2=%.1f%%",
			a2*100, b1*100, b2*100)
	}
	// A floor on the easiest text: if it drops this low, either the prose
	// has drifted away from the curriculum or deinflection has regressed.
	if a2 < 0.65 {
		t.Errorf("A2 text covers only %.1f%% of tokens with curated vocabulary, want >= 65%%", a2*100)
	}
}

// Closed-class words are kept out of WordFrequency (they are grammar, not
// vocabulary) but still counted, because the comprehension thresholds the
// coverage figure is compared against are defined over all running words.
// Both halves of that have to hold, or the ratio is wrong in one direction
// or the other.
func TestFunctionTokensAreCountedButNotVocabulary(t *testing.T) {
	for _, m := range Catalog.List() {
		var vocabTokens int
		for _, n := range m.WordFrequency {
			vocabTokens += n
		}
		if m.FunctionTokenCount == 0 {
			t.Errorf("%s: no function tokens counted", m.ID)
		}
		if vocabTokens+m.FunctionTokenCount+m.ProperTokenCount > m.TokenCount {
			t.Errorf("%s: %d vocabulary + %d function + %d proper tokens exceeds %d total",
				m.ID, vocabTokens, m.FunctionTokenCount, m.ProperTokenCount, m.TokenCount)
		}
		// None of the closed-class words may appear as a vocabulary entry.
		for id := range m.WordFrequency {
			lemma, _, _ := strings.Cut(id, "|")
			if isFunctionWord(lemma) {
				t.Errorf("%s: function word %q appears in WordFrequency", m.ID, lemma)
			}
		}
	}
}

// Character names are not vocabulary. Before mid-sentence capitalisation
// was used as evidence, names that collide with an ordinary word were read
// as that word — Spanish "Clara" as the adjective claro, "Carmen" as a
// common noun — which both glossed them wrongly and charged them to the
// reader's unknown-word count.
func TestCharacterNamesAreNotCountedAsVocabulary(t *testing.T) {
	for _, m := range Catalog.List() {
		doc, _ := Catalog.Get(m.ID)
		if len(doc.ProperNouns) == 0 {
			continue // a text may legitimately contain no names
		}
		analyze := reading.ProperNounAware(Provider.Analyzer, doc.ProperNouns)
		for name := range doc.ProperNouns {
			as := analyze(name)
			if len(as) == 0 {
				continue // not in the dictionary at all; harmless
			}
			if !as[0].Proper {
				continue // not a known name; nothing to assert
			}
			if _, counted := m.WordFrequency[as[0].WordID]; counted {
				t.Errorf("%s: name %q is counted as vocabulary", m.ID, name)
			}
		}
	}
}
