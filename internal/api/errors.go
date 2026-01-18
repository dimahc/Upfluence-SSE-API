package api

import "errors"

// Validation errors.
var (
	ErrMissingDuration  = errors.New("missing required parameter: duration")
	ErrInvalidDuration  = errors.New("invalid duration format (use 5s, 30s, 5m, 1h, etc.)")
	ErrDurationTooShort = errors.New("duration too short (minimum: 5s)")
	ErrDurationTooLong  = errors.New("duration too long (maximum: 24h)")
	ErrMissingDimension = errors.New("missing required parameter: dimension")
	ErrInvalidDimension = errors.New("invalid dimension (allowed: likes, comments, favorites, retweets)")
)
