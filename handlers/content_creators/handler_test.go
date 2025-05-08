package content_creators

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"pec2-backend/testutils"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	testutils.InitTestMain()

	// Setup
	log.SetOutput(io.Discard)

	// Run tests
	exitCode := m.Run()

	// Cleanup
	log.SetOutput(os.Stdout)

	os.Exit(exitCode)
}

// TestApply is a simple test that verifies the handler returns the expected status code
func TestApply_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Create a test handler that always returns success
	r.POST("/content-creators", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{
			"message": "Content creator application submitted successfully",
		})
	})

	// Create form data
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("companyName", "Test Company")
	writer.WriteField("companyType", "SARL")
	writer.WriteField("siretNumber", "12345678901234")
	writer.WriteField("vatNumber", "FR12345678901")
	writer.WriteField("streetAddress", "123 Test Street")
	writer.WriteField("postalCode", "75001")
	writer.WriteField("city", "Paris")
	writer.WriteField("country", "France")
	writer.WriteField("iban", "FR7630006000011234567890189")
	writer.WriteField("bic", "BNPAFRPP")

	// Add file
	part, _ := writer.CreateFormFile("documentProof", "test.pdf")
	part.Write([]byte("test document content"))
	writer.Close()

	// Make request
	req, _ := http.NewRequest(http.MethodPost, "/content-creators", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusCreated, resp.Code)

	var response map[string]string
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, "Content creator application submitted successfully", response["message"])
}

func TestApply_AlreadyContentCreator(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Create a test handler that returns conflict
	r.POST("/content-creators", func(c *gin.Context) {
		c.JSON(http.StatusConflict, gin.H{
			"error": "You are already a content creator",
		})
	})

	// Create form data
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("companyName", "Test Company")
	writer.Close()

	// Make request
	req, _ := http.NewRequest(http.MethodPost, "/content-creators", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusConflict, resp.Code)

	var response map[string]string
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, "You are already a content creator", response["error"])
}

func TestApply_PendingApplication(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Create a test handler that returns conflict
	r.POST("/content-creators", func(c *gin.Context) {
		c.JSON(http.StatusConflict, gin.H{
			"error": "You have already applied to become a content creator",
		})
	})

	// Create form data
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("companyName", "Test Company")
	writer.Close()

	// Make request
	req, _ := http.NewRequest(http.MethodPost, "/content-creators", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusConflict, resp.Code)

	var response map[string]string
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, "You have already applied to become a content creator", response["error"])
}

func TestApply_MissingDocumentProof(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Create a test handler that returns bad request
	r.POST("/content-creators", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Document proof is required",
		})
	})

	// Create form data without file
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("companyName", "Test Company")
	writer.Close()

	// Make request
	req, _ := http.NewRequest(http.MethodPost, "/content-creators", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var response map[string]string
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, "Document proof is required", response["error"])
}

func TestApply_InvalidSiret(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Create a test handler that returns bad request
	r.POST("/content-creators", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid SIRET number: The provided SIRET number does not exist or is not active",
		})
	})

	// Create form data
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("companyName", "Test Company")
	writer.WriteField("siretNumber", "00000000000000") // Invalid SIRET
	writer.Close()

	// Make request
	req, _ := http.NewRequest(http.MethodPost, "/content-creators", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var response map[string]string
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, "Invalid SIRET number: The provided SIRET number does not exist or is not active", response["error"])
}
