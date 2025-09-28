package creator

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockCreator struct {
	mock.Mock
}

func (m *MockCreator) CreateTasksBatch(ctx context.Context, count int) ([]uuid.UUID, error) {
	args := m.Called(ctx, count)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}
