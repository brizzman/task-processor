package failedtaskrepo

import (
	"context"
	"task-processor/internal/domain"

	"github.com/stretchr/testify/mock"
)

type MockFailedTaskRepo struct { 
	mock.Mock 
}

func (m *MockFailedTaskRepo) Create(ctx context.Context, task *domain.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}