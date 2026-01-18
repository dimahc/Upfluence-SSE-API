package ingestion

import (
	"context"
	"log"

	"github.com/dimahc/upfluence-sse-api/internal/model"
	"github.com/dimahc/upfluence-sse-api/internal/sse"
)

// Collector reads from an SSE stream.
type Collector struct {
	streamURL string
}

// NewCollector sets up a collector.
func NewCollector(streamURL string) *Collector {
	return &Collector{streamURL: streamURL}
}

// Collect processes events until ctx is done.
func (c *Collector) Collect(ctx context.Context, handler func(*model.Post)) error {
	client := sse.NewClient(c.streamURL)
	messages := make(chan []byte, 100)

	go func() {
		if err := client.Consume(ctx, messages); err != nil {
			if ctx.Err() == nil {
				log.Printf("Ingestion: stream error: %v", err)
			}
		}
		close(messages)
	}()

	for {
		select {
		case msg, ok := <-messages:
			if !ok {
				return nil
			}
			post, err := model.Parse(msg)
			if err != nil {
				continue
			}
			handler(post)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
