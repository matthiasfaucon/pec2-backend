package testutils

import (
	"database/sql/driver"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"pec2-backend/db"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Erreur lors de la création de la connexion SQL mock: %s", err)
	}

	newLogger := logger.New(
		log.New(io.Discard, "", log.LstdFlags),
		logger.Config{
			LogLevel: logger.Silent,
		},
	)

	dialector := postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		t.Fatalf("Erreur lors de l'ouverture de la connexion GORM: %s", err)
	}

	originalDB := db.DB
	db.DB = gormDB

	cleanup := func() {
		db.DB = originalDB
		sqlDB.Close()
	}

	return gormDB, mock, cleanup
}

func SetupTestRouter() *gin.Engine {
	r := gin.New()
	return r
}

func InitTestMain() {
	gin.SetMode(gin.TestMode)
}

type Result struct {
	lastInsertID int64
	rowsAffected int64
}

func (r Result) LastInsertId() (int64, error) {
	return r.lastInsertID, nil
}

func (r Result) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

func NewResult(lastInsertID int64, rowsAffected int64) driver.Result {
	return Result{lastInsertID: lastInsertID, rowsAffected: rowsAffected}
}

// GenerateTestToken génère un token JWT pour les tests avec un ID numérique
func GenerateTestToken(userID uint, role string) (string, error) {
	// On utilise une clé secrète fixe pour les tests
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		// Utiliser une clé par défaut si aucune n'est définie
		jwtSecret = []byte("test_secret_key")
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// GenerateTestTokenString génère un token JWT pour les tests avec un ID chaîne de caractères (UUID)
func GenerateTestTokenString(userID string, role string) (string, error) {
	// On utilise une clé secrète fixe pour les tests
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		// Utiliser une clé par défaut si aucune n'est définie
		jwtSecret = []byte("test_secret_key")
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
