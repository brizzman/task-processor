package task

import (
	"context"
)

// TaskProcessor defines the interface for processing tasks in the system.
type TaskProcessor interface {
	// ProcessTasks processes a batch of tasks according to the given request parameters.
	ProcessTasks(ctx context.Context, request *ProcessTasksRequest) (*ProcessTasksResponse, error)
}

