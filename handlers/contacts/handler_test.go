package contacts

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"pec2-backend/testutils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	testutils.InitTestMain()

	log.SetOutput(io.Discard)

	exitCode := m.Run()

	log.SetOutput(os.Stdout)

	os.Exit(exitCode)
}

func TestCreateContact_Success(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "contacts" (.+) RETURNING "id"`).
		WillReturnRows(mock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow("123e4567-e89b-12d3-a456-426614174000", time.Now(), time.Now()))
	mock.ExpectCommit()

	r := testutils.SetupTestRouter()
	r.POST("/contact", CreateContact)

	contactData := map[string]string{
		"firstName": "Jean",
		"lastName":  "Dupont",
		"email":     "jean.dupont@example.com",
		"subject":   "Demande d'information",
		"message":   "J'aimerais avoir plus d'informations sur vos services.",
	}
	jsonData, _ := json.Marshal(contactData)

	req, _ := http.NewRequest(http.MethodPost, "/contact", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)

	var respBody map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.Equal(t, "Contact request submitted successfully", respBody["message"])
	assert.NotNil(t, respBody["id"])
}

func TestCreateContact_EmptyFirstName(t *testing.T) {
	r := testutils.SetupTestRouter()
	r.POST("/contact", CreateContact)

	contactData := map[string]string{
		"firstName": "",
		"lastName":  "Dupont",
		"email":     "jean.dupont@example.com",
		"subject":   "Demande d'information",
		"message":   "J'aimerais avoir plus d'informations sur vos services.",
	}
	jsonData, _ := json.Marshal(contactData)

	req, _ := http.NewRequest(http.MethodPost, "/contact", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var respBody map[string]string
	json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.Contains(t, respBody["error"], "Field validation for 'FirstName' failed")
}

func TestCreateContact_EmptyLastName(t *testing.T) {
	r := testutils.SetupTestRouter()
	r.POST("/contact", CreateContact)

	contactData := map[string]string{
		"firstName": "Jean",
		"lastName":  "",
		"email":     "jean.dupont@example.com",
		"subject":   "Demande d'information",
		"message":   "J'aimerais avoir plus d'informations sur vos services.",
	}
	jsonData, _ := json.Marshal(contactData)

	req, _ := http.NewRequest(http.MethodPost, "/contact", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var respBody map[string]string
	json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.Contains(t, respBody["error"], "Field validation for 'LastName' failed")
}

func TestCreateContact_EmptyEmail(t *testing.T) {
	r := testutils.SetupTestRouter()
	r.POST("/contact", CreateContact)

	contactData := map[string]string{
		"firstName": "Jean",
		"lastName":  "Dupont",
		"email":     "",
		"subject":   "Demande d'information",
		"message":   "J'aimerais avoir plus d'informations sur vos services.",
	}
	jsonData, _ := json.Marshal(contactData)

	req, _ := http.NewRequest(http.MethodPost, "/contact", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var respBody map[string]string
	json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.Contains(t, respBody["error"], "Field validation for 'Email' failed")
}

func TestCreateContact_InvalidEmailFormat(t *testing.T) {
	r := testutils.SetupTestRouter()
	r.POST("/contact", CreateContact)

	contactData := map[string]string{
		"firstName": "Jean",
		"lastName":  "Dupont",
		"email":     "invalid-email",
		"subject":   "Demande d'information",
		"message":   "J'aimerais avoir plus d'informations sur vos services.",
	}
	jsonData, _ := json.Marshal(contactData)

	req, _ := http.NewRequest(http.MethodPost, "/contact", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var respBody map[string]string
	json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.Contains(t, respBody["error"], "Field validation for 'Email' failed")
}

func TestCreateContact_EmptySubject(t *testing.T) {
	r := testutils.SetupTestRouter()
	r.POST("/contact", CreateContact)

	contactData := map[string]string{
		"firstName": "Jean",
		"lastName":  "Dupont",
		"email":     "jean.dupont@example.com",
		"subject":   "",
		"message":   "J'aimerais avoir plus d'informations sur vos services.",
	}
	jsonData, _ := json.Marshal(contactData)

	req, _ := http.NewRequest(http.MethodPost, "/contact", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var respBody map[string]string
	json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.Contains(t, respBody["error"], "Field validation for 'Subject' failed")
}

func TestCreateContact_EmptyMessage(t *testing.T) {
	r := testutils.SetupTestRouter()
	r.POST("/contact", CreateContact)

	contactData := map[string]string{
		"firstName": "Jean",
		"lastName":  "Dupont",
		"email":     "jean.dupont@example.com",
		"subject":   "Demande d'information",
		"message":   "",
	}
	jsonData, _ := json.Marshal(contactData)

	req, _ := http.NewRequest(http.MethodPost, "/contact", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var respBody map[string]string
	json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.Contains(t, respBody["error"], "Field validation for 'Message' failed")
}

func TestCreateContact_DatabaseError(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "contacts" (.+) RETURNING "id"`).
		WillReturnError(gorm.ErrInvalidDB)
	mock.ExpectRollback()

	r := testutils.SetupTestRouter()
	r.POST("/contact", CreateContact)

	contactData := map[string]string{
		"firstName": "Jean",
		"lastName":  "Dupont",
		"email":     "jean.dupont@example.com",
		"subject":   "Demande d'information",
		"message":   "J'aimerais avoir plus d'informations sur vos services.",
	}
	jsonData, _ := json.Marshal(contactData)

	req, _ := http.NewRequest(http.MethodPost, "/contact", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestGetAllContacts_Success(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery(`SELECT \* FROM "contacts" ORDER BY submitted_at DESC`).
		WillReturnRows(
			mock.NewRows([]string{"id", "first_name", "last_name", "email", "subject", "message", "submitted_at", "created_at", "updated_at", "deleted_at"}).
				AddRow("123e4567-e89b-12d3-a456-426614174000", "Jean", "Dupont", "jean.dupont@example.com", "Sujet 1", "Message 1", now, now, now, nil).
				AddRow("223e4567-e89b-12d3-a456-426614174000", "Marie", "Martin", "marie.martin@example.com", "Sujet 2", "Message 2", now, now, now, nil),
		)

	r := testutils.SetupTestRouter()
	r.GET("/contacts", GetAllContacts)

	req, _ := http.NewRequest(http.MethodGet, "/contacts", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var contacts []map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &contacts)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(contacts), "there should be 2 contacts in the response")

	if len(contacts) >= 2 {
		assert.Equal(t, "Jean", contacts[0]["firstName"])
		assert.Equal(t, "Dupont", contacts[0]["lastName"])
		assert.Equal(t, "jean.dupont@example.com", contacts[0]["email"])

		assert.Equal(t, "Marie", contacts[1]["firstName"])
		assert.Equal(t, "Martin", contacts[1]["lastName"])
		assert.Equal(t, "marie.martin@example.com", contacts[1]["email"])
	}
}

func TestGetAllContacts_EmptyList(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT \* FROM "contacts" ORDER BY submitted_at DESC`).
		WillReturnRows(mock.NewRows([]string{"id", "first_name", "last_name", "email", "subject", "message", "submitted_at", "created_at", "updated_at", "deleted_at"}))

	r := testutils.SetupTestRouter()
	r.GET("/contacts", GetAllContacts)

	req, _ := http.NewRequest(http.MethodGet, "/contacts", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var contacts []map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &contacts)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(contacts), "the contacts list should be empty")
}

func TestGetAllContacts_DatabaseError(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT \* FROM "contacts" ORDER BY submitted_at DESC`).
		WillReturnError(gorm.ErrInvalidDB)

	r := testutils.SetupTestRouter()
	r.GET("/contacts", GetAllContacts)

	req, _ := http.NewRequest(http.MethodGet, "/contacts", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	var respBody map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.NoError(t, err)

	errorMsg, exists := respBody["error"]
	assert.True(t, exists, "the key 'error' should exist in the response")
	assert.Contains(t, errorMsg, "invalid db")
}
