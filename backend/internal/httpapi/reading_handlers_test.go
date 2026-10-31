package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func getReading(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	NewRouter().ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out T
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decoding response: %v\nbody: %s", err, rec.Body.String())
	}
	return out
}

func TestHandleTextsListsFrenchCorpus(t *testing.T) {
	texts := decode[[]ReadingTextDTO](t, getReading(t, "/api/texts?lang=fr"))
	if len(texts) != 3 {
		t.Fatalf("got %d texts, want 3", len(texts))
	}
	for _, x := range texts {
		if x.ID == "" || x.Title == "" || x.License == "" || x.Blurb == "" {
			t.Errorf("%+v: incomplete library card", x)
		}
		if x.SectionCount < 1 || x.TokenCount < 1 {
			t.Errorf("%s: SectionCount=%d TokenCount=%d", x.ID, x.SectionCount, x.TokenCount)
		}
		// Articles and clitics are roughly a quarter of running French.
		// A zero here means the flag is not reaching the DTO, which
		// silently deflates every coverage figure on the library screen.
		if x.FunctionTokenCount < x.TokenCount/10 {
			t.Errorf("%s: functionTokenCount=%d of %d tokens — implausibly low",
				x.ID, x.FunctionTokenCount, x.TokenCount)
		}
		if x.FunctionTokenCount >= x.TokenCount {
			t.Errorf("%s: functionTokenCount=%d exceeds tokenCount=%d",
				x.ID, x.FunctionTokenCount, x.TokenCount)
		}
		// Without this the client cannot compute how much of the text the
		// reader already knows, which is the whole difficulty signal.
		if len(x.WordFrequency) == 0 {
			t.Errorf("%s: empty wordFrequency", x.ID)
		}
		// Keyed by SRS card id so the client can intersect it with the
		// deck directly rather than guessing at a curated word's category.
		for id := range x.WordFrequency {
			if !strings.Contains(id, "|") {
				t.Errorf("%s: frequency key %q is not a card id", x.ID, id)
			}
		}
	}
}

// Every studied language now has a corpus. The registry still models
// reading as a subset of the supported languages — adding a fourth
// language's grammar should not oblige anyone to write it a corpus on the
// same day — so the distinction that matters is kept under test: a known
// language without reading would be a 404, an unknown one is a 400.
func TestHandleTextsLanguageAvailability(t *testing.T) {
	for _, lang := range []string{"fr", "es", "de"} {
		texts := decode[[]ReadingTextDTO](t, getReading(t, "/api/texts?lang="+lang))
		if len(texts) != 3 {
			t.Errorf("%s: got %d texts, want 3", lang, len(texts))
		}
		for _, x := range texts {
			if x.TokenCount == 0 || len(x.WordFrequency) == 0 || x.FunctionTokenCount == 0 {
				t.Errorf("%s/%s: not analyzed (%d tokens, %d vocab, %d function)",
					lang, x.ID, x.TokenCount, len(x.WordFrequency), x.FunctionTokenCount)
			}
		}
	}
	if rec := getReading(t, "/api/texts?lang=xx"); rec.Code != http.StatusBadRequest {
		t.Errorf("unknown language: status = %d, want 400", rec.Code)
	}
	// French is the default language, so an absent lang must still work.
	if rec := getReading(t, "/api/texts"); rec.Code != http.StatusOK {
		t.Errorf("default language: status = %d, want 200", rec.Code)
	}
}

// Each language's reader has to actually work end to end, not merely list
// texts: tokens resolving to glosses is the whole feature.
func TestEachLanguageServesAReadableSection(t *testing.T) {
	for _, tc := range []struct{ lang, id, word string }{
		{"fr", "le-marche-du-samedi", "fenêtre"},
		{"es", "el-mercado-del-sabado", "ventana"},
		{"de", "der-samstagsmarkt", "Fenster"},
	} {
		t.Run(tc.lang, func(t *testing.T) {
			sec := decode[ReadingSectionDTO](t, getReading(t, "/api/texts/"+tc.id+"?lang="+tc.lang))
			if len(sec.Paragraphs) == 0 || len(sec.Lexicon) == 0 {
				t.Fatalf("%d paragraphs, %d lexicon entries", len(sec.Paragraphs), len(sec.Lexicon))
			}
			var curated int
			for _, e := range sec.Lexicon {
				if e.Curated {
					curated++
				}
			}
			if curated == 0 {
				t.Error("no curated words resolved, so nothing links back to the vocabulary deck")
			}

			got := decode[[]ReadingEntryDTO](t, getReading(t, "/api/lookup?lang="+tc.lang+"&q="+tc.word))
			if len(got) == 0 || len(got[0].Senses) == 0 {
				t.Errorf("lookup of %q returned nothing usable", tc.word)
			}
		})
	}
}

