package httpapi

// The reading DTOs. Two shapes here are worth explaining, because both are
// chosen for how the reader actually behaves rather than for what reads
// most naturally as JSON.
//
// First, a paragraph ships its text verbatim plus an array of integer
// spans, not a pre-rendered list of token objects. The frontend renders by
// slicing the original string, which means whitespace, punctuation and
// typography survive untouched, and — more importantly — the DOM ends up as
// one span per word over the real text, so the browser's native selection
// and Range APIs map directly onto token indices. That is what makes
// drag-selecting a phrase work at all.
//
// Second, the lexicon for a whole section ships with that section rather
// than being fetched per word. Hovering has to feel instantaneous, and a
// network round trip per word cannot. Distinct forms are pooled and
// referenced by index, so a 600-word section carries a few hundred entries
// rather than 600 copies.

// ReadingTextDTO is one text's library card. It carries no content.
type ReadingTextDTO struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Author       string `json:"author,omitempty"`
	Level        string `json:"level"`
	Blurb        string `json:"blurb"`
	Source       string `json:"source,omitempty"`
	License      string `json:"license"`
	SectionCount int    `json:"sectionCount"`
	TokenCount   int    `json:"tokenCount"`
	// FunctionTokenCount is how many tokens are closed-class grammatical
	// words, which WordFrequency deliberately omits. The client adds it to
	// both sides of the coverage ratio; see the field's doc comment in
	// reading.DocumentMeta.
	FunctionTokenCount int `json:"functionTokenCount"`
	// ProperTokenCount is how many tokens are proper nouns, likewise
	// omitted from WordFrequency and likewise counted as known.
	ProperTokenCount int `json:"properTokenCount"`

	// WordFrequency maps SRS card id -> number of word tokens resolving to
	// it. The client intersects this with its own deck to show what share
	// of the text it already knows, which is a far better difficulty
	// signal for a given reader than the static CEFR Level above.
	WordFrequency map[string]int `json:"wordFrequency"`
}

// ReadingSectionDTO is one chapter-sized unit, the quantum of delivery.
// Paginating here is what keeps an imported 300-page book the same size on
// the wire as a 600-word article.
type ReadingSectionDTO struct {
	TextID       string `json:"textId"`
	Title        string `json:"title"`
	SectionTitle string `json:"sectionTitle"`
	Index        int    `json:"index"`
	SectionCount int    `json:"sectionCount"`
	License      string `json:"license"`
	Source       string `json:"source,omitempty"`

	Paragraphs []ReadingParagraphDTO `json:"paragraphs"`

	// Lexicon is the pool of distinct analyses referenced by the tokens in
	// this section. Token spans index into it.
	Lexicon []ReadingEntryDTO `json:"lexicon"`
}

// ReadingParagraphDTO is one block of running text.
type ReadingParagraphDTO struct {
	Text string `json:"text"`
	// Spans is a flat triple-per-token array: start byte, end byte, and
	// either an index into the section Lexicon or -1 for a span with no
	// dictionary entry (punctuation, whitespace, proper nouns). Flat
	// integers rather than objects because this is the single largest part
	// of the payload and the frontend only ever walks it in order.
	Spans []int `json:"spans"`
}

// SpanStride is the number of integers per token in ReadingParagraphDTO.Spans.
const SpanStride = 3

// NoEntry is the Spans lexicon index meaning "nothing to look up here".
const NoEntry = -1

// ReadingEntryDTO is one dictionary entry as the reader shows it.
type ReadingEntryDTO struct {
	// Surface is the form as it appears in the text, lowercased. Present
	// so the client can key its lookup-count tracking without re-deriving
	// it from byte offsets.
	Surface string `json:"surface"`
	Lemma   string `json:"lemma"`
	POS     string `json:"pos"`
	Gender  string `json:"gender,omitempty"`
	// FormNote explains the inflection, e.g. "first-person plural present
	// of parler". Empty when the surface is already the lemma. Showing it
	// is what connects a form in the text back to the conjugation table
	// the learner studied, instead of presenting it as a new word.
	FormNote string            `json:"formNote,omitempty"`
	Senses   []ReadingSenseDTO `json:"senses"`
	// WordID is the SRS card id to use when saving this word. Always minted
	// server-side: a word saved while reading must touch the same card the
	// vocabulary tab schedules, never a duplicate with its own history.
	WordID string `json:"wordId"`
	// Curated marks a word as part of the taught curriculum rather than
	// merely present in the dictionary. Drives both ranking and how the
	// word is presented.
	Curated bool `json:"curated"`
	// Function marks a closed-class grammatical word, which the reader
	// renders as known and leaves undecorated.
	Function bool `json:"function,omitempty"`
	// Proper marks a proper noun, likewise left undecorated.
	Proper bool `json:"proper,omitempty"`
	// Uncertain marks an analysis that rule-based fallback guessed, so the
	// UI can present it as a guess rather than as a dictionary fact.
	Uncertain bool `json:"uncertain,omitempty"`
	// Alternatives are other readings of the same surface, best-first,
	// e.g. "est" as the noun "east" alongside the verb form of être.
	Alternatives []ReadingAltDTO `json:"alternatives,omitempty"`
}

// ReadingSenseDTO is one English gloss.
type ReadingSenseDTO struct {
	Gloss string `json:"gloss"`
}

// ReadingAltDTO is a secondary reading of a surface form, shown in the
// pinned panel so a reader can correct a mis-ranked homograph.
type ReadingAltDTO struct {
	Lemma    string            `json:"lemma"`
	POS      string            `json:"pos"`
	Gender   string            `json:"gender,omitempty"`
	FormNote string            `json:"formNote,omitempty"`
	Senses   []ReadingSenseDTO `json:"senses"`
	WordID   string            `json:"wordId"`
	Curated  bool              `json:"curated"`
}
