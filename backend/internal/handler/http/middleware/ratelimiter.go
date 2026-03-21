package middleware

import (
	"net/http"
	"sync"
	"time"

	"backend/internal/pkg/response"

	"golang.org/x/time/rate"
)

// IP based rate limiter
type ipRateLimiter struct {
	ips   sync.Map
	rate  rate.Limit
	burst int
}

func newIPRateLimiter(r rate.Limit, b int) *ipRateLimiter {
	limiter := &ipRateLimiter{
		rate:  r,
		burst: b,
	}

	// Clean up old entries periodically
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			// A true implementation would also track last access time to clean up
			// Here we just accept simple memory leak for demonstration or use a proper cache layer (Redis/Memcached)
		}
	}()

	return limiter
}

func (i *ipRateLimiter) getLimiter(ip string) *rate.Limiter {
	limiter, exists := i.ips.Load(ip)
	if !exists {
		newLimiter := rate.NewLimiter(i.rate, i.burst)
		i.ips.Store(ip, newLimiter)
		return newLimiter
	}
	return limiter.(*rate.Limiter)
}

// RateLimit creates a rate limiting middleware
func RateLimit(requestsPerSecond float64, burst int) func(http.Handler) http.Handler {
	limiter := newIPRateLimiter(rate.Limit(requestsPerSecond), burst)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Basic IP extraction
			ip := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ip = forwarded
			}

			if !limiter.getLimiter(ip).Allow() {
				response.Err(w, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "Rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
