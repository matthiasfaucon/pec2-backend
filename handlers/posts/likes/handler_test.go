package likes

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"pec2-backend/testutils"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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

// Test pour ajouter un like (cas de succès)
func TestToggleLike_Add(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	postID := "123e4567-e89b-12d3-a456-426614174000"
	userID := "abc12345-e89b-12d3-a456-426614174000"

	// Mock pour vérifier si le post existe
	mock.ExpectQuery(`SELECT (.+) FROM "posts" WHERE id = \$1 AND .+"posts"\."deleted_at" IS NULL LIMIT 1`).
		WithArgs(postID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).
			AddRow(postID, "Test Post"))

	// Mock pour vérifier si le like existe déjà
	mock.ExpectQuery(`SELECT (.+) FROM "likes" WHERE post_id = \$1 AND user_id = \$2 AND .+"likes"\."deleted_at" IS NULL LIMIT 1`).
		WithArgs(postID, userID).
		WillReturnError(gorm.ErrRecordNotFound)

	// Mock pour créer un nouveau like
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "likes" (.+) RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("like123"))
	mock.ExpectCommit()

	r := testutils.SetupTestRouter()
	r.POST("/posts/:id/like", func(c *gin.Context) {
		// Simuler l'authentification
		c.Set("user_id", userID)
		ToggleLike(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/posts/"+postID+"/like", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var respBody map[string]string
	json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.Equal(t, "Like added successfully", respBody["message"])
}

// Test pour supprimer un like (cas de succès)
func TestToggleLike_Remove(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	postID := "123e4567-e89b-12d3-a456-426614174000"
	userID := "abc12345-e89b-12d3-a456-426614174000"
	likeID := "like123"

	// Mock pour vérifier si le post existe
	mock.ExpectQuery(`SELECT (.+) FROM "posts" WHERE id = \$1 AND .+"posts"\."deleted_at" IS NULL LIMIT 1`).
		WithArgs(postID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).
			AddRow(postID, "Test Post"))

	// Mock pour vérifier si le like existe déjà
	mock.ExpectQuery(`SELECT (.+) FROM "likes" WHERE post_id = \$1 AND user_id = \$2 AND .+"likes"\."deleted_at" IS NULL LIMIT 1`).
		WithArgs(postID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "post_id", "user_id"}).
			AddRow(likeID, postID, userID))

	// Mock pour supprimer le like
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "likes" WHERE "likes"."id" = \$1`).
		WithArgs(likeID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	r := testutils.SetupTestRouter()
	r.POST("/posts/:id/like", func(c *gin.Context) {
		// Simuler l'authentification
		c.Set("user_id", userID)
		ToggleLike(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/posts/"+postID+"/like", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var respBody map[string]string
	json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.Equal(t, "Like removed successfully", respBody["message"])
}

// Test pour un post inexistant (cas d'échec)
func TestToggleLike_PostNotFound(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	postID := "non-existent-id"
	userID := "abc12345-e89b-12d3-a456-426614174000"

	// Mock pour vérifier si le post existe - retourne qu'il n'existe pas
	mock.ExpectQuery(`SELECT (.+) FROM "posts" WHERE id = \$1 AND .+"posts"\."deleted_at" IS NULL LIMIT 1`).
		WithArgs(postID).
		WillReturnError(gorm.ErrRecordNotFound)

	r := testutils.SetupTestRouter()
	r.POST("/posts/:id/like", func(c *gin.Context) {
		// Simuler l'authentification
		c.Set("user_id", userID)
		ToggleLike(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/posts/"+postID+"/like", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNotFound, resp.Code)

	var respBody map[string]string
	json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.Contains(t, respBody["error"], "Post not found")
}

// Test pour un utilisateur non authentifié (cas d'échec)
func TestToggleLike_Unauthorized(t *testing.T) {
	r := testutils.SetupTestRouter()
	r.POST("/posts/:id/like", ToggleLike)

	postID := "123e4567-e89b-12d3-a456-426614174000"

	req, _ := http.NewRequest(http.MethodPost, "/posts/"+postID+"/like", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)

	var respBody map[string]string
	json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.Contains(t, respBody["error"], "User not found in token")
}
