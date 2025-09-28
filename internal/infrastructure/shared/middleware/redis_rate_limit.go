package middleware

import (
	"encoding/json"
	"log"
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
}

func NewRedisRateLimiter(redisClient *redis.Client, rps int) *RedisRateLimiter {
	return &RedisRateLimiter{
		limiter: redis_rate.NewLimiter(redisClient),
		rps:     rps,
	}
}

func (r *RedisRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Get key for rate limiting by IP
		key := getRateLimitKey(req)
		
		// Check the rate limit
		res, err := r.limiter.Allow(req.Context(), key, redis_rate.PerSecond(r.rps))
		if err != nil {
			// On Redis error, let the request pass (fail open)
			next.ServeHTTP(w, req)
			return
		}

		// Rate limit exceeded
		if res.Allowed == 0 {
			// Prepare response body first
			response := map[string]string{
				"error": rateLimitExceeded,
			}
			
			body, err := json.Marshal(response)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Set headers and status
			w.Header().Set("Content-Type", contentTypeJSON)
			w.Header().Set("Retry-After", strconv.Itoa(int(res.RetryAfter.Seconds())))
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(r.rps))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
			w.Header().Set("X-RateLimit-Reset", time.Now().Add(res.RetryAfter).Format(time.RFC1123))
			w.WriteHeader(http.StatusTooManyRequests)

			// Write pre-prepared body
			if _, err := w.Write(body); err != nil {
				log.Printf("Failed to write rate limit response: %v", err)
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

// getRateLimitKey determines the key for rate limiting by IP only
func getRateLimitKey(req *http.Request) string {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		host = req.RemoteAddr 
	}
	return "ip:" + host
}