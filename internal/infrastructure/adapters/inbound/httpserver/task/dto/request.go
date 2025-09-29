package dto

import "task-processor/internal/application/ports/inbound/tasksprocessor"

// @Description Request payload for task processing
type ProcessTasksRequest struct {
	// @Description Number of tasks to process (1-50)
	// @Example     10
	Limit int `json:"limit" validate:"min=1,max=50"`
	
	// @Description Minimum processing delay in milliseconds
	// @Example     100
	MinDelayMS int `json:"min_delay_ms" validate:"min=0"`
	
	// @Description Maximum processing delay in milliseconds
	// @Example     500
	MaxDelayMS int `json:"max_delay_ms" validate:"min=0,gtefield=MinDelayMS"`
	
	// @Description Success rate probability (0.0 - 1.0)
	// @Example     0.8
	SuccessRate float64 `json:"success_rate" validate:"min=0,max=1"`
}

// ToDomain converts HTTP DTO to domain request
func (r *ProcessTasksRequest) ToDomainProcess() *tasksprocessor.ProcessTasksRequest {
	return &tasksprocessor.ProcessTasksRequest{
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
	Count int `json:"count" validate:"required,min=1,max=50"`
}

// ToDomain converts HTTP DTO to domain request (use case input)
func (r *BatchCreateTasksRequest) ToDomainBatchCreate() *tasksprocessor.BatchCreateTasksRequest {
	return &tasksprocessor.BatchCreateTasksRequest{
		Count: r.Count,
	}
}
