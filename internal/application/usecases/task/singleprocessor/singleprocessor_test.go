package singleprocessor

import (
	"context"
	"task-processor/internal/application/ports/inbound/random"
	"task-processor/internal/application/ports/outbound/persistence/failedtaskrepo"
	"task-processor/internal/application/ports/outbound/persistence/taskrepo"
	"task-processor/internal/application/ports/outbound/persistence/txmanager"
	"task-processor/internal/application/ports/inbound/tasksprocessor"
	"task-processor/internal/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessTask_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(taskrepo.MockTaskRepository)
	mockFailedRepo := new(failedtaskrepo.MockFailedTaskRepo)
	mockTx := new(txmanager.MockTxManager)
	mockRand := new(random.MockRandom)

	task := &domain.Task{ID: uuid.New(), Attempts: 0, MaxAttempts: 3}
	req := &tasksprocessor.ProcessTasksRequest{SuccessRate: 1.0, MinDelayMS: 0, MaxDelayMS: 0}

	mockRand.On("Float64").Return(0.5)
	mockRepo.On("MarkAsProcessed", ctx, task.ID).Return(nil)

	pr := NewSingleProcessor(mockRepo, mockFailedRepo, mockTx, mockRand)
	success, err := pr.ProcessTask(ctx, task, req)

	assert.True(t, success)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockRand.AssertExpectations(t)
}

func TestProcessTask_Failure(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(taskrepo.MockTaskRepository)
	mockFailedRepo := new(failedtaskrepo.MockFailedTaskRepo)
	mockTx := new(txmanager.MockTxManager)
	mockRand := new(random.MockRandom)

	task := &domain.Task{ID: uuid.New(), Attempts: 1, MaxAttempts: 3}
	req := &tasksprocessor.ProcessTasksRequest{SuccessRate: 0.0, MinDelayMS: 0, MaxDelayMS: 0}

	mockRand.On("Float64").Return(1.0)
	mockRepo.On("MarkAsFailed", ctx, task.ID, mock.Anything).Return(nil)

	pr := NewSingleProcessor(mockRepo, mockFailedRepo, mockTx, mockRand)
	success, err := pr.ProcessTask(ctx, task, req)

	assert.False(t, success)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockRand.AssertExpectations(t)
}

func TestProcessTask_MaxAttemptsExceeded(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(taskrepo.MockTaskRepository)
	mockFailedRepo := new(failedtaskrepo.MockFailedTaskRepo)
	mockTx := new(txmanager.MockTxManager)
	mockRand := new(random.MockRandom)

	task := &domain.Task{ID: uuid.New(), Attempts: 3, MaxAttempts: 3}

	mockTx.On("WithTransaction", ctx, mock.Anything).Return(nil)
	mockRepo.On("Delete", ctx, task.ID).Return(nil)
	mockFailedRepo.On("Create", ctx, task).Return(nil)

	pr := NewSingleProcessor(mockRepo, mockFailedRepo, mockTx, mockRand)
	success, err := pr.ProcessTask(ctx, task, &tasksprocessor.ProcessTasksRequest{})

	assert.False(t, success)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockFailedRepo.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}