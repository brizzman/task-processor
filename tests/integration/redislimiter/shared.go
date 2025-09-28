package redislimiter

import (
	"net/http"
	"testing"

	"task-processor/internal/infrastructure/config"
	rd "task-processor/internal/infrastructure/adapters/outbound/redis"
	"github.com/redis/go-redis/v9"
)

func GetRedisClient(t *testing.T) *redis.Client {
	// Load application config
	cfg := config.GetConfig()

	// Initialize Redis client and handle errors
	rdb, err := rd.NewRedisClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}

	return rdb.Client()
}

// TestHandler regular handler for tests
func TestHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
