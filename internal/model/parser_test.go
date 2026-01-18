package model

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		name          string
		data          []byte
		wantErr       error
		wantTimestamp int64
		wantLikes     *int
		wantComments  *int
		wantRetweets  *int
		wantFavorites *int
	}{
		{
			name:          "instagram media",
			data:          []byte(`{"instagram_media":{"id":123,"likes":27,"comments":42,"timestamp":1234567890}}`),
			wantTimestamp: 1234567890,
			wantLikes:     intPtr(27),
			wantComments:  intPtr(42),
		},
		{
			name:          "tweet",
			data:          []byte(`{"tweet":{"id":123,"retweets":10,"favorites":25,"timestamp":1234567890}}`),
			wantTimestamp: 1234567890,
			wantRetweets:  intPtr(10),
			wantFavorites: intPtr(25),
		},
		{
			name:          "twitch stream",
			data:          []byte(`{"twitch_stream":{"timestamp":1768512121,"avg_viewers":324,"peak_viewers":325}}`),
			wantTimestamp: 1768512121,
		},
		{
			name:    "invalid json",
			data:    []byte(`{invalid}`),
			wantErr: errAny,
		},
		{
			name:    "empty object",
			data:    []byte(`{}`),
			wantErr: ErrNoPostData,
		},
		{
			name:    "missing timestamp",
			data:    []byte(`{"tweet":{"id":123}}`),
			wantErr: ErrMissingTimestamp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Parse(tt.data)

			if tt.wantErr != nil {
				if tt.wantErr == errAny {
					if err == nil {
						t.Error("expected error, got nil")
					}
				} else if err != tt.wantErr {
					t.Errorf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.Timestamp != tt.wantTimestamp {
				t.Errorf("Timestamp = %d, want %d", p.Timestamp, tt.wantTimestamp)
			}
			checkMetric(t, "Likes", p.Metrics.Likes, tt.wantLikes)
			checkMetric(t, "Comments", p.Metrics.Comments, tt.wantComments)
			checkMetric(t, "Retweets", p.Metrics.Retweets, tt.wantRetweets)
			checkMetric(t, "Favorites", p.Metrics.Favorites, tt.wantFavorites)
		})
	}
}

var errAny = &struct{ error }{}

func intPtr(v int) *int { return &v }

func checkMetric(t *testing.T, name string, got, want *int) {
	t.Helper()
	if want == nil {
		return
	}
	if got == nil {
		t.Errorf("%s = nil, want %d", name, *want)
		return
	}
	if *got != *want {
		t.Errorf("%s = %d, want %d", name, *got, *want)
	}
}

func TestMetrics_GetDimension(t *testing.T) {
	likes, comments := 100, 50
	m := Metrics{Likes: &likes, Comments: &comments}

	tests := []struct {
		dim    string
		want   int
		wantOk bool
	}{
		{"likes", 100, true},
		{"comments", 50, true},
		{"retweets", 0, false},
		{"invalid", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.dim, func(t *testing.T) {
			got, ok := m.GetDimension(tt.dim)
			if ok != tt.wantOk || got != tt.want {
				t.Errorf("GetDimension(%q) = %d, %v; want %d, %v", tt.dim, got, ok, tt.want, tt.wantOk)
			}
		})
	}
}
