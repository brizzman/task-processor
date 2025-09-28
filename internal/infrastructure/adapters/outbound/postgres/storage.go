package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"task-processor/internal/application/ports/outbound/persistence/failedtaskrepo"
	"task-processor/internal/application/ports/outbound/persistence/taskrepo"
	"task-processor/internal/application/ports/outbound/persistence/txmanager"
	"task-processor/internal/infrastructure/adapters/outbound/circuitbreaker"
	"task-processor/internal/infrastructure/adapters/outbound/postgres/txManager"
	"task-processor/internal/infrastructure/config"
	"task-processor/internal/infrastructure/shared/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

// Storage contains database connection pool and repositories
type Storage struct {
	pool     	  *pgxpool.Pool
	TxManager      txmanager.TxManager
	TaskRepo  	   taskrepo.TaskRepository
	FailedTaskRepo failedtaskrepo.FailedTaskRepository
}

// NewStorage initializes PostgreSQL storage with optional Circuit Breaker protection
func NewStorage(
	ctx 	context.Context,
	logger  logger.Logger,
	cfg    *config.Config,
) (*Storage, error) {

	pool, err := createConnectionPool(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

    if err := runMigrations(cfg, logger); err != nil {
        pool.Close()
        return nil, fmt.Errorf("failed to run migrations: %w", err)
    }

	txManager := txManager.NewTxManager(pool)

	taskRepo, err := createTaskRepository(pool, logger, cfg)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to create task repository: %w", err)
	}

	failedTaskRepo, err := createFailedTaskRepository(pool, logger, cfg)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to create failedTask repository: %w", err)
	}

	return &Storage{
		pool:     		pool,
		TxManager: 	    txManager,
		TaskRepo: 		taskRepo,
		FailedTaskRepo: failedTaskRepo,
	}, nil
}


// HealthCheck verifies database connection availability
func (s *Storage) HealthCheck(ctx context.Context) error {
	if s.pool == nil {
		return fmt.Errorf("connection pool not initialized")
	}
	return s.pool.Ping(ctx)
}

func (s *Storage) Pool() *pgxpool.Pool {
	return s.pool
}

// Close releases all database connections and resources
func (s *Storage) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// runMigrations applies database migrations using goose
func runMigrations(cfg *config.Config, logger logger.Logger) error {
	goose.SetBaseFS(Migrations)
    db, err := sql.Open("pgx", cfg.PG.URL)
    if err != nil {
        return fmt.Errorf("failed to open sql.DB: %w", err)
    }
    defer db.Close()
	
    if err := goose.SetDialect("postgres"); err != nil {
        return fmt.Errorf("failed to set goose dialect: %w", err)
    }

    if err := goose.Up(db, "migrations"); err != nil {
        return fmt.Errorf("failed to run migrations: %w", err)
    }

    logger.Info("Database migrations applied successfully")
    return nil
}	

// createConnectionPool establishes PostgreSQL connection pool with configured settings
func createConnectionPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, cfg.PG.ConnectTimeout)
	defer cancel()

	pgxCfg, err := pgxpool.ParseConfig(cfg.PG.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	pgxCfg.MaxConns = cfg.PG.MaxPoolSize
	pgxCfg.MinConns = cfg.PG.MinPoolSize
	pgxCfg.MaxConnLifetime = cfg.PG.MaxConnLife
	pgxCfg.MaxConnIdleTime = cfg.PG.MaxConnIdle

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize connection pool: %w", err)
	}

	return pool, nil
}

// createTaskRepository initializes task repository with optional Circuit Breaker wrapper
func createTaskRepository(pool *pgxpool.Pool, logger logger.Logger, cfg  *config.Config) (taskrepo.TaskRepository, error) {
	baseRepo := NewTaskRepo(pool)

	if cfg.CircuitBreaker.Enabled && logger != nil {
		return circuitbreaker.NewTaskRepoDecorator(baseRepo, cfg, logger, "postgres-task-repo"), nil
	}

	return baseRepo, nil
}

// createFailedTaskRepository initializes failedTask repository with optional Circuit Breaker wrapper
func createFailedTaskRepository(pool *pgxpool.Pool, logger logger.Logger, cfg  *config.Config) (failedtaskrepo.FailedTaskRepository, error) {
	baseRepo := NewFailedTaskRepo(pool)

	if cfg.CircuitBreaker.Enabled && logger != nil {
		return circuitbreaker.NewFailedTaskRepoDecorator(baseRepo, cfg, logger, "postgres-failedTask-repo"), nil
	}

	return baseRepo, nil
}