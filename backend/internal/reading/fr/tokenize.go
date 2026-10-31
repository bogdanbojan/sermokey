// Package fr implements French reading support: a tokenizer that
// understands French orthography, a layered analyzer over the curated
// vocabulary plus a Wiktionary-derived lexicon, and the bundled corpus.
//
// French tokenization is not "split on non-letters". Three features of the
// orthography break that naive rule, and each one is handled explicitly in
// tokenize.go:
//
//   - Elision. "l'homme" is the determiner le plus the noun homme and must
//     split, so that hovering homme finds a noun rather than nothing. But
//     "aujourd'hui" is a single lexeme that merely looks elided, and
//     splitting it produces two nonsense fragments. The distinction is
//     lexical, not structural, so it needs a list.
//   - Hyphens. "est-ce", "parle-t-il" and "vas-y" are several words joined
//     for phonological reasons and should split; "peut-être",
//     "après-midi" and "c'est-à-dire" are single lexemes and should not.
//     Again lexical.
//   - Apostrophe characters. Real French text uses U+2019 (the typographic
//     apostrophe), keyboards produce U+0027, and both appear in the same
//     document. They must behave identically.
package fr

import (
	"strings"
	"unicode"

	"github.com/bogdanbojan/sermokey/backend/internal/reading"
)

// Tokenizer is the French tokenizer. It holds no state, so the zero value
// is ready to use and it is safe for concurrent use.
type Tokenizer struct{}

var _ reading.Tokenizer = Tokenizer{}

// elisionExceptions are words that contain an apostrophe but are a single
// lexeme: the apostrophe is frozen into the spelling rather than marking a
// live elision. Splitting them yields fragments that are in no dictionary.
//
// Matched case-insensitively against the whole apostrophe-containing run.
var elisionExceptions = map[string]bool{
	"aujourd'hui":  true,
	"quelqu'un":    true,
	"quelqu'une":   true,
	"prud'homme":   true,
	"prud'hommes":  true,
	"presqu'île":   true,
	"presqu'îles":  true,
	"entr'acte":    true,
	"entr'actes":   true,
	"grand'mère":   true,
	"grand'mères":  true,
	"c'est-à-dire": true,
	"jusqu'alors":  true,
	"d'abord":      true,
	"d'accord":     true,
	"d'ailleurs":   true,
	"d'après":      true,
}

// hyphenExceptions are words written with a hyphen that are nonetheless a
// single lexeme. Everything else containing a hyphen is split, which is the
// right default because the productive uses of the French hyphen —
// inversion ("parle-t-il"), clitics ("vas-y", "donne-moi"), and
// "est-ce que" — all join words that each deserve their own lookup.
var hyphenExceptions = map[string]bool{
	"peut-être":      true,
	"après-midi":     true,
	"après-midis":    true,
	"c'est-à-dire":   true,
	"au-dessus":      true,
	"au-dessous":     true,
	"au-delà":        true,
	"là-bas":         true,
	"là-haut":        true,
	"là-dedans":      true,
	"vis-à-vis":      true,
	"rendez-vous":    true,
	"week-end":       true,
	"week-ends":      true,
	"porte-monnaie":  true,
	"chef-d'œuvre":   true,
	"aujourd'hui":    true,
	"arc-en-ciel":    true,
	"petit-déjeuner": true,
	"grand-père":     true,
	"grand-mère":     true,
	"grands-parents": true,
	"beaux-parents":  true,
	"belle-mère":     true,
	"beau-père":      true,
	"tout-à-fait":    true,
	"quelque-chose":  true,
	"sous-sol":       true,
	"avant-hier":     true,
	"self-service":   true,
}

// IsApostrophe reports whether r is any of the characters French text uses
// as an apostrophe. U+2019 is the typographically correct one and what
// published text contains; U+0027 is what keyboards produce; U+02BC turns
// up in text converted from other systems. Treating them differently would
// make the same word look up successfully in one document and fail in
// another.
func IsApostrophe(r rune) bool {
	return r == '\'' || r == '\u2019' || r == '\u02BC'
}

// isHyphen reports whether r joins words. U+2010 and U+2011 appear in
// typeset text; the en and em dashes do not join words and are punctuation.
func isHyphen(r rune) bool {
	return r == '-' || r == '\u2010' || r == '\u2011'
}

// isWordRune reports whether r can appear inside a French word. Marks are
// included so that decomposed accents (e + U+0301 rather than é) do not
// split a word in half; digits are included so "1920" is one token.
func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Mn, r)
}

