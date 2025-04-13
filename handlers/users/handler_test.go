package users

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"pec2-backend/db"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB configure une base de données mock pour les tests
func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	// Crée une connexion SQL mock
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Erreur lors de la création de la connexion SQL mock: %s", err)
	}

	// Désactiver les logs pour les tests
	newLogger := logger.New(
		log.New(io.Discard, "", log.LstdFlags), // Output à io.Discard (remplace ioutil.Discard)
		logger.Config{
			LogLevel: logger.Silent, // Log level silencieux
		},
	)

	// Configure GORM pour utiliser cette connexion
	dialector := postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{
		Logger: newLogger, // Utiliser le logger silencieux
	})
	if err != nil {
		t.Fatalf("Erreur lors de l'ouverture de la connexion GORM: %s", err)
	}

	// Sauvegarde l'instance DB originale et la remplace par notre mock
	originalDB := db.DB
	db.DB = gormDB

	// Retourne une fonction de nettoyage pour restaurer db.DB après le test
	cleanup := func() {
		db.DB = originalDB
		sqlDB.Close()
	}

	return gormDB, mock, cleanup
}

func TestMain(m *testing.M) {
	// Désactiver le logging de Gin
	gin.SetMode(gin.TestMode)
	// Redirection des logs standards pendant les tests
	log.SetOutput(io.Discard)

	// Exécution de tous les tests
	exitCode := m.Run()

	// Restauration des logs standard
	log.SetOutput(os.Stdout)

	os.Exit(exitCode)
}

func setupTestRouter() *gin.Engine {
	// gin.SetMode déjà défini dans TestMain
	r := gin.New() // Utiliser gin.New() au lieu de gin.Default() pour ne pas avoir de logs
	return r
}

func TestCreateUser_Success(t *testing.T) {
	_, mock, cleanup := setupTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT (.+) FROM "users" WHERE email = (.+) AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT (.+)`).
		WithArgs("test@example.com", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users" (.+) RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	r := setupTestRouter()
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
	r := setupTestRouter()
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
	r := setupTestRouter()
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
	r := setupTestRouter()
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
	r := setupTestRouter()
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
			_, mock, cleanup := setupTestDB(t)
			defer cleanup()

			mock.ExpectQuery(`SELECT (.+) FROM "users" WHERE email = (.+) AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT (.+)`).
				WithArgs("test@example.com", 1).
				WillReturnError(gorm.ErrRecordNotFound)

			r := setupTestRouter()
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
	_, mock, cleanup := setupTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT (.+) FROM "users" WHERE email = (.+) AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT (.+)`).
		WithArgs("existing@example.com", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow(1, "existing@example.com"))

	r := setupTestRouter()
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
	_, mock, cleanup := setupTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT (.+) FROM "users" WHERE email = (.+) AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT (.+)`).
		WithArgs("test@example.com", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users" (.+) RETURNING "id"`).
		WillReturnError(gorm.ErrInvalidDB)
	mock.ExpectRollback()

	r := setupTestRouter()
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
	hashed := hashPassword("Password123")
	assert.NotEmpty(t, hashed)

	assert.NotEqual(t, "Password123", hashed)

	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte("Password123"))
	assert.NoError(t, err)

	err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte("WrongPassword"))
	assert.Error(t, err)
}
