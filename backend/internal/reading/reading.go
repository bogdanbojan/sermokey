// Package reading holds the language-agnostic core of the reader: the
// document model, the tokenizer/analyzer interfaces, and the lexicon shapes
// that the HTTP layer serves to the frontend.
//
// The package is deliberately split along four seams, because the end goal
// is to read an imported EPUB with the same UI that reads the three bundled
// texts:
//
//	Catalog    where documents come from       (bundled files -> .epub)
//	Tokenizer  running text -> word spans      (fr -> es/de)
//	Analyzer   surface form -> lemma + senses  (curated -> wiktionary -> rules)
//	SRS        review scheduling               (already exists, in the frontend)
//
// Only the first two are language-aware; only the first is source-aware.
// Adding a language means adding a Tokenizer and an Analyzer. Adding EPUB
// support means adding a Catalog. Neither requires touching this file, the
// HTTP handlers, or the frontend.
//
// Note that reading is NOT part of the grammar package's languageProvider
// abstraction. Reading availability is deliberately a subset of grammar
// availability (French ships first), and forcing es/de to stub out methods
// they cannot implement would be dishonest about that. See registry.go.
package reading

// Token is one span of a paragraph. Start and End are byte offsets into the
// paragraph's Text, so Text[Start:End] reproduces the token exactly and
// concatenating every token in order reproduces the paragraph exactly — an
// invariant the tokenizer tests assert directly, because silently dropping
// or mangling characters in a reading surface is the worst class of bug here.
//
// Non-word tokens (whitespace, punctuation) are emitted too rather than
// skipped, both to keep that round-trip property checkable and so the
// frontend can render a paragraph by walking tokens alone.
type Token struct {
	Start int
	End   int
	Word  bool
}

// Surface returns the token's text given the paragraph it came from.
func (t Token) Surface(paragraph string) string {
	return paragraph[t.Start:t.End]
}

// Paragraph is one block of running text plus its tokenization.
type Paragraph struct {
	Text   string
	Tokens []Token
}

// Section is a chapter-sized unit of a document. Sections are the unit of
// delivery: the API serves one section at a time, which keeps the payload
// bounded for a 300-page EPUB exactly as it does for a 600-word article.
type Section struct {
	Title      string
	Paragraphs []Paragraph
}

// Document is a complete readable text.
type Document struct {
	Meta     DocumentMeta
	Sections []Section
	// ProperNouns holds the lower-cased surfaces this document uses as
	// names, determined by mid-sentence capitalisation. Computed once per
	// document because the evidence is document-wide: a name seen
	// mid-sentence in one line is still a name when it opens the next.
	ProperNouns map[string]bool
}

// Level is a coarse, author-assigned difficulty hint. It is intentionally
// secondary to the per-reader coverage figure the frontend computes from
// LemmaFrequency: CEFR bands describe a text in the abstract, whereas the
// share of tokens a specific reader already knows is what actually predicts
// whether they will understand it (Hu & Nation 2000; Laufer 1989).
type Level string

const (
	LevelA2 Level = "A2"
	LevelB1 Level = "B1"
	LevelB2 Level = "B2"
)

// DocumentMeta is everything needed to render a text's card in the library
// without loading the text itself.
type DocumentMeta struct {
	ID       string
	Title    string
	Author   string
	Level    Level
	Blurb    string
	Source   string // provenance, e.g. "Original" or a Wikisource URL
	License  string // e.g. "CC BY-SA 4.0", "Public domain"
	Sections []SectionMeta

	// TokenCount is the number of word tokens across the whole document.
	TokenCount int

	// FunctionTokenCount is how many of those are closed-class grammatical
	// words. They are kept out of WordFrequency — they are not vocabulary
	// to learn — but the coverage figure still has to account for them,
	// because the 95%/98% comprehension thresholds are defined over all
	// running words, articles included. Reporting the total separately
	// lets the client add them to both sides of the ratio without
	// carrying an entry per article.
	FunctionTokenCount int

	// ProperTokenCount is how many are proper nouns. Also excluded from
	// WordFrequency and also counted as known; kept separate from
	// FunctionTokenCount because they are different things and conflating
	// them would make the numbers impossible to reason about later.
	ProperTokenCount int

	// WordFrequency maps SRS card id -> number of word tokens resolving to
	// it. The frontend intersects this with its deck and known-word set to
	// compute token-weighted known coverage. Token-weighted, not
	// type-weighted, is what the 95%/98% comprehension thresholds in the
	// literature refer to.
	//
	// Keyed by card id rather than by lemma so the client can do that
	// intersection directly. A curated word's id is "{lemma}|{category}",
	// which cannot be reconstructed from the lemma alone, so a
	// lemma-keyed table would force every consumer to guess at the
	// category — and guess wrong for exactly the curated words that
	// matter most.
	WordFrequency map[string]int
}

// SectionMeta describes one section without its content.
type SectionMeta struct {
	Title      string
	TokenCount int
}

// Catalog is the source of readable documents for one language. The bundled
// three-text corpus and a future EPUB importer are both just Catalogs.
type Catalog interface {
	// List returns metadata for every document, in display order.
	List() []DocumentMeta
	// Get returns one document by ID.
	Get(id string) (Document, bool)
}

// Provider bundles everything needed to serve reading for one language.
type Provider struct {
	Code      string
	Tokenizer Tokenizer
	Analyzer  Analyzer
	// Catalog is what handlers read from.
	Catalog Catalog
}
