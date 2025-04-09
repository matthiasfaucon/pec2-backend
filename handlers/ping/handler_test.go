package ping

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"pec2-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandlePing(t *testing.T) {
	// Switch to test mode
	gin.SetMode(gin.TestMode)

	// Setup
	r := gin.New()
	handler := New()
	r.GET("/ping", handler.HandlePing)

	// Create test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)

	// Perform request
	r.ServeHTTP(w, req)

	// Assert status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)

	// Assert response structure
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Ping successful", response.Message)

	// Assert response data
	data, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "pong", data["message"])
}
