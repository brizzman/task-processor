package task

import (
	"context"
	"task-processor/internal/application/ports/inbound/random"
	"task-processor/internal/application/ports/inbound/tasksprocessor"
	"task-processor/internal/application/ports/outbound/persistence/failedtaskrepo"
	"task-processor/internal/application/ports/outbound/persistence/taskrepo"
	"task-processor/internal/application/ports/outbound/persistence/txmanager"
	"task-processor/internal/application/usecases/task/acquirer"
	"task-processor/internal/application/usecases/task/creator"
	"task-processor/internal/application/usecases/task/singleprocessor"
	"task-processor/internal/domain"

	"github.com/google/uuid"
)

type UseCases struct {
	Creator   		 Creator
	Acquirer  	 	 Acquirer
	SingleProcessor  SingleProcessor
}

func NewUseCases(
	taskRepo taskrepo.TaskRepository,
	failedTaskRepo failedtaskrepo.FailedTaskRepository,
	txManager txmanager.TxManager,
	randomProvider random.RandomProvider,
) *UseCases {

	return &UseCases{
		Creator:   creator.NewCreator(taskRepo),
		Acquirer:  acquirer.NewAcquirer(taskRepo),
		SingleProcessor: singleprocessor.NewSingleProcessor(taskRepo, failedTaskRepo, txManager, randomProvider),
	}
}

type Creator interface {
	CreateTasksBatch(ctx context.Context, count int) ([]uuid.UUID, error)
}
type Acquirer interface {
	AcquireTasks(ctx context.Context, limit int) ([]*domain.Task, error) 
}
type SingleProcessor interface {
	ProcessTask(ctx context.Context, task *domain.Task, request *tasksprocessor.ProcessTasksRequest) (bool, error)
}
