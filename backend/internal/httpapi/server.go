package httpapi

import (
	"log"
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
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("GET /api/languages", handleLanguages)
	mux.HandleFunc("GET /api/tenses", handleTenses)
	mux.HandleFunc("GET /api/theory/{id}", handleTheory)
	mux.HandleFunc("GET /api/verbs", handleVerbs)
	mux.HandleFunc("GET /api/exercises", handleExercises)
	mux.HandleFunc("GET /api/vocab-categories", handleVocabCategories)
	mux.HandleFunc("GET /api/vocabulary", handleVocabulary)
	mux.HandleFunc("GET /api/texts", handleTexts)
	mux.HandleFunc("GET /api/texts/{id}", handleTextSection)
	mux.HandleFunc("GET /api/lookup", handleLookup)
	// The only non-GET routes in the API. An imported book is content
	// rather than user state, and is far too large to live in the browser
	// alongside the decks; see reading_import.go.
	mux.HandleFunc("POST /api/texts/import", handleImportText)
	mux.HandleFunc("DELETE /api/texts/{id}", handleDeleteText)

	return withLogging(withCORS(mux))
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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
