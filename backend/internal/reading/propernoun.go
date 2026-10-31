package reading

import (
	"strings"
	"unicode"
)

// ProperNounAnalyzer is an Analyzer that can be told a surface was seen
// capitalised in the middle of a sentence.
//
// That position is much stronger evidence than capitalisation alone. Every
// sentence-initial word is capitalised, so Analyze has to stay cautious and
// refuse to promote a word that also has an ordinary reading — otherwise
// French "Maison" at the start of a sentence becomes a surname. Mid-sentence
// there is no such ambiguity, and the caution becomes the problem instead:
// it is exactly why Spanish "Clara" was read as the adjective claro and
// "Carmen" as a common noun, in texts where both are people.
type ProperNounAnalyzer interface {
	AnalyzeProperNoun(surface string) []Analysis
}

// sentenceEnders terminate a sentence, so the next word is capitalised for
// positional reasons and carries no information.
const sentenceEnders = ".!?…:;"

// quoteOpeners begin quoted speech. The first word inside is capitalised
// like a sentence opening, so it is treated as one.
const quoteOpeners = `«"“„'‘([`

// FindProperNouns returns the lower-cased surfaces that appear capitalised
// mid-sentence anywhere in the given sections.
//
// The decision is made per document rather than per occurrence. A name
// usually appears somewhere other than the start of a sentence — "Chaque
// samedi matin, Claire se lève" — and once that is seen, every other
// occurrence of the same surface is a name too, including the ones that do
// open a sentence. Deciding occurrence by occurrence would read "Claire"
// correctly in one line and as an adjective in the next.
//
// This is also what makes imported books work. A novel is full of character
// names that collide with ordinary words, and they are precisely the words
// a reader should not be asked to add to a vocabulary deck.
func FindProperNouns(text string, tokens []Token) map[string]bool {
	out := map[string]bool{}
	addFrom(out, text, tokens)
	return out
}

// addFrom accumulates into an existing set, for walking a whole document.
func addFrom(set map[string]bool, text string, tokens []Token) {
	// A paragraph opens a sentence.
	atSentenceStart := true
	for _, t := range tokens {
		surface := t.Surface(text)
		if !t.Word {
			// Punctuation and spaces decide where the next sentence begins.
			if strings.ContainsAny(surface, sentenceEnders) {
				atSentenceStart = true
			} else if strings.ContainsAny(surface, quoteOpeners) {
				atSentenceStart = true
			}
			continue
		}
		if !atSentenceStart && startsUpper(surface) {
			set[strings.ToLower(surface)] = true
		}
		atSentenceStart = false
	}
}

// FindProperNounsIn walks every paragraph of every section.
func FindProperNounsIn(sections []Section) map[string]bool {
	set := map[string]bool{}
	for _, s := range sections {
		for _, p := range s.Paragraphs {
			addFrom(set, p.Text, p.Tokens)
		}
	}
	return set
}

func startsUpper(s string) bool {
	for _, r := range s {
		return unicode.IsUpper(r)
	}
	return false
}
