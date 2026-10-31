package reading

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

// MemoryCatalog is a Catalog over documents already in memory. The bundled
// per-language corpora use it; so will an EPUB importer, once it has parsed
// a file into Documents.
type MemoryCatalog struct {
	order []string
	byID  map[string]Document
}

// NewMemoryCatalog builds a catalog from documents, preserving their order.
// Panics on a duplicate or empty ID, which can only be an authoring bug in
// the bundled corpus.
func NewMemoryCatalog(docs ...Document) *MemoryCatalog {
	c := &MemoryCatalog{byID: make(map[string]Document, len(docs))}
	for _, d := range docs {
		if d.Meta.ID == "" {
			panic("reading: document with empty ID")
		}
		if _, dup := c.byID[d.Meta.ID]; dup {
			panic("reading: duplicate document ID " + d.Meta.ID)
		}
		c.byID[d.Meta.ID] = d
		c.order = append(c.order, d.Meta.ID)
	}
	return c
}

// List implements Catalog.
func (c *MemoryCatalog) List() []DocumentMeta {
	out := make([]DocumentMeta, 0, len(c.order))
	for _, id := range c.order {
		out = append(out, c.byID[id].Meta)
	}
	return out
}

// Get implements Catalog.
func (c *MemoryCatalog) Get(id string) (Document, bool) {
	d, ok := c.byID[id]
	return d, ok
}

// Build tokenizes every paragraph of a document and computes the derived
// metadata (token counts, lemma frequencies) that the library view needs.
//
// Doing this once at package init rather than per request keeps the hover
// path free of work: by the time a reader asks for a section, its tokens
// and its lemma frequency table already exist.
func Build(meta DocumentMeta, tok Tokenizer, an Analyzer, sections []Section) Document {
	meta.WordFrequency = map[string]int{}
	meta.FunctionTokenCount = 0
	meta.ProperTokenCount = 0
	meta.Sections = make([]SectionMeta, 0, len(sections))
	meta.TokenCount = 0

	// Tokenize everything first: which surfaces are names is a property of
	// the whole document, and it has to be known before any of it is
	// analyzed or the counts will disagree with what the reader is shown.
	built := make([]Section, 0, len(sections))
	for _, s := range sections {
		paras := make([]Paragraph, 0, len(s.Paragraphs))
		for _, p := range s.Paragraphs {
			p.Tokens = tok.Tokenize(p.Text)
			paras = append(paras, p)
		}
		built = append(built, Section{Title: s.Title, Paragraphs: paras})
	}
	properNouns := FindProperNounsIn(built)
	analyze := properNounAware(an, properNouns)

	for si := range built {
		s := built[si]
		sectionTokens := 0
		for _, p := range s.Paragraphs {
			for _, t := range p.Tokens {
				if !t.Word {
					continue
				}
				sectionTokens++
				if as := analyze(t.Surface(p.Text)); len(as) > 0 {
					switch {
					case as[0].Function:
						meta.FunctionTokenCount++
					case as[0].Proper:
						meta.ProperTokenCount++
					default:
						meta.WordFrequency[as[0].WordID]++
					}
				}
			}
		}
		meta.Sections = append(meta.Sections, SectionMeta{Title: s.Title, TokenCount: sectionTokens})
		meta.TokenCount += sectionTokens
	}
	return Document{Meta: meta, Sections: built, ProperNouns: properNouns}
}

// ProperNounAware returns an analysis function that applies a document's
// name set, falling back to plain analysis for analyzers that do not
// support the hint.
func ProperNounAware(an Analyzer, properNouns map[string]bool) func(string) []Analysis {
	return properNounAware(an, properNouns)
}

func properNounAware(an Analyzer, properNouns map[string]bool) func(string) []Analysis {
	hinter, ok := an.(ProperNounAnalyzer)
	if !ok || len(properNouns) == 0 {
		return an.Analyze
	}
	return func(surface string) []Analysis {
		if properNouns[strings.ToLower(surface)] {
			return hinter.AnalyzeProperNoun(surface)
		}
		return an.Analyze(surface)
	}
}

