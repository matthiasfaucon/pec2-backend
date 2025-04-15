package users

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"pec2-backend/models"
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

func TestGetAllUsers_Success(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	createdAt := time.Now()
	rows := mock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "email", "password", "user_name", "role", "bio", "profile_picture", "stripe_customer_id", "subscription_price", "enable", "subscription_enable", "comments_enable", "message_enable", "email_verified_at", "siret"}).
		AddRow("user-uuid-1", createdAt, createdAt, nil, "user1@example.com", "hashedpassword1", "User1", "USER", "Bio 1", "", "", 0, true, false, true, true, nil, "").
		AddRow("user-uuid-2", createdAt.Add(-time.Hour), createdAt.Add(-time.Hour), nil, "user2@example.com", "hashedpassword2", "User2", "ADMIN", "Bio 2", "", "", 0, true, false, true, true, nil, "")

	mock.ExpectQuery(`SELECT \* FROM "users" ORDER BY created_at DESC`).
		WillReturnRows(rows)

	r := testutils.SetupTestRouter()
	r.GET("/users", GetAllUsers)

	req, _ := http.NewRequest(http.MethodGet, "/users", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string][]models.User
	json.Unmarshal(resp.Body.Bytes(), &response)

	users := response["users"]
	assert.Len(t, users, 2)
	assert.Equal(t, "user1@example.com", users[0].Email)
	assert.Equal(t, "user2@example.com", users[1].Email)

	assert.Empty(t, users[0].Password)
	assert.Empty(t, users[1].Password)
}

func TestGetAllUsers_DatabaseError(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT \* FROM "users" ORDER BY created_at DESC`).
		WillReturnError(gorm.ErrInvalidDB)

	r := testutils.SetupTestRouter()
	r.GET("/users", GetAllUsers)

	req, _ := http.NewRequest(http.MethodGet, "/users", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	var respBody map[string]string
	json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.Contains(t, respBody["error"], "invalid db")
}

func TestGetUserByID_Success(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	createdAt := time.Now()
	rows := mock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "email", "password", "user_name", "role", "bio", "profile_picture", "stripe_customer_id", "subscription_price", "enable", "subscription_enable", "comments_enable", "message_enable", "email_verified_at", "siret"}).
		AddRow("user-uuid-1", createdAt, createdAt, nil, "user1@example.com", "hashedpassword1", "User1", "USER", "Bio 1", "", "", 0, true, false, true, true, nil, "")

	mock.ExpectQuery(`SELECT \* FROM "users" WHERE id = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs("user-uuid-1", 1).
		WillReturnRows(rows)

	r := testutils.SetupTestRouter()
	r.GET("/users/:id", GetUserByID)

	req, _ := http.NewRequest(http.MethodGet, "/users/user-uuid-1", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]models.User
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err, "Erreur lors de la désérialisation de la réponse JSON: %s", resp.Body.String())

	user := response["user"]
	assert.Equal(t, "user1@example.com", user.Email)
	assert.Equal(t, "User1", user.UserName)

	assert.Empty(t, user.Password)
}

func TestGetUserByID_UserNotFound(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT \* FROM "users" WHERE id = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs("nonexistent-uuid", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	r := testutils.SetupTestRouter()
	r.GET("/users/:id", GetUserByID)

	req, _ := http.NewRequest(http.MethodGet, "/users/nonexistent-uuid", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNotFound, resp.Code)

	var respBody map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.NoError(t, err, "Erreur lors de la désérialisation de la réponse JSON: %s", resp.Body.String())
	assert.Equal(t, "Utilisateur non trouvé", respBody["error"])
}
