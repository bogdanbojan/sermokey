package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/bogdanbojan/sermokey/backend/internal/reading"
)

// maxLookupRunes bounds the ad-hoc lookup endpoint. Phrase selection in the
// reader is capped at a few tokens — formulaic sequences like "tout à coup"
// or "avoir besoin de" are short — so anything longer is a client bug or an
// attempt to use the endpoint as something it is not.
const maxLookupRunes = 120

// readingProviderFromRequest resolves the ?lang= parameter to a reading
// provider, writing the error response itself if it cannot.
//
// Reading is deliberately available for fewer languages than grammar is, so
// a valid language with no corpus is a 404 rather than a 400: the language
// is fine, the reading resource simply does not exist yet.
func readingProviderFromRequest(w http.ResponseWriter, r *http.Request) (reading.Provider, bool) {
	lang, ok := languageFromRequest(w, r)
	if !ok {
		return reading.Provider{}, false
	}
	p, ok := reading.For(lang.Code())
	if !ok {
		writeError(w, http.StatusNotFound, "reading is not available for language "+lang.Code()+" yet")
		return reading.Provider{}, false
	}
	return p, true
}

// handleTexts lists the library for one language.
func handleTexts(w http.ResponseWriter, r *http.Request) {
	p, ok := readingProviderFromRequest(w, r)
	if !ok {
		return
	}
	metas := p.Catalog.List()
	out := make([]ReadingTextDTO, 0, len(metas))
	for _, m := range metas {
		out = append(out, toReadingTextDTO(m, p.Library.IsImported(m.ID)))
	}
	writeJSON(w, http.StatusOK, out)
}

// toReadingTextDTO is shared by the list and import endpoints so the two
// cannot drift: an imported book must look exactly like a bundled one on
// the library screen, bar the control that removes it.
func toReadingTextDTO(m reading.DocumentMeta, imported bool) ReadingTextDTO {
	return ReadingTextDTO{
		ID:                 m.ID,
		Title:              m.Title,
		Author:             m.Author,
		Level:              string(m.Level),
		Blurb:              m.Blurb,
		Source:             m.Source,
		License:            m.License,
		SectionCount:       len(m.Sections),
		TokenCount:         m.TokenCount,
		FunctionTokenCount: m.FunctionTokenCount,
		ProperTokenCount:   m.ProperTokenCount,
		WordFrequency:      m.WordFrequency,
		Imported:           imported,
	}
}

