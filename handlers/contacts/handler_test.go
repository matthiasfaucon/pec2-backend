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

	// Simuler l'insertion d'un contact
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "contacts" (.+) RETURNING "id"`).
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow("123e4567-e89b-12d3-a456-426614174000"))
	mock.ExpectCommit()

	r := testutils.SetupTestRouter()
	r.POST("/contact", CreateContact)

	// Données valides pour la création d'un contact
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

	// Vérification de la réponse
	assert.Equal(t, http.StatusCreated, resp.Code)

	// Vérification du corps de la réponse
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
