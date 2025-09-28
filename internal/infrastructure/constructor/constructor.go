package constructor

import (
	"net/http"
	"sync/atomic"
	"task-processor/internal/application/ports/inbound/tasksprocessor"
	"task-processor/internal/application/usecases/task"
	"task-processor/internal/infrastructure/adapters/inbound/httpserver/health"
	"task-processor/internal/infrastructure/adapters/inbound/httpserver/swagger"
	tsk "task-processor/internal/infrastructure/adapters/inbound/httpserver/task"
	"task-processor/internal/infrastructure/adapters/outbound/postgres"
	"task-processor/internal/infrastructure/config"
	"task-processor/internal/infrastructure/shared/logger"
	mdlware "task-processor/internal/infrastructure/shared/middleware"
	"task-processor/internal/infrastructure/shared/validator"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
)

// InfraDeps contains low-level infrastructure dependencies
type InfraDeps struct {
	Logger  	    logger.Logger
	Config  	   *config.Config
	PG     		   *postgres.Storage
	Redis   	   *redis.Client 
	Validator      *validator.Validator
}

// AppDeps contains application-level services
type AppDeps struct {
	TaskUseCases   *task.UseCases
	TasksProcessor  tasksprocessor.TasksProcessor
	IsShuttingDown *atomic.Bool
}

// Dependencies is a root container for router setup
type Dependencies struct {
	Infra InfraDeps
	App   AppDeps
}

func Construct(router *chi.Mux, deps Dependencies) {
	registerMiddleware(router, deps)
	registerHealthController(router, deps)
	registerTaskController(router, deps)
	registerSwaggerController(router)
}

func registerMiddleware(router *chi.Mux, deps Dependencies) {
	middlewares := []func(http.Handler) http.Handler{
		middleware.RequestID,
		mdlware.LoggerMiddleware(deps.Infra.Logger),
		middleware.Recoverer,
		mdlware.NewRedisRateLimiter(
			deps.Infra.Redis,
			deps.Infra.Config.RateLimit.RPS,
		).Middleware,
	}

	for _, mw := range middlewares {
		router.Use(mw)
	}
}

func registerHealthController(router *chi.Mux, deps Dependencies) {
	healthController := health.NewController(
		deps.App.IsShuttingDown,
		NewHealthChecker(
			deps.Infra.PG,
			deps.Infra.Redis,
			deps.Infra.Logger,
			deps.Infra.Config.HealthCheck.Timeout,
		).Check,
	)
	healthController.RegisterRoutes(router)
}

func registerTaskController(router *chi.Mux, deps Dependencies) {
	taskController := tsk.NewController(
		deps.Infra.Validator, 
		deps.App.TasksProcessor, 
		deps.App.TaskUseCases,
	)
	taskController.RegisterRoutes(router)
}

func registerSwaggerController(router *chi.Mux) {
	swaggerUI := swagger.NewController()
	swaggerUI.RegisterRoutes(router)
}

