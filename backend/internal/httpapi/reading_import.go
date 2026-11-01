package httpapi

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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

	// Serialised against other imports; see importSlots.
	if !acquireImportSlot(r) {
		return
	}
	defer releaseImportSlot()

	// ParseMultipartForm's argument is maxMemory, *not* a body cap:
	// everything above it is spooled to temporary files on disk, with no
	// limit, and the size check below only runs once that has finished. An
	// unauthenticated client could therefore write an arbitrary number of
	// bytes to the server's disk. MaxBytesReader is the actual cap.
	r.Body = http.MaxBytesReader(w, r.Body, MaxImportBytes+(1<<20))

	if err := r.ParseMultipartForm(8 << 20); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge,
				fmt.Sprintf("the file is too large (limit %d MB)", MaxImportBytes>>20))
			return
		}
		writeError(w, http.StatusBadRequest, "expected a multipart form upload")
		return
	}
	// ParseMultipartForm may have spooled part of the body to disk.
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()
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

	// Defence in depth behind MaxBytesReader. (header.Size is computed by
	// ReadForm from the bytes it actually read, so it is not in itself
	// client-controlled — the original comment here claimed otherwise.)
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

	// Parsing, tokenizing and analyzing a full-length book is minutes of
	// CPU. Without a check the work continues after the client has gone.
	if err := r.Context().Err(); err != nil {
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
		// Sentinel errors, not substring matching on the message. The
		// previous `strings.Contains(err.Error(), "bundled")` also matched
		// the *not found* message, which interpolates the caller's id — so
		// DELETE /api/texts/bundled returned 403 instead of 404.
		status := http.StatusNotFound
		if errors.Is(err, reading.ErrBundledText) {
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
