package reading

import (
	"bytes"
	"compress/gzip"
	"io"
	"sort"
	"strings"
)

// Lexicon is a language's generated dictionary: headwords with glosses,
// plus an index from inflected forms back to them.
//
// The tables are produced by cmd/lexgen from a kaikki.org Wiktionary
// extract and embedded per language. Parsing and indexing them is identical
// for every language — the differences between French, Spanish and German
// live in their tokenizers and closed-class sets, not in how a TSV is read
// — so this lives here rather than being written three times.
//
// Lookup is binary search over the decompressed table rather than a Go map.
// That is not premature cleverness: German alone has 1.3 million inflected
// forms, and a map[string][]FormEntry over them costs roughly 350 MB in
// headers and bucket overhead before any text is stored. With all three
// languages resident that was over a gigabyte. The generator already sorts
// its output, so searching it directly costs one extra offset table —
// about 4 bytes per row — and turns the whole structure back into roughly
// the size of the text itself.
type Lexicon struct {
	entries table // headwords, keyed by lower-cased lemma
	forms   table // inflected forms, keyed by lower-cased surface
}

// LexEntry is one headword: `lemma \t pos \t gender \t glosses`.
type LexEntry struct {
	Lemma  string
	POS    string
	Gender string
	// Glosses is "; "-separated and split on demand, because most entries
	// in a 100k-word lexicon are never read in a given session.
	Glosses string
}

// FormEntry is one inflected form: `surface \t lemma \t pos \t note`.
type FormEntry struct {
	Lemma string
	POS   string
	Note  string
}

// LoadLexicon inflates and indexes a pair of generated tables.
//
// They are stored gzipped because they are highly repetitive text — tens
// of megabytes that compress to a few — which is a worthwhile trade for
// one decompression at process start.
//
// Panics on malformed input: the data is committed alongside the code, so
// a failure here is a build problem rather than a runtime condition, the
// same stance the vocabulary CSV parser takes.
func LoadLexicon(lexiconGz, formsGz []byte) *Lexicon {
	return &Lexicon{
		entries: newTable(mustGunzip(lexiconGz)),
		forms:   newTable(mustGunzip(formsGz)),
	}
}

// Entries returns the readings of a lower-cased headword. Column 0 is the
// lower-cased lookup key; column 1 is the real spelling.
func (l *Lexicon) Entries(lemma string) []LexEntry {
	rows := l.entries.lookup(lemma, lexiconFields)
	out := make([]LexEntry, 0, len(rows))
	for _, f := range rows {
		out = append(out, LexEntry{Lemma: f[1], POS: f[2], Gender: f[3], Glosses: f[4]})
	}
	return out
}

// Forms returns every analysis of a lower-cased inflected form. There can
// be more than one: German "Hause" is both the dative of Haus and a form
// of the verb hausen, and keeping only the first would make "nach Hause"
// mean "to dwell".
func (l *Lexicon) Forms(surface string) []FormEntry {
	rows := l.forms.lookup(surface, formFields)
	out := make([]FormEntry, 0, len(rows))
	for _, f := range rows {
		out = append(out, FormEntry{Lemma: f[1], POS: f[2], Note: f[3]})
	}
	return out
}

// Size reports the row counts, for startup logging and tests.
func (l *Lexicon) Size() (lemmas, forms int) {
	return len(l.entries.offsets), len(l.forms.offsets)
}

// table is a sorted, tab-separated blob plus an index of line starts.
type table struct {
	data    string
	offsets []uint32
}

// Column counts of the two generated tables.
const (
	lexiconFields = 5 // key, lemma, pos, gender, glosses
	formFields    = 4 // surface, lemma, pos, note
)

func newTable(data string) table {
	t := table{data: data}
	// One pass to count, one to fill: a single exactly-sized allocation
	// beats letting a multi-million-element slice grow.
	n := strings.Count(data, "\n")
	t.offsets = make([]uint32, 0, n+1)
	start := 0
	for i := 0; i < len(data); i++ {
		if data[i] != '\n' {
			continue
		}
		if i > start {
			t.offsets = append(t.offsets, uint32(start))
		}
		start = i + 1
	}
	if start < len(data) {
		t.offsets = append(t.offsets, uint32(start))
	}
	return t
}

// line returns row i without its trailing newline.
func (t table) line(i int) string {
	start := int(t.offsets[i])
	end := len(t.data)
	if i+1 < len(t.offsets) {
		end = int(t.offsets[i+1]) - 1
	}
	if end > len(t.data) {
		end = len(t.data)
	}
	return t.data[start:end]
}

// lookup returns the split fields of every row whose first column equals
// key.
//
// Rows sharing a first column are contiguous because the generator sorts
// whole lines and the separator is a tab: "est\tnoun" sorts before "esta"
// since 0x09 is below every letter. So one binary search for the first
// match plus a short forward scan finds them all.
func (t table) lookup(key string, fields int) [][]string {
	if key == "" || len(t.offsets) == 0 {
		return nil
	}
	prefix := key + "\t"
	i := sort.Search(len(t.offsets), func(i int) bool { return t.line(i) >= prefix })

	var out [][]string
	for ; i < len(t.offsets); i++ {
		line := t.line(i)
		if !strings.HasPrefix(line, prefix) {
			break
		}
		f := strings.SplitN(line, "\t", fields)
		if len(f) == fields {
			out = append(out, f)
		}
	}
	return out
}

// mustGunzip inflates an embedded table into a string without the copies
// the obvious implementation makes.
//
// io.ReadAll grows its buffer by doubling and string(b) then copies the
// result, so a 95 MB table peaks at roughly three times its own size —
// which, across three languages, was most of the process's memory.
// strings.Builder with the size taken from the gzip footer allocates
// exactly once and hands back the string without copying it.
func mustGunzip(b []byte) string {
	zr, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		panic("reading: corrupt embedded lexicon table: " + err.Error())
	}
	defer zr.Close()

	var sb strings.Builder
	sb.Grow(gzipSize(b))
	if _, err := io.Copy(&sb, zr); err != nil {
		panic("reading: could not inflate embedded lexicon table: " + err.Error())
	}
	return sb.String()
}

// gzipSize reads ISIZE, the uncompressed length modulo 2^32 stored in the
// last four bytes of a gzip stream. It is a hint for preallocation, so a
// wrong answer costs a reallocation rather than correctness.
func gzipSize(b []byte) int {
	if len(b) < 4 {
		return 0
	}
	tail := b[len(b)-4:]
	return int(uint32(tail[0]) | uint32(tail[1])<<8 | uint32(tail[2])<<16 | uint32(tail[3])<<24)
}

// SplitGlosses turns a stored "a; b; c" into individual glosses.
func SplitGlosses(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, "; ")
	out := parts[:0]
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
