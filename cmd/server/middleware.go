package main

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rs/cors"
)

// newCORS creates a configured CORS handler for the given allowed origin.
func newCORS(allowedOrigin string) *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:       []string{allowedOrigin},
		AllowedMethods:       []string{"GET", "POST", "OPTIONS", "PUT", "DELETE"},
		AllowedHeaders:       []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Connect-Protocol-Version"},
		AllowCredentials:     true,
		OptionsSuccessStatus: http.StatusOK,
	})
}

// securityHeaders adds standard security response headers to every response.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("X-XSS-Protection", "0") // Disabled in favour of CSP
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=()")
		next.ServeHTTP(w, r)
	})
}

// ipLimiter holds per-IP token buckets.
type ipLimiter struct {
	mu       sync.Mutex
	limiters map[string]*tokenBucket
	rps      float64
	burst    int
}

type tokenBucket struct {
	tokens   float64
	maxBurst float64
	rps      float64
	lastTime time.Time
}

func (b *tokenBucket) allow() bool {
	now := time.Now()
	elapsed := now.Sub(b.lastTime).Seconds()
	b.lastTime = now
	b.tokens += elapsed * b.rps
	if b.tokens > b.maxBurst {
		b.tokens = b.maxBurst
	}
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func newIPLimiter(rps float64, burst int) *ipLimiter {
	return &ipLimiter{
		limiters: make(map[string]*tokenBucket),
		rps:      rps,
		burst:    burst,
	}
}

func (l *ipLimiter) getLimiter(ip string) *tokenBucket {
	l.mu.Lock()
	defer l.mu.Unlock()
	if b, ok := l.limiters[ip]; ok {
		return b
	}
	b := &tokenBucket{
		tokens:   float64(l.burst),
		maxBurst: float64(l.burst),
		rps:      l.rps,
		lastTime: time.Now(),
	}
	l.limiters[ip] = b
	return b
}

// rateLimiter returns a middleware that rate-limits by client IP.
func rateLimiter(rps float64, burst int) func(http.Handler) http.Handler {
	limiter := newIPLimiter(rps, burst)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			// Strip port: "1.2.3.4:5678" → "1.2.3.4"
			for i := len(ip) - 1; i >= 0; i-- {
				if ip[i] == ':' {
					ip = ip[:i]
					break
				}
			}
			if !limiter.getLimiter(ip).allow() {
				http.Error(w, `{"error":"too many requests"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// newRateLimiterFromEnv reads RATE_LIMIT_RPS and RATE_LIMIT_BURST env vars.
func newRateLimiterFromEnv() func(http.Handler) http.Handler {
	rps := 100.0
	burst := 200
	if v := os.Getenv("RATE_LIMIT_RPS"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			rps = f
		}
	}
	if v := os.Getenv("RATE_LIMIT_BURST"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			burst = n
		}
	}
	return rateLimiter(rps, burst)
}
