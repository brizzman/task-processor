package constructor

import (
	"context"
	"task-processor/internal/infrastructure/adapters/outbound/postgres"
	"task-processor/internal/infrastructure/shared/logger"
	"time"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type HealthChecker struct {
	pg     	*postgres.Storage
	redis  	*redis.Client
	log    	*logger.Logger
	timeout  time.Duration
}

func NewHealthChecker(
	pg 		*postgres.Storage, 
	redis 	*redis.Client, 
	log 	*logger.Logger,
	timeout  time.Duration,
) *HealthChecker {
	return &HealthChecker{
		pg: pg, 
		redis: redis, 
		log: log,
		timeout: timeout,
	}
}

func (h *HealthChecker) Check() bool {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	if err := h.pg.HealthCheck(ctx); err != nil {
		h.log.Warn("Postgres health check failed", zap.Error(err))
		return false
	}
	if err := h.redis.Ping(ctx).Err(); err != nil {
		h.log.Warn("Redis health check failed", zap.Error(err))
		return false
	}
	return true
}