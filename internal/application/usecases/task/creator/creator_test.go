package creator

import (
	"context"
	"errors"
	"task-processor/internal/application/ports/outbound/persistence/taskrepo"
	"task-processor/internal/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTaskCreator_CreateTasksBatch(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(taskrepo.MockTaskRepository)

	expectedIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	taskCount := len(expectedIDs)

	mockRepo.On("BatchCreate", ctx, mock.MatchedBy(func(tasks []*domain.Task) bool {
		return len(tasks) == taskCount
	})).Return(expectedIDs, nil)

	creator := NewCreator(mockRepo)

	ids, err := creator.CreateTasksBatch(ctx, taskCount)

	assert.NoError(t, err)
	assert.Equal(t, expectedIDs, ids)

	mockRepo.AssertExpectations(t)
}

func TestTaskCreator_CreateTasksBatch_Error(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(taskrepo.MockTaskRepository)

	taskCount := 3

	mockRepo.On("BatchCreate", ctx, mock.Anything).Return([]uuid.UUID(nil), errors.New("db error"))

	creator := NewCreator(mockRepo)

	ids, err := creator.CreateTasksBatch(ctx, taskCount)

	assert.Error(t, err)
	assert.Nil(t, ids)
	mockRepo.AssertExpectations(t)
}

func TestTaskCreator_CreateTasksBatch_ZeroTasks(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(taskrepo.MockTaskRepository)

	mockRepo.On("BatchCreate", ctx, []*domain.Task{}).Return([]uuid.UUID{}, nil)
	
	creator := NewCreator(mockRepo)

	ids, err := creator.CreateTasksBatch(ctx, 0)

	assert.NoError(t, err)
	assert.Empty(t, ids)
	mockRepo.AssertExpectations(t)
}

func TestTaskCreator_CreateTasksBatch_StatusCheck(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(taskrepo.MockTaskRepository)

	expectedIDs := []uuid.UUID{uuid.New(), uuid.New()}
	taskCount := len(expectedIDs)

	mockRepo.On("BatchCreate", ctx, mock.MatchedBy(func(tasks []*domain.Task) bool {
		for _, t := range tasks {
			if t.Status != domain.StatusNew {
				return false
			}
		}
		return true
	})).Return(expectedIDs, nil)

	creator := NewCreator(mockRepo)

	ids, err := creator.CreateTasksBatch(ctx, taskCount)

	assert.NoError(t, err)
	assert.Equal(t, expectedIDs, ids)
	mockRepo.AssertExpectations(t)
}
