package epub

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bogdanbojan/sermokey/backend/internal/reading"
)

// ErrWrongLanguage reports an EPUB imported into the wrong language.
type ErrWrongLanguage struct {
	Declared, Importing string
}

func (e *ErrWrongLanguage) Error() string {
	return fmt.Sprintf("this EPUB is in %s, but you are importing it into %s — "+
		"switch language first, or the text will be analysed with the wrong rules",
		e.Declared, e.Importing)
}

// checkLanguage refuses an import whose declared language is a different
// language the reader also supports.
//
// Analysing French prose with German rules produces nothing useful, and
// the resulting book would look broken rather than misfiled. The check is
// deliberately narrow: EPUB language metadata is often missing or wrong,
// so only a confident mismatch — a declared language that is itself a
// supported reading language — is refused. Anything absent or unrecognised
// is allowed through on the assumption that the reader knows better than
// the file does.
func checkLanguage(importing, declared string) error {
	if declared == "" || declared == importing {
		return nil
	}
	if _, supported := reading.For(declared); !supported {
		return nil
	}
	return &ErrWrongLanguage{Declared: declared, Importing: importing}
}

// Store persists imported EPUB files on disk, one directory per language.
//
// The original file is kept as the source of truth rather than a serialized
// Document. Parsing a book takes milliseconds, so re-reading it at startup
// costs nothing worth optimising, and it means a change to the tokenizer or
// the lexicon takes effect on books already imported — which, for a tool
// whose whole job is linguistic analysis, is the behaviour you want.
// Caching the analysis instead would quietly freeze every imported book at
// the version of the engine that first read it.
type Store struct {
	root string
}

// NewStore returns a store rooted at dir. An empty dir disables
// persistence, which is what the tests use and what keeps the server
// runnable in a read-only container.
func NewStore(dir string) *Store { return &Store{root: dir} }

// Enabled reports whether this store writes anything.
func (s *Store) Enabled() bool { return s != nil && s.root != "" }

func (s *Store) langDir(lang string) string { return filepath.Join(s.root, lang) }

// Save writes an imported file so it comes back after a restart.
func (s *Store) Save(lang, id string, data []byte) error {
	if !s.Enabled() {
		return nil
	}
	dir := s.langDir(lang)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// The ID is content-derived hex, so it is already safe as a filename;
	// the check guards against a future ID scheme quietly introducing a
	// path traversal.
	if strings.ContainsAny(id, `/\.`) {
		return fmt.Errorf("refusing to write unsafe id %q", id)
	}
	return os.WriteFile(filepath.Join(dir, id+".epub"), data, 0o644)
}

// Delete removes a persisted file. A file that is already gone is not an
// error: the caller's intent is satisfied either way.
func (s *Store) Delete(lang, id string) error {
	if !s.Enabled() {
		return nil
	}
	err := os.Remove(filepath.Join(s.langDir(lang), id+".epub"))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// LoadInto parses every persisted file for a language and adds it to the
// provider's library. It returns the number restored plus any per-file
// errors, rather than stopping at the first failure: one unreadable book
// should not cost the reader the rest of their library.
func (s *Store) LoadInto(p reading.Provider) (int, []error) {
	if !s.Enabled() {
		return 0, nil
	}
	entries, err := os.ReadDir(s.langDir(p.Code))
	if err != nil {
		if !os.IsNotExist(err) {
			return 0, []error{err}
		}
		return 0, nil
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".epub") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names) // deterministic restore order

	var errs []error
	var n int
	for _, name := range names {
		path := filepath.Join(s.langDir(p.Code), name)
		data, err := os.ReadFile(path)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
			continue
		}
		doc, err := BuildDocument(p, data)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
			continue
		}
		if err := p.Library.Add(doc); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
			continue
		}
		n++
	}
	return n, errs
}

// BuildDocument parses EPUB bytes and runs them through a provider's
// tokenizer and analyzer, producing a Document indistinguishable from a
// bundled text.
//
// This is the whole of the EPUB integration. Everything downstream — the
// section endpoint, the hover lexicon, status colouring, phrase selection,
// saving to the deck — was written against Document and needs no knowledge
// that books exist.
func BuildDocument(p reading.Provider, data []byte) (reading.Document, error) {
	book, err := Parse(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return reading.Document{}, err
	}
	if err := checkLanguage(p.Code, book.Language); err != nil {
		return reading.Document{}, err
	}
	book.Meta.ID = reading.DocumentID(data)
	book.Meta.Title = reading.SanitizeTitle(book.Meta.Title)
	return p.Build(book.Meta, book.Sections), nil
}
