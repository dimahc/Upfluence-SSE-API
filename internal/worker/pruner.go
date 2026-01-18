package worker

import (
	"context"
	"log"
	"time"

	"github.com/dimahc/upfluence-sse-api/internal/ingestion"
)

// Pruner removes stale data periodically.
type Pruner struct {
	store    *ingestion.Store
	interval time.Duration
}

// NewPruner wires up a pruner.
func NewPruner(store *ingestion.Store, interval time.Duration) *Pruner {
	return &Pruner{store: store, interval: interval}
}

// Start runs until ctx is cancelled.
func (p *Pruner) Start(ctx context.Context) error {
	log.Printf("Pruner: Starting (interval=%v)", p.interval)
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Pruner: Shutting down")
			return ctx.Err()
		case <-ticker.C:
			if pruned := p.store.Prune(); pruned > 0 {
				log.Printf("Pruner: Removed %d stale buckets", pruned)
			}
		}
	}
}
