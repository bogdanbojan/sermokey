package httpapi

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// --- Response writing -------------------------------------------------------

// buildVersion identifies this binary's embedded content. Bundled texts,
// theory pages and vocabulary are compiled in, so the build is the only
// thing that can change them — which makes it a sound ETag component.
var buildVersion = sync.OnceValue(func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				return s.Value
			}
		}
	}
	return "dev"
})

// writeJSONCached marshals a response, serves a conditional 304 when the
// client already has it, and sets validators.
//
// Marshalling into a buffer before touching the header is also what makes
// the status honest: writeJSON commits the status first and then streams,
// so an encode failure part-way through a multi-megabyte section payload
// left a 200 with truncated JSON on the wire.
func writeJSONCached(w http.ResponseWriter, r *http.Request, status int, v any, cacheable bool) {
	body, err := json.Marshal(v)
	if err != nil {
		log.Printf("httpapi: failed to encode response: %v", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	h := w.Header()
	h.Set("Content-Type", "application/json; charset=utf-8")
	setVary(h)

	if cacheable && status == http.StatusOK {
		sum := sha256.Sum256(body)
		etag := `"` + hex.EncodeToString(sum[:16]) + `"`
		h.Set("ETag", etag)
		// Content is immutable for a given build: it is compiled into the
		// binary. Revalidation is still required because an imported book
		// can be deleted, so max-age stays short and the ETag does the work.
		h.Set("Cache-Control", "private, max-age=60, must-revalidate")
		if match := r.Header.Get("If-None-Match"); match != "" && strings.Contains(match, etag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	// Section payloads are verbatim prose plus a flat integer array plus a
	// gloss table — JSON that compresses roughly 5-10x, re-fetched on every
	// navigation.
	if gzipAccepted(r) && len(body) > 1024 {
		h.Set("Content-Encoding", "gzip")
		w.WriteHeader(status)
		zw := gzip.NewWriter(w)
		defer zw.Close()
		if _, err := zw.Write(body); err != nil {
			log.Printf("httpapi: failed to write gzip response: %v", err)
		}
		return
	}

	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		log.Printf("httpapi: failed to write response: %v", err)
	}
}

// setVary declares everything that changes the representation.
//
// It is set here rather than split between the CORS middleware and the
// encoder because http.TimeoutHandler copies the inner handler's headers
// over the outer ones with `dst[k] = vv` — a whole-slice assignment, not an
// append. Any header written by both layers therefore loses the outer
// value entirely, which silently dropped "Vary: Origin" the moment the
// encoder added "Vary: Accept-Encoding".
func setVary(h http.Header) {
	h.Set("Vary", "Accept-Encoding, Origin")
}

func gzipAccepted(r *http.Request) bool {
	for _, enc := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		if strings.TrimSpace(strings.SplitN(enc, ";", 2)[0]) == "gzip" {
			return true
		}
	}
	return false
}

// --- Middleware -------------------------------------------------------------

// withRecovery turns a panic into a 500 with a JSON body.
//
// The codebase panics deliberately in several places (registry, catalog,
// lexicon loading) on invariants that should hold at startup, and there are
// reachable index-risk sites besides. net/http recovers per connection, so
// the process survives either way — but without this the client gets an
// aborted response rather than an error, and withLogging never records the
// request at all because the handler never returns.
func withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				// A panic after the client disconnected is not a bug worth
				// a stack trace.
				if rec == http.ErrAbortHandler {
					panic(rec)
				}
				log.Printf("httpapi: panic serving %s %q: %v\n%s",
					r.Method, r.URL.Path, rec, debug.Stack())
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"internal server error"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// statusRecorder captures what was actually sent, so the access log can
// report it.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	n, err := s.ResponseWriter.Write(b)
	s.bytes += n
	return n, err
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		if rec.status == 0 {
			rec.status = http.StatusOK
		}
		// %q on the path: it is attacker-controlled and was being written
		// to the log unescaped, so a request to "/api/%0a%0afake" injected
		// newlines into the log stream.
		log.Printf("%s %q %d %dB %s", r.Method, r.URL.Path, rec.status, rec.bytes, time.Since(start))
	})
}

// --- Concurrency limiting ---------------------------------------------------

// importSlots bounds concurrent imports. Each one holds the whole upload in
// memory plus a fully built Document, and spends minutes of CPU tokenizing
// and analyzing; letting them run in parallel multiplies both.
var importSlots = make(chan struct{}, 2)

func acquireImportSlot(r *http.Request) bool {
	select {
	case importSlots <- struct{}{}:
		return true
	case <-r.Context().Done():
		return false
	}
}

func releaseImportSlot() { <-importSlots }
