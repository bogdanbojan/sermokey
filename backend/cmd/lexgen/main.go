// Command lexgen distills a kaikki.org Wiktionary extract into the two
// compact tables the reader embeds: a lemma lexicon and an inflected-form
// index.
//
// This is a developer tool, not part of the server. It is run by hand when
// the lexicon needs refreshing, and its output is committed, so that
// building and testing the project never requires a 600 MB download:
//
//	curl -LO https://kaikki.org/dictionary/French/kaikki.org-dictionary-French.jsonl
//	go run ./cmd/lexgen -in kaikki.org-dictionary-French.jsonl \
//	    -out internal/reading/fr/data
//
// Output (both gzipped, tab-separated, sorted for a stable diff):
//
//	lexicon.tsv.gz  lemma \t pos \t gender \t gloss; gloss; gloss
//	forms.tsv.gz    surface \t lemma \t pos \t note
//
// The source data is Wiktionary, dual-licensed CC BY-SA and GFDL. The
// attribution that obligation requires is rendered in the reader's word
// panel; see frontend/src/components/WordPanel.vue.
package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

// entry is the subset of a kaikki record this tool needs. encoding/json
// ignores the ~40 other fields (etymology trees, pronunciation, categories,
// translations), which is most of the file's bulk.
type entry struct {
	Word   string   `json:"word"`
	POS    string   `json:"pos"`
	Tags   []string `json:"tags"`
	Senses []sense  `json:"senses"`
	Forms  []form   `json:"forms"`
}

type sense struct {
	Glosses []string `json:"glosses"`
	Tags    []string `json:"tags"`
	FormOf  []formOf `json:"form_of"`
}

type formOf struct {
	Word string `json:"word"`
}

type form struct {
	Form string   `json:"form"`
	Tags []string `json:"tags"`
}

const (
	maxSenses    = 3
	maxGlossRune = 110
)

func main() {
	in := flag.String("in", "", "path to the kaikki.org JSONL extract")
	out := flag.String("out", "", "directory to write lexicon.tsv.gz and forms.tsv.gz into")
	flag.Parse()
	if *in == "" || *out == "" {
		flag.Usage()
		os.Exit(2)
	}
	if err := run(*in, *out); err != nil {
		log.Fatal(err)
	}
}

