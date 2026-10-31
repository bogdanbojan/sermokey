package httpapi

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/bogdanbojan/sermokey/backend/internal/reading"
	"github.com/bogdanbojan/sermokey/backend/internal/reading/epub"
)

// MaxImportBytes caps an uploaded EPUB. Generous enough for an illustrated
// novel, small enough that a single request cannot exhaust memory — the
// file is read whole because the zip reader needs random access.
const MaxImportBytes = 32 << 20

// libraryStore persists imported books between restarts. The directory is
// configurable so the server stays runnable read-only; when it is empty,
// imports live only as long as the process, which is the right default for
// tests and for a container with no volume.
var libraryStore = epub.NewStore(os.Getenv("SERMOKEY_LIBRARY_DIR"))

// RestoreLibrary re-imports everything previously saved. Called once at
// startup, after the language packages have registered themselves.
func RestoreLibrary() {
	if !libraryStore.Enabled() {
		return
	}
	for _, code := range reading.Languages() {
		p, ok := reading.For(code)
		if !ok {
			continue
		}
		n, errs := libraryStore.LoadInto(p)
		for _, err := range errs {
			log.Printf("library: could not restore %s book: %v", code, err)
		}
		if n > 0 {
			log.Printf("library: restored %d imported %s book(s)", n, code)
		}
	}
}

// handleImportText accepts an EPUB upload and turns it into a readable
// text.
//
// This is the one endpoint in the API that is not a GET. The rest of the
// app is deliberately read-only — all user state lives in the browser —
// but an imported book is different in kind: it is content, it is far too
// large for localStorage, and it has to be tokenized and analyzed by the
// same engine that serves the bundled corpus. Keeping it server-side is
// what lets the reader screen stay completely unchanged.
func handleImportText(w http.ResponseWriter, r *http.Request) {
	p, ok := readingProviderFromRequest(w, r)
	if !ok {
		return
	}

	if err := r.ParseMultipartForm(8 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "expected a multipart form upload")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "no file was uploaded (expected a 'file' field)")
		return
	}
	defer file.Close()

	if header.Size > MaxImportBytes {
		writeError(w, http.StatusRequestEntityTooLarge,
			fmt.Sprintf("the file is too large (limit %d MB)", MaxImportBytes>>20))
		return
	}

	// LimitReader rather than trusting header.Size, which is client-supplied.
	data, err := io.ReadAll(io.LimitReader(file, MaxImportBytes+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, "could not read the uploaded file")
		return
	}
	if len(data) > MaxImportBytes {
		writeError(w, http.StatusRequestEntityTooLarge,
			fmt.Sprintf("the file is too large (limit %d MB)", MaxImportBytes>>20))
		return
	}

	doc, err := epub.BuildDocument(p, data)
	if err != nil {
		// A file the reader chose that turns out not to be a usable EPUB is
		// their mistake to correct, so the parser's own message is passed
		// through rather than flattened into "bad request".
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := p.Library.Add(doc); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	if err := libraryStore.Save(p.Code, doc.Meta.ID, data); err != nil {
		// The book is already importable; failing the request over a
		// persistence problem would be worse than reading it this session.
		log.Printf("library: could not persist %s: %v", doc.Meta.ID, err)
	}

	writeJSON(w, http.StatusCreated, toReadingTextDTO(doc.Meta, true))
}

// handleDeleteText removes an imported book. Bundled texts are not
// removable, which the Library enforces.
func handleDeleteText(w http.ResponseWriter, r *http.Request) {
	p, ok := readingProviderFromRequest(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if err := p.Library.Remove(id); err != nil {
		status := http.StatusNotFound
		if strings.Contains(err.Error(), "bundled") {
			status = http.StatusForbidden
		}
		writeError(w, status, err.Error())
		return
	}
	if err := libraryStore.Delete(p.Code, id); err != nil {
		log.Printf("library: could not delete %s from disk: %v", id, err)
	}
	w.WriteHeader(http.StatusNoContent)
}
