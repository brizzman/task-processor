package domain

import (
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	StatusNew         TaskStatus = "NEW"
	StatusProcessing  TaskStatus = "PROCESSING"
	StatusProcessed   TaskStatus = "PROCESSED"
	StatusFailed      TaskStatus = "FAILED"
)

type Task struct {
    // Unique identifier for the task (UUID for distributed systems)
    ID                  uuid.UUID   
    
    // Current state of the task (NEW, PROCESSING, PROCESSED, FAILED)
    Status              TaskStatus  
    
    // When the task was created (for sorting)
    CreatedAt           time.Time   
    
    // Last modification timestamp (for detecting stuck tasks)
    UpdatedAt           time.Time 

    // Number of processing attempts (for retry logic)
    Attempts    int
    
    // Maximum allowed attempts (prevents infinite retries)
    MaxAttempts int

    // Last error which happened.
    ErrorMessage        string
}