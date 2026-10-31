package fr

import (
	"embed"
	_ "embed"

	vocab "github.com/bogdanbojan/sermokey/backend/internal/grammar/fr"
	"github.com/bogdanbojan/sermokey/backend/internal/reading"
)

// The generated lexicon, produced by cmd/lexgen from a kaikki.org
// Wiktionary extract. Committed so that building and testing never requires
// a 600 MB download; see the README for how to regenerate it.
//
// Source: English Wiktionary, dual-licensed CC BY-SA 4.0 and GFDL. The
// attribution that requires is shown in the reader's word panel.
var (
	//go:embed data/lexicon.tsv.gz
	lexiconGz []byte
	//go:embed data/forms.tsv.gz
	formsGz []byte
)

// The bundled corpus. Adding a text means dropping a file in here — the
// glob picks it up and the header inside it supplies the metadata.
//
//go:embed texts/*.txt
var textFS embed.FS

// ReadingCategory is the synthetic vocabulary category that words saved
// while reading fall into when they are not part of the curated
// curriculum. Curated words keep their real category, so that saving
// "maison" from a text touches exactly the card the vocabulary tab already
// schedules rather than creating a second one with its own review state.
const ReadingCategory = "lecture"

// priorityLemmas are the auxiliaries and modals. They break ties between
// readings the curriculum ranks equally — "suis" is the first person of
// both être and suivre, and both are tier-1 verbs. See
// reading.LexicalConfig.PriorityLemmas.
var priorityLemmas = map[string]bool{
	"être":    true,
	"avoir":   true,
	"aller":   true,
	"faire":   true,
	"pouvoir": true,
	"vouloir": true,
	"devoir":  true,
}

var (
	// Lexicon is the generated dictionary and inflected-form index.
	Lexicon *reading.Lexicon
	// Analyzer resolves French surface forms.
	Analyzer reading.Analyzer
	// Catalog is the bundled corpus, tokenized and analyzed at startup so
	// that serving a section never does linguistic work on the request path.
	Catalog *reading.MemoryCatalog
	// Provider bundles French reading support for the HTTP layer.
	Provider reading.Provider
)

// init is the package's single entry point, and the order below is load
// bearing: building the catalog tokenizes and analyzes every bundled text,
// so the lexicon and analyzer have to exist first. Leaving these as
// package-level variable initializers would hand the ordering to Go's
// dependency analysis, which cannot see through the Analyzer interface
// that reading.Build calls and would silently produce a corpus with no
// word frequencies at all.
func init() {
	Lexicon = reading.LoadLexicon(lexiconGz, formsGz)
	Analyzer = reading.NewStandardAnalyzer(Lexicon, reading.LexicalConfig{
		Curated:         vocab.Vocabulary,
		ClosedClass:     closedClass,
		Elisions:        elisions,
		Normalize:       NormalizeSurface,
		ReadingCategory: ReadingCategory,
		PriorityLemmas:  priorityLemmas,
		// French capitalises only names and sentence openings, so an
		// initial capital is real evidence of a proper noun.
		CapitalSignalsProperNoun: true,
	})

	// PhraseTokenizer, not the bare Tokenizer: multi-word items have to
	// become single tokens before anything downstream sees them.
	tok := reading.PhraseTokenizer{
		Base:     Tokenizer{},
		Phrases:  reading.NewVocabPhrases(vocab.Vocabulary, NormalizeSurface),
		Analyzer: Analyzer,
	}

	Catalog = reading.LoadCorpus(textFS, "texts", tok, Analyzer)
	Provider = reading.Provider{
		Code:      "fr",
		Tokenizer: tok,
		Analyzer:  Analyzer,
		Catalog:   Catalog,
	}
	reading.Register(Provider)
}