func run(inPath, outDir string) error {
	f, err := os.Open(inPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Records are one JSON object per line and some are very large
	// (etymology chains run to hundreds of kilobytes), so the scanner needs
	// a buffer far bigger than the default 64 KB.
	sc := bufio.NewScanner(bufio.NewReaderSize(f, 1<<20))
	sc.Buffer(make([]byte, 1<<20), 1<<26)

	lemmas := map[string]*lemmaRow{}
	forms := map[string][]*formRow{}
	var lines, skipped int

	for sc.Scan() {
		lines++
		var e entry
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			skipped++
			continue
		}
		collect(&e, lemmas, forms)
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("reading %s: %w", inPath, err)
	}

	// lemmas is keyed by word+POS to keep homographs apart, so membership
	// tests by bare word need their own set.
	lemmaWords := make(map[string]bool, len(lemmas))
	for key, row := range lemmas {
		if len(row.glosses) > 0 {
			w, _ := splitLemmaKey(key)
			lemmaWords[strings.ToLower(w)] = true
		}
	}

	resolveChains(forms, lemmaWords)

	// A form pointing at a lemma that was dropped is a dangling reference:
	// it would resolve to a headword with no definition.
	//
	// Note what is deliberately NOT removed here: forms whose surface is
	// also a lemma. "est" is both the noun "east" and the third-person
	// singular of être, and dropping the inflection would lose the single
	// most common verb form in the language. Both analyses are kept and
	// the analyzer ranks them; see fr/analyze.go.
	for surface, rows := range forms {
		kept := rows[:0]
		for _, fr := range rows {
			if lemmaWords[strings.ToLower(fr.lemma)] {
				kept = append(kept, fr)
			}
		}
		if len(kept) == 0 {
			delete(forms, surface)
		} else {
			forms[surface] = kept
		}
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	nLex, err := writeTSV(filepath.Join(outDir, "lexicon.tsv.gz"), lexiconLines(lemmas))
	if err != nil {
		return err
	}
	nForm, err := writeTSV(filepath.Join(outDir, "forms.tsv.gz"), formLines(forms))
	if err != nil {
		return err
	}

	log.Printf("read %d records (%d unparseable)", lines, skipped)
	log.Printf("wrote %d lemmas, %d inflected forms", nLex, nForm)
	return nil
}

type lemmaRow struct {
	pos     string
	gender  string
	glosses []string
}

type formRow struct {
	lemma string
	pos   string
	note  string
	// prose records that this came from a form-of sense written by an
	// editor rather than assembled from an inflection table's tags. Prose
	// wins, because it is both more reliable and more readable.
	prose bool
}

// maxFormsPerSurface caps how many analyses one inflected form keeps.
//
// One is not enough. German "Hause" is both the dative of Haus and a form
// of the verb hausen, and whichever Wiktionary happens to list first would
// otherwise become the only reachable reading — so "nach Hause" is glossed
// "to dwell". Keeping a second lets the analyzer rank them, and ranking
// gets it right because Haus is in the curriculum and hausen is not.
//
// Two rather than all: the tail is long and almost entirely obscure, and
// the table is already the largest thing in the binary.
const maxFormsPerSurface = 2

// resolveChains collapses multi-step inflection chains. Wiktionary records
// "allés" as the masculine plural of "allé", and "allé" as the past
// participle of "aller"; only "aller" is a headword with a definition, so a
// reader hovering "allés" must be walked the whole way there.
//
// The hop limit guards against cycles, which do occur in crowd-edited data.
func resolveChains(forms map[string][]*formRow, lemmaWords map[string]bool) {
	const maxHops = 4
	for _, rows := range forms {
		for _, fr := range rows {
			for hop := 0; hop < maxHops; hop++ {
				target := strings.ToLower(fr.lemma)
				if lemmaWords[target] {
					break // reached a real headword
				}
				next, ok := forms[target]
				if !ok || len(next) == 0 || next[0] == fr {
					break
				}
				fr.lemma = next[0].lemma
				// Keep the first (most specific) note; appending every hop
				// produces unreadable chains like "masculine plural of allé
				// — past participle of aller".
				if fr.note == "" {
					fr.note = next[0].note
				}
			}
		}
	}
}

func collect(e *entry, lemmas map[string]*lemmaRow, forms map[string][]*formRow) {
	pos, ok := normalizePOS(e.POS)
	if !ok || e.Word == "" {
		return
	}

	for _, s := range e.Senses {
		if len(s.Glosses) == 0 {
			continue
		}
		// A "form-of" sense says this page documents an inflection, not a
		// headword. It belongs in the form index, pointing at its lemma.
		if base := formOfLemma(&s); base != "" {
			addForm(forms, e.Word, base, pos, inflectionNote(&s), true)
			continue
		}
		if hasAny(s.Tags, "obsolete", "archaic", "dated", "rare", "misspelling") {
			continue
		}
		key := lemmaKey(e.Word, pos)
		row := lemmas[key]
		if row == nil {
			row = &lemmaRow{pos: pos}
			lemmas[key] = row
		}
		if row.gender == "" {
			row.gender = genderOf(pos, e.Tags, s.Tags)
		}
		if len(row.glosses) < maxSenses {
			if g := cleanGloss(s.Glosses); g != "" && !slices.Contains(row.glosses, g) {
				row.glosses = append(row.glosses, g)
			}
		}
	}

	// The inflection table on a headword catches forms that have no page of
	// their own. These carry tags rather than prose, so the note is built
	// from the tags.
	for _, fm := range e.Forms {
		if fm.Form == "" || fm.Form == e.Word || !usableFormTags(fm.Tags) {
			continue
		}
		addForm(forms, fm.Form, e.Word, pos, tagNote(fm.Tags, e.Word), false)
	}
}

func addForm(forms map[string][]*formRow, surface, lemma, pos, note string, prose bool) {
	surface = strings.ToLower(strings.TrimSpace(surface))
	if surface == "" || lemma == "" || surface == lemma {
		return
	}
	rows := forms[surface]
	for _, r := range rows {
		if strings.EqualFold(r.lemma, lemma) {
			// Already known. Upgrade a tag-assembled note to editor prose,
			// which is both more reliable and more readable.
			if prose && !r.prose {
				r.note, r.prose = note, true
			}
			return
		}
	}
	row := &formRow{lemma: lemma, pos: pos, note: note, prose: prose}
	if len(rows) < maxFormsPerSurface {
		forms[surface] = append(rows, row)
		return
	}
	// Full: an editor-written analysis may displace a tag-assembled one,
	// otherwise the first two win so the output stays deterministic.
	if prose {
		for i, r := range rows {
			if !r.prose {
				rows[i] = row
				return
			}
		}
	}
}

// lemmaKey keeps homographs with different parts of speech apart, so that
// "est" the verb form and "est" the compass point stay distinct entries.
func lemmaKey(word, pos string) string { return word + "\x00" + pos }

func splitLemmaKey(k string) (word, pos string) {
	w, p, _ := strings.Cut(k, "\x00")
	return w, p
}

// normalizePOS maps Wiktionary's fine-grained part of speech onto the five
// values the rest of the app already uses (see grammar.VocabWord.POS), and
// reports whether the entry is worth keeping at all. Affixes, individual
// characters and punctuation never appear as tokens a reader would hover.
func normalizePOS(p string) (string, bool) {
	switch p {
	case "noun":
		return "noun", true
	case "name":
		// Kept distinct from "noun". Proper nouns are not vocabulary to
		// learn, and coverage studies count them as known (Nation 2006),
		// so the reader has to be able to tell them apart.
		return "name", true
	case "verb":
		return "verb", true
	case "adj":
		return "adj", true
	case "adv":
		return "adv", true
	case "prefix", "suffix", "infix", "interfix", "character", "punct",
		"romanization", "syllable", "symbol":
		return "", false
	default:
		// det, pron, prep, conj, intj, num, article, phrase, proverb, ...
		// These are exactly the function words a learner most needs, so
		// they are kept — just not given their own POS bucket.
		return "other", true
	}
}

func formOfLemma(s *sense) string {
	if len(s.FormOf) > 0 && s.FormOf[0].Word != "" {
		return s.FormOf[0].Word
	}
	return ""
}

// inflectionNote turns a form-of sense into one readable line. kaikki splits
// these across several glosses ("inflection of parler:", "first-person
// plural present indicative"); joined and tidied they read as a sentence.
func inflectionNote(s *sense) string {
	parts := make([]string, 0, len(s.Glosses))
	for _, g := range s.Glosses {
		g = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(g), ":"))
		if g != "" {
			parts = append(parts, g)
		}
	}
	return truncate(strings.Join(parts, " — "), maxGlossRune)
}

