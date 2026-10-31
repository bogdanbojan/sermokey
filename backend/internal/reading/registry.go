package reading

import "sort"

// registry holds the Provider for each language that supports reading.
//
// This is deliberately separate from httpapi's languageProvider registry.
// Reading is available for a subset of the languages that have grammar and
// vocabulary — French first — and modelling that as "look it up, it might
// not be there" is honest, whereas bolting a Reading() method onto
// languageProvider would force es/de to return nil and push the real
// nil-check into every handler anyway.
var registry = map[string]Provider{}

// Register adds a language's reading provider. It is called from each
// language package's init, so importing the package is what enables it.
// Panics on a duplicate or incomplete registration: both are build-time
// wiring bugs, consistent with how the vocabulary CSV parser treats
// malformed data.
func Register(p Provider) {
	if p.Code == "" {
		panic("reading: Register called with empty language code")
	}
	if p.Tokenizer == nil || p.Analyzer == nil || p.Catalog == nil {
		panic("reading: Register(" + p.Code + ") with a nil Tokenizer, Analyzer or Catalog")
	}
	if _, dup := registry[p.Code]; dup {
		panic("reading: duplicate Register for language " + p.Code)
	}
	registry[p.Code] = p
}

// For returns the reading provider for a language code, if it has one.
func For(code string) (Provider, bool) {
	p, ok := registry[code]
	return p, ok
}

// Languages lists every language code with reading support, sorted so the
// output is stable across runs (Go map iteration order is not).
func Languages() []string {
	out := make([]string, 0, len(registry))
	for code := range registry {
		out = append(out, code)
	}
	sort.Strings(out)
	return out
}
