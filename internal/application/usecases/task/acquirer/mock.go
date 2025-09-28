package acquirer

import (
	"context"
	"task-processor/internal/domain"

	"github.com/stretchr/testify/mock"
)

type MockAcquirer struct {
	mock.Mock
}

func (m *MockAcquirer) AcquireTasks(ctx context.Context, limit int) ([]*domain.Task, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*domain.Task), args.Error(1)
}
