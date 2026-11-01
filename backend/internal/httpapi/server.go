package httpapi

import (
	"net/http"
	"os"
	"time"

	// Importing a language's reading package is what registers it in
	// reading.Registry and therefore what makes /api/texts serve it.
	_ "github.com/bogdanbojan/sermokey/backend/internal/reading/de"
	_ "github.com/bogdanbojan/sermokey/backend/internal/reading/es"
	_ "github.com/bogdanbojan/sermokey/backend/internal/reading/fr"
)

// NewRouter builds the complete HTTP handler for the backend: routes,
// request logging, and CORS for the local Vite dev server.
func NewRouter() http.Handler {
	// Bring back any EPUBs imported in a previous run, before the first
	// request can ask for them.
	RestoreLibrary()

	mux := http.NewServeMux()

	// Read-only routes are bounded individually rather than by the
	// server-wide WriteTimeout, which has to stay long enough for an
	// import to finish.
	const readTimeout = 30 * time.Second
	get := func(pattern string, h http.HandlerFunc) {
		mux.Handle(pattern, http.TimeoutHandler(h, readTimeout, `{"error":"request timed out"}`))
	}

	get("GET /api/health", handleHealth)
	get("GET /api/languages", handleLanguages)
	get("GET /api/tenses", handleTenses)
	get("GET /api/theory/{id}", handleTheory)
	get("GET /api/verbs", handleVerbs)
	get("GET /api/exercises", handleExercises)
	get("GET /api/vocab-categories", handleVocabCategories)
	get("GET /api/vocabulary", handleVocabulary)
	get("GET /api/texts", handleTexts)
	get("GET /api/texts/{id}", handleTextSection)
	get("GET /api/lookup", handleLookup)
	// The only non-GET routes in the API. An imported book is content
	// rather than user state, and is far too large to live in the browser
	// alongside the decks; see reading_import.go.
	mux.HandleFunc("POST /api/texts/import", handleImportText)
	mux.HandleFunc("DELETE /api/texts/{id}", handleDeleteText)

	// Recovery sits inside logging so a panicking request is still logged
	// with its status, and inside CORS so the error response carries the
	// headers the browser needs in order to surface it at all.
	return withLogging(withCORS(withRecovery(mux)))
}

// withCORS allows the configured frontend origin (default: the Vite dev
// server) to call the API from the browser.
func withCORS(next http.Handler) http.Handler {
	allowedOrigin := os.Getenv("FRONTEND_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5173"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, If-None-Match")
		w.Header().Set("Access-Control-Expose-Headers", "ETag")
		// Vary is declared by the response encoder, not here: see setVary
		// for why splitting it across the two layers does not survive
		// http.TimeoutHandler.
		if r.Method == http.MethodOptions {
			w.Header().Set("Vary", "Accept-Encoding, Origin")
			// Without this every cross-origin request pays a preflight
			// round trip.
			w.Header().Set("Access-Control-Max-Age", "600")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
