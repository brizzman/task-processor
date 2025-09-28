package taskrepo

import (
	"context"
	"task-processor/internal/domain"

	"github.com/google/uuid"
)


// TaskRepository defines the interface for task data access operations
type TaskRepository interface {

	// BatchCreate creates multiple tasks in a single operation
	BatchCreate(ctx context.Context, tasks []*domain.Task) ([]uuid.UUID, error)
	
	// AcquireTasks acquires tasks for processing with pessimistic locking
	AcquireTasks(ctx context.Context, limit int) ([]*domain.Task, error)
	
	// MarkAsProcessed marks task as processed
	MarkAsProcessed(ctx context.Context, taskID uuid.UUID) error
	
	// MarkAsFailed marks task as failed and records error message
	MarkAsFailed(ctx context.Context, taskID uuid.UUID, errorMsg string) error

	// Delete removes row from table
	Delete(ctx context.Context, taskID uuid.UUID) error
}