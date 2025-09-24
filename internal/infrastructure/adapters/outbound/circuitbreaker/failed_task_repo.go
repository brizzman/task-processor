package circuitbreaker

import (
	"context"
	"task-processor/internal/application/ports/outbound/persistence"
	"task-processor/internal/domain"
	"task-processor/internal/infrastructure/config"
	"task-processor/internal/infrastructure/shared/logger"

)

type FailedTaskRepoDecorator struct {
	repository persistence.FailedTaskRepository
	base       *BaseDecorator
}

func NewFailedTaskRepoDecorator(
	repository persistence.FailedTaskRepository,
	cfg *config.Config,
	logger *logger.Logger,
	name string,
) *FailedTaskRepoDecorator {
	
	base := NewBaseDecorator(cfg, logger, name)
	base.AddCircuitBreaker("Create", base.CreateSettings(cfg, "Create"))

	return &FailedTaskRepoDecorator{
		repository: repository,
		base:       base,
	}
}

func (d *FailedTaskRepoDecorator) Create(ctx context.Context, task *domain.Task) error {
	_, err := d.base.ExecuteWithCB("Create", func() (any, error) {
		return nil, d.repository.Create(ctx, task)
	})
	return err
}