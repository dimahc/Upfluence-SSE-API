package model

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Parse errors.
var (
	ErrNoPostData       = errors.New("no post data found")
	ErrInvalidFormat    = errors.New("invalid format")
	ErrMissingTimestamp = errors.New("missing timestamp")
)

// Parse decodes SSE JSON into a Post.
func Parse(data []byte) (*Post, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidFormat, err)
	}
	if len(raw) != 1 {
		return nil, ErrNoPostData
	}
	for _, content := range raw {
		return parseContent(content)
	}
	return nil, ErrNoPostData
}

func parseContent(content json.RawMessage) (*Post, error) {
	var data struct {
		Timestamp   int64 `json:"timestamp"`
		Likes       *int  `json:"likes"`
		Comments    *int  `json:"comments"`
		Retweets    *int  `json:"retweets"`
		Favorites   *int  `json:"favorites"`
		Shares      *int  `json:"shares"`
		Plays       *int  `json:"plays"`
		Views       *int  `json:"views"`
		Saves       *int  `json:"saves"`
		Repins      *int  `json:"repins"`
		Dislikes    *int  `json:"dislikes"`
		AvgViewers  *int  `json:"avg_viewers"`
		PeakViewers *int  `json:"peak_viewers"`
	}
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidFormat, err)
	}
	if data.Timestamp == 0 {
		return nil, ErrMissingTimestamp
	}
	return &Post{
		Timestamp: data.Timestamp,
		Metrics: Metrics{
			Likes:       data.Likes,
			Comments:    data.Comments,
			Retweets:    data.Retweets,
			Favorites:   data.Favorites,
			Shares:      data.Shares,
			Plays:       data.Plays,
			Views:       data.Views,
			Saves:       data.Saves,
			Repins:      data.Repins,
			Dislikes:    data.Dislikes,
			AvgViewers:  data.AvgViewers,
			PeakViewers: data.PeakViewers,
		},
	}, nil
}
