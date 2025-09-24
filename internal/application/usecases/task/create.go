package task

import (
	"context"
	"fmt"
	"task-processor/internal/application/ports/outbound/persistence"
	"task-processor/internal/domain"
	"github.com/google/uuid"
)

type TaskCreator struct {
	taskRepo persistence.TaskRepository
}

func NewTaskCreator(taskRepo persistence.TaskRepository) *TaskCreator {
	return &TaskCreator{taskRepo: taskRepo}
}

func (uc *TaskCreator) CreateTasksBatch(ctx context.Context, count int) ([]uuid.UUID, error) {
	tasks := make([]*domain.Task, count)
	for i := 0; i < count; i++ {
		tasks[i] = &domain.Task{Status: domain.StatusNew}
	}

	ids, err := uc.taskRepo.BatchCreate(ctx, tasks)
	if err != nil {
		return ids, fmt.Errorf("failed to create tasks batch: %w", err)
	}

	return ids, nil
}