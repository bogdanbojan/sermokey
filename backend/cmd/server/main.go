// Command server runs the French grammar practice backend: a small JSON API
// over the grammar package, meant to be consumed by the local Vue frontend.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/bogdanbojan/sermokey/backend/internal/httpapi"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Building the router loads every language's lexicon, which churns
	// through a few hundred megabytes of decompression buffers. Hand the
	// pages back before serving rather than carrying them for the life of
	// the process.
	handler := httpapi.NewRouter()
	debug.FreeOSMemory()

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		// Read and write timeouts are deliberately generous rather than
		// 10s. They cover the whole body and the whole handler, and an
		// EPUB import is a 32 MB upload followed by minutes of
		// tokenization — so a 10s write timeout killed the connection
		// *after* the book had already been added to the library and
		// written to disk. The learner saw a failure, retried, and paid
		// for the work twice. Read-only routes are bounded individually by
		// httpapi.NewRouter instead.
		ReadTimeout:    5 * time.Minute,
		WriteTimeout:   5 * time.Minute,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
		ErrorLog:       log.Default(),
	}

	// A bind failure used to call log.Fatalf from inside a goroutine, which
	// calls os.Exit and skips the shutdown path entirely.
	serveErr := make(chan error, 1)
	go func() {
		log.Printf("sermokey backend listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	select {
	case err := <-serveErr:
		if err != nil {
			log.Printf("server error: %v", err)
			os.Exit(1)
		}
		return
	case <-stop:
	}
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	log.Println("server stopped")
}