// AssertRoundTrip verifies the tokenizer's core invariant: the tokens of a
// string, concatenated in order, reproduce that string exactly. A reader
// that silently drops a character — or worse, reorders one — corrupts the
// text the learner is trying to understand, so this is checked over the
// entire bundled corpus on every test run.
func AssertRoundTrip(tok Tokenizer, text string) error {
	tokens := tok.Tokenize(text)
	var b strings.Builder
	prevEnd := 0
	for i, t := range tokens {
		if t.Start != prevEnd {
			return fmt.Errorf("token %d starts at %d, expected %d (gap or overlap)", i, t.Start, prevEnd)
		}
		if t.End < t.Start || t.End > len(text) {
			return fmt.Errorf("token %d has invalid range [%d,%d) for text of length %d", i, t.Start, t.End, len(text))
		}
		if t.Start == t.End {
			return fmt.Errorf("token %d is empty at offset %d", i, t.Start)
		}
		b.WriteString(text[t.Start:t.End])
		prevEnd = t.End
	}
	if prevEnd != len(text) {
		return fmt.Errorf("tokens cover %d of %d bytes", prevEnd, len(text))
	}
	if got := b.String(); got != text {
		return fmt.Errorf("round-trip mismatch:\n  want %q\n  got  %q", text, got)
	}
	return nil
}

// CorpusCheck configures ValidateCorpus.
type CorpusCheck struct {
	Provider Provider
	// MinCoverage is the minimum share of word tokens that must resolve to
	// a lemma, per document. Set below 1 to tolerate proper nouns.
	MinCoverage float64
	// CuratedWordIDs is the set of SRS ids that exist in the vocabulary
	// deck. Any curated analysis pointing outside it would silently create
	// an orphan card, so it is checked here — the same class of bug
	// grammar.ValidateVocab guards against with its duplicate-ID check.
	CuratedWordIDs map[string]bool
}

// ValidateCorpus checks every bundled document against the tokenizer and
// analyzer that will serve it. It returns every problem found rather than
// the first, so a single test run reports all of them.
func ValidateCorpus(c CorpusCheck) []error {
	var errs []error
	p := c.Provider

	for _, meta := range p.Catalog.List() {
		doc, ok := p.Catalog.Get(meta.ID)
		if !ok {
			errs = append(errs, fmt.Errorf("%s: listed but not retrievable", meta.ID))
			continue
		}
		if len(doc.Sections) == 0 {
			errs = append(errs, fmt.Errorf("%s: has no sections", meta.ID))
		}
		if meta.Title == "" || meta.License == "" {
			errs = append(errs, fmt.Errorf("%s: missing title or license", meta.ID))
		}

		var words, resolved int
		for si, s := range doc.Sections {
			if len(s.Paragraphs) == 0 {
				errs = append(errs, fmt.Errorf("%s section %d: has no paragraphs", meta.ID, si))
			}
			for pi, para := range s.Paragraphs {
				if strings.TrimSpace(para.Text) == "" {
					errs = append(errs, fmt.Errorf("%s section %d paragraph %d: is blank", meta.ID, si, pi))
				}
				if err := AssertRoundTrip(p.Tokenizer, para.Text); err != nil {
					errs = append(errs, fmt.Errorf("%s section %d paragraph %d: %w", meta.ID, si, pi, err))
				}
				for _, t := range para.Tokens {
					if !t.Word {
						continue
					}
					words++
					as := p.Analyzer.Analyze(t.Surface(para.Text))
					if len(as) == 0 {
						continue
					}
					resolved++
					if a := as[0]; a.Source == SourceCurated && c.CuratedWordIDs != nil {
						if !c.CuratedWordIDs[a.WordID] {
							errs = append(errs, fmt.Errorf(
								"%s: curated analysis of %q yields WordID %q, which is not in the vocabulary deck",
								meta.ID, t.Surface(para.Text), a.WordID))
						}
					}
				}
			}
		}

		if words == 0 {
			errs = append(errs, fmt.Errorf("%s: contains no word tokens", meta.ID))
			continue
		}
		if cov := float64(resolved) / float64(words); cov < c.MinCoverage {
			errs = append(errs, fmt.Errorf(
				"%s: analyzer resolves %.1f%% of %d word tokens, want at least %.1f%%",
				meta.ID, cov*100, words, c.MinCoverage*100))
		}
	}
	return errs
}

// LoadCorpus parses, tokenizes and analyzes every bundled text in a
// directory of an embedded filesystem.
//
// Adding a text is therefore a data change: drop a file in, and the glob
// finds it. The numeric filename prefix sets display order. Panics on a
// malformed file, which can only be an authoring bug in content we ship.
func LoadCorpus(fsys fs.FS, dir string, tok Tokenizer, an Analyzer) *MemoryCatalog {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		panic("reading: cannot read embedded corpus " + dir + ": " + err.Error())
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	docs := make([]Document, 0, len(names))
	for _, name := range names {
		raw, err := fs.ReadFile(fsys, path.Join(dir, name))
		if err != nil {
			panic("reading: cannot read " + name + ": " + err.Error())
		}
		meta, sections := ParseDocument(name, string(raw))
		docs = append(docs, Build(meta, tok, an, sections))
	}
	return NewMemoryCatalog(docs...)
}
