package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"sync/atomic"
	"syscall"
	"task-processor/internal/application/usecases/task"
	"task-processor/internal/infrastructure/adapters/inbound/httpserver"
	"task-processor/internal/infrastructure/adapters/inbound/taskprocessor"
	"task-processor/internal/infrastructure/adapters/outbound/postgres"
	"task-processor/internal/infrastructure/adapters/outbound/redis"
	"task-processor/internal/infrastructure/config"
	"task-processor/internal/infrastructure/constructor"
	"task-processor/internal/infrastructure/shared/logger"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/go-chi/chi/v5"
	"github.com/oklog/run"
	"go.uber.org/zap"

	_ "task-processor/internal/infrastructure/adapters/inbound/httpserver/docs"

	_ "go.uber.org/automaxprocs"
)


var isShuttingDown atomic.Bool

func main() {
	if err := runApp(); err != nil {
		panic(fmt.Sprintf("Application failed: %v", err))
	}
}

func runApp() error {
	var g run.Group

	// --- Root context with OS signals ---
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Context used as base for all incoming HTTP requests
	ongoingCtx, stopOngoing := context.WithCancel(context.Background())
	defer stopOngoing()
	
	// --- Init config & logger ---
	cfg := config.GetConfig()
	
	log := logger.GetLogger()
	
	log.Info("initializing postgresql storage")
	store, err := postgres.NewStorage(rootCtx, log, cfg)
	if err != nil {
		return fmt.Errorf("postgres.NewStorage failed: %w", err)
	}
	defer store.Close()

	log.Info("initializing redis client")
	rdb, err := redis.NewRedisClient(cfg)
	if err != nil {
		return fmt.Errorf("redis.NewRedisStorage failed: %w", err)
	}
	defer rdb.Close()

	// --  Init worker pool ---
	wp := workerpool.New(cfg.WorkerPool.MaxWorkers)
	defer wp.StopWait()

	// --- Init task processor ---
	taskProcessor := taskprocessor.NewConcurrentTaskProcessor(log, store, wp)
	
	// --- Init usecases ---
	taskCreator := task.NewTaskCreator(store.TaskRepo)

	// --- Init & Construct chi-router ---
	router := chi.NewRouter()
	constructorDeps := constructor.Dependencies{
		Infra: constructor.InfraDeps{
			Logger: log,
			Config: cfg,
			PG:     store,
			Redis:  rdb.Client(),
		},
		App: constructor.AppDeps{
			TaskProcessor:   taskProcessor,
			TaskCreator: 	 taskCreator,
			IsShuttingDown: &isShuttingDown,
		},
	}
	constructor.Construct(router, constructorDeps)

	// --- Init HTTP server ---
	httpSrv := httpserver.NewHTTPServer(cfg, router, ongoingCtx)

	// --- OS signal listener (SIGINT, SIGTERM) ---
	g.Add(
		func() error {
			// Wait for termination signal (e.g., Ctrl+C or Kubernetes pod shutdown)
			<-rootCtx.Done()
			// Returning nil will trigger the shutdown sequence in other run.Group actors
			return nil
		},
		func(err error) {
			// Cleanup logic for the signal handler
			stop()
			isShuttingDown.Store(true)
			log.Info("shutdown signal received")
			// Allow time for readiness probe updates to propagate to external load balancers
			time.Sleep(cfg.Shutdown.ReadinessDrain)
		},
	)

	// --- HTTP server lifecycle ---
	g.Add(
		func() error {
			log.Info("starting HTTP server", zap.String("addr", httpSrv.Addr()))
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Error("http server error", zap.Error(err))
				return err
			}
			return nil
		},
		func(err error) {
			log.Info("shutting down http server")
			// Begin graceful shutdown process with timeout
			ctx, cancel := context.WithTimeout(context.Background(), cfg.Shutdown.HTTPTimeout)
			defer cancel()
			// Attempt graceful shutdown of the server (waits for handlers to complete)
			shutdownErr := httpSrv.Shutdown(ctx)
			// Notify all active handlers via context cancellation
			stopOngoing()
			if shutdownErr != nil {
				log.Error("graceful shutdown failed, forcing close", zap.Error(shutdownErr))
				time.Sleep(cfg.Shutdown.HardPeriod)
			}

			log.Info("http server stopped gracefully.")
		},
	)

	// --- Group execution ---
	if err := g.Run(); err != nil {
		log.Error("application exited with error", zap.Error(err))
	} else {
		log.Info("application stopped gracefully")
	}

	return nil
}
