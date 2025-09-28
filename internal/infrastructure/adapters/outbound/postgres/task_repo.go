package postgres

import (
	"context"
	"errors"
	"fmt"
	"task-processor/internal/application/ports/outbound/persistence/taskrepo"
	"task-processor/internal/domain"
	"task-processor/internal/infrastructure/adapters/outbound/postgres/txManager"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTaskNotFound = errors.New("task not found")

// TaskRepo implements persistence.TaskRepository
type TaskRepo struct {
	pool *pgxpool.Pool
}

// NewTaskRepo creates new repository instance
func NewTaskRepo(pool *pgxpool.Pool) taskrepo.TaskRepository {
	return &TaskRepo{pool: pool}
}

// BatchCreate creates multiple tasks in a single operation
func (r *TaskRepo) BatchCreate(ctx context.Context, tasks []*domain.Task) ([]uuid.UUID, error) {
	if len(tasks) == 0 {
		return []uuid.UUID{}, nil
	}

	querier := txManager.GetQuerier(ctx, r.pool)
	batch := &pgx.Batch{}

	for _, task := range tasks {
		batch.Queue(`
			INSERT INTO tasks (status)
			VALUES ($1)
			RETURNING id
		`, task.Status)
	}

	results := querier.SendBatch(ctx, batch)
	defer results.Close()

	ids := make([]uuid.UUID, 0, len(tasks))
	var errs []error

	for i := 0; i < len(tasks); i++ {
		var id uuid.UUID
		if err := results.QueryRow().Scan(&id); err != nil {
			errs = append(errs, fmt.Errorf("task %d: %w", i, err))
			continue
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("failed to insert tasks: %w", errors.Join(errs...))
	}

	if len(errs) > 0 {
		return ids, fmt.Errorf("inserted %d out of %d tasks, errors: %v", len(ids), len(tasks), errs)
	}

	return ids, nil
}

// AcquireTasks acquires tasks for processing with pessimistic locking
func (r *TaskRepo) AcquireTasks(ctx context.Context, limit int) ([]*domain.Task, error) {
	querier := txManager.GetQuerier(ctx, r.pool)

	query := `
		UPDATE tasks 
		SET 
			status = $1,
			attempts = attempts + 1
		WHERE id IN (
			SELECT id FROM tasks 
			WHERE status IN ($2, $3) 
			AND attempts < max_attempts
			ORDER BY created_at ASC
			LIMIT $4
			FOR UPDATE SKIP LOCKED
		)
		RETURNING 
			id, status, created_at, updated_at, 
			attempts, max_attempts, error_message	
		`

	rows, err := querier.Query(ctx, query, 
		domain.StatusProcessing, 
		domain.StatusNew, 
		domain.StatusFailed, 
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]*domain.Task, 0, limit)

	for rows.Next() {
		var task domain.Task
		var errorMsg *string

		err := rows.Scan(
			&task.ID,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.Attempts,
			&task.MaxAttempts,
			&errorMsg,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		if errorMsg != nil {
			task.ErrorMessage = *errorMsg
		} else {
			task.ErrorMessage = ""
		}
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through tasks: %w", err)
	}

	return tasks, nil
}

// MarkAsProcessed marks task as processed
func (r *TaskRepo) MarkAsProcessed(ctx context.Context, taskID uuid.UUID) error {
	querier := txManager.GetQuerier(ctx, r.pool)
	
	tag, err := querier.Exec(ctx, `
		UPDATE tasks
		SET status = $1
		WHERE id = $2
	`, domain.StatusProcessed, taskID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrTaskNotFound
	}
	return nil
}

// MarkAsFailed marks task as failed and records error message
func (r *TaskRepo) MarkAsFailed(ctx context.Context, taskID uuid.UUID, errorMsg string) error {
	querier := txManager.GetQuerier(ctx, r.pool)

	tag, err := querier.Exec(ctx, `
		UPDATE tasks
		SET status = $1,
		    error_message = $2
		WHERE id = $3
	`, domain.StatusFailed, errorMsg, taskID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrTaskNotFound
	}
	return nil
}

// Delete removes task from table
func (r *TaskRepo) Delete(ctx context.Context, taskID uuid.UUID) error {
	querier := txManager.GetQuerier(ctx, r.pool)

	tag, err := querier.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTaskNotFound
	}
	return nil
}
