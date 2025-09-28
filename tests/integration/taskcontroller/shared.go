package taskcontroller

import (
	"context"
	"net/http"
	taskUseCases "task-processor/internal/application/usecases/task"
	"task-processor/internal/infrastructure/adapters/inbound/httpserver/task"
	"task-processor/internal/infrastructure/adapters/inbound/random"
	"task-processor/internal/infrastructure/adapters/inbound/tasksprocessor"
	"task-processor/internal/infrastructure/adapters/outbound/postgres"
	"task-processor/internal/infrastructure/adapters/outbound/redis"
	"task-processor/internal/infrastructure/config"
	"task-processor/internal/infrastructure/shared/logger"
	"task-processor/internal/infrastructure/shared/validator"
	"testing"

	"github.com/gammazero/workerpool"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

// setupTestDependencies initializes storage, logger, and controller for integration tests
func setupTestDependencies(t *testing.T) (*task.Controller, context.Context, func()) {
	ctx := context.Background()

	// Load test config	
	cfg := config.GetConfig()
	log := logger.GetLogger()

	// Initialize PostgreSQL storage
	storage, err := postgres.NewStorage(ctx, log, cfg)
	require.NoError(t, err)

	// Initialize Redis client
	rdb, err := redis.NewRedisClient(cfg)
	require.NoError(t, err)

	// Initialize validator 
	validator := validator.New()

	// Initialize worker pool
	workerpool := workerpool.New(cfg.WorkerPool.MaxWorkers)

	// Initialize task use cases and concurrent processor
	taskUseCases := taskUseCases.NewUseCases(
		storage.TaskRepo, 
		storage.FailedTaskRepo, 
		storage.TxManager, 
		random.NewCryptoRandomProvider(),
	)
	ccProcessor := tasksprocessor.NewConcurrentTasksProcessor(log, workerpool, taskUseCases)

	// Initialize controller
	controller := task.NewController(validator, ccProcessor, taskUseCases)

	// Cleanup function
	cleanup := func() {
		storage.Close()
		_ = rdb.Close()
		workerpool.Stop()
	}

	return controller, ctx, cleanup
}

// setupRouter registers controller routes into chi router
func setupRouter(controller *task.Controller) http.Handler {
	r := chi.NewRouter()
	controller.RegisterRoutes(r)
	return r
}