package reading

import (
	"strings"

	"github.com/bogdanbojan/sermokey/backend/internal/grammar"
)

// PhraseIndex recognises multi-word lexical items — "pomme de terre",
// "parce que", "se lever".
//
// These have to be matched as units or the reader glosses them wrongly in
// the most misleading way available: "parce" and "que" separately, or
// "terre" as earth in the middle of a potato. They are also exactly the
// formulaic sequences that are learned and stored whole (Wray 2002), so
// splitting them is pedagogically backwards as well as inaccurate.
type PhraseIndex interface {
	// HasPhrase reports whether a normalized string is a known multi-word
	// headword.
	HasPhrase(normalized string) bool
	// MaxWords is the longest phrase in the index, in word tokens. Bounds
	// the matching window.
	MaxWords() int
	// Normalize prepares a surface for lookup, applying whatever
	// case and punctuation folding the language needs.
	Normalize(surface string) string
}

// MergePhrases rewrites a token list so that each known multi-word item is
// a single word token.
//
// Doing it here, as a post-pass over tokenization, is what keeps the rest
// of the system unaware that phrases exist: the merged token is just a
// token with a wider byte range, so rendering, hovering, selection, saving
// and the round-trip invariant all keep working untouched.
//
// Matching is greedy longest-first, the standard approach for dictionary
// segmentation: "salle de bains" must win over "salle".
func MergePhrases(text string, tokens []Token, idx PhraseIndex, an Analyzer) []Token {
	maxWords := idx.MaxWords()
	if maxWords < 2 || len(tokens) == 0 {
		return tokens
	}

	// Positions of the word tokens, so a phrase window can be expressed in
	// words while still covering the punctuation and spaces between them.
	wordAt := make([]int, 0, len(tokens))
	for i, t := range tokens {
		if t.Word {
			wordAt = append(wordAt, i)
		}
	}

	out := make([]Token, 0, len(tokens))
	consumedUpTo := 0 // index into tokens
	for w := 0; w < len(wordAt); w++ {
		start := wordAt[w]
		if start < consumedUpTo {
			continue
		}

		matchedEnd := -1
		for n := min(maxWords, len(wordAt)-w); n >= 2; n-- {
			end := wordAt[w+n-1]
			if idx.HasPhrase(idx.Normalize(text[tokens[start].Start:tokens[end].End])) ||
				idx.HasPhrase(lemmaKey(text, tokens, wordAt[w:w+n], an)) {
				matchedEnd = end
				break
			}
		}

		// Everything before this word that has not been emitted yet.
		for i := consumedUpTo; i < start; i++ {
			out = append(out, tokens[i])
		}
		if matchedEnd < 0 {
			out = append(out, tokens[start])
			consumedUpTo = start + 1
			continue
		}
		out = append(out, Token{Start: tokens[start].Start, End: tokens[matchedEnd].End, Word: true})
		consumedUpTo = matchedEnd + 1
	}
	out = append(out, tokens[consumedUpTo:]...)
	return out
}

// lemmaKey joins the lemmas of a run of word tokens.
//
// Running text inflects its phrases — "les pommes de terre", "elle se
// lève" — so matching raw surfaces alone would find the headword only in
// the one form that happens to be the citation form. Going through the
// analyzer first means the plural and the conjugated forms reach the same
// entry.
func lemmaKey(text string, tokens []Token, words []int, an Analyzer) string {
	parts := make([]string, 0, len(words))
	for _, i := range words {
		surface := tokens[i].Surface(text)
		if as := an.Analyze(surface); len(as) > 0 {
			parts = append(parts, as[0].Lemma)
		} else {
			parts = append(parts, strings.ToLower(surface))
		}
	}
	return strings.Join(parts, " ")
}

// vocabPhrases is a PhraseIndex over a curated vocabulary list.
type vocabPhrases struct {
	set       map[string]bool
	maxWords  int
	normalize func(string) string
}

// NewVocabPhrases builds the multi-word index from a language's curated
// vocabulary.
//
// Deliberately from the curated list only, even though the generated
// lexicon holds tens of thousands of multi-word entries. Those include a
// great many incidental collocations, and merging greedily against all of
// them would swallow ordinary word sequences — a reader meeting "de la" or
// "tout le monde" inside a longer sentence should still be able to hover
// each word. The curated set is a few dozen hand-checked items a learner is
// meant to acquire whole, which is exactly the set worth merging.
func NewVocabPhrases(curated []grammar.VocabWord, normalize func(string) string) PhraseIndex {
	if normalize == nil {
		normalize = func(s string) string { return strings.ToLower(strings.TrimSpace(s)) }
	}
	p := &vocabPhrases{set: map[string]bool{}, normalize: normalize}
	for _, w := range curated {
		n := len(strings.Fields(w.Lemma))
		if n < 2 {
			continue
		}
		p.set[p.Normalize(w.Lemma)] = true
		if n > p.maxWords {
			p.maxWords = n
		}
	}
	return p
}

func (p *vocabPhrases) HasPhrase(normalized string) bool { return p.set[normalized] }
func (p *vocabPhrases) MaxWords() int                    { return p.maxWords }

// Normalize collapses internal whitespace too, so a phrase broken across a
// line in a source file still matches its single-spaced entry.
func (p *vocabPhrases) Normalize(surface string) string {
	return strings.Join(strings.Fields(p.normalize(surface)), " ")
}

// PhraseTokenizer wraps a plain tokenizer with multi-word merging. It is a
// separate type so each language's tokenizer stays a pure description of
// its orthography and the lexical layer stays opt-in.
type PhraseTokenizer struct {
	Base     Tokenizer
	Phrases  PhraseIndex
	Analyzer Analyzer
}

// Tokenize implements Tokenizer.
func (p PhraseTokenizer) Tokenize(text string) []Token {
	return MergePhrases(text, p.Base.Tokenize(text), p.Phrases, p.Analyzer)
}
