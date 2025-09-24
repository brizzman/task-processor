package redis

import (
	"context"
	"fmt"
	"task-processor/internal/infrastructure/config"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps Redis client
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient initializes Redis storage
func NewRedisClient(cfg *config.Config) (*RedisClient, error) {
	opts := &redis.Options{
		Addr:         cfg.Redis.Address,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	}

	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisClient{client: client}, nil
}

func (s *RedisClient) Client() *redis.Client {
	return s.client
}

func (s *RedisClient) HealthCheck(ctx context.Context) error {
	if s.client == nil {
		return fmt.Errorf("redis client is not initialized")
	}
	return s.client.Ping(ctx).Err()
}
 
func (s *RedisClient) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}
