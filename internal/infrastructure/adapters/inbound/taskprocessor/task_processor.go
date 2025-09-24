package taskprocessor

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"task-processor/internal/application/ports/inbound/task"
	"task-processor/internal/domain"
	"task-processor/internal/infrastructure/adapters/outbound/postgres"
	"task-processor/internal/infrastructure/shared/logger"
	"time"

	"github.com/gammazero/workerpool"
	"go.uber.org/zap"
)

// ConcurrentTaskProcessor implements task.TaskProcessor with concurrency
type ConcurrentTaskProcessor struct {
	log        	    *logger.Logger
	store 			*postgres.Storage
	workerPool 		*workerpool.WorkerPool
}

// NewConcurrentTaskProcessor creates a new concurrent processor
func NewConcurrentTaskProcessor(
	log        	    *logger.Logger,
	store 			*postgres.Storage,
	workerPool 		*workerpool.WorkerPool,
) *ConcurrentTaskProcessor {
	return &ConcurrentTaskProcessor{
		log: 			log,
		store: 			store,
		workerPool: 	workerPool,
	}
}

// ProcessTasks acquires tasks and processes them concurrently
func (p *ConcurrentTaskProcessor) ProcessTasks(
	ctx context.Context, 
	request *task.ProcessTasksRequest,
) (*task.ProcessTasksResponse, error) {
	p.log.Debug("acquiring tasks", zap.Int("limit", request.Limit))
	// Acquire tasks for processing
	tasks, err := p.store.TaskRepo.AcquireTasks(ctx, request.Limit)
	if err != nil {
		p.log.Error("failed to acquire tasks", zap.Error(err))
		return nil, fmt.Errorf("failed to acquire tasks: %w", err)
	}

	// Return early if no tasks are available
	if len(tasks) == 0 {
		p.log.Debug("no tasks available for processing")
		return &task.ProcessTasksResponse{
			ProcessedCount: 0,
			SuccessCount:   0,
			FailedCount:    0,
		}, nil
	}

	p.log.Info("processing tasks", zap.Int("count", len(tasks)))

	results := make(chan processingResult, len(tasks))
	var wg sync.WaitGroup

	for _, t := range tasks {
		wg.Add(1)
		p.workerPool.Submit(func() {
			defer wg.Done()
			success, err := p.processSingleTask(ctx, t, request)
			results <- processingResult{success: success, err: err}
		})
	}

    go func() {
        wg.Wait()
        close(results)
    }()

	var successCount, failedCount int
	for result := range results {
		if result.err != nil || !result.success {
			failedCount++
		} else {
			successCount++
		}
	}

	p.log.Info("tasks processing completed", 
		zap.Int("processed", successCount+failedCount),
		zap.Int("success", successCount),
		zap.Int("failed", failedCount),
	)

	return &task.ProcessTasksResponse{
		ProcessedCount: successCount + failedCount,
		SuccessCount:   successCount,
		FailedCount:    failedCount,
	}, nil
}

// processingResult represents the result of processing a single task
type processingResult struct {
	success bool
	err     error
}

// processSingleTask processes a single task with optional delay and success probability
func (p *ConcurrentTaskProcessor) processSingleTask(
	ctx context.Context, 
	task *domain.Task, 
	request *task.ProcessTasksRequest,
) (bool, error) {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	if task.Attempts >= task.MaxAttempts {
		p.log.Warn("task exceeded max attempts", 
			zap.String("task_id", task.ID.String()),
			zap.Int("attempts", task.Attempts),
			zap.Int("max_attempts", task.MaxAttempts),
		)
		err := p.store.TxManager.WithTransaction(ctx, func(ctx context.Context) error {
			if err := p.store.TaskRepo.Delete(ctx, task.ID); err != nil {
				return fmt.Errorf("failed to delete task after moving to failed tasks: %w", err)
			}
			if err := p.store.FailedTaskRepo.Create(ctx, task); err != nil {
				return fmt.Errorf("failed to create failed task record: %w", err)
			}
			return nil
		})
		if err != nil {
			p.log.Error("failed to move task to failed tasks", 
				zap.String("task_id", task.ID.String()),
				zap.Error(err),
			)
		}
		return false, err
	}

	// Apply optional delay if specified
	if request.MinDelayMS > 0 || request.MaxDelayMS > 0 {
		minDelay := max(request.MinDelayMS, 0)
		maxDelay := max(request.MaxDelayMS, minDelay)
		
		delayMS := random.Intn(maxDelay-minDelay+1) + minDelay
		select {
		case <-ctx.Done(): // respect context cancellation
			return false, ctx.Err()
		case <-time.After(time.Duration(delayMS) * time.Millisecond):
		}
	}

	// Determine success based on success rate probability
	isSuccess := random.Float64() <= request.SuccessRate

	if isSuccess {
		p.log.Debug("task processed successfully", zap.String("task_id", task.ID.String()))
		// Mark task as successfully processed
		if err := p.store.TaskRepo.MarkAsProcessed(ctx, task.ID); err != nil {
			return false, fmt.Errorf("failed to mark task as processed: %w", err)
		}
		return true, nil
	} else {
		p.log.Debug("task processing failed", 
			zap.String("task_id", task.ID.String()),
			zap.Int("attempt", task.Attempts),
		)
		// Mark task as failed with error message
		errorMsg := fmt.Sprintf("processing failed according to success rate %.2f (attempt %d/%d)", 
			request.SuccessRate, task.Attempts, task.MaxAttempts)
		
		if err := p.store.TaskRepo.MarkAsFailed(ctx, task.ID, errorMsg); err != nil {
			return false, fmt.Errorf("failed to mark task as failed: %w", err)
		}
		return false, nil
	}
}
