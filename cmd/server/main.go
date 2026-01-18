package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dimahc/upfluence-sse-api/internal/api"
	"github.com/dimahc/upfluence-sse-api/internal/app"
	"github.com/dimahc/upfluence-sse-api/internal/ingestion"
	"github.com/dimahc/upfluence-sse-api/internal/worker"
)

const (
	defaultAddr     = ":8080"
	streamURL       = "https://stream.upfluence.co/stream"
	pruneInterval   = time.Minute
	shutdownTimeout = 10 * time.Second
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = defaultAddr
	}

	log.Println("Starting Upfluence SSE API server")
	log.Printf("HTTP Address: %s", addr)

	store := ingestion.NewStore()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	collector := ingestion.NewCollector(streamURL)
	w := worker.NewWorker(collector, store)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := w.Start(ctx); err != nil && err != context.Canceled {
			log.Fatalf("Worker failed: %v", err)
		}
	}()

	pruner := worker.NewPruner(store, pruneInterval)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := pruner.Start(ctx); err != nil && err != context.Canceled {
			log.Fatalf("Pruner failed: %v", err)
		}
	}()

	service := app.NewService(store, streamURL)
	handler := api.NewHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("/analysis", handler.AnalysisHandler)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server started on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-sigCh
	log.Println("Shutting down...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	wg.Wait()
	log.Println("Shutdown complete")
}
