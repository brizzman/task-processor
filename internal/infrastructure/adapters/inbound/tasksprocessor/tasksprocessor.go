package tasksprocessor

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"task-processor/internal/application/ports/inbound/tasksprocessor"
	"task-processor/internal/application/usecases/task"
	"task-processor/internal/infrastructure/shared/logger"

	"github.com/gammazero/workerpool"
	"go.uber.org/zap"
)

type ConcurrentTasksProcessor struct {
	log        	     logger.Logger
	workerPool 		*workerpool.WorkerPool
	taskUseCases    *task.UseCases
}

func NewConcurrentTasksProcessor(
	log        	     logger.Logger,
	workerPool 		*workerpool.WorkerPool,
	taskUseCases    *task.UseCases,
) tasksprocessor.TasksProcessor {
	return &ConcurrentTasksProcessor{
		log: 			  log,
		workerPool: 	  workerPool,
		taskUseCases:     taskUseCases,
	}
}

func (a *ConcurrentTasksProcessor) ProcessTasks(
	ctx context.Context, 
	req *tasksprocessor.ProcessTasksRequest,
) (*tasksprocessor.ProcessTasksResponse, error) {
	a.log.Debug("acquiring tasks", zap.Int("limit", req.Limit))

	tasks, err := a.taskUseCases.Acquirer.AcquireTasks(ctx, req.Limit)
	if err != nil {
		a.log.Error("failed to acquire tasks", zap.Error(err))
		return nil, fmt.Errorf("failed to acquire tasks: %w", err)
	}

	if len(tasks) == 0 {
		a.log.Debug("no tasks available for processing")
		return &tasksprocessor.ProcessTasksResponse{}, nil
	}

	a.log.Info("processing tasks", zap.Int("count", len(tasks)))

	var successCount, failedCount int64
	var wg sync.WaitGroup

	for _, task := range tasks {
		wg.Add(1)
		a.workerPool.Submit(func() {
			defer wg.Done()

			success, err := a.taskUseCases.SingleProcessor.ProcessTask(ctx, task, req)

			if err != nil || !success {
				atomic.AddInt64(&failedCount, 1)
				a.log.Warn("task processing error", zap.String("task_id", task.ID.String()), zap.Error(err))
			} else {
				atomic.AddInt64(&successCount, 1)
				a.log.Debug("task processed successfully", zap.String("task_id", task.ID.String()))
			}
		})
	}

	wg.Wait()

	a.log.Info("tasks processing completed",
		zap.Int("processed", int(successCount + failedCount)),
		zap.Int("success", int(successCount)),
		zap.Int("failed", int(failedCount)),
	)

	return &tasksprocessor.ProcessTasksResponse{
		ProcessedCount: int(successCount + failedCount),
		SuccessCount:   int(successCount),
		FailedCount:    int(failedCount),
	}, nil
}