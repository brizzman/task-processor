package taskrepo

import (
	"context"
	"task-processor/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) BatchCreate(ctx context.Context, tasks []*domain.Task) ([]uuid.UUID, error) {
	args := m.Called(ctx, tasks)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (m *MockTaskRepository) AcquireTasks(ctx context.Context, limit int) ([]*domain.Task, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*domain.Task), args.Error(1)
}

func (m *MockTaskRepository) MarkAsProcessed(ctx context.Context, taskID uuid.UUID) error {
	args := m.Called(ctx, taskID)
	return args.Error(0)
}

func (m *MockTaskRepository) MarkAsFailed(ctx context.Context, taskID uuid.UUID, errorMsg string) error {
	args := m.Called(ctx, taskID, errorMsg)
	return args.Error(0)
}

func (m *MockTaskRepository) Delete(ctx context.Context, taskID uuid.UUID) error {
	args := m.Called(ctx, taskID)
	return args.Error(0)
}