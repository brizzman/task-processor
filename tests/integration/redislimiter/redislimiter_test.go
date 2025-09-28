package redislimiter

import (
	"net/http"
	"net/http/httptest"
	"testing"
    "time"

	"task-processor/internal/infrastructure/shared/middleware"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// TestRedisRateLimiter_AllowedRequests tests allowed requests within limit
func TestRedisRateLimiter_AllowedRequests(t *testing.T) {
	redisClient := GetRedisClient(t)
	defer redisClient.Close()

	// Create limiter with 10 requests per second
	limiter := middleware.NewRedisRateLimiter(redisClient, 10)
	handler := limiter.Middleware(http.HandlerFunc(TestHandler))

	// Make 5 requests - all should pass
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "OK", rr.Body.String())
		assert.Contains(t, rr.Header().Get("X-RateLimit-Limit"), "10")
		assert.NotEmpty(t, rr.Header().Get("X-RateLimit-Remaining"))
	}
}

// TestRedisRateLimiter_RateLimitExceeded tests rate limit exceeded scenario
func TestRedisRateLimiter_RateLimitExceeded(t *testing.T) {
	redisClient := GetRedisClient(t)
	defer redisClient.Close()

	// Limit 2 requests per second
	limiter := middleware.NewRedisRateLimiter(redisClient, 2)
	handler := limiter.Middleware(http.HandlerFunc(TestHandler))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.2:12345"

	// First 2 requests should pass
	for i := 0; i < 2; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	}

	// Third request should be rejected
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	assert.Contains(t, rr.Body.String(), "rate_limit_exceeded")
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.NotEmpty(t, rr.Header().Get("Retry-After"))
	assert.Equal(t, "2", rr.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "0", rr.Header().Get("X-RateLimit-Remaining"))
}

// TestRedisRateLimiter_DifferentIPs tests rate limiting for different IP addresses
func TestRedisRateLimiter_DifferentIPs(t *testing.T) {
	redisClient := GetRedisClient(t)
	defer redisClient.Close()

	// Limit 1 request per second
	limiter := middleware.NewRedisRateLimiter(redisClient, 1)
	handler := limiter.Middleware(http.HandlerFunc(TestHandler))

	// Request from first IP
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.RemoteAddr = "192.168.1.10:12345"

	// Request from second IP
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "192.168.1.20:12345"

	// Both requests should pass since they're from different IPs
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusOK, rr1.Code)

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)
}

// TestRedisRateLimiter_ResetAfterTime tests rate limit reset after time window
func TestRedisRateLimiter_ResetAfterTime(t *testing.T) {
	redisClient := GetRedisClient(t)
	defer redisClient.Close()

	// Limit 1 request per second
	limiter := middleware.NewRedisRateLimiter(redisClient, 1)
	handler := limiter.Middleware(http.HandlerFunc(TestHandler))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.30:12345"

	// First request - OK
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req)
	assert.Equal(t, http.StatusOK, rr1.Code)

	// Second request immediately - Too Many Requests
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)
	assert.Equal(t, http.StatusTooManyRequests, rr2.Code)

	// Wait for rate limit to reset (1 second + small buffer)
	time.Sleep(1100 * time.Millisecond)

	// Third request after waiting - should be OK again
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req)
	assert.Equal(t, http.StatusOK, rr3.Code)
}

// TestRedisRateLimiter_RedisFailure tests behavior when Redis is unavailable
func TestRedisRateLimiter_RedisFailure(t *testing.T) {
	// Modify config to use invalid Redis connection
	invalidCfg := &redis.Options{Addr: "invalid-addr"}

	redisClient := redis.NewClient(invalidCfg)
	defer redisClient.Close()

	limiter := middleware.NewRedisRateLimiter(redisClient, 10)
	handler := limiter.Middleware(http.HandlerFunc(TestHandler))

	// Request should pass (fail-open strategy) even with Redis connection issues
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.40:12345"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "OK", rr.Body.String())
}
