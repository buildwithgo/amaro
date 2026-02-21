package middlewares

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/buildwithgo/amaro"
)

type rateLimiter struct {
	rate      float64 // tokens per second
	burst     int
	tokens    float64
	lastCheck time.Time
	mu        sync.Mutex
}

// Allow checks if a token is available
func (l *rateLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastCheck).Seconds()
	l.lastCheck = now

	// Refill
	l.tokens += elapsed * l.rate
	if l.tokens > float64(l.burst) {
		l.tokens = float64(l.burst)
	}

	// Consume
	if l.tokens >= 1.0 {
		l.tokens -= 1.0
		return true
	}
	return false
}

// RateLimiter implements a simple token bucket rate limiter.
// In a real scenario, you'd map IP addresses to limiters.
// This example limits globally or per IP if extended map logic added.
// For "all possible middlewares", let's make it a simple IP-based limiter.
func RateLimiter(requestsPerSecond float64, burst int) amaro.Middleware {
	type client struct {
		limiter  *rateLimiter
		lastSeen atomic.Int64 // UnixNano
	}

	var mu sync.RWMutex
	clients := make(map[string]*client)

	// Cleanup routine (leak prevention) - strictly primitive
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			limit := int64(3 * time.Minute)

			// Snapshot expired keys
			mu.RLock()
			var toDelete []string
			now := time.Now().UnixNano()
			for ip, c := range clients {
				if now - c.lastSeen.Load() > limit {
					toDelete = append(toDelete, ip)
				}
			}
			mu.RUnlock()

			if len(toDelete) > 0 {
				mu.Lock()
				now = time.Now().UnixNano()
				for _, ip := range toDelete {
					if c, ok := clients[ip]; ok {
						if now - c.lastSeen.Load() > limit {
							delete(clients, ip)
						}
					}
				}
				mu.Unlock()
			}
		}
	}()

	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
			ip := c.Request.RemoteAddr
			// Simplified IP matching

			mu.RLock()
			cli, exists := clients[ip]
			mu.RUnlock()

			if !exists {
				mu.Lock()
				cli, exists = clients[ip]
				if !exists {
					cli = &client{
						limiter: &rateLimiter{
							rate:      requestsPerSecond,
							burst:     burst,
							tokens:    float64(burst),
							lastCheck: time.Now(),
						},
					}
					cli.lastSeen.Store(time.Now().UnixNano())
					clients[ip] = cli
				}
				mu.Unlock()
			}

			cli.lastSeen.Store(time.Now().UnixNano())

			if !cli.limiter.Allow() {
				c.String(http.StatusTooManyRequests, "Too Many Requests")
				return nil
			}

			return next(c)
		}
	}
}
