package httpapi

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// sampleEPUB builds a small but structurally real French EPUB.
func sampleEPUB(t *testing.T, title, body string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	write := func(name, content string) {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	write("META-INF/container.xml", `<?xml version="1.0"?>
<container xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles><rootfile full-path="OEBPS/content.opf"/></rootfiles></container>`)
	write("OEBPS/content.opf", fmt.Sprintf(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>%s</dc:title><dc:creator>Anonyme</dc:creator><dc:language>fr</dc:language>
  </metadata>
  <manifest><item id="c1" href="ch1.xhtml" media-type="application/xhtml+xml"/></manifest>
  <spine><itemref idref="c1"/></spine>
</package>`, title))
	write("OEBPS/ch1.xhtml", `<html><body><h1>Chapitre un</h1><p>`+body+`</p></body></html>`)
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func uploadEPUB(t *testing.T, router http.Handler, filename string, data []byte) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	fw.Write(data)
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/texts/import?lang=fr", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// The point of the whole EPUB design: an imported book must be served by
// exactly the same section endpoint, with the same tokenization, the same
// lexicon and the same card ids as a bundled text. If this passes, nothing
// downstream — the reader screen, hovering, saving — needed to change.
func TestImportedEPUBBehavesLikeABundledText(t *testing.T) {
	router := NewRouter()
	data := sampleEPUB(t, "Le Petit Livre", "Elle se lève et regarde la maison.")

	rec := uploadEPUB(t, router, "livre.epub", data)
	if rec.Code != http.StatusCreated {
		t.Fatalf("import status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var imported ReadingTextDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &imported); err != nil {
		t.Fatal(err)
	}
	if imported.Title != "Le Petit Livre" || !imported.Imported {
		t.Errorf("imported DTO = %+v", imported)
	}
	if imported.TokenCount == 0 || len(imported.WordFrequency) == 0 {
		t.Errorf("book was not analyzed: %+v", imported)
	}
	// Function words must be counted for a book exactly as for a bundled
	// text, or imported books would all report an implausible coverage.
	if imported.FunctionTokenCount == 0 {
		t.Error("no function tokens counted for the imported book")
	}

	// It appears in the library, ahead of the bundled corpus.
	list := decode[[]ReadingTextDTO](t, getReading(t, "/api/texts?lang=fr"))
	if len(list) == 0 || list[0].ID != imported.ID {
		t.Errorf("imported book is not first in the library listing")
	}

	// And it reads through the ordinary section endpoint.
	sec := decode[ReadingSectionDTO](t, getReading(t, "/api/texts/"+imported.ID+"?lang=fr"))
	if sec.SectionTitle != "Chapitre un" || len(sec.Paragraphs) == 0 {
		t.Fatalf("section = %+v", sec)
	}
	if len(sec.Lexicon) == 0 {
		t.Fatal("imported section carries no lexicon, so hovering would do nothing")
	}
	// Curated words inside a book must resolve to the same deck cards as
	// anywhere else, or saving from a book would fork review history.
	var sawCurated bool
	for _, e := range sec.Lexicon {
		if e.Curated && e.WordID != "" {
			sawCurated = true
		}
	}
	if !sawCurated {
		t.Error("no curated words resolved inside the imported book")
	}

	// Phrase merging applies too: "se lève" is one lexical item.
	var sawPhrase bool
	for _, e := range sec.Lexicon {
		if strings.Contains(e.Surface, " ") {
			sawPhrase = true
		}
	}
	if !sawPhrase {
		t.Error("multi-word items were not merged inside the imported book")
	}
}

// Importing the same file twice must update one entry, not accumulate
// copies: the ID is derived from the file's bytes precisely so that
// reading positions and saved words keep pointing at the same book.
func TestImportIsIdempotent(t *testing.T) {
	router := NewRouter()
	data := sampleEPUB(t, "Deux fois", "Le chien dort.")

	first := uploadEPUB(t, router, "a.epub", data)
	second := uploadEPUB(t, router, "renamed.epub", data)
	if first.Code != http.StatusCreated || second.Code != http.StatusCreated {
		t.Fatalf("statuses = %d, %d", first.Code, second.Code)
	}
	var a, b ReadingTextDTO
	json.Unmarshal(first.Body.Bytes(), &a)
	json.Unmarshal(second.Body.Bytes(), &b)
	if a.ID != b.ID {
		t.Errorf("same file produced two ids: %q and %q", a.ID, b.ID)
	}

	list := decode[[]ReadingTextDTO](t, getReading(t, "/api/texts?lang=fr"))
	var n int
	for _, x := range list {
		if x.ID == a.ID {
			n++
		}
	}
	if n != 1 {
		t.Errorf("book appears %d times in the library, want 1", n)
	}
}

func TestDeleteImportedText(t *testing.T) {
	router := NewRouter()
	data := sampleEPUB(t, "À supprimer", "Le livre est court.")
	rec := uploadEPUB(t, router, "x.epub", data)
	var imported ReadingTextDTO
	json.Unmarshal(rec.Body.Bytes(), &imported)

	req := httptest.NewRequest(http.MethodDelete, "/api/texts/"+imported.ID+"?lang=fr", nil)
	del := httptest.NewRecorder()
	router.ServeHTTP(del, req)
	if del.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, body = %s", del.Code, del.Body.String())
	}
	if rec := getReading(t, "/api/texts/"+imported.ID+"?lang=fr"); rec.Code != http.StatusNotFound {
		t.Errorf("deleted book still readable: status %d", rec.Code)
	}
}

// Bundled texts are part of the product, not the reader's library.
func TestDeleteRefusesBundledTexts(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/texts/la-parure?lang=fr", nil)
	rec := httptest.NewRecorder()
	NewRouter().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
	if rec := getReading(t, "/api/texts/la-parure?lang=fr"); rec.Code != http.StatusOK {
		t.Error("a refused delete must leave the bundled text readable")
	}
}

func TestImportRejectsBadUploads(t *testing.T) {
	router := NewRouter()

	t.Run("not an epub", func(t *testing.T) {
		rec := uploadEPUB(t, router, "notes.txt", []byte("just some text"))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
		// The reader chose this file, so the message has to say what is
		// wrong with it rather than just "bad request".
		if !strings.Contains(strings.ToLower(rec.Body.String()), "epub") {
			t.Errorf("unhelpful error: %s", rec.Body.String())
		}
	})

	t.Run("no file field", func(t *testing.T) {
		var body bytes.Buffer
		mw := multipart.NewWriter(&body)
		mw.WriteField("wrong", "field")
		mw.Close()
		req := httptest.NewRequest(http.MethodPost, "/api/texts/import?lang=fr", &body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("language without reading support", func(t *testing.T) {
		data := sampleEPUB(t, "T", "Hola.")
		var body bytes.Buffer
		mw := multipart.NewWriter(&body)
		fw, _ := mw.CreateFormFile("file", "x.epub")
		fw.Write(data)
		mw.Close()
		req := httptest.NewRequest(http.MethodPost, "/api/texts/import?lang=xx", &body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	})
}

// The browser calls these cross-origin from the Vite dev server, so the
// preflight has to allow the methods the import UI uses.
func TestCORSAllowsImportMethods(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/api/texts/import?lang=fr", nil)
	rec := httptest.NewRecorder()
	NewRouter().ServeHTTP(rec, req)
	allow := rec.Header().Get("Access-Control-Allow-Methods")
	for _, m := range []string{"POST", "DELETE"} {
		if !strings.Contains(allow, m) {
			t.Errorf("Access-Control-Allow-Methods = %q, missing %s", allow, m)
		}
	}
}

// Analysing French prose with German rules produces nothing useful, and
// the imported book would look broken rather than misfiled. A confident
// mismatch is refused; metadata that is absent or unrecognised is allowed
// through, because EPUB language tags are frequently wrong and the reader
// knows better than the file does.
func TestImportRefusesAConfidentLanguageMismatch(t *testing.T) {
	router := NewRouter()
	data := sampleEPUB(t, "Un livre", "Elle se lève et regarde la maison.") // declares fr

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("file", "livre.epub")
	fw.Write(data)
	mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/texts/import?lang=de", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	for _, want := range []string{"fr", "de"} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("error should name both languages, got %s", rec.Body.String())
		}
	}

	// The same file into its own language is fine.
	if rec := uploadEPUB(t, router, "livre.epub", data); rec.Code != http.StatusCreated {
		t.Errorf("same file into French: status = %d, body = %s", rec.Code, rec.Body.String())
	}
}
