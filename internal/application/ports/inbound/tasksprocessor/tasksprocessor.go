package tasksprocessor

import (
	"context"
)

// TasksProcessor defines the interface for processing tasks in the system.
type TasksProcessor interface {
	// ProcessTasks processes a batch of tasks according to the given request parameters.
	ProcessTasks(ctx context.Context, request *ProcessTasksRequest) (*ProcessTasksResponse, error)
}

