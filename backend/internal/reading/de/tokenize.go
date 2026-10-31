// Package de implements German reading support.
//
// German orthography differs from the Romance languages in ways that
// matter here:
//
//   - Every noun is capitalised, so capitalisation carries no information
//     about proper nouns. The shared analyzer's capitalisation rule has to
//     be suppressed, or every noun in a text would be promoted to a name
//     and dropped from the vocabulary count.
//   - Compounds are written solid ("Geschwindigkeitsbegrenzung"), so an
//     unknown word is frequently a transparent compound of known parts
//     rather than new vocabulary. Splitting them is a research problem of
//     its own and is deliberately out of scope; the lexicon covers the
//     common ones and the rest are simply unknown words, which is honest.
//   - Separable-prefix verbs put their prefix at the end of the clause
//     ("ich rufe dich an" / anrufen). Reuniting them needs clause-level
//     parsing, so "an" is glossed on its own. The generated form index does
//     recognise the joined citation forms.
package de

import (
	"strings"
	"unicode"

	"github.com/bogdanbojan/sermokey/backend/internal/reading"
)

// Tokenizer is the German tokenizer. Stateless and safe for concurrent use.
type Tokenizer struct{}

var _ reading.Tokenizer = Tokenizer{}

func isHyphen(r rune) bool { return r == '-' || r == '\u2010' || r == '\u2011' }

// isWordRune reports whether r can appear inside a German word. ß and the
// umlauted vowels are ordinary letters to unicode.IsLetter; marks cover
// decomposed accents and digits keep "1920" whole.
func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Mn, r)
}

// Tokenize implements reading.Tokenizer.
//
// Hyphenated compounds are kept whole rather than split. German hyphenates
// to avoid an unreadable run of letters ("Kaffee-Ersatz") or to coordinate
// a shared head ("Ein- und Ausgang"), and in the first case the hyphenated
// whole is the word the dictionary lists. Keeping it together also keeps
// the token aligned with what the reader perceives as one item.
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
		out = append(out, reading.Token{Start: offsets[i], End: offsets[j], Word: true})
		i = j
	}
	return out
}

// NormalizeSurface prepares a surface form for dictionary lookup.
//
// Lower-casing is safe despite German capitalising every noun, because the
// lexicon is indexed by lower-cased key and stores the real spelling: a
// hovered "Haus" finds the entry whose lemma is "Haus". Umlauts are not
// folded to ae/oe/ue — they are lexical ("schon" vs "schön", "Mutter" vs
// "Mütter") — and ß is left alone for the same reason.
func NormalizeSurface(s string) string {
	return strings.TrimRight(strings.ToLower(strings.TrimSpace(s)), "-")
}
