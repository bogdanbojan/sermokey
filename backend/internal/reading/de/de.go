package de

import (
	"embed"
	_ "embed"
	"strings"

	"github.com/bogdanbojan/sermokey/backend/internal/grammar"
	vocab "github.com/bogdanbojan/sermokey/backend/internal/grammar/de"
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

// ReadingCategory is the synthetic vocabulary category for words saved
// while reading that are not part of the curated curriculum.
const ReadingCategory = "lektuere"

// germanArticles are the three definite articles the vocabulary list
// prefixes to every noun.
var germanArticles = []string{"der ", "die ", "das "}

// bareNoun strips the article the curated vocabulary stores with each noun.
//
// The list deliberately records "der Hund" rather than "Hund", because for
// German the gender is half of what has to be memorised and a flashcard
// without it teaches an incomplete word. Running text, of course, contains
// the bare noun. Registering the stripped form as an alias keeps the two in
// agreement: hovering "Hund" finds the curated entry and saves to the very
// card the vocabulary tab schedules, instead of minting a second one and
// splitting that word's review history.
func bareNoun(w grammar.VocabWord) []string {
	for _, art := range germanArticles {
		if rest, found := strings.CutPrefix(w.Lemma, art); found {
			return []string{rest}
		}
	}
	return nil
}

// germanSuffixes replaces the Romance defaults. German pluralises with -e,
// -en, -er and umlaut rather than -s, and stripping a final -s from a
// German word is almost always wrong: it is far more often a genitive or
// simply part of the stem.
var germanSuffixes = []reading.SuffixRule{
	{Suffix: "en", Replacement: "", Note: "plural or inflected form"},
	{Suffix: "er", Replacement: "", Note: "plural or comparative"},
	{Suffix: "e", Replacement: "", Note: "plural or inflected form"},
	{Suffix: "n", Replacement: "", Note: "plural or inflected form"},
	{Suffix: "es", Replacement: "", Note: "genitive"},
	{Suffix: "s", Replacement: "", Note: "genitive"},
}

// priorityLemmas are the auxiliaries and copulas. They break ties between
// readings the curriculum ranks equally — see reading.LexicalConfig.
var priorityLemmas = map[string]bool{
	"sein":   true,
	"haben":  true,
	"werden": true,
	"können": true,
	"müssen": true,
	"wollen": true,
}

// phraseCandidates is the vocabulary the multi-word index is built from.
//
// It excludes the article-prefixed noun lemmas, which make up most of the
// list. "das Fenster" is two words on the page — a determiner and a noun,
// each worth hovering — and is written that way in the deck only so the
// flashcard teaches the gender. Feeding it to the phrase index merged the
// pair into a single token, so a reader could no longer look up either
// half. Genuine German multi-word entries such as "es gibt" are kept.
func phraseCandidates() []grammar.VocabWord {
	out := make([]grammar.VocabWord, 0, len(vocab.Vocabulary))
	for _, w := range vocab.Vocabulary {
		if bareNoun(w) != nil {
			continue // article + noun, not a multi-word expression
		}
		out = append(out, w)
	}
	return out
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
		CuratedAliases:  bareNoun,
		DerivedSuffixes: germanSuffixes,
		// German capitalises every noun, so an initial capital says
		// nothing about whether a word is a name. Acting on it would
		// promote half the words on the page to proper nouns and drop
		// them from the vocabulary count entirely.
		CapitalSignalsProperNoun: false,
	})

	tok := reading.PhraseTokenizer{
		Base:     Tokenizer{},
		Phrases:  reading.NewVocabPhrases(phraseCandidates(), NormalizeSurface),
		Analyzer: Analyzer,
	}

	Catalog = reading.LoadCorpus(textFS, "texts", tok, Analyzer)
	Provider = reading.Provider{
		Code:      "de",
		Tokenizer: tok,
		Analyzer:  Analyzer,
		Catalog:   Catalog,
	}
	reading.Register(Provider)
}
