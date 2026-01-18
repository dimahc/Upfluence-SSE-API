package aggregation

import (
	"testing"

	"github.com/dimahc/upfluence-sse-api/internal/model"
)

func TestAggregate(t *testing.T) {
	likes100 := 100
	likes50 := 50

	tests := []struct {
		name      string
		posts     []*model.Post
		dimension string
		wantTotal int
		wantP50   int
		wantMinTS int64
		wantMaxTS int64
		p50Range  [2]int // [min, max] for approximate checks, ignored if both 0
	}{
		{
			name:      "empty slice",
			posts:     nil,
			dimension: "likes",
			wantTotal: 0,
			wantP50:   0,
		},
		{
			name: "single post",
			posts: []*model.Post{
				{Timestamp: 1000, Metrics: model.Metrics{Likes: &likes100}},
			},
			dimension: "likes",
			wantTotal: 1,
			wantP50:   100,
			wantMinTS: 1000,
			wantMaxTS: 1000,
		},
		{
			name: "missing dimension",
			posts: []*model.Post{
				{Timestamp: 1000, Metrics: model.Metrics{Likes: &likes100}},
			},
			dimension: "comments",
			wantTotal: 1,
			wantP50:   0,
		},
		{
			name: "nil posts filtered",
			posts: []*model.Post{
				nil,
				{Timestamp: 1000, Metrics: model.Metrics{Likes: &likes50}},
				nil,
			},
			dimension: "likes",
			wantTotal: 3,
			wantP50:   50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Aggregate(tt.posts, tt.dimension)

			if result.TotalPosts != tt.wantTotal {
				t.Errorf("TotalPosts = %d, want %d", result.TotalPosts, tt.wantTotal)
			}
			if tt.p50Range[0] == 0 && tt.p50Range[1] == 0 {
				if result.P50 != tt.wantP50 {
					t.Errorf("P50 = %d, want %d", result.P50, tt.wantP50)
				}
			}
			if tt.wantMinTS != 0 && result.MinTimestamp != tt.wantMinTS {
				t.Errorf("MinTimestamp = %d, want %d", result.MinTimestamp, tt.wantMinTS)
			}
			if tt.wantMaxTS != 0 && result.MaxTimestamp != tt.wantMaxTS {
				t.Errorf("MaxTimestamp = %d, want %d", result.MaxTimestamp, tt.wantMaxTS)
			}
		})
	}
}

func TestAggregate_Percentiles(t *testing.T) {
	// 100 posts with values 1-100
	posts := make([]*model.Post, 100)
	for i := 0; i < 100; i++ {
		likes := i + 1
		posts[i] = &model.Post{
			Timestamp: int64(1000 + i),
			Metrics:   model.Metrics{Likes: &likes},
		}
	}

	result := Aggregate(posts, "likes")

	if result.TotalPosts != 100 {
		t.Errorf("TotalPosts = %d, want 100", result.TotalPosts)
	}
	if result.P50 < 45 || result.P50 > 55 {
		t.Errorf("P50 = %d, want ~50", result.P50)
	}
	if result.P90 < 85 || result.P90 > 95 {
		t.Errorf("P90 = %d, want ~90", result.P90)
	}
	if result.P99 < 95 {
		t.Errorf("P99 = %d, want >= 95", result.P99)
	}
	if result.MinTimestamp != 1000 || result.MaxTimestamp != 1099 {
		t.Errorf("timestamps = %d-%d, want 1000-1099", result.MinTimestamp, result.MaxTimestamp)
	}
}
