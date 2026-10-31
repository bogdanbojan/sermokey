// Package es implements Spanish reading support.
//
// Spanish orthography is kinder to a tokenizer than French: there is no
// elision, so no apostrophes to split, and the hyphen is rare outside
// compound adjectives ("teórico-práctico"). Two features do need handling,
// and they are the only reasons this is not the generic tokenizer:
//
//   - Inverted punctuation. "¿" and "¡" open a question or exclamation and
//     are punctuation, not word characters. Getting this wrong attaches
//     them to the following word and every question in a text becomes a
//     token the dictionary has never heard of.
//   - Enclitic pronouns. "dámelo" is dar + me + lo written as one word.
//     They are deliberately NOT split: unlike French elision the pieces
//     are not separable in the orthography, Wiktionary lists the combined
//     forms, and splitting on a guess would mangle ordinary words that
//     merely end in the same letters ("pueblo" is not "puebl" + "lo").
package es

import (
	"strings"
	"unicode"

	"github.com/bogdanbojan/sermokey/backend/internal/reading"
)

// Tokenizer is the Spanish tokenizer. Stateless, so the zero value is
// ready to use and it is safe for concurrent use.
type Tokenizer struct{}

var _ reading.Tokenizer = Tokenizer{}

// hyphenExceptions are words whose hyphen is part of the spelling. Spanish
// uses the hyphen far less than French does, so the list is short; anything
// not here is split, which is right for coordinated compounds like
// "franco-alemán" where each half is a word a learner may want to look up.
var hyphenExceptions = map[string]bool{
	"ciencia-ficción": true,
	"veinte-treinta":  true,
}

func isHyphen(r rune) bool { return r == '-' || r == '\u2010' || r == '\u2011' }

// isWordRune reports whether r can appear inside a Spanish word. Marks are
// included so decomposed accents (e + U+0301 rather than é) do not split a
// word; digits so "1920" is one token.
func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Mn, r)
}

// Tokenize implements reading.Tokenizer.
func (Tokenizer) Tokenize(text string) []reading.Token {
	var out []reading.Token
	runes := []rune(text)
	offsets := make([]int, len(runes)+1)
	{
		b := 0
		for i, r := range runes {
			offsets[i] = b
			b += len(string(r))
		}
		offsets[len(runes)] = len(text)
	}

	i := 0
	for i < len(runes) {
		if !isWordRune(runes[i]) {
			j := i
			for j < len(runes) && !isWordRune(runes[j]) {
				j++
			}
			out = append(out, reading.Token{Start: offsets[i], End: offsets[j], Word: false})
			i = j
			continue
		}

		// A maximal word run, absorbing a hyphen only when a word
		// character follows, so a dash used as punctuation stays punctuation.
		j := i
		for j < len(runes) {
			if isWordRune(runes[j]) {
				j++
				continue
			}
			if isHyphen(runes[j]) && j+1 < len(runes) && isWordRune(runes[j+1]) {
				j++
				continue
			}
			break
		}
		out = append(out, splitRun(runes, offsets, i, j)...)
		i = j
	}
	return out
}

// splitRun decides how many tokens a word run is. Every byte is emitted
// exactly once regardless, which is what preserves the round-trip invariant.
func splitRun(runes []rune, offsets []int, start, end int) []reading.Token {
	run := strings.ToLower(string(runes[start:end]))
	if hyphenExceptions[run] || !strings.ContainsFunc(run, isHyphen) {
		return []reading.Token{{Start: offsets[start], End: offsets[end], Word: true}}
	}

	// Split at each hyphen, emitting the hyphen itself as punctuation: it
	// joins two words and belongs to neither.
	var out []reading.Token
	segStart := start
	for k := start; k < end; k++ {
		if !isHyphen(runes[k]) {
			continue
		}
		if k > segStart {
			out = append(out, reading.Token{Start: offsets[segStart], End: offsets[k], Word: true})
		}
		out = append(out, reading.Token{Start: offsets[k], End: offsets[k+1], Word: false})
		segStart = k + 1
	}
	if segStart < end {
		out = append(out, reading.Token{Start: offsets[segStart], End: offsets[end], Word: true})
	}
	return out
}

// NormalizeSurface prepares a surface form for dictionary lookup.
//
// Case is folded because a sentence-initial capital is an artefact of
// position. Accents are deliberately NOT stripped: in Spanish they are
// lexical and often the only thing distinguishing two words — "el" (the)
// from "él" (he), "si" (if) from "sí" (yes), "esta" from "está".
func NormalizeSurface(s string) string {
	return strings.TrimRight(strings.ToLower(strings.TrimSpace(s)), "-")
}
