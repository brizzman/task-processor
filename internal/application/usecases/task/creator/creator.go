package creator

import (
	"context"
	"fmt"
	"task-processor/internal/application/ports/outbound/persistence/taskrepo"
	"task-processor/internal/domain"
	"github.com/google/uuid"
)

type Creator struct {
	taskRepo taskrepo.TaskRepository
}

func NewCreator(taskRepo taskrepo.TaskRepository) *Creator {
	return &Creator{taskRepo: taskRepo}
}

func (c *Creator) CreateTasksBatch(ctx context.Context, count int) ([]uuid.UUID, error) {
	tasks := make([]*domain.Task, count)
	for i := 0; i < count; i++ {
		tasks[i] = &domain.Task{Status: domain.StatusNew}
	}

	ids, err := c.taskRepo.BatchCreate(ctx, tasks)
	if err != nil {
		return ids, fmt.Errorf("failed to create tasks batch: %w", err)
	}

	return ids, nil
}