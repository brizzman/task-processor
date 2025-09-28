package tasksprocessor

import (
	"context"
	"errors"
	"testing"

	"task-processor/internal/application/ports/inbound/tasksprocessor"
	"task-processor/internal/application/usecases/task"
	"task-processor/internal/application/usecases/task/acquirer"
	"task-processor/internal/application/usecases/task/singleprocessor"
	"task-processor/internal/domain"
	"task-processor/internal/infrastructure/shared/logger"

	"github.com/gammazero/workerpool"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

func TestProcessTasks_NoTasks(t *testing.T) {
	log := &logger.ZapLogger{Logger: zaptest.NewLogger(t)}
	workerPool := workerpool.New(1)

	mockAcquirer := &acquirer.MockAcquirer{}
	mockAcquirer.On("AcquireTasks", mock.Anything, 10).Return([]*domain.Task{}, nil)

	taskUseCases := &task.UseCases{
		Acquirer:        mockAcquirer,
		SingleProcessor: &singleprocessor.MockSingleProcessor{},
	}

	processor := NewConcurrentTasksProcessor(log, workerPool, taskUseCases)
	resp, err := processor.ProcessTasks(context.Background(), &tasksprocessor.ProcessTasksRequest{Limit: 10})

	assert.NoError(t, err)
	assert.Equal(t, 0, resp.ProcessedCount)
	assert.Equal(t, 0, resp.SuccessCount)
	assert.Equal(t, 0, resp.FailedCount)
}

func TestProcessTasks_AllSuccess(t *testing.T) {
	log := &logger.ZapLogger{Logger: zaptest.NewLogger(t)}
	workerPool := workerpool.New(5)

	tasks := []*domain.Task{
		{ID: uuid.New()},
		{ID: uuid.New()},
		{ID: uuid.New()},
	}

	mockAcquirer := &acquirer.MockAcquirer{}
	mockAcquirer.On("AcquireTasks", mock.Anything, 3).Return(tasks, nil)

	mockProcessor := &singleprocessor.MockSingleProcessor{}
	for _, t := range tasks {
		mockProcessor.On("ProcessTask", mock.Anything, t, mock.Anything).Return(true, nil)
	}

	taskUseCases := &task.UseCases{
		Acquirer:        mockAcquirer,
		SingleProcessor: mockProcessor,
	}

	processor := NewConcurrentTasksProcessor(log, workerPool, taskUseCases)
	resp, err := processor.ProcessTasks(context.Background(), &tasksprocessor.ProcessTasksRequest{Limit: 3})

	assert.NoError(t, err)
	assert.Equal(t, 3, resp.ProcessedCount)
	assert.Equal(t, 3, resp.SuccessCount)
	assert.Equal(t, 0, resp.FailedCount)
}

func TestProcessTasks_AllFailed(t *testing.T) {
	log := &logger.ZapLogger{Logger: zaptest.NewLogger(t)}
	workerPool := workerpool.New(3)

	tasks := []*domain.Task{
		{ID: uuid.New()},
		{ID: uuid.New()},
	}

	mockAcquirer := &acquirer.MockAcquirer{}
	mockAcquirer.On("AcquireTasks", mock.Anything, 2).Return(tasks, nil)

	mockProcessor := &singleprocessor.MockSingleProcessor{}
	for _, t := range tasks {
		mockProcessor.On("ProcessTask", mock.Anything, t, mock.Anything).Return(false, errors.New("fail"))
	}

	taskUseCases := &task.UseCases{
		Acquirer:        mockAcquirer,
		SingleProcessor: mockProcessor,
	}

	processor := NewConcurrentTasksProcessor(log, workerPool, taskUseCases)
	resp, err := processor.ProcessTasks(context.Background(), &tasksprocessor.ProcessTasksRequest{Limit: 2})

	assert.NoError(t, err)
	assert.Equal(t, 2, resp.ProcessedCount)
	assert.Equal(t, 0, resp.SuccessCount)
	assert.Equal(t, 2, resp.FailedCount)
}

func TestProcessTasks_PartialSuccess(t *testing.T) {
	log := &logger.ZapLogger{Logger: zaptest.NewLogger(t)}
	workerPool := workerpool.New(4)

	tasks := []*domain.Task{
		{ID: uuid.New()},
		{ID: uuid.New()},
		{ID: uuid.New()},
	}

	mockAcquirer := &acquirer.MockAcquirer{}
	mockAcquirer.On("AcquireTasks", mock.Anything, 3).Return(tasks, nil)

	mockProcessor := &singleprocessor.MockSingleProcessor{}
	mockProcessor.On("ProcessTask", mock.Anything, tasks[0], mock.Anything).Return(true, nil)
	mockProcessor.On("ProcessTask", mock.Anything, tasks[1], mock.Anything).Return(false, errors.New("fail"))
	mockProcessor.On("ProcessTask", mock.Anything, tasks[2], mock.Anything).Return(true, nil)

	taskUseCases := &task.UseCases{
		Acquirer:        mockAcquirer,
		SingleProcessor: mockProcessor,
	}

	processor := NewConcurrentTasksProcessor(log, workerPool, taskUseCases)
	resp, err := processor.ProcessTasks(context.Background(), &tasksprocessor.ProcessTasksRequest{Limit: 3})

	assert.NoError(t, err)
	assert.Equal(t, 3, resp.ProcessedCount)
	assert.Equal(t, 2, resp.SuccessCount)
	assert.Equal(t, 1, resp.FailedCount)
}

func TestProcessTasks_AcquireError(t *testing.T) {
	log := &logger.ZapLogger{Logger: zaptest.NewLogger(t)}
	workerPool := workerpool.New(1)

	mockAcquirer := &acquirer.MockAcquirer{}
	mockAcquirer.On("AcquireTasks", mock.Anything, 5).Return([]*domain.Task{}, errors.New("acquire fail"))

	taskUseCases := &task.UseCases{
		Acquirer:        mockAcquirer,
		SingleProcessor: &singleprocessor.MockSingleProcessor{},
	}

	processor := NewConcurrentTasksProcessor(log, workerPool, taskUseCases)
	resp, err := processor.ProcessTasks(context.Background(), &tasksprocessor.ProcessTasksRequest{Limit: 5})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to acquire tasks")
}

func TestProcessTasks_ContextCanceled(t *testing.T) {
	log := &logger.ZapLogger{Logger: zaptest.NewLogger(t)}
	workerPool := workerpool.New(2)

	tasks := []*domain.Task{
		{ID: uuid.New()},
	}

	mockAcquirer := &acquirer.MockAcquirer{}
	mockAcquirer.On("AcquireTasks", mock.Anything, 1).Return(tasks, nil)

	mockProcessor := &singleprocessor.MockSingleProcessor{}
	mockProcessor.On("ProcessTask", mock.Anything, tasks[0], mock.Anything).Return(false, errors.New("fail"))

	taskUseCases := &task.UseCases{
		Acquirer:        mockAcquirer,
		SingleProcessor: mockProcessor,
	}

	processor := NewConcurrentTasksProcessor(log, workerPool, taskUseCases)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() 

	resp, err := processor.ProcessTasks(ctx, &tasksprocessor.ProcessTasksRequest{Limit: 1})
	assert.NotNil(t, resp)
	assert.NoError(t, err)
}
