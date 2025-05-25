package categories

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"pec2-backend/models"
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

func TestCreateCategory_Success(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	// Mock pour vérifier si la catégorie existe déjà
	mock.ExpectQuery(`SELECT \* FROM "categories" WHERE name = \$1 ORDER BY "categories"."id" LIMIT \$2`).
		WithArgs("Test Category", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// Mock pour l'insertion
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "categories" (.+) RETURNING "id"`).
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow("category-uuid"))
	mock.ExpectCommit()

	r := testutils.SetupTestRouter()
	r.POST("/categories", CreateCategory)

	categoryData := map[string]string{
		"name":       "Test Category",
		"pictureUrl": "http://example.com/test-image.jpg",
	}
	jsonData, _ := json.Marshal(categoryData)

	req, _ := http.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)

	var category models.Category
	json.Unmarshal(resp.Body.Bytes(), &category)
	assert.Equal(t, "Test Category", category.Name)
	assert.Equal(t, "http://example.com/test-image.jpg", category.PictureURL)
}