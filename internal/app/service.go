package app

import (
	"context"
	"errors"
	"time"

	"github.com/dimahc/upfluence-sse-api/internal/aggregation"
	"github.com/dimahc/upfluence-sse-api/internal/api"
	"github.com/dimahc/upfluence-sse-api/internal/ingestion"
	"github.com/dimahc/upfluence-sse-api/internal/model"
)

// Errors returned by Analyze.
var (
	ErrNoDataCollected = errors.New("no data collected during the specified duration")
	ErrNoDataAvailable = errors.New("no data available for requested dimension and duration")
)

const realtimeThreshold = 60 * time.Second

// Service handles analysis requests.
type Service struct {
	store     *ingestion.Store
	streamURL string
}

// NewService wires up a service.
func NewService(store *ingestion.Store, streamURL string) *Service {
	return &Service{store: store, streamURL: streamURL}
}

// Analyze uses realtime mode for durations ≤60s, historical otherwise.
func (s *Service) Analyze(ctx context.Context, req *model.Request) (*api.AnalysisResponse, error) {
	if req.Duration <= realtimeThreshold {
		return s.analyzeRealtime(ctx, req)
	}
	return s.analyzeHistorical(req)
}

func (s *Service) analyzeRealtime(parentCtx context.Context, req *model.Request) (*api.AnalysisResponse, error) {
	ctx, cancel := context.WithTimeout(parentCtx, req.Duration)
	defer cancel()

	collector := ingestion.NewCollector(s.streamURL)
	var posts []*model.Post

	err := collector.Collect(ctx, func(p *model.Post) {
		posts = append(posts, p)
	})
	if err != nil && err != context.DeadlineExceeded {
		return nil, err
	}
	if len(posts) == 0 {
		return nil, ErrNoDataCollected
	}

	agg := aggregation.Aggregate(posts, req.Dimension)
	return &api.AnalysisResponse{
		Result: &model.Result{
			TotalPosts:   agg.TotalPosts,
			MinTimestamp: agg.MinTimestamp,
			MaxTimestamp: agg.MaxTimestamp,
			P50:          agg.P50,
			P90:          agg.P90,
			P99:          agg.P99,
		},
		Mode: "REALTIME",
	}, nil
}

func (s *Service) analyzeHistorical(req *model.Request) (*api.AnalysisResponse, error) {
	posts := s.store.Query(req.Duration)
	if len(posts) == 0 {
		return nil, ErrNoDataAvailable
	}

	agg := aggregation.Aggregate(posts, req.Dimension)
	if agg.P50 == 0 && agg.P90 == 0 && agg.P99 == 0 && agg.TotalPosts > 0 {
		return nil, ErrNoDataAvailable
	}

	return &api.AnalysisResponse{
		Result: &model.Result{
			TotalPosts:   agg.TotalPosts,
			MinTimestamp: agg.MinTimestamp,
			MaxTimestamp: agg.MaxTimestamp,
			P50:          agg.P50,
			P90:          agg.P90,
			P99:          agg.P99,
		},
		Mode: "HISTORICAL",
	}, nil
}

// GetMinDuration reports min allowed duration.
func (s *Service) GetMinDuration() time.Duration { return s.store.MinDuration() }

// GetMaxDuration reports max allowed duration.
func (s *Service) GetMaxDuration() time.Duration { return s.store.MaxDuration() }
