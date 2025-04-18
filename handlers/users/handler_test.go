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

	"github.com/gin-gonic/gin"
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
	birthDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)

	rows := mock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "email", "password", "user_name", "first_name", "last_name", "birth_day_date", "sexe", "role", "bio", "profile_picture", "stripe_customer_id", "subscription_price", "enable", "subscription_enable", "comments_enable", "message_enable", "email_verified_at", "siret"}).
		AddRow("user-uuid-1", createdAt, createdAt, nil, "user1@example.com", "hashedpassword1", "User1", "John", "Doe", birthDate, "MAN", "USER", "Bio 1", "", "", 0, true, false, true, true, nil, "").
		AddRow("user-uuid-2", createdAt.Add(-time.Hour), createdAt.Add(-time.Hour), nil, "user2@example.com", "hashedpassword2", "User2", "Jane", "Smith", birthDate.AddDate(-2, 0, 0), "WOMAN", "ADMIN", "Bio 2", "", "", 0, true, false, true, true, nil, "")

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
	assert.Equal(t, "John", users[0].FirstName)
	assert.Equal(t, "Jane", users[1].FirstName)
	assert.Equal(t, "Doe", users[0].LastName)
	assert.Equal(t, "Smith", users[1].LastName)
	assert.Equal(t, models.Sexe("MAN"), users[0].Sexe)
	assert.Equal(t, models.Sexe("WOMAN"), users[1].Sexe)

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

func TestGetUserProfile_Success(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	userID := "user-uuid-1"
	createdAt := time.Now()
	birthDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)

	rows := mock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "email", "password", "user_name", "first_name", "last_name", "birth_day_date", "sexe", "role", "bio", "profile_picture", "stripe_customer_id", "subscription_price", "enable", "subscription_enable", "comments_enable", "message_enable", "email_verified_at", "siret"}).
		AddRow(userID, createdAt, createdAt, nil, "user1@example.com", "hashedpassword1", "User1", "John", "Doe", birthDate, "MAN", "USER", "Bio 1", "profile.jpg", "", 0, true, false, true, true, nil, "")

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	r := testutils.SetupTestRouter()
	r.GET("/users/profile", func(c *gin.Context) {
		c.Set("user_id", userID)
		GetUserProfile(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/users/profile", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]models.User
	json.Unmarshal(resp.Body.Bytes(), &response)

	user := response["user"]
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "user1@example.com", user.Email)
	assert.Equal(t, "User1", user.UserName)
	assert.Equal(t, "John", user.FirstName)
	assert.Equal(t, "Doe", user.LastName)
	assert.Equal(t, birthDate.UTC().Format(time.RFC3339), user.BirthDayDate.UTC().Format(time.RFC3339))
	assert.Equal(t, models.Sexe("MAN"), user.Sexe)
	assert.Equal(t, "Bio 1", user.Bio)
	assert.Equal(t, "profile.jpg", user.ProfilePicture)
	assert.Empty(t, user.Password)
}

func TestGetUserProfile_UserNotFound(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	userID := "non-existent-user-id"

	mock.ExpectQuery("SELECT").WillReturnError(gorm.ErrRecordNotFound)

	r := testutils.SetupTestRouter()
	r.GET("/users/profile", func(c *gin.Context) {
		c.Set("user_id", userID)
		GetUserProfile(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/users/profile", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNotFound, resp.Code)

	var respBody map[string]string
	json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.Contains(t, respBody["error"], "User not found")
}
