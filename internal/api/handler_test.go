package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dimahc/upfluence-sse-api/internal/model"
)

type mockAnalyzer struct {
	response    *AnalysisResponse
	err         error
	minDuration time.Duration
	maxDuration time.Duration
}

func (m *mockAnalyzer) Analyze(ctx context.Context, req *model.Request) (*AnalysisResponse, error) {
	return m.response, m.err
}

func (m *mockAnalyzer) GetMinDuration() time.Duration { return m.minDuration }
func (m *mockAnalyzer) GetMaxDuration() time.Duration { return m.maxDuration }

type errNoData struct{}

func (e errNoData) Error() string { return "no data collected during the specified duration" }

func TestAnalysisHandler(t *testing.T) {
	successResponse := &AnalysisResponse{
		Result: &model.Result{
			TotalPosts:   10,
			MinTimestamp: 1000,
			MaxTimestamp: 2000,
			P50:          50,
			P90:          90,
			P99:          99,
		},
		Mode: "HISTORICAL",
	}

	tests := []struct {
		name       string
		method     string
		url        string
		response   *AnalysisResponse
		err        error
		wantStatus int
		checkBody  bool
	}{
		{
			name:       "success",
			method:     "GET",
			url:        "/analysis?duration=5m&dimension=likes",
			response:   successResponse,
			wantStatus: http.StatusOK,
			checkBody:  true,
		},
		{
			name:       "missing duration",
			method:     "GET",
			url:        "/analysis?dimension=likes",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing dimension",
			method:     "GET",
			url:        "/analysis?duration=5m",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid dimension",
			method:     "GET",
			url:        "/analysis?duration=5m&dimension=invalid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "duration too short",
			method:     "GET",
			url:        "/analysis?duration=1s&dimension=likes",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "duration too long",
			method:     "GET",
			url:        "/analysis?duration=48h&dimension=likes",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "no data",
			method:     "GET",
			url:        "/analysis?duration=5m&dimension=likes",
			err:        errNoData{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "method not allowed",
			method:     "POST",
			url:        "/analysis?duration=5m&dimension=likes",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := &mockAnalyzer{
				response:    tt.response,
				err:         tt.err,
				minDuration: 5 * time.Second,
				maxDuration: 24 * time.Hour,
			}
			handler := NewHandler(analyzer)

			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			handler.AnalysisHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.checkBody {
				var result map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if result["total_posts"] == nil {
					t.Error("expected total_posts in response")
				}
			}
		})
	}
}

func TestIsValidDimension(t *testing.T) {
	tests := []struct {
		dimension string
		want      bool
	}{
		{"likes", true},
		{"comments", true},
		{"favorites", true},
		{"retweets", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.dimension, func(t *testing.T) {
			if got := model.IsValidDimension(tt.dimension); got != tt.want {
				t.Errorf("IsValidDimension(%q) = %v, want %v", tt.dimension, got, tt.want)
			}
		})
	}
}
