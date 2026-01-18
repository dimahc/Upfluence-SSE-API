package model

import "time"

// Request captures analysis params.
type Request struct {
	Duration  time.Duration
	Dimension string
}

// Result holds analysis output.
type Result struct {
	TotalPosts   int
	MinTimestamp int64
	MaxTimestamp int64
	P50          int
	P90          int
	P99          int
}

// ToJSON formats for HTTP response.
func (r *Result) ToJSON(dimension string) map[string]interface{} {
	return map[string]interface{}{
		"total_posts":       r.TotalPosts,
		"minimum_timestamp": r.MinTimestamp,
		"maximum_timestamp": r.MaxTimestamp,
		dimension + "_p50":  r.P50,
		dimension + "_p90":  r.P90,
		dimension + "_p99":  r.P99,
	}
}
