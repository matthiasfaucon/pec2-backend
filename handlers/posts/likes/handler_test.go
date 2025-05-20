package likes

import (
	"database/sql/driver"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"pec2-backend/testutils"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// AnyTime est un type personnalisé pour matcher n'importe quelle valeur de temps dans les tests
type AnyTime struct{}

// Match satisfait l'interface driver.Value pour matcher n'importe quelle valeur de temps
func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func TestMain(m *testing.M) {
	testutils.InitTestMain()

	log.SetOutput(io.Discard)

	exitCode := m.Run()

	log.SetOutput(os.Stdout)

	os.Exit(exitCode)
}

// Test l'ajout d'un like à un post
func TestToggleLike_Add(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	postID := "post-uuid"
	userID := "user-uuid"

	// Mock pour vérifier si le post existe
	postRows := mock.NewRows([]string{"id", "user_id", "name", "picture_url", "is_free", "enable"}).
		AddRow(postID, "author-uuid", "Test Post", "http://example.com/image.jpg", true, true)
	mock.ExpectQuery(`SELECT \* FROM "posts" WHERE id = \$1 ORDER BY "posts"."id" LIMIT \$2`).
		WithArgs(postID, 1).
		WillReturnRows(postRows)

	// Mock pour vérifier si le like existe déjà
	mock.ExpectQuery(`SELECT \* FROM "likes" WHERE post_id = \$1 AND user_id = \$2 ORDER BY "likes"."id" LIMIT \$3`).
		WithArgs(postID, userID, 1).
		WillReturnError(gorm.ErrRecordNotFound)	// Mock pour l'insertion du like
	mock.ExpectBegin()	
	mock.ExpectQuery(`INSERT INTO "likes" \("post_id","user_id","created_at"\) VALUES \(\$1,\$2,\$3\) RETURNING "id"`).
		WithArgs(postID, userID, AnyTime{}).
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow("like-uuid"))
	mock.ExpectCommit()

	// Mock pour compter les likes après ajout
	mock.ExpectQuery(`SELECT count\(\*\) FROM "likes" WHERE post_id = \$1`).
		WithArgs(postID).
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(1))

	r := testutils.SetupTestRouter()
	r.POST("/posts/:id/like", func(c *gin.Context) {
		c.Set("user_id", userID)
		ToggleLike(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/posts/"+postID+"/like", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, "Like added successfully", response["message"])
	assert.Equal(t, "added", response["action"])
	assert.Equal(t, float64(1), response["likesCount"])
}

// Test la suppression d'un like existant
func TestToggleLike_Remove(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	postID := "post-uuid"
	userID := "user-uuid"
	likeID := "like-uuid"
	// Mock pour vérifier si le post existe
	postRows := mock.NewRows([]string{"id", "user_id", "name", "picture_url", "is_free", "enable"}).
		AddRow(postID, "author-uuid", "Test Post", "http://example.com/image.jpg", true, true)
	mock.ExpectQuery(`SELECT \* FROM "posts" WHERE id = \$1 ORDER BY "posts"."id" LIMIT \$2`).
		WithArgs(postID, 1).
		WillReturnRows(postRows)
	// Mock pour vérifier si le like existe déjà
	likeRows := mock.NewRows([]string{"id", "post_id", "user_id", "created_at"}).
		AddRow(likeID, postID, userID, time.Now())
	mock.ExpectQuery(`SELECT \* FROM "likes" WHERE post_id = \$1 AND user_id = \$2 ORDER BY "likes"."id" LIMIT \$3`).
		WithArgs(postID, userID, 1).
		WillReturnRows(likeRows)
	// Mock pour la suppression du like
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "likes" WHERE "likes"."id" = \$1`).
		WithArgs(likeID).
		WillReturnResult(testutils.NewResult(1, 1))
	mock.ExpectCommit()

	// Mock pour compter les likes après suppression
	mock.ExpectQuery(`SELECT count\(\*\) FROM "likes" WHERE post_id = \$1`).
		WithArgs(postID).
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(0))

	r := testutils.SetupTestRouter()
	r.POST("/posts/:id/like", func(c *gin.Context) {
		c.Set("user_id", userID)
		ToggleLike(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/posts/"+postID+"/like", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, "Like removed successfully", response["message"])
	assert.Equal(t, "removed", response["action"])
	assert.Equal(t, float64(0), response["likesCount"])
}

// Test le cas où le post n'existe pas
func TestToggleLike_PostNotFound(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	postID := "non-existent-post-uuid"
	userID := "user-uuid"

	// Mock pour vérifier si le post existe (ne le trouve pas)
	mock.ExpectQuery(`SELECT \* FROM "posts" WHERE id = \$1 ORDER BY "posts"."id" LIMIT \$2`).
		WithArgs(postID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	r := testutils.SetupTestRouter()
	r.POST("/posts/:id/like", func(c *gin.Context) {
		c.Set("user_id", userID)
		ToggleLike(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/posts/"+postID+"/like", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNotFound, resp.Code)

	var response map[string]string
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, "Post not found", response["error"])
}

// Test le cas où l'utilisateur n'est pas authentifié
func TestToggleLike_Unauthorized(t *testing.T) {
	_, _, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	postID := "post-uuid"

	r := testutils.SetupTestRouter()
	r.POST("/posts/:id/like", ToggleLike)

	req, _ := http.NewRequest(http.MethodPost, "/posts/"+postID+"/like", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)

	var response map[string]string
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, "User not found in token", response["error"])
}