// usableFormTags filters out the scaffolding rows kaikki emits inside
// inflection tables ("no-table-tags", "fr-conj-auto") and the variants that
// are not inflections of the headword at all.
func usableFormTags(tags []string) bool {
	if len(tags) == 0 {
		return false
	}
	return !hasAny(tags, "table-tags", "inflection-template", "alternative",
		"obsolete", "archaic", "dated", "misspelling", "error-unknown-tag")
}

func tagNote(tags []string, lemma string) string {
	keep := make([]string, 0, len(tags))
	for _, t := range tags {
		if t == "" || t == "form-of" {
			continue
		}
		keep = append(keep, strings.ReplaceAll(t, "-", " "))
	}
	if len(keep) == 0 {
		return ""
	}
	return truncate(strings.Join(keep, " ")+" of "+lemma, maxGlossRune)
}

func genderOf(pos string, entryTags, senseTags []string) string {
	if pos != "noun" {
		return ""
	}
	for _, tags := range [][]string{entryTags, senseTags} {
		switch {
		case hasAny(tags, "feminine"):
			return "f"
		case hasAny(tags, "masculine"):
			return "m"
		}
	}
	return ""
}

// cleanGloss joins a sense's gloss lines and strips the characters that
// would corrupt a tab-separated file.
func cleanGloss(glosses []string) string {
	g := strings.Join(glosses, " ")
	g = strings.NewReplacer("\t", " ", "\n", " ", "\r", " ", ";", ",").Replace(g)
	return truncate(strings.Join(strings.Fields(g), " "), maxGlossRune)
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return strings.TrimSpace(string(r[:n])) + "…"
}

func hasAny(tags []string, want ...string) bool {
	for _, t := range tags {
		if slices.Contains(want, t) {
			return true
		}
	}
	return false
}

func lexiconLines(lemmas map[string]*lemmaRow) []string {
	out := make([]string, 0, len(lemmas))
	for key, row := range lemmas {
		if len(row.glosses) == 0 {
			continue
		}
		word, _ := splitLemmaKey(key)
		// A lower-cased lookup key leads, with the real spelling kept in
		// the next column. The reader always searches in lower case, but
		// the display form has to survive — "Sophie" and "Marseille" are
		// headwords whose capitals are part of the word.
		out = append(out, strings.Join([]string{
			strings.ToLower(word), word, row.pos, row.gender, strings.Join(row.glosses, "; "),
		}, "\t"))
	}
	sort.Strings(out)
	return out
}

func formLines(forms map[string][]*formRow) []string {
	out := make([]string, 0, len(forms))
	for surface, rows := range forms {
		for _, row := range rows {
			out = append(out, strings.Join([]string{surface, row.lemma, row.pos, row.note}, "\t"))
		}
	}
	sort.Strings(out)
	return out
}

func writeTSV(path string, lines []string) (int, error) {
	f, err := os.Create(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	zw, err := gzip.NewWriterLevel(f, gzip.BestCompression)
	if err != nil {
		return 0, err
	}
	bw := bufio.NewWriterSize(zw, 1<<20)
	for _, l := range lines {
		if _, err := io.WriteString(bw, l+"\n"); err != nil {
			return 0, err
		}
	}
	if err := bw.Flush(); err != nil {
		return 0, err
	}
	if err := zw.Close(); err != nil {
		return 0, err
	}
	return len(lines), nil
}
