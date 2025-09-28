package taskcontroller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"task-processor/internal/infrastructure/adapters/inbound/httpserver/task/dto"
	"task-processor/internal/infrastructure/adapters/inbound/httpserver/utils"

	"github.com/stretchr/testify/require"
)

func TestBatchCreateHandler_Success(t *testing.T) {
	controller, _, cleanup := setupTestDependencies(t)
	defer cleanup()
	router := setupRouter(controller)

	reqBody := dto.BatchCreateTasksRequest{Count: 5}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/batch-create", bytes.NewReader(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Result().StatusCode)

	// First, decode the HTTPResponse wrapper
	var httpResp utils.HTTPResponse
	err := json.NewDecoder(w.Body).Decode(&httpResp)
	require.NoError(t, err)

	// Then, decode the Data field into BatchCreateTasksResponse
	dataBytes, err := json.Marshal(httpResp.Data)
	require.NoError(t, err)

	var resp dto.BatchCreateTasksResponse
	err = json.Unmarshal(dataBytes, &resp)
	require.NoError(t, err)

	// Assert that exactly 5 IDs were created
	require.Len(t, resp.IDs, 5)
}


func TestBatchCreateHandler_InvalidJSON(t *testing.T) {
	controller, _, cleanup := setupTestDependencies(t)
	defer cleanup()
	router := setupRouter(controller)

	// Send invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/batch-create", bytes.NewReader([]byte(`invalid-json`)))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Expect HTTP 400 Bad Request
	require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

	// Optionally, check error message in HTTPResponse
	var httpResp utils.HTTPResponse
	err := json.NewDecoder(w.Body).Decode(&httpResp)
	require.NoError(t, err)
	require.False(t, httpResp.Success)
	require.Contains(t, httpResp.Message, "Invalid JSON")
}
func TestBatchCreateHandler_InvalidCount(t *testing.T) {
	controller, _, cleanup := setupTestDependencies(t)
	defer cleanup()
	router := setupRouter(controller)

	invalidCounts := []int{0, 51}

	for _, count := range invalidCounts {
		// Prepare request with invalid count
		reqBody := dto.BatchCreateTasksRequest{Count: count}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/batch-create", bytes.NewReader(body))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Expect HTTP 400 Bad Request
		require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

		// Decode HTTPResponse to check validation errors
		var httpResp utils.HTTPResponse
		err := json.NewDecoder(w.Body).Decode(&httpResp)
		require.NoError(t, err)
		require.False(t, httpResp.Success)
		require.Contains(t, httpResp.Message, "Validation failed")
	}
}

func TestBatchCreateHandler_LargeCount(t *testing.T) {
	controller, _, cleanup := setupTestDependencies(t)
	defer cleanup()
	router := setupRouter(controller)

	reqBody := dto.BatchCreateTasksRequest{Count: 50} 
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/batch-create", bytes.NewReader(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Result().StatusCode)

	// Decode the HTTPResponse wrapper
	var httpResp utils.HTTPResponse
	err := json.NewDecoder(w.Body).Decode(&httpResp)
	require.NoError(t, err)
	require.True(t, httpResp.Success)

	// Decode the Data field into BatchCreateTasksResponse
	dataBytes, err := json.Marshal(httpResp.Data)
	require.NoError(t, err)

	var resp dto.BatchCreateTasksResponse
	err = json.Unmarshal(dataBytes, &resp)
	require.NoError(t, err)

	// Assert that all 50 IDs were created
	require.Len(t, resp.IDs, 50)
}
