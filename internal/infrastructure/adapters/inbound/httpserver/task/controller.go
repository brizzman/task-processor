package task

import (
	"encoding/json"
	"net/http"

	proc "task-processor/internal/application/ports/inbound/task"
	"task-processor/internal/application/usecases/task"
	"task-processor/internal/infrastructure/adapters/inbound/httpserver/task/dto"
	"task-processor/internal/infrastructure/adapters/inbound/httpserver/utils"

	"github.com/go-chi/chi/v5"
)

// Controller handles HTTP requests for task processing
type Controller struct {
	taskProcessor  proc.TaskProcessor
	taskCreator   *task.TaskCreator
}

// NewController creates a new task controller
func NewController(taskProcessor proc.TaskProcessor, taskCreator *task.TaskCreator) *Controller {
	return &Controller{
		taskProcessor: taskProcessor,
		taskCreator:   taskCreator,
	}
}

// RegisterRoutes registers routes for Controller
func (c *Controller) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/tasks", func(r chi.Router) {
		r.Post("/process", c.ProcessTasksHandler)
		r.Post("/batch-create", c.BatchCreateHandler)
	})
}

// @Summary      Process multiple tasks
// @Description  Acquires and processes tasks with configurable parameters
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Param        request body dto.ProcessTasksRequest true "Processing parameters"
// @Success      200 {object} dto.ProcessTasksResponse
// @Failure      400 {object} utils.HTTPResponse
// @Failure      500 {object} utils.HTTPResponse
// @Router       /api/v1/tasks/process [post]
func (c *Controller) ProcessTasksHandler(w http.ResponseWriter, r *http.Request) {
	// Parse and validate request
	var req dto.ProcessTasksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, r, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		utils.SendValidationError(w, r, err)
		return
	}

	// Convert to domain request
	domainReq := req.ToDomainProcess()

	// Process tasks
	response, err := c.taskProcessor.ProcessTasks(r.Context(), domainReq)
	if err != nil {
		utils.SendError(w, r, "Processing failed", http.StatusInternalServerError)
		return
	}

	// Convert to HTTP response
	httpResponse := dto.FromDomainProcess(response)

	// Send success response
	utils.SendSuccess(w, r, httpResponse, http.StatusOK)
}

// @Summary      Batch create tasks
// @Description  Creates multiple tasks in a single operation
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Param        request body dto.BatchCreateTasksRequest true "Tasks to create"
// @Success      200 {object} dto.BatchCreateTasksResponse
// @Failure      400 {object} utils.HTTPResponse
// @Failure      500 {object} utils.HTTPResponse
// @Router       /api/v1/tasks/batch-create [post]
func (c *Controller) BatchCreateHandler(w http.ResponseWriter, r *http.Request) {
	var req dto.BatchCreateTasksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, r, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.SendValidationError(w, r, err)
		return
	}

	ids, err := c.taskCreator.CreateTasksBatch(r.Context(), req.Count)
	if err != nil {
		utils.SendError(w, r, "Failed to create tasks", http.StatusInternalServerError)
		return
	}

	httpResponse := dto.FromDomainBatchCreate(ids)

	utils.SendSuccess(w, r, httpResponse, http.StatusOK)
}
