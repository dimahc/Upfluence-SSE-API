package ingestion

import (
	"sync"
	"time"

	"github.com/dimahc/upfluence-sse-api/internal/model"
)

const (
	bucketGranularity = 5 * time.Second
	maxRetention      = 24 * time.Hour
)

// Store keeps posts in time buckets.
type Store struct {
	buckets map[int64]*bucket
	mu      sync.RWMutex
}

// NewStore initializes an empty store.
func NewStore() *Store {
	return &Store{buckets: make(map[int64]*bucket)}
}

// Add inserts a post into the current bucket.
func (s *Store) Add(p *model.Post) {
	if p == nil {
		return
	}
	now := time.Now().Unix()
	key := (now / 5) * 5

	s.mu.Lock()
	b, exists := s.buckets[key]
	if !exists {
		b = &bucket{posts: make([]*model.Post, 0)}
		s.buckets[key] = b
	}
	s.mu.Unlock()

	b.add(p)
}

// Query fetches posts within the given duration.
func (s *Store) Query(duration time.Duration) []*model.Post {
	now := time.Now().Unix()
	cutoff := now - int64(duration.Seconds())

	s.mu.RLock()
	defer s.mu.RUnlock()

	var posts []*model.Post
	for ts, b := range s.buckets {
		if ts >= cutoff {
			posts = append(posts, b.getPosts()...)
		}
	}
	return posts
}

// Prune deletes expired buckets.
func (s *Store) Prune() int {
	cutoff := time.Now().Unix() - int64(maxRetention.Seconds())

	s.mu.Lock()
	defer s.mu.Unlock()

	pruned := 0
	for ts := range s.buckets {
		if ts < cutoff {
			delete(s.buckets, ts)
			pruned++
		}
	}
	return pruned
}

// BucketCount returns active bucket count.
func (s *Store) BucketCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.buckets)
}

// TotalPosts returns total stored posts.
func (s *Store) TotalPosts() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	total := 0
	for _, b := range s.buckets {
		total += len(b.posts)
	}
	return total
}

// MinDuration is the smallest queryable window.
func (s *Store) MinDuration() time.Duration { return bucketGranularity }

// MaxDuration is the largest queryable window.
func (s *Store) MaxDuration() time.Duration { return maxRetention }

type bucket struct {
	posts []*model.Post
	mu    sync.RWMutex
}

func (b *bucket) add(p *model.Post) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.posts = append(b.posts, p)
}

func (b *bucket) getPosts() []*model.Post {
	b.mu.RLock()
	defer b.mu.RUnlock()
	result := make([]*model.Post, len(b.posts))
	copy(result, b.posts)
	return result
}
