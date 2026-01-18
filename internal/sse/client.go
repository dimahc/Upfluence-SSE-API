package sse

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ErrHTTPStatus signals a non-200 response.
var ErrHTTPStatus = errors.New("invalid HTTP status")

// Client connects to an SSE endpoint.
type Client struct {
	client  http.Client
	baseURL string
}

// NewClient creates an SSE client.
func NewClient(baseURL string) *Client {
	return &Client{
		client:  http.Client{Transport: &http.Transport{DisableKeepAlives: true}},
		baseURL: baseURL,
	}
}

// Consume streams events until ctx is done.
func (c *Client) Consume(ctx context.Context, messages chan<- []byte) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %d", ErrHTTPStatus, resp.StatusCode)
	}

	parser := NewParser(resp.Body)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		data, err := parser.NextEvent()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		select {
		case messages <- data:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
