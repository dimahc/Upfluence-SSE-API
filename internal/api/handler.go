package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dimahc/upfluence-sse-api/internal/model"
)

// AnalysisResponse pairs a result with its execution mode.
type AnalysisResponse struct {
	Result *model.Result
	Mode   string
}

// Analyzer computes stream statistics.
type Analyzer interface {
	Analyze(ctx context.Context, req *model.Request) (*AnalysisResponse, error)
	GetMinDuration() time.Duration
	GetMaxDuration() time.Duration
}

// Handler serves the /analysis endpoint.
type Handler struct {
	analyzer Analyzer
}

// NewHandler wires up a Handler.
func NewHandler(analyzer Analyzer) *Handler {
	return &Handler{analyzer: analyzer}
}

// AnalysisHandler handles GET /analysis.
func (h *Handler) AnalysisHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req, err := h.parseRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Processing analysis request (duration=%v, dimension=%s)", req.Duration, req.Dimension)
	response, err := h.analyzer.Analyze(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response.Result.ToJSON(req.Dimension)); err != nil {
		log.Printf("Failed to encode response: %v", err)
		return
	}
	log.Printf("Request completed: %d posts, duration=%v, mode=%s", response.Result.TotalPosts, req.Duration, response.Mode)
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	msg := err.Error()
	if msg == "no data collected during the specified duration" ||
		msg == "no data available for requested dimension and duration" {
		log.Printf("No data: %v", err)
		http.Error(w, msg, http.StatusNotFound)
		return
	}
	log.Printf("Service error: %v", err)
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}

func (h *Handler) parseRequest(r *http.Request) (*model.Request, error) {
	query := r.URL.Query()

	durationStr := query.Get("duration")
	if durationStr == "" {
		return nil, ErrMissingDuration
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil || duration <= 0 {
		return nil, ErrInvalidDuration
	}

	if duration < h.analyzer.GetMinDuration() {
		return nil, ErrDurationTooShort
	}
	if duration > h.analyzer.GetMaxDuration() {
		return nil, ErrDurationTooLong
	}

	dimension := query.Get("dimension")
	if dimension == "" {
		return nil, ErrMissingDimension
	}
	if !model.IsValidDimension(dimension) {
		return nil, ErrInvalidDimension
	}

	return &model.Request{Duration: duration, Dimension: dimension}, nil
}
