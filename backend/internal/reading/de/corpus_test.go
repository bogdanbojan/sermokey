package de

import (
	"strings"
	"testing"

	vocab "github.com/bogdanbojan/sermokey/backend/internal/grammar/de"
	"github.com/bogdanbojan/sermokey/backend/internal/reading"
)

func curatedWordIDs() map[string]bool {
	ids := make(map[string]bool, len(vocab.Vocabulary))
	for _, w := range vocab.Vocabulary {
		ids[w.ID] = true
	}
	return ids
}

// The corpus check runs the real tokenizer and analyzer over every bundled
// text and reports every problem at once, so authoring a new text cannot
// silently ship a paragraph the reader renders wrongly.
//
// The coverage floor keeps the lexicon and the corpus honest in both
// directions: a text too obscure for the shipped dictionary fails, and so
// does a regression that breaks deinflection.
func TestCorpusIsWellFormed(t *testing.T) {
	for _, err := range reading.ValidateCorpus(reading.CorpusCheck{
		Provider:       Provider,
		MinCoverage:    0.95,
		CuratedWordIDs: curatedWordIDs(),
	}) {
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

// Closed-class words are kept out of WordFrequency (they are grammar, not
// vocabulary) but still counted, because the comprehension thresholds the
// coverage figure is compared against are defined over all running words.
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
			t.Errorf("%s: %d vocabulary + %d function + %d proper exceeds %d total",
				m.ID, vocabTokens, m.FunctionTokenCount, m.ProperTokenCount, m.TokenCount)
		}
		for id := range m.WordFrequency {
			lemma, _, _ := strings.Cut(id, "|")
			if isFunctionWord(lemma) {
				t.Errorf("%s: function word %q appears in WordFrequency", m.ID, lemma)
			}
		}
	}
}

// The graded texts exist to be comprehensible against the curated deck and
// the literary one to prove the lexicon survives authentic prose. Asserting
// the gradient stops the corpus drifting into three texts of one difficulty.
func TestGradedTextsAreEasierThanTheLiteraryOne(t *testing.T) {
	curated := curatedWordIDs()
	coverage := func(m reading.DocumentMeta) float64 {
		known, total := m.FunctionTokenCount+m.ProperTokenCount, m.FunctionTokenCount+m.ProperTokenCount
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
	var cov []float64
	for _, m := range Catalog.List() {
		c := coverage(m)
		cov = append(cov, c)
		t.Logf("%-28s %s  coverage %.1f%%", m.ID, m.Level, c*100)
	}
	if !(cov[0] > cov[2] && cov[1] > cov[2]) {
		t.Errorf("the literary text is not the hardest: %.1f%% / %.1f%% / %.1f%%",
			cov[0]*100, cov[1]*100, cov[2]*100)
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
