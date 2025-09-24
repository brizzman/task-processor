package persistence

import (
	"context"
	"task-processor/internal/domain"
)


// FailedTaskRepository defines the interface for failed task data access operations
type FailedTaskRepository interface {
    Create(ctx context.Context, task *domain.Task) error
}