// Tokenize implements reading.Tokenizer.
//
// It works in two passes: scan maximal runs of word characters (plus any
// internal apostrophes and hyphens), then decide whether each run is one
// token or several. Keeping the splitting decision separate from the
// scanning means the round-trip invariant is a property of the scan alone
// and cannot be broken by a splitting rule.
func (Tokenizer) Tokenize(text string) []reading.Token {
	var out []reading.Token
	runes := []rune(text)
	// offsets[i] is the byte offset of runes[i]; offsets[len] is len(text).
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
			// Consume a maximal run of non-word characters as one token.
			j := i
			for j < len(runes) && !isWordRune(runes[j]) {
				j++
			}
			out = append(out, reading.Token{Start: offsets[i], End: offsets[j], Word: false})
			i = j
			continue
		}

		// Consume a maximal word run, absorbing an apostrophe or hyphen
		// only when a word character follows it — so the apostrophe in
		// "'bonjour'" or the dash in "mot -- mot" stays punctuation.
		j := i
		for j < len(runes) {
			if isWordRune(runes[j]) {
				j++
				continue
			}
			if (IsApostrophe(runes[j]) || isHyphen(runes[j])) &&
				j+1 < len(runes) && isWordRune(runes[j+1]) {
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

// splitRun decides how many tokens the word run runes[start:end] is, and
// emits them. Every byte of the run is emitted exactly once regardless of
// the decision, which is what preserves the round-trip invariant.
func splitRun(runes []rune, offsets []int, start, end int) []reading.Token {
	run := string(runes[start:end])
	folded := foldApostrophes(strings.ToLower(run))

	// A single lexeme that merely looks compound: emit it whole.
	if elisionExceptions[folded] || hyphenExceptions[folded] {
		return []reading.Token{{Start: offsets[start], End: offsets[end], Word: true}}
	}

	// Otherwise split at every apostrophe and hyphen. The two separators
	// are treated differently, because they are different things:
	//
	//   - An apostrophe is part of the elided word's spelling. The clitic
	//     is "l'", not "l" — and keeping it is what lets the analyzer tell
	//     the determiner "l'" from the letter "l".
	//   - A hyphen belongs to neither side; it is a joiner. "est-ce" is
	//     "est" + "ce", and emitting "est-" would match nothing. It is
	//     emitted as its own non-word token so the bytes still round-trip.
	//
	// Segments are collected in rune indices first, because the euphonic-t
	// rule below needs to look at neighbours.
	type seg struct {
		from, to int // rune indices
		word     bool
	}
	var segs []seg
	segStart := start
	for k := start; k < end; k++ {
		switch {
		case IsApostrophe(runes[k]):
			segs = append(segs, seg{segStart, k + 1, true})
			segStart = k + 1
		case isHyphen(runes[k]):
			if k > segStart {
				segs = append(segs, seg{segStart, k, true})
			}
			segs = append(segs, seg{k, k + 1, false})
			segStart = k + 1
		}
	}
	if segStart < end {
		segs = append(segs, seg{segStart, end, true})
	}
	if len(segs) == 0 {
		// No separator found: the whole run is one token.
		return []reading.Token{{Start: offsets[start], End: offsets[end], Word: true}}
	}

	// The "t" in inverted forms like "demande-t-il" is a liaison consonant
	// inserted purely so the phrase is pronounceable. It is not a morpheme
	// and has no dictionary entry, so leaving it as a word token would
	// produce a clickable span resolving to a letter name or an acronym —
	// noise on exactly the kind of common interrogative a learner probes.
	isHyphenSeg := func(s seg) bool { return !s.word && isHyphen(runes[s.from]) }
	for i := 1; i < len(segs)-1; i++ {
		s := segs[i]
		if !s.word || s.to-s.from != 1 {
			continue
		}
		if r := runes[s.from]; r != 't' && r != 'T' {
			continue
		}
		if isHyphenSeg(segs[i-1]) && isHyphenSeg(segs[i+1]) {
			segs[i].word = false
		}
	}

	out := make([]reading.Token, 0, len(segs))
	for _, s := range segs {
		out = append(out, reading.Token{Start: offsets[s.from], End: offsets[s.to], Word: s.word})
	}
	return out
}

// foldApostrophes rewrites every apostrophe variant to U+0027 so the
// exception tables need only one spelling of each entry.
func foldApostrophes(s string) string {
	if !strings.ContainsAny(s, "\u2019\u02BC") {
		return s
	}
	return strings.Map(func(r rune) rune {
		if IsApostrophe(r) {
			return '\''
		}
		return r
	}, s)
}

// NormalizeSurface prepares a surface form for dictionary lookup: lowercase,
// apostrophes folded to one variant, and any trailing separator dropped.
//
// Case is folded because sentence-initial capitals are an artefact of
// position, not of the word. Accents are deliberately NOT stripped: in
// French they are lexical ("ou" vs "où", "a" vs "à", "sur" vs "sûr"), so
// folding them would merge genuinely different words.
func NormalizeSurface(s string) string {
	s = foldApostrophes(strings.ToLower(strings.TrimSpace(s)))
	return strings.TrimRight(s, "'-")
}
