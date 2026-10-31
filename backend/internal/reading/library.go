package reading

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
)

// Library is a Catalog that layers imported documents over a fixed
// bundled corpus.
//
// Every registered language gets one automatically (see Register), which is
// what makes "import a book" a capability of the reading system rather than
// of any one language: a language only has to supply a tokenizer, an
// analyzer and its bundled texts, and import works for it immediately.
//
// It is safe for concurrent use. Reads vastly outnumber writes — a request
// per section against the occasional upload — so a plain RWMutex is the
// right shape, and the documents it hands out are already-built values that
// callers never mutate.
type Library struct {
	bundled Catalog

	mu       sync.RWMutex
	imported map[string]Document
	order    []string // import order, newest last
}

// NewLibrary wraps a bundled catalog so documents can be added to it.
func NewLibrary(bundled Catalog) *Library {
	return &Library{bundled: bundled, imported: map[string]Document{}}
}

// List implements Catalog. Imported books come first: they are what the
// reader just added and what they are most likely to be returning to,
// whereas the bundled corpus is static and always findable below.
func (l *Library) List() []DocumentMeta {
	l.mu.RLock()
	defer l.mu.RUnlock()

	out := make([]DocumentMeta, 0, len(l.order))
	for i := len(l.order) - 1; i >= 0; i-- {
		out = append(out, l.imported[l.order[i]].Meta)
	}
	return append(out, l.bundled.List()...)
}

// Get implements Catalog.
func (l *Library) Get(id string) (Document, bool) {
	l.mu.RLock()
	doc, ok := l.imported[id]
	l.mu.RUnlock()
	if ok {
		return doc, true
	}
	return l.bundled.Get(id)
}

// Add stores an imported document, replacing any previous import with the
// same ID. Re-importing the same file is therefore idempotent rather than
// a way to accumulate duplicates — see DocumentID.
func (l *Library) Add(doc Document) error {
	if doc.Meta.ID == "" {
		return fmt.Errorf("cannot import a document with no ID")
	}
	if _, clash := l.bundled.Get(doc.Meta.ID); clash {
		return fmt.Errorf("%q collides with a bundled text", doc.Meta.ID)
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if _, exists := l.imported[doc.Meta.ID]; !exists {
		l.order = append(l.order, doc.Meta.ID)
	}
	l.imported[doc.Meta.ID] = doc
	return nil
}

// Remove deletes an imported document. Bundled texts cannot be removed,
// and asking to remove one is reported rather than silently ignored.
func (l *Library) Remove(id string) error {
	if _, bundled := l.bundled.Get(id); bundled {
		return fmt.Errorf("%q is a bundled text and cannot be removed", id)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.imported[id]; !ok {
		return fmt.Errorf("unknown text: %s", id)
	}
	delete(l.imported, id)
	for i, existing := range l.order {
		if existing == id {
			l.order = append(l.order[:i], l.order[i+1:]...)
			break
		}
	}
	return nil
}

// IsImported reports whether an ID belongs to an imported document rather
// than the bundled corpus. The UI uses it to decide whether to offer a
// delete control.
func (l *Library) IsImported(id string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	_, ok := l.imported[id]
	return ok
}

// DocumentID derives a stable ID for an imported document from the bytes of
// the file it came from.
//
// Content-addressing rather than a counter or a timestamp means importing
// the same book twice updates one entry instead of creating a second copy
// with its own review history — and it means the ID survives a restart,
// which matters because reading positions and saved words reference it.
func DocumentID(fileBytes []byte) string {
	sum := sha256.Sum256(fileBytes)
	return "import-" + hex.EncodeToString(sum[:8])
}

// SanitizeTitle trims an imported title to something a card can show.
func SanitizeTitle(title string) string {
	title = strings.Join(strings.Fields(title), " ")
	const max = 120
	if r := []rune(title); len(r) > max {
		return strings.TrimSpace(string(r[:max])) + "…"
	}
	return title
}
