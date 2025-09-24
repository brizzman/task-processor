package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

const (
    rateLimitExceeded = "rate_limit_exceeded"
    contentTypeJSON   = "application/json"
)

type RedisRateLimiter struct {
	limiter *redis_rate.Limiter
	rps     int
	burst   int
	period  time.Duration
}

func NewRedisRateLimiter(redisClient *redis.Client, rps int, burst int, period time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{
		limiter: redis_rate.NewLimiter(redisClient),
		rps:     rps,
		burst:   burst,
		period:  period,
	}
}

func (r *RedisRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Get key for rate limiting (IP or user ID)
		key := getRateLimitKey(req)
		
		// Check the rate limit
		res, err := r.limiter.Allow(req.Context(), key, redis_rate.Limit{
			Rate:   r.rps,
			Burst:  r.burst,
			Period: r.period,
		})

		if err != nil {
			// On Redis error, let the request pass (fail open)
			next.ServeHTTP(w, req)
			return
		}

		// Rate limit exceeded
		if res.Allowed == 0 {
			w.Header().Set("Content-Type", contentTypeJSON)
			w.Header().Set("Retry-After", strconv.Itoa(int(res.RetryAfter.Seconds())))
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(r.rps))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
			w.Header().Set("X-RateLimit-Reset", time.Now().Add(res.RetryAfter).Format(time.RFC1123))
			w.WriteHeader(http.StatusTooManyRequests)

			if err := json.NewEncoder(w).Encode(map[string]string{
				"error": rateLimitExceeded,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			
			return
		}

		// Add headers with rate limit information
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(r.rps))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
		w.Header().Set("X-RateLimit-Reset", time.Now().Add(res.RetryAfter).Format(time.RFC1123))

		next.ServeHTTP(w, req)
	})
}

// getRateLimitKey determines the key for rate limiting
func getRateLimitKey(req *http.Request) string {
	// 1. By IP (default)
    host, _, err := net.SplitHostPort(req.RemoteAddr)
    if err != nil {
        host = req.RemoteAddr 
    }
	
	// 2. By API key (if provided)
    if apiKey := req.Header.Get("X-API-Key"); apiKey != "" {
        return "api_key:" + apiKey
    }
	
	return "ip:" + host
}
