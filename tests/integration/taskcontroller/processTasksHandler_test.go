package taskcontroller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"task-processor/internal/infrastructure/adapters/inbound/httpserver/task/dto"
	"task-processor/internal/infrastructure/adapters/inbound/httpserver/utils"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProcessTasksHandler_Success(t *testing.T) {
	controller, ctx, cleanup := setupTestDependencies(t)
	defer cleanup()
	router := setupRouter(controller)

	// Pre-create tasks so there is something to process
	ids, err := controller.TaskUseCases.Creator.CreateTasksBatch(ctx, 5)
	require.NoError(t, err)
	require.Len(t, ids, 5)

	// Prepare a valid request payload
	reqBody := dto.ProcessTasksRequest{
		Limit:       5,
		MinDelayMS:  10,
		MaxDelayMS:  50,
		SuccessRate: 1.0,
	}
	body, _ := json.Marshal(reqBody)

	// Create HTTP POST request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/process", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Expect HTTP 200 OK
	require.Equal(t, http.StatusOK, w.Result().StatusCode)

	// Decode HTTPResponse wrapper
	var httpResp utils.HTTPResponse
	err = json.NewDecoder(w.Body).Decode(&httpResp)
	require.NoError(t, err)
	require.True(t, httpResp.Success)

	// Decode Data field into ProcessTasksResponse
	dataBytes, err := json.Marshal(httpResp.Data)
	require.NoError(t, err)

	var resp dto.ProcessTasksResponse
	err = json.Unmarshal(dataBytes, &resp)
	require.NoError(t, err)

	// Assert that all tasks were processed successfully
	require.Equal(t, 5, resp.ProcessedCount)
	require.Equal(t, 5, resp.SuccessCount)
	require.Equal(t, 0, resp.FailedCount)
}

func TestProcessTasksHandler_InvalidJSON(t *testing.T) {
	controller, _, cleanup := setupTestDependencies(t)
	defer cleanup()
	router := setupRouter(controller)

	// Send invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/process", bytes.NewReader([]byte(`invalid-json`)))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Expect HTTP 400 Bad Request
	require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

	// Decode HTTPResponse to check error message
	var httpResp utils.HTTPResponse
	err := json.NewDecoder(w.Body).Decode(&httpResp)
	require.NoError(t, err)
	require.False(t, httpResp.Success)
	require.Contains(t, httpResp.Message, "Invalid JSON")
}

func TestProcessTasksHandler_InvalidPayload(t *testing.T) {
	controller, _, cleanup := setupTestDependencies(t)
	defer cleanup()
	router := setupRouter(controller)

	// Prepare invalid payloads for different validation errors
	invalidPayloads := []dto.ProcessTasksRequest{
		{Limit: 0, MinDelayMS: 10, MaxDelayMS: 50, SuccessRate: 0.5},  // Limit < 1
		{Limit: 5, MinDelayMS: -1, MaxDelayMS: 50, SuccessRate: 0.5},  // MinDelayMS < 0
		{Limit: 5, MinDelayMS: 10, MaxDelayMS: -10, SuccessRate: 0.5}, // MaxDelayMS < 0
		{Limit: 5, MinDelayMS: 10, MaxDelayMS: 50, SuccessRate: -0.1}, // SuccessRate < 0
		{Limit: 5, MinDelayMS: 10, MaxDelayMS: 50, SuccessRate: 1.1},  // SuccessRate > 1
	}

	for _, payload := range invalidPayloads {
		// Marshal invalid payload to JSON
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/process", bytes.NewReader(body))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Expect HTTP 400 Bad Request
		require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

		// Decode HTTPResponse to check validation error message
		var httpResp utils.HTTPResponse
		err := json.NewDecoder(w.Body).Decode(&httpResp)
		require.NoError(t, err)
		require.False(t, httpResp.Success)
		require.Contains(t, httpResp.Message, "Validation failed")
	}
}

func TestProcessTasksHandler_MultipleRequests(t *testing.T) {
	controller, ctx, cleanup := setupTestDependencies(t)
	defer cleanup()
	router := setupRouter(controller)

	// Pre-create tasks to be processed
	ids, err := controller.TaskUseCases.Creator.CreateTasksBatch(ctx, 10)
	require.NoError(t, err)
	require.Len(t, ids, 10)

	// Send multiple requests in a loop
	for i := 0; i < 2; i++ {
		reqBody := dto.ProcessTasksRequest{
			Limit:       5,
			MinDelayMS:  10,
			MaxDelayMS:  50,
			SuccessRate: 0.8,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/process", bytes.NewReader(body))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Expect HTTP 200 OK
		require.Equal(t, http.StatusOK, w.Result().StatusCode)

		// Decode HTTPResponse wrapper
		var httpResp utils.HTTPResponse
		err := json.NewDecoder(w.Body).Decode(&httpResp)
		require.NoError(t, err)
		require.True(t, httpResp.Success)

		// Decode Data field
		dataBytes, err := json.Marshal(httpResp.Data)
		require.NoError(t, err)

		var resp dto.ProcessTasksResponse
		err = json.Unmarshal(dataBytes, &resp)
		require.NoError(t, err)

		// Assert correct number of tasks processed in each request
		require.Equal(t, 5, resp.ProcessedCount)
	}
}
