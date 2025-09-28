package acquirer

import (
	"context"
	"task-processor/internal/application/ports/outbound/persistence/taskrepo"
	"task-processor/internal/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAcquireTasks_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(taskrepo.MockTaskRepository)
	
	tasks := []*domain.Task{{ID: uuid.New()}, {ID: uuid.New()}}
	mockRepo.On("AcquireTasks", ctx, 2).Return(tasks, nil)

	aq := NewAcquirer(mockRepo)
	result, err := aq.AcquireTasks(ctx, 2)

	assert.NoError(t, err)
	assert.Equal(t, tasks, result)
	mockRepo.AssertExpectations(t)
}
