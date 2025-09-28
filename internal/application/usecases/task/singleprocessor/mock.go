package singleprocessor

import (
	"context"
	"task-processor/internal/application/ports/inbound/tasksprocessor"
	"task-processor/internal/domain"

	"github.com/stretchr/testify/mock"
)

type MockSingleProcessor struct {
	mock.Mock
}

func (m *MockSingleProcessor) ProcessTask(
	ctx context.Context, 
	task *domain.Task, 
	req *tasksprocessor.ProcessTasksRequest,
) (bool, error) {
	args := m.Called(ctx, task, req)
	return args.Bool(0), args.Error(1)
}
