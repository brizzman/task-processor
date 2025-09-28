package singleprocessor

import (
	"context"
	"fmt"
	"task-processor/internal/application/ports/outbound/persistence/taskrepo"
	"task-processor/internal/application/ports/outbound/persistence/failedtaskrepo"
	"task-processor/internal/application/ports/outbound/persistence/txmanager"
	"task-processor/internal/application/ports/inbound/random"
	"task-processor/internal/application/ports/inbound/tasksprocessor"
	"task-processor/internal/domain"
	"time"
)

type SingleProcessor struct {
	taskRepo           taskrepo.TaskRepository
	failedTaskRepo     failedtaskrepo.FailedTaskRepository
	txManager 		   txmanager.TxManager
	randomProvider     random.RandomProvider
}

func NewSingleProcessor(
	taskRepo 	   taskrepo.TaskRepository,
	failedTaskRepo failedtaskrepo.FailedTaskRepository,
	txManager 	   txmanager.TxManager,
	randomProvider random.RandomProvider,
) *SingleProcessor {
	return &SingleProcessor{
		taskRepo:           taskRepo,
		failedTaskRepo:     failedTaskRepo,
		txManager: 			txManager,
		randomProvider:     randomProvider,
	}
}

func (s *SingleProcessor) ProcessTask(
	ctx context.Context,
	task *domain.Task,
	request *tasksprocessor.ProcessTasksRequest,
) (bool, error) {
	if task.Attempts >= task.MaxAttempts {
		return s.handleMaxAttemptsExceeded(ctx, task)
	}

	if err := s.applyProcessingDelay(ctx, request); err != nil {
		return false, err
	}

	isSuccess := s.randomProvider.Float64() <= request.SuccessRate

	if isSuccess {
		return s.handleSuccessfulProcessing(ctx, task)
	} else {
		return s.handleFailedProcessing(ctx, task, request)
	}
}

func (s *SingleProcessor) handleMaxAttemptsExceeded(
	ctx context.Context,
	task *domain.Task,
) (bool, error) {
	err := s.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.taskRepo.Delete(ctx, task.ID); err != nil {
			return fmt.Errorf("failed to delete task: %w", err)
		}
		if err := s.failedTaskRepo.Create(ctx, task); err != nil {
			return fmt.Errorf("failed to create failed task record: %w", err)
		}
		return nil
	})

	return false, err
}

func (s *SingleProcessor) applyProcessingDelay(
	ctx context.Context,
	request *tasksprocessor.ProcessTasksRequest,
) error {
	if request.MinDelayMS > 0 || request.MaxDelayMS > 0 {
		minDelay := max(request.MinDelayMS, 0)
		maxDelay := max(request.MaxDelayMS, minDelay)
		
		delayMS := s.randomProvider.Intn(maxDelay-minDelay+1) + minDelay
		
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(delayMS) * time.Millisecond):
		}
	}
	return nil
}

func (s *SingleProcessor) handleSuccessfulProcessing(
	ctx context.Context,
	task *domain.Task,
) (bool, error) {
	if err := s.taskRepo.MarkAsProcessed(ctx, task.ID); err != nil {
		return false, fmt.Errorf("failed to mark task as processed: %w", err)
	}
	return true, nil
}

func (s *SingleProcessor) handleFailedProcessing(
	ctx context.Context,
	task *domain.Task,
	request *tasksprocessor.ProcessTasksRequest,
) (bool, error) {
	errorMsg := fmt.Sprintf("processing failed according to success rate %.2f (attempt %d/%d)",
		request.SuccessRate, task.Attempts, task.MaxAttempts)

	if err := s.taskRepo.MarkAsFailed(ctx, task.ID, errorMsg); err != nil {
		return false, fmt.Errorf("failed to mark task as failed: %w", err)
	}
	return false, nil
}