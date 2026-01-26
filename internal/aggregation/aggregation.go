package aggregation

import (
	"sort"

	"github.com/dimahc/upfluence-sse-api/internal/model"
)

// Result holds computed percentiles.
type Result struct {
	TotalPosts   int
	MinTimestamp int64
	MaxTimestamp int64
	P50          int
	P90          int
	P99          int
}

// Aggregate computes p50/p90/p99 for a dimension.
func Aggregate(posts []*model.Post, dimension string) Result {
	if len(posts) == 0 {
		return Result{}
	}

	var values []int
	var minTS, maxTS int64
	first := true

	for _, p := range posts {
		if p == nil {
			continue
		}
		if first {
			minTS, maxTS = p.Timestamp, p.Timestamp
			first = false
		} else {
			if p.Timestamp < minTS {
				minTS = p.Timestamp
			}
			if p.Timestamp > maxTS {
				maxTS = p.Timestamp
			}
		}
		if val, ok := p.Metrics.GetDimension(dimension); ok {
			values = append(values, val)
		}
	}

	if len(values) == 0 {
		return Result{TotalPosts: len(posts), MinTimestamp: minTS, MaxTimestamp: maxTS}
	}

	sort.Ints(values)
	return Result{
		TotalPosts:   len(posts),
		MinTimestamp: minTS,
		MaxTimestamp: maxTS,
		P50:          percentile(values, 50),
		P90:          percentile(values, 90),
		P99:          percentile(values, 99),
	}
}

func percentile(sorted []int, p int) int {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return sorted[0]
	}
	rank := (p * (n - 1)) / 100
	if rank >= n {
		rank = n - 1
	}
	return sorted[rank]
}
