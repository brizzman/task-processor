package txmanager

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockTxManager struct{ 
	mock.Mock 
}

func (m *MockTxManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	m.Called(ctx, fn)
	return fn(ctx)
}