package testutils

import (
	"io"
	"log"
	"testing"

	"pec2-backend/db"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
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
