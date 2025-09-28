package acquirer

import (
	"context"
	"fmt"
	"task-processor/internal/application/ports/outbound/persistence/taskrepo"
	"task-processor/internal/domain"
)

type Acquirer struct {
	taskRepo           taskrepo.TaskRepository
}

func NewAcquirer(
	taskRepo 	   taskrepo.TaskRepository,
) *Acquirer {
	return &Acquirer{
		taskRepo:           taskRepo,
	}
}

func (a *Acquirer) AcquireTasks(
	ctx context.Context,
	limit int,
) ([]*domain.Task, error) {
	tasks, err := a.taskRepo.AcquireTasks(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire tasks: %w", err)
	}
	return tasks, nil
}