func TestHandleTextSection(t *testing.T) {
	sec := decode[ReadingSectionDTO](t, getReading(t, "/api/texts/le-marche-du-samedi?lang=fr"))
	if sec.TextID != "le-marche-du-samedi" || sec.Index != 0 {
		t.Fatalf("got %s section %d", sec.TextID, sec.Index)
	}
	if sec.SectionCount < 2 {
		t.Errorf("SectionCount = %d, want at least 2", sec.SectionCount)
	}
	if len(sec.Paragraphs) == 0 || len(sec.Lexicon) == 0 {
		t.Fatalf("%d paragraphs, %d lexicon entries", len(sec.Paragraphs), len(sec.Lexicon))
	}
}

// The client reconstructs the page by slicing paragraph text with these
// offsets. If a span were out of range or out of order it would render
// garbled or truncated prose, so the contract is checked explicitly.
func TestTextSectionSpansAreValidAndContiguous(t *testing.T) {
	sec := decode[ReadingSectionDTO](t, getReading(t, "/api/texts/la-parure?lang=fr"))
	for pi, p := range sec.Paragraphs {
		if len(p.Spans)%SpanStride != 0 {
			t.Fatalf("paragraph %d: %d span integers, not a multiple of %d", pi, len(p.Spans), SpanStride)
		}
		var rebuilt strings.Builder
		prevEnd := 0
		for i := 0; i < len(p.Spans); i += SpanStride {
			start, end, ref := p.Spans[i], p.Spans[i+1], p.Spans[i+2]
			if start != prevEnd {
				t.Fatalf("paragraph %d: span at %d starts at %d, want %d", pi, i, start, prevEnd)
			}
			if end <= start || end > len(p.Text) {
				t.Fatalf("paragraph %d: invalid span [%d,%d) over %d bytes", pi, start, end, len(p.Text))
			}
			if ref != NoEntry && (ref < 0 || ref >= len(sec.Lexicon)) {
				t.Fatalf("paragraph %d: lexicon ref %d out of range (%d entries)", pi, ref, len(sec.Lexicon))
			}
			rebuilt.WriteString(p.Text[start:end])
			prevEnd = end
		}
		if rebuilt.String() != p.Text {
			t.Errorf("paragraph %d: spans do not reconstruct the text", pi)
		}
	}
}

// Pooling is what keeps the payload small enough to ship the whole
// section's dictionary up front, which is in turn what makes hover
// instantaneous. If entries were emitted per occurrence this would fail.
func TestTextSectionLexiconIsPooled(t *testing.T) {
	sec := decode[ReadingSectionDTO](t, getReading(t, "/api/texts/la-parure?lang=fr"))
	var wordSpans int
	for _, p := range sec.Paragraphs {
		for i := 2; i < len(p.Spans); i += SpanStride {
			if p.Spans[i] != NoEntry {
				wordSpans++
			}
		}
	}
	if len(sec.Lexicon) >= wordSpans {
		t.Errorf("%d lexicon entries for %d word spans — not deduplicated", len(sec.Lexicon), wordSpans)
	}
	seen := map[string]bool{}
	for _, e := range sec.Lexicon {
		if seen[e.Surface] {
			t.Errorf("duplicate lexicon entry for %q", e.Surface)
		}
		seen[e.Surface] = true
		if len(e.Senses) == 0 {
			t.Errorf("%q: entry with no senses", e.Surface)
		}
		if e.WordID == "" {
			t.Errorf("%q: no WordID, so it cannot be saved to the deck", e.Surface)
		}
	}
}

