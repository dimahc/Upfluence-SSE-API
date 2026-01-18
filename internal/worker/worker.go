package worker

import (
	"context"
	"log"
	"time"

	"github.com/dimahc/upfluence-sse-api/internal/ingestion"
	"github.com/dimahc/upfluence-sse-api/internal/model"
)

// Worker ingests SSE events into the store.
type Worker struct {
	collector *ingestion.Collector
	store     *ingestion.Store
}

// NewWorker wires up a worker.
func NewWorker(collector *ingestion.Collector, store *ingestion.Store) *Worker {
	return &Worker{collector: collector, store: store}
}

// Start runs until ctx is cancelled.
func (w *Worker) Start(ctx context.Context) error {
	log.Println("Worker: Starting SSE stream collection")
	for {
		select {
		case <-ctx.Done():
			log.Println("Worker: Shutting down")
			return ctx.Err()
		default:
			if err := w.collect(ctx); err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				log.Printf("Worker: Stream error: %v, reconnecting in 5s...", err)
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (w *Worker) collect(ctx context.Context) error {
	count := 0
	return w.collector.Collect(ctx, func(p *model.Post) {
		w.store.Add(p)
		count++
		if count%100 == 0 {
			log.Printf("Worker: Processed %d posts | Buckets: %d | Total: %d",
				count, w.store.BucketCount(), w.store.TotalPosts())
		}
	})
}
