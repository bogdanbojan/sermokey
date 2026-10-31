package es

import (
	"embed"
	_ "embed"
	"strings"

	"github.com/bogdanbojan/sermokey/backend/internal/grammar"
	vocab "github.com/bogdanbojan/sermokey/backend/internal/grammar/es"
	"github.com/bogdanbojan/sermokey/backend/internal/reading"
)

// The generated lexicon, produced by cmd/lexgen from a kaikki.org
// Wiktionary extract. Source: English Wiktionary, dual-licensed CC BY-SA
// 4.0 and GFDL; the attribution that requires is shown in the reader's
// word panel.
var (
	//go:embed data/lexicon.tsv.gz
	lexiconGz []byte
	//go:embed data/forms.tsv.gz
	formsGz []byte
)

//go:embed texts/*.txt
var textFS embed.FS

// ReadingCategory is the synthetic vocabulary category that words saved
// while reading fall into when they are not part of the curated
// curriculum. Curated words keep their real category, so saving a word
// from a text touches the card the vocabulary tab already schedules.
const ReadingCategory = "lectura"

// priorityLemmas are the auxiliaries and copulas. They break ties between
// readings the curriculum ranks equally — see reading.LexicalConfig.
var priorityLemmas = map[string]bool{
	"ser":   true,
	"estar": true,
	"haber": true,
	"ir":    true,
	"hacer": true,
	"tener": true,
	"poder": true,
}

// reflexiveBase strips the enclitic pronoun from a reflexive infinitive.
//
// Spanish lemmatises reflexives with the pronoun attached — "levantarse" —
// but in finite use the clitic detaches and precedes the verb: "Clara se
// levanta". That analyses to the bare infinitive "levantar", which matches
// no curated entry, so the word missed the curriculum and was charged to
// the reader as unknown vocabulary. Registering the bare stem as an alias
// is the same fix German needs for its article-prefixed nouns.
func reflexiveBase(w grammar.VocabWord) []string {
	if w.POS != "verb" {
		return nil
	}
	for _, suffix := range []string{"arse", "erse", "irse"} {
		if strings.HasSuffix(w.Lemma, suffix) {
			return []string{strings.TrimSuffix(w.Lemma, "se")}
		}
	}
	return nil
}

var (
	Lexicon  *reading.Lexicon
	Analyzer reading.Analyzer
	Catalog  *reading.MemoryCatalog
	Provider reading.Provider
)

// init order is load bearing: building the catalog tokenizes and analyzes
// every bundled text, so the lexicon and analyzer have to exist first.
func init() {
	Lexicon = reading.LoadLexicon(lexiconGz, formsGz)
	Analyzer = reading.NewStandardAnalyzer(Lexicon, reading.LexicalConfig{
		Curated:         vocab.Vocabulary,
		ClosedClass:     closedClass,
		Normalize:       NormalizeSurface,
		ReadingCategory: ReadingCategory,
		PriorityLemmas:  priorityLemmas,
		CuratedAliases:  reflexiveBase,
		// Spanish capitalises only names and sentence openings, so an
		// initial capital is real evidence of a proper noun.
		CapitalSignalsProperNoun: true,
	})

	tok := reading.PhraseTokenizer{
		Base:     Tokenizer{},
		Phrases:  reading.NewVocabPhrases(vocab.Vocabulary, NormalizeSurface),
		Analyzer: Analyzer,
	}

	Catalog = reading.LoadCorpus(textFS, "texts", tok, Analyzer)
	Provider = reading.Provider{
		Code:      "es",
		Tokenizer: tok,
		Analyzer:  Analyzer,
		Catalog:   Catalog,
	}
	reading.Register(Provider)
}
