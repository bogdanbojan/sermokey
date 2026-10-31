package reading

import (
	"sort"
	"strings"
	"testing"
	"unicode"
)

// A whitespace-and-letters tokenizer, enough to exercise the sentence
// walk without pulling in a language package.
type simpleTok struct{}

func (simpleTok) Tokenize(text string) []Token {
	var out []Token
	runes := []rune(text)
	offs := make([]int, len(runes)+1)
	b := 0
	for i, r := range runes {
		offs[i] = b
		b += len(string(r))
	}
	offs[len(runes)] = len(text)
	isWord := func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }
	i := 0
	for i < len(runes) {
		j := i
		word := isWord(runes[i])
		for j < len(runes) && isWord(runes[j]) == word {
			j++
		}
		out = append(out, Token{Start: offs[i], End: offs[j], Word: word})
		i = j
	}
	return out
}

func names(t *testing.T, text string) []string {
	t.Helper()
	set := FindProperNouns(text, simpleTok{}.Tokenize(text))
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func TestFindProperNouns(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		// The whole point: a capital in the middle of a sentence is not
		// positional, so it means what it appears to mean.
		{"mid-sentence name", "Chaque samedi matin, Claire se lève.", []string{"claire"}},
		// A capital at the start of a sentence says nothing — every
		// sentence has one.
		{"sentence-initial only", "Maison est grand. Chaque jour passe.", nil},
		// Both positions: one sighting mid-sentence settles it, and the
		// set is per document so the sentence-initial one is covered too.
		{"both positions", "Claire part. Chaque jour, Claire revient.", []string{"claire"}},
		// A word opening quoted speech is capitalised like a sentence.
		{"after an opening quote", "Il dit : « Bonjour madame. »", nil},
		{"several names", "Hier, Claire et Sophie ont vu Marseille.", []string{"claire", "marseille", "sophie"}},
		// Other sentence terminators behave like a full stop.
		{"after a question mark", "Qui vient ? Claire vient.", nil},
		{"empty", "", nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := names(t, tc.in)
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// The set has to be document-wide, not per paragraph: a name introduced in
// one paragraph is still a name where it opens a sentence in the next.
func TestFindProperNounsAcrossSections(t *testing.T) {
	sections := []Section{
		{Paragraphs: []Paragraph{{Text: "Claire ouvre la porte."}}},
		{Paragraphs: []Paragraph{{Text: "Le matin, Claire sort."}}},
	}
	for si := range sections {
		for pi := range sections[si].Paragraphs {
			p := &sections[si].Paragraphs[pi]
			p.Tokens = simpleTok{}.Tokenize(p.Text)
		}
	}
	if got := FindProperNounsIn(sections); !got["claire"] {
		t.Errorf("Claire not recognised across sections: %v", got)
	}
}
