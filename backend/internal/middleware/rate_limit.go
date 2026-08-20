package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"starter-backend/internal/respond"
)

type rateEntry struct {
	count     int
	windowEnd time.Time
	lastSeen  time.Time
}

type RateLimiter struct {
	mu         sync.Mutex
	entries    map[string]*rateEntry
	limit      int
	trustProxy bool
	now        func() time.Time
}

func NewRateLimiter(limit int, trustProxy bool) *RateLimiter {
	return &RateLimiter{entries: make(map[string]*rateEntry), limit: limit, trustProxy: trustProxy, now: time.Now}
}

func (l *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := l.now()
		key := ClientIP(r, l.trustProxy)
		l.mu.Lock()
		entry, ok := l.entries[key]
		if !ok || !now.Before(entry.windowEnd) {
			entry = &rateEntry{windowEnd: now.Add(time.Minute)}
			l.entries[key] = entry
		}
		entry.count++
		entry.lastSeen = now
		count := entry.count
		reset := entry.windowEnd
		if len(l.entries) > 10000 {
			for candidate, value := range l.entries {
				if now.Sub(value.lastSeen) > 10*time.Minute {
					delete(l.entries, candidate)
				}
			}
		}
		l.mu.Unlock()

		remaining := l.limit - count
		if remaining < 0 {
			remaining = 0
		}
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(l.limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		if count > l.limit {
			retry := int(time.Until(reset).Seconds())
			if retry < 1 {
				retry = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(retry))
			respond.Error(w, r, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests; try again later", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}
