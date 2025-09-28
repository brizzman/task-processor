package tasksprocessor

type ProcessTasksRequest struct {
	// Limit defines the maximum number of tasks to acquire and process.
	Limit          int   
	// MinDelayMS specifies the minimum delay in milliseconds to simulate 
	// processing time for each task. If set to 0, no delay is applied.
	MinDelayMS     int    
	// MaxDelayMS specifies the maximum delay in milliseconds to simulate
	// processing time for each task. If set to 0, no delay is applied.
	MaxDelayMS     int    
	// SuccessRate determines the probability of successful task processing.
	// Value must be between 0.0 (0% success) and 1.0 (100% success).
	SuccessRate    float64 
}

type ProcessTasksResponse struct {
	// ProcessedCount indicates the total number of tasks that were 
	// successfully acquired and attempted to be processed.
	// This count includes both successful and failed processing attempts.
	ProcessedCount int 
	// SuccessCount shows how many tasks were successfully processed
	// and marked with PROCESSED status in the database.
	SuccessCount   int 
	// FailedCount indicates how many tasks failed during processing
	// and were marked with FAILED status in the database.
	FailedCount    int 
}

// BatchCreateTasksRequest defines the input for creating multiple tasks at once.
type BatchCreateTasksRequest struct {
	// Count specifies the number of tasks to create in a single batch.
	// Must be greater than 0.
	Count int
}