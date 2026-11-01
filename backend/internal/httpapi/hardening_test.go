package httpapi

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bogdanbojan/sermokey/backend/internal/reading"
	"github.com/bogdanbojan/sermokey/backend/internal/reading/epub"
)

// A panic in a handler must become a 500 with a JSON body, not a dropped
// connection. The codebase panics deliberately on startup invariants, and
// net/http's per-connection recovery leaves the client with a truncated
// response and the access log with no record of the request at all.
func TestPanicBecomesJSON500(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /boom", func(http.ResponseWriter, *http.Request) {
		panic("deliberate")
	})
	h := withRecovery(mux)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want JSON", ct)
	}
	if !strings.Contains(rec.Body.String(), `"error"`) {
		t.Errorf("body = %q, want an error object", rec.Body.String())
	}
}

// An upload larger than the cap must be rejected without the body first
// being spooled to disk in full. ParseMultipartForm's argument is
// maxMemory, not a limit, so before MaxBytesReader this wrote the entire
// request to $TMPDIR before any size check ran.
func TestImportRejectsOversizedBody(t *testing.T) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", "big.epub")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(fw, io.LimitReader(zeroReader{}, MaxImportBytes+(4<<20))); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/texts/import?lang=fr", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	NewRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", rec.Code)
	}
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 'a'
	}
	return len(p), nil
}

// A zip declaring an implausible number of members is rejected before the
// central directory is turned into a map and sorted.
func TestEPUBRejectsTooManyEntries(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for i := 0; i < epub.MaxZipEntries+1; i++ {
		w, err := zw.Create(fmt.Sprintf("f%d.txt", i))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte("x")); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	data := buf.Bytes()
	_, err := epub.Parse(bytes.NewReader(data), int64(len(data)))
	if err == nil || !strings.Contains(err.Error(), "too many files") {
		t.Fatalf("err = %v, want a too-many-files rejection", err)
	}
}

// 403 is for a bundled text; a nonexistent id is 404 even when its name
// contains the word the old substring check looked for.
func TestDeleteStatusDistinguishesBundledFromMissing(t *testing.T) {
	r := NewRouter()

	// An id that contains "bundled" but does not exist. The previous
	// implementation chose its status with
	// strings.Contains(err.Error(), "bundled"), and the not-found message
	// interpolates the caller's id — so this returned 403.
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/texts/bundled-nonsense?lang=fr", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("missing text: status = %d, want 404", rec.Code)
	}

	// A genuinely bundled text is 403.
	p, ok := reading.For("fr")
	if !ok {
		t.Skip("french reading not registered")
	}
	metas := p.Catalog.List()
	if len(metas) == 0 {
		t.Skip("no bundled texts")
	}
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/texts/"+metas[0].ID+"?lang=fr", nil))
	if rec.Code != http.StatusForbidden {
		t.Errorf("bundled text: status = %d, want 403", rec.Code)
	}
}

func TestLibraryRemoveReturnsSentinels(t *testing.T) {
	p, ok := reading.For("fr")
	if !ok {
		t.Skip("french reading not registered")
	}
	if err := p.Library.Remove("definitely-not-a-text"); !errors.Is(err, reading.ErrNoSuchText) {
		t.Errorf("err = %v, want ErrNoSuchText", err)
	}
	if metas := p.Catalog.List(); len(metas) > 0 {
		if err := p.Library.Remove(metas[0].ID); !errors.Is(err, reading.ErrBundledText) {
			t.Errorf("err = %v, want ErrBundledText", err)
		}
	}
}

// Immutable content carries a validator and honours a conditional request.
func TestSectionIsConditionallyCacheable(t *testing.T) {
	r := NewRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/texts/la-parure?lang=fr", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	etag := rec.Header().Get("ETag")
	if etag == "" {
		t.Fatal("no ETag on an immutable section payload")
	}

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/texts/la-parure?lang=fr", nil)
	req2.Header.Set("If-None-Match", etag)
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNotModified {
		t.Errorf("conditional request: status = %d, want 304", rec2.Code)
	}
	if rec2.Body.Len() != 0 {
		t.Errorf("304 carried a %d-byte body", rec2.Body.Len())
	}
}

func TestSectionIsCompressed(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/texts/la-parure?lang=fr", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	NewRouter().ServeHTTP(rec, req)

	if got := rec.Header().Get("Content-Encoding"); got != "gzip" {
		t.Errorf("Content-Encoding = %q, want gzip", got)
	}
	if !strings.Contains(rec.Header().Get("Vary"), "Accept-Encoding") {
		t.Errorf("Vary = %q, want it to include Accept-Encoding", rec.Header().Get("Vary"))
	}
}

// http.TimeoutHandler copies the inner handler's headers over the outer
// ones with a whole-slice assignment, so a header written by both layers
// loses the outer value. Vary is the one that mattered: CORS set "Origin",
// the encoder set "Accept-Encoding", and only the second survived.
func TestVarySurvivesTimeoutHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/texts/la-parure?lang=fr", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Origin", "http://localhost:5173")
	NewRouter().ServeHTTP(rec, req)

	vary := rec.Header().Get("Vary")
	for _, want := range []string{"Accept-Encoding", "Origin"} {
		if !strings.Contains(vary, want) {
			t.Errorf("Vary = %q, want it to include %q", vary, want)
		}
	}
}
