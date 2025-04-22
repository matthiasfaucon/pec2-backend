package categories

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pec2-backend/models"
	"pec2-backend/testutils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCreateCategory_Success(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "categories" (.+) RETURNING "id"`).
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow("category-uuid"))
	mock.ExpectCommit()

	r := testutils.SetupTestRouter()
	r.POST("/categories", CreateCategory)

	categoryData := map[string]string{
		"name": "Test Category",
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
}

func TestGetAllCategories_Success(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	now := time.Now()
	rows := mock.NewRows([]string{"id", "name", "created_at"}).
		AddRow("category-uuid-1", "Category 1", now).
		AddRow("category-uuid-2", "Category 2", now)

	mock.ExpectQuery(`SELECT \* FROM "categories" ORDER BY name ASC`).
		WillReturnRows(rows)

	r := testutils.SetupTestRouter()
	r.GET("/categories", GetAllCategories)

	req, _ := http.NewRequest(http.MethodGet, "/categories", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var categories []models.Category
	json.Unmarshal(resp.Body.Bytes(), &categories)
	assert.Len(t, categories, 2)
	assert.Equal(t, "Category 1", categories[0].Name)
	assert.Equal(t, "Category 2", categories[1].Name)
}

func TestUpdateCategory_Success(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery(`SELECT \* FROM "categories" WHERE id = (.+)`).
		WithArgs("category-uuid").
		WillReturnRows(mock.NewRows([]string{"id", "name", "created_at"}).
			AddRow("category-uuid", "Old Category", now))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "categories" SET (.+) WHERE (.+)`).
		WillReturnResult(testutils.NewResult(1, 1))
	mock.ExpectCommit()

	r := testutils.SetupTestRouter()
	r.PUT("/categories/:id", UpdateCategory)

	categoryData := map[string]string{
		"name": "Updated Category",
	}
	jsonData, _ := json.Marshal(categoryData)

	req, _ := http.NewRequest(http.MethodPut, "/categories/category-uuid", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var category models.Category
	json.Unmarshal(resp.Body.Bytes(), &category)
	assert.Equal(t, "Updated Category", category.Name)
}

func TestUpdateCategory_NotFound(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT \* FROM "categories" WHERE id = (.+)`).
		WithArgs("non-existent-uuid").
		WillReturnError(gorm.ErrRecordNotFound)

	r := testutils.SetupTestRouter()
	r.PUT("/categories/:id", UpdateCategory)

	categoryData := map[string]string{
		"name": "Updated Category",
	}
	jsonData, _ := json.Marshal(categoryData)

	req, _ := http.NewRequest(http.MethodPut, "/categories/non-existent-uuid", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestDeleteCategory_Success(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery(`SELECT \* FROM "categories" WHERE id = (.+)`).
		WithArgs("category-uuid").
		WillReturnRows(mock.NewRows([]string{"id", "name", "created_at"}).
			AddRow("category-uuid", "Test Category", now))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "post_categories" SET (.+) WHERE (.+)`).
		WillReturnResult(testutils.NewResult(0, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "categories" WHERE (.+)`).
		WithArgs("category-uuid").
		WillReturnResult(testutils.NewResult(1, 1))
	mock.ExpectCommit()

	r := testutils.SetupTestRouter()
	r.DELETE("/categories/:id", DeleteCategory)

	req, _ := http.NewRequest(http.MethodDelete, "/categories/category-uuid", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]string
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, "Category deleted successfully", response["message"])
}

func TestDeleteCategory_NotFound(t *testing.T) {
	_, mock, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT \* FROM "categories" WHERE id = (.+)`).
		WithArgs("non-existent-uuid").
		WillReturnError(gorm.ErrRecordNotFound)

	r := testutils.SetupTestRouter()
	r.DELETE("/categories/:id", DeleteCategory)

	req, _ := http.NewRequest(http.MethodDelete, "/categories/non-existent-uuid", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNotFound, resp.Code)
}