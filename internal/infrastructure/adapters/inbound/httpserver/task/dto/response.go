package dto

import (
	"task-processor/internal/application/ports/inbound/tasksprocessor"
	"github.com/google/uuid"

)

// @Description Response after task processing
type ProcessTasksResponse struct {
	// @Description Total number of tasks processed
	// @Example     10
	ProcessedCount int `json:"processed_count"`
	
	// @Description Number of successfully processed tasks
	// @Example     8
	SuccessCount int `json:"success_count"`
	
	// @Description Number of failed tasks
	// @Example     2
	FailedCount int `json:"failed_count"`
}

// FromDomain converts domain response to HTTP DTO
func FromDomainProcess(domainResponse *tasksprocessor.ProcessTasksResponse) *ProcessTasksResponse {
	return &ProcessTasksResponse{
		ProcessedCount: domainResponse.ProcessedCount,
		SuccessCount:   domainResponse.SuccessCount,
		FailedCount:    domainResponse.FailedCount,
	}
}

// @Description Response payload for batch task creation
type BatchCreateTasksResponse struct {
	// @Description List of created task IDs
	IDs []string  `json:"ids"`
}

func FromDomainBatchCreate(ids []uuid.UUID) *BatchCreateTasksResponse {
    strIDs := make([]string, len(ids))
    for i, id := range ids {
        strIDs[i] = id.String()
    }
    return &BatchCreateTasksResponse{
        IDs: strIDs,
    }
}