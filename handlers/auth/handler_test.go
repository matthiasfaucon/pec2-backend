package auth

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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	// Initialiser l'environnement de test
	testutils.InitTestMain()

	// Redirection des logs standards pendant les tests
	log.SetOutput(io.Discard)

	// Exécution de tous les tests
	exitCode := m.Run()

	// Restauration des logs standard
	log.SetOutput(os.Stdout)

	os.Exit(exitCode)
}

func TestCreateUser_Success(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT (.+) FROM "users" WHERE email = (.+) AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT (.+)`).
		WithArgs("test@example.com", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users" (.+) RETURNING "id"`).
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	r := testutils.SetupTestRouter()
	r.POST("/user", CreateUser)

	userData := map[string]string{
		"email":    "test@example.com",
		"password": "Password123",
	}
	jsonData, _ := json.Marshal(userData)

	req, _ := http.NewRequest(http.MethodPost, "/user", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var respBody map[string]string
	json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.Equal(t, "User created successfully", respBody["message"])
	assert.Equal(t, "test@example.com", respBody["email"])
}

func TestCreateUser_EmptyEmail(t *testing.T) {
	r := testutils.SetupTestRouter()
	r.POST("/user", CreateUser)

	userData := map[string]string{
		"email":    "",
		"password": "Password123",
	}
	jsonData, _ := json.Marshal(userData)

	req, _ := http.NewRequest(http.MethodPost, "/user", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var respBody map[string]string
	json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.Contains(t, respBody["error"], "Field validation for 'Email' failed")
}

func TestCreateUser_InvalidEmailFormat(t *testing.T) {
	r := testutils.SetupTestRouter()
	r.POST("/user", CreateUser)

	userData := map[string]string{
		"email":    "invalid-email",
		"password": "Password123",
	}
	jsonData, _ := json.Marshal(userData)

	req, _ := http.NewRequest(http.MethodPost, "/user", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var respBody map[string]string
	json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.Contains(t, respBody["error"], "Field validation for 'Email' failed")
}

func TestCreateUser_EmptyPassword(t *testing.T) {
	// Configuration du routeur
	r := testutils.SetupTestRouter()
	r.POST("/user", CreateUser)

	// Données avec mot de passe vide
	userData := map[string]string{
		"email":    "test@example.com",
		"password": "",
	}
	jsonData, _ := json.Marshal(userData)

	req, _ := http.NewRequest(http.MethodPost, "/user", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var respBody map[string]string
	json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.Contains(t, respBody["error"], "Field validation for 'Password' failed")
}

func TestCreateUser_ShortPassword(t *testing.T) {
	r := testutils.SetupTestRouter()
	r.POST("/user", CreateUser)

	userData := map[string]string{
		"email":    "test@example.com",
		"password": "Abc1",
	}
	jsonData, _ := json.Marshal(userData)

	req, _ := http.NewRequest(http.MethodPost, "/user", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var respBody map[string]string
	json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.Contains(t, respBody["error"], "Field validation for 'Password' failed")
}

func TestCreateUser_WeakPassword(t *testing.T) {
	testCases := []struct {
		name          string
		password      string
		expectedError string
	}{
		{"OnlyLowercase", "password123", "The password must contain at least one lowercase, one uppercase and one digit"},
		{"OnlyUppercase", "PASSWORD123", "The password must contain at least one lowercase, one uppercase and one digit"},
		{"NoDigits", "PasswordOnly", "The password must contain at least one lowercase, one uppercase and one digit"},
		{"OnlyDigits", "12345678", "The password must contain at least one lowercase, one uppercase and one digit"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, mock, cleanup := testutils.SetupTestDB(t)
			defer cleanup()

			mock.ExpectQuery(`SELECT (.+) FROM "users" WHERE email = (.+) AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT (.+)`).
				WithArgs("test@example.com", 1).
				WillReturnError(gorm.ErrRecordNotFound)

			r := testutils.SetupTestRouter()
			r.POST("/user", CreateUser)

			userData := map[string]string{
				"email":    "test@example.com",
				"password": tc.password,
			}
			jsonData, _ := json.Marshal(userData)

			req, _ := http.NewRequest(http.MethodPost, "/user", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			r.ServeHTTP(resp, req)

			assert.Equal(t, http.StatusBadRequest, resp.Code)

			var respBody map[string]string
			json.Unmarshal(resp.Body.Bytes(), &respBody)
			assert.Contains(t, respBody["error"], tc.expectedError)
		})
	}
}

func TestCreateUser_EmailAlreadyExists(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT (.+) FROM "users" WHERE email = (.+) AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT (.+)`).
		WithArgs("existing@example.com", 1).
		WillReturnRows(mock.NewRows([]string{"id", "email"}).AddRow(1, "existing@example.com"))

	r := testutils.SetupTestRouter()
	r.POST("/user", CreateUser)

	userData := map[string]string{
		"email":    "existing@example.com",
		"password": "Password123",
	}
	jsonData, _ := json.Marshal(userData)

	req, _ := http.NewRequest(http.MethodPost, "/user", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusConflict, resp.Code)

	var respBody map[string]string
	json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.Contains(t, respBody["error"], "This email is already used")
}

func TestCreateUser_DatabaseError(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT (.+) FROM "users" WHERE email = (.+) AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT (.+)`).
		WithArgs("test@example.com", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users" (.+) RETURNING "id"`).
		WillReturnError(gorm.ErrInvalidDB)
	mock.ExpectRollback()

	r := testutils.SetupTestRouter()
	r.POST("/user", CreateUser)

	userData := map[string]string{
		"email":    "test@example.com",
		"password": "Password123",
	}
	jsonData, _ := json.Marshal(userData)

	req, _ := http.NewRequest(http.MethodPost, "/user", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestHashPassword(t *testing.T) {
	hashed, _ := hashPassword("Password123")
	assert.NotEmpty(t, hashed)

	assert.NotEqual(t, "Password123", hashed)

	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte("Password123"))
	assert.NoError(t, err)

	err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte("WrongPassword"))
	assert.Error(t, err)
}

func TestSamePassword_Correct(t *testing.T) {
	hashed := samePassword("Test123!", "$2a$10$8b9qfHvbQVnP1IgEyd/AX.X5PCNGO/ZVE13NZS8xg3wDo6f4rWpiW")
	assert.True(t, hashed)

}

func TestSamePassword_Incorrect(t *testing.T) {
	hashed := samePassword("Test123!!", "$2a$10$8b9qfHvbQVnP1IgEyd/AX.X5PCNGO/ZVE13NZS8xg3wDo6f4rWpiW")
	assert.False(t, hashed)

}

//TODO test for login
// 1. user avec un bon mail et un bon pwd
// 2. user avec un bon mail et un mauvais pwd
//3. user avec un email inexistant