// handleTextSection serves one section of one text, with every dictionary
// entry its tokens need. Everything here is precomputed at startup; the
// request path only pools and marshals.
func handleTextSection(w http.ResponseWriter, r *http.Request) {
	p, ok := readingProviderFromRequest(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	doc, found := p.Catalog.Get(id)
	if !found {
		writeError(w, http.StatusNotFound, "unknown text: "+id)
		return
	}

	index := 0
	if raw := r.URL.Query().Get("section"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 || n >= len(doc.Sections) {
			writeError(w, http.StatusBadRequest,
				"section must be between 0 and "+strconv.Itoa(len(doc.Sections)-1))
			return
		}
		index = n
	}

	section := doc.Sections[index]
	// The pool uses the document's name set, so the lexicon the reader
	// hovers agrees with the counts on the library card. Analyzing the
	// same word two different ways in the two places would show a name
	// glossed as a common word while the card claimed it was neither.
	pool := newEntryPool(reading.ProperNounAware(p.Analyzer, doc.ProperNouns))
	paragraphs := make([]ReadingParagraphDTO, 0, len(section.Paragraphs))
	for _, para := range section.Paragraphs {
		spans := make([]int, 0, len(para.Tokens)*SpanStride)
		for _, t := range para.Tokens {
			ref := NoEntry
			if t.Word {
				ref = pool.ref(t.Surface(para.Text))
			}
			spans = append(spans, t.Start, t.End, ref)
		}
		paragraphs = append(paragraphs, ReadingParagraphDTO{Text: para.Text, Spans: spans})
	}

	writeJSON(w, http.StatusOK, ReadingSectionDTO{
		TextID:       doc.Meta.ID,
		Title:        doc.Meta.Title,
		SectionTitle: section.Title,
		Index:        index,
		SectionCount: len(doc.Sections),
		License:      doc.Meta.License,
		Source:       doc.Meta.Source,
		Paragraphs:   paragraphs,
		Lexicon:      pool.entries,
	})
}

// handleLookup resolves an arbitrary surface form or phrase. The reader
// uses it for drag-selected spans, which are not single tokens and so
// cannot come from the section lexicon.
//
// It is a GET so the whole API stays read-only and cacheable; phrases are
// short enough to sit comfortably in a query string.
func handleLookup(w http.ResponseWriter, r *http.Request) {
	p, ok := readingProviderFromRequest(w, r)
	if !ok {
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeError(w, http.StatusBadRequest, "q is required")
		return
	}
	if len([]rune(q)) > maxLookupRunes {
		writeError(w, http.StatusBadRequest,
			"q must be at most "+strconv.Itoa(maxLookupRunes)+" characters")
		return
	}

	analyses := p.Analyzer.Analyze(q)
	if len(analyses) == 0 {
		// Not an error: "this phrase is not in the dictionary" is a normal,
		// expected answer, and the reader renders it as such. Returning 404
		// here would make an ordinary outcome look like a failure.
		writeJSON(w, http.StatusOK, []ReadingEntryDTO{})
		return
	}
	writeJSON(w, http.StatusOK, []ReadingEntryDTO{toEntryDTO(strings.ToLower(q), analyses)})
}

// entryPool deduplicates dictionary entries within one section. The same
// word occurs many times on a page, and shipping a copy per occurrence
// would dominate the payload.
type entryPool struct {
	analyze func(string) []reading.Analysis
	entries []ReadingEntryDTO
	byForm  map[string]int
}

func newEntryPool(analyze func(string) []reading.Analysis) *entryPool {
	return &entryPool{analyze: analyze, byForm: map[string]int{}}
}

// ref returns the pool index for a surface form, analyzing it on first
// sight, or NoEntry when the dictionary has nothing for it.
func (p *entryPool) ref(surface string) int {
	key := strings.ToLower(surface)
	if i, ok := p.byForm[key]; ok {
		return i
	}
	analyses := p.analyze(surface)
	if len(analyses) == 0 {
		p.byForm[key] = NoEntry
		return NoEntry
	}
	p.entries = append(p.entries, toEntryDTO(key, analyses))
	i := len(p.entries) - 1
	p.byForm[key] = i
	return i
}

// toEntryDTO renders the best analysis as the primary entry and the rest as
// alternatives, so the hover popup can show one answer while the pinned
// panel still offers the others — which is how a reader corrects a
// mis-ranked homograph.
func toEntryDTO(surface string, analyses []reading.Analysis) ReadingEntryDTO {
	best := analyses[0]
	dto := ReadingEntryDTO{
		Surface:   surface,
		Lemma:     best.Lemma,
		POS:       best.POS,
		Gender:    best.Gender,
		FormNote:  best.FormNote,
		Senses:    toSenseDTOs(best.Senses),
		WordID:    best.WordID,
		Curated:   best.Source == reading.SourceCurated,
		Function:  best.Function,
		Proper:    best.Proper,
		Uncertain: best.Source == reading.SourceDerived,
	}
	for _, a := range analyses[1:] {
		dto.Alternatives = append(dto.Alternatives, ReadingAltDTO{
			Lemma:    a.Lemma,
			POS:      a.POS,
			Gender:   a.Gender,
			FormNote: a.FormNote,
			Senses:   toSenseDTOs(a.Senses),
			WordID:   a.WordID,
			Curated:  a.Source == reading.SourceCurated,
		})
	}
	return dto
}

func toSenseDTOs(senses []reading.Sense) []ReadingSenseDTO {
	out := make([]ReadingSenseDTO, 0, len(senses))
	for _, s := range senses {
		if s.Gloss == "" {
			continue
		}
		out = append(out, ReadingSenseDTO{Gloss: s.Gloss})
	}
	return out
}
