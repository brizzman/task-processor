package circuitbreaker

import (
	"context"
	"errors"
	"task-processor/internal/application/ports/outbound/persistence/taskrepo"
	"task-processor/internal/domain"
	"task-processor/internal/infrastructure/config"
	"task-processor/internal/infrastructure/shared/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TaskRepoDecorator struct {
	repository taskrepo.TaskRepository
	base       *BaseDecorator
}

func NewTaskRepoDecorator(
	repository taskrepo.TaskRepository,
	cfg 	  *config.Config,
	logger    logger.Logger,
	name       string,
) *TaskRepoDecorator {
	
	base := NewBaseDecorator(cfg, logger, name)
	
	operations := []string{"BatchCreate", "AcquireTasks", "MarkAsProcessed", "MarkAsFailed", "Delete"}
	for _, op := range operations {
		base.AddCircuitBreaker(op, base.CreateSettings(cfg, op))
	}

	return &TaskRepoDecorator{
		repository: repository,
		base:       base,
	}
}

func (d *TaskRepoDecorator) BatchCreate(ctx context.Context, tasks []*domain.Task) ([]uuid.UUID, error) {
	result, err := d.base.ExecuteWithCB("BatchCreate", func() (any, error) {
		return d.repository.BatchCreate(ctx, tasks)
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]uuid.UUID)
	if !ok {
		d.base.logger.Error("type assertion failed",
			zap.String("operation", "BatchCreate"),
			zap.String("expected", "[]uuid.UUID"))
		return nil, errors.New("type assertion error")
	}

	return ids, nil
}

func (d *TaskRepoDecorator) AcquireTasks(ctx context.Context, limit int) ([]*domain.Task, error) {
	result, err := d.base.ExecuteWithCB("AcquireTasks", func() (any, error) {
		return d.repository.AcquireTasks(ctx, limit)
	})
	if err != nil {
		return nil, err
	}

	tasks, ok := result.([]*domain.Task)
	if !ok {
		d.base.logger.Error("type assertion failed",
			zap.String("operation", "AcquireTasks"),
			zap.String("expected", "[]*domain.Task"))
		return nil, errors.New("type assertion error")
	}

	return tasks, nil
}

func (d *TaskRepoDecorator) MarkAsProcessed(ctx context.Context, taskID uuid.UUID) error {
	_, err := d.base.ExecuteWithCB("MarkAsProcessed", func() (any, error) {
		return nil, d.repository.MarkAsProcessed(ctx, taskID)
	})
	return err
}

func (d *TaskRepoDecorator) MarkAsFailed(ctx context.Context, taskID uuid.UUID, errorMsg string) error {
	_, err := d.base.ExecuteWithCB("MarkAsFailed", func() (any, error) {
		return nil, d.repository.MarkAsFailed(ctx, taskID, errorMsg)
	})
	return err
}

func (d *TaskRepoDecorator) Delete(ctx context.Context, taskID uuid.UUID) error {
	_, err := d.base.ExecuteWithCB("Delete", func() (any, error) {
		return nil, d.repository.Delete(ctx, taskID)
	})
	return err
}