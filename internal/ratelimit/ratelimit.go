package ratelimit

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	kept := rl.requests[key][:0]
	for _, t := range rl.requests[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}

	if len(kept) >= rl.limit {
		rl.requests[key] = kept
		return false
	}
	// Drop keys with no recent requests so the map doesn't grow forever
	// with every client that ever connected.
	for k, ts := range rl.requests {
		if k != key && (len(ts) == 0 || !ts[len(ts)-1].After(cutoff)) {
			delete(rl.requests, k)
		}
	}

	rl.requests[key] = append(kept, now)
	return true
}

// Middleware rejects requests over the limit with 429. Key is the client IP
// without the port, so new connections from the same client share a bucket.
// Behind a proxy (Render) this is the proxy's address; see server.md TODO.
func (rl *RateLimiter) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			key = r.RemoteAddr
		}
		if !rl.allow(key) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"too many requests, try again later"}`))
			return
		}
		next(w, r)
	}
}
