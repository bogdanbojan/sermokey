package reading

// Tokenizer splits running text into tokens. Implementations are
// language-specific: French needs elision handling ("l'homme" is two words,
// "aujourd'hui" is one), Spanish does not; German needs compound awareness
// that French does not.
//
// Every implementation must satisfy the round-trip property — concatenating
// Text[t.Start:t.End] for the returned tokens, in order, must reproduce the
// input byte for byte. AssertRoundTrip checks it, and every language's
// tokenizer test calls it over the whole bundled corpus.
type Tokenizer interface {
	Tokenize(text string) []Token
}

// Analyzer resolves a surface form as it appears in running text to the
// dictionary entries a reader needs. A single surface can have several
// analyses — "est" is a form of both être and est (the compass point), and
// "la" is a determiner, a pronoun, and a note of the scale — so the result
// is a slice, ordered best-first.
type Analyzer interface {
	Analyze(surface string) []Analysis
}

// Source records where an analysis came from. It drives both ranking (a
// curated entry always outranks a scraped one) and the frontend's visual
// treatment: words from the curated curriculum are the ones with a real
// place in the vocabulary deck.
type Source uint8

const (
	// SourceUnknown is the zero value and never appears in a real analysis.
	SourceUnknown Source = iota
	// SourceCurated comes from the hand-built vocabulary CSV. These carry a
	// WordID that already exists in the SRS deck.
	SourceCurated
	// SourceLexicon comes from the generated Wiktionary-derived lexicon.
	SourceLexicon
	// SourceDerived comes from rule-based fallback deinflection — a guess,
	// shown to the reader as such.
	SourceDerived
)

func (s Source) String() string {
	switch s {
	case SourceCurated:
		return "curated"
	case SourceLexicon:
		return "lexicon"
	case SourceDerived:
		return "derived"
	default:
		return "unknown"
	}
}

// Analysis is one candidate reading of a surface form.
type Analysis struct {
	// Lemma is the dictionary headword, e.g. "parler" for "parlons".
	Lemma string
	// POS is the part of speech: noun, verb, adj, adv, other.
	POS string
	// Gender is "m", "f", "n" or "" — only meaningful for nouns.
	Gender string
	// FormNote describes the inflection, e.g. "first-person plural present
	// of parler". Empty when the surface is already the lemma. Showing the
	// deinflection chain is what lets a learner connect "parlons" to the
	// conjugation table they studied, rather than treating it as a new word.
	FormNote string
	// Senses are English glosses, best-first.
	Senses []Sense
	// Function marks a closed-class grammatical word: articles, clitic
	// pronouns, negation particles, basic prepositions. These are acquired
	// as grammar rather than learned as vocabulary — "le → the" is not a
	// flashcard anyone needs — so the reader treats them as known and the
	// coverage figure counts them rather than holding them against the
	// learner. See fr/closed.go for the list and why it can be a list.
	Function bool
	// Proper marks a proper noun. Like function words these are not
	// vocabulary to learn, and coverage studies conventionally count them
	// as known (Nation 2006) because a reader simply reads a name — so
	// leaving them to drag the coverage figure down would misreport how
	// hard a text actually is.
	Proper bool
	// WordID is the canonical SRS card id for this lemma. Saving a word
	// while reading must touch the SAME card as studying it in the
	// vocabulary tab, never a duplicate with its own review state, so this
	// id is always minted server-side and never constructed by the client.
	WordID string
	// Source records provenance; see Source.
	Source Source
}

// Sense is one English gloss of a lemma.
type Sense struct {
	Gloss   string
	Example string
}

// LayeredAnalyzer tries each analyzer in order and returns the first
// non-empty result. This is the extensibility seam for lookup: adding a new
// dictionary means appending an Analyzer, not editing one.
//
// First-match rather than merge-all is deliberate. A reader hovering a word
// wants the best answer, not the union of every source's opinion; and
// because the curated layer comes first, any word in the curriculum is
// glossed with the translation the learner has already seen on its
// flashcard, keeping the two surfaces consistent.
type LayeredAnalyzer []Analyzer

// Analyze implements Analyzer.
func (l LayeredAnalyzer) Analyze(surface string) []Analysis {
	for _, a := range l {
		if out := a.Analyze(surface); len(out) > 0 {
			return out
		}
	}
	return nil
}
