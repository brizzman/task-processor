package postgres

import (
	"context"
	"fmt"
	"task-processor/internal/application/ports/outbound/persistence"
	"task-processor/internal/domain"
	"task-processor/internal/infrastructure/adapters/outbound/postgres/txManager"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TaskRepo implements persistence.FailedTaskRepository
type FailedTaskRepo struct {
	pool *pgxpool.Pool
}

// NewFailedTaskRepo creates new repository instance
func NewFailedTaskRepo(pool *pgxpool.Pool) persistence.FailedTaskRepository {
	return &FailedTaskRepo{pool: pool}
}

// Create inserts task into failed_tasks table
func (r *FailedTaskRepo) Create(ctx context.Context, task *domain.Task) error {
	querier := txManager.GetQuerier(ctx, r.pool)

	_, err := querier.Exec(ctx, `
		INSERT INTO failed_tasks (
			id, status, created_at, updated_at, attempts, max_attempts, error_message
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, task.ID, task.Status, task.CreatedAt, task.UpdatedAt, task.Attempts, task.MaxAttempts, task.ErrorMessage)
	if err != nil {
		return fmt.Errorf("failed to insert into failed_tasks: %w", err)
	}
	return nil
}