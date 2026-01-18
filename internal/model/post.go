package model

import "slices"

// ValidDimensions enumerates allowed metrics.
var ValidDimensions = []string{"likes", "comments", "favorites", "retweets"}

// IsValidDimension checks if dim is allowed.
func IsValidDimension(dimension string) bool {
	return slices.Contains(ValidDimensions, dimension)
}

// Post is a single social media entry.
type Post struct {
	Timestamp int64
	Metrics   Metrics
}

// Metrics holds engagement metrics. Nil values indicate missing data.
type Metrics struct {
	Likes       *int
	Comments    *int
	Retweets    *int
	Favorites   *int
	Shares      *int
	Plays       *int
	Views       *int
	Saves       *int
	Repins      *int
	Dislikes    *int
	AvgViewers  *int
	PeakViewers *int
}

// GetDimension extracts a metric by name.
func (m *Metrics) GetDimension(dimension string) (int, bool) {
	switch dimension {
	case "likes":
		return derefInt(m.Likes)
	case "comments":
		return derefInt(m.Comments)
	case "retweets":
		return derefInt(m.Retweets)
	case "favorites":
		return derefInt(m.Favorites)
	case "shares":
		return derefInt(m.Shares)
	case "plays":
		return derefInt(m.Plays)
	case "views":
		return derefInt(m.Views)
	case "saves":
		return derefInt(m.Saves)
	case "repins":
		return derefInt(m.Repins)
	case "dislikes":
		return derefInt(m.Dislikes)
	default:
		return 0, false
	}
}

func derefInt(v *int) (int, bool) {
	if v == nil {
		return 0, false
	}
	return *v, true
}
