package dto

import "task-processor/internal/application/ports/inbound/task"

// @Description Request payload for task processing
type ProcessTasksRequest struct {
	// @Description Number of tasks to process (1-1000)
	// @Example     10
	Limit int `json:"limit" validate:"required,min=1,max=1000"`
	
	// @Description Minimum processing delay in milliseconds
	// @Example     100
	MinDelayMS int `json:"min_delay_ms" validate:"min=0"`
	
	// @Description Maximum processing delay in milliseconds
	// @Example     500
	MaxDelayMS int `json:"max_delay_ms" validate:"min=0"`
	
	// @Description Success rate probability (0.0 - 1.0)
	// @Example     0.8
	SuccessRate float64 `json:"success_rate" validate:"required,min=0,max=1"`
}

// ToDomain converts HTTP DTO to domain request
func (r *ProcessTasksRequest) ToDomainProcess() *task.ProcessTasksRequest {
	return &task.ProcessTasksRequest{
		Limit:       r.Limit,
		MinDelayMS:  r.MinDelayMS,
		MaxDelayMS:  r.MaxDelayMS,
		SuccessRate: r.SuccessRate,
	}
}

// @Description Request payload for batch task creation
type BatchCreateTasksRequest struct {
	// @Description Number of tasks to create
	// @Example     5
	Count int `json:"count" validate:"required,min=1"`
}

// ToDomain converts HTTP DTO to domain request (use case input)
func (r *BatchCreateTasksRequest) ToDomainBatchCreate() *task.BatchCreateTasksRequest {
	return &task.BatchCreateTasksRequest{
		Count: r.Count,
	}
}
