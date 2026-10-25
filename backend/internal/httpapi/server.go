package httpapi

import (
	"log"
	"net/http"
	"os"
	"time"
)

// NewRouter builds the complete HTTP handler for the backend: routes,
// request logging, and CORS for the local Vite dev server.
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("GET /api/languages", handleLanguages)
	mux.HandleFunc("GET /api/tenses", handleTenses)
	mux.HandleFunc("GET /api/theory/{id}", handleTheory)
	mux.HandleFunc("GET /api/verbs", handleVerbs)
	mux.HandleFunc("GET /api/exercises", handleExercises)
	mux.HandleFunc("GET /api/vocab-categories", handleVocabCategories)
	mux.HandleFunc("GET /api/vocabulary", handleVocabulary)

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
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
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