func TestHandleTextSectionPagination(t *testing.T) {
	first := decode[ReadingSectionDTO](t, getReading(t, "/api/texts/une-lettre-de-marseille?lang=fr&section=0"))
	last := decode[ReadingSectionDTO](t, getReading(t,
		"/api/texts/une-lettre-de-marseille?lang=fr&section="+itoa(first.SectionCount-1)))
	if last.Index != first.SectionCount-1 {
		t.Errorf("last section index = %d, want %d", last.Index, first.SectionCount-1)
	}
	if first.SectionTitle == last.SectionTitle {
		t.Error("first and last sections have the same title")
	}
	for _, bad := range []string{"-1", "99", "abc"} {
		if rec := getReading(t, "/api/texts/une-lettre-de-marseille?lang=fr&section="+bad); rec.Code != http.StatusBadRequest {
			t.Errorf("section=%s: status = %d, want 400", bad, rec.Code)
		}
	}
}

func TestHandleTextSectionUnknownText(t *testing.T) {
	if rec := getReading(t, "/api/texts/nope?lang=fr"); rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestHandleLookup(t *testing.T) {
	got := decode[[]ReadingEntryDTO](t, getReading(t, "/api/lookup?lang=fr&q=parlons"))
	if len(got) != 1 {
		t.Fatalf("got %d entries, want 1", len(got))
	}
	if got[0].Lemma != "parler" {
		t.Errorf("lemma = %q, want parler", got[0].Lemma)
	}
	if got[0].FormNote == "" {
		t.Error("no form note, so the UI cannot explain the inflection")
	}
}

// Phrase selection is the reason this endpoint exists: a drag-selected span
// is not a single token and so cannot come from the section lexicon.
func TestHandleLookupPhrase(t *testing.T) {
	got := decode[[]ReadingEntryDTO](t, getReading(t, "/api/lookup?lang=fr&q=tout%20%C3%A0%20coup"))
	if len(got) == 0 {
		t.Fatal("phrase lookup returned nothing")
	}
}

// "Not in the dictionary" is an ordinary outcome the reader renders inline,
// not a failure. Returning 404 would make it look like one.
func TestHandleLookupUnknownIsEmptyNotAnError(t *testing.T) {
	rec := getReading(t, "/api/lookup?lang=fr&q=zzzqqqxx")
	got := decode[[]ReadingEntryDTO](t, rec)
	if len(got) != 0 {
		t.Errorf("got %d entries, want 0", len(got))
	}
}

func TestHandleLookupValidation(t *testing.T) {
	if rec := getReading(t, "/api/lookup?lang=fr"); rec.Code != http.StatusBadRequest {
		t.Errorf("missing q: status = %d, want 400", rec.Code)
	}
	long := "/api/lookup?lang=fr&q=" + strings.Repeat("a", maxLookupRunes+1)
	if rec := getReading(t, long); rec.Code != http.StatusBadRequest {
		t.Errorf("over-long q: status = %d, want 400", rec.Code)
	}
}

// A word in the curriculum must be saveable onto the card the vocabulary
// tab already schedules. A different id here would split one word's review
// history across two cards.
func TestLookupCuratedWordsReuseTheVocabularyDeckID(t *testing.T) {
	for _, tc := range []struct{ q, wantID string }{
		{"maison", "maison|maison"},
		{"maisons", "maison|maison"},
		{"parlons", "parler|verbes"},
	} {
		got := decode[[]ReadingEntryDTO](t, getReading(t, "/api/lookup?lang=fr&q="+tc.q))
		if len(got) == 0 {
			t.Fatalf("%s: no entry", tc.q)
		}
		if !got[0].Curated {
			t.Errorf("%s: not marked curated", tc.q)
		}
		if got[0].WordID != tc.wantID {
			t.Errorf("%s: WordID = %q, want %q", tc.q, got[0].WordID, tc.wantID)
		}
	}
}

// "est" is both a curated noun ("east") and the third-person singular of
// être. The verb must be primary and the noun must still be reachable, so
// a reader can correct the ranking from the panel.
func TestLookupOffersAlternativeReadings(t *testing.T) {
	got := decode[[]ReadingEntryDTO](t, getReading(t, "/api/lookup?lang=fr&q=est"))
	if len(got) == 0 {
		t.Fatal("no entry for est")
	}
	if got[0].Lemma != "être" {
		t.Errorf("primary lemma = %q, want être", got[0].Lemma)
	}
	var sawEast bool
	for _, alt := range got[0].Alternatives {
		if alt.Lemma == "est" {
			sawEast = true
		}
	}
	if !sawEast {
		t.Error("the noun reading of est is not offered as an alternative")
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
