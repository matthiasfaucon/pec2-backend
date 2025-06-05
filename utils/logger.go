package utils

import (
	"context"
	"io"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

var Logger = logrus.New()

func init() {
	// Configuration pour Grafana (format JSON avec champs normalisés)
	Logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})
	Logger.SetLevel(logrus.InfoLevel)

	// Création du dossier logs si besoin
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		err = os.Mkdir("logs", 0755)
		if err != nil {
			log.Printf("Erreur lors de la création du dossier logs: %v", err)
		}
	}

	file, err := os.OpenFile("logs/app.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Impossible d'ouvrir le fichier de log: %v", err)
		Logger.SetOutput(os.Stdout)
	} else {
		// Log au format JSON dans le fichier, et coloré dans la console
		Logger.SetOutput(io.MultiWriter(os.Stdout, file))
	}
}

func LogWriter() io.Writer {
	file, err := os.OpenFile("logs/gin.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Impossible d'ouvrir le fichier de log: %v", err)
		return &ginLogWriter{out: os.Stdout}
	}
	return &ginLogWriter{out: io.MultiWriter(os.Stdout, file)}
}

// ginLogWriter est un wrapper pour formater les logs de Gin au format JSON
type ginLogWriter struct {
	out io.Writer
}

// Write implémente io.Writer pour les logs de Gin
func (w *ginLogWriter) Write(p []byte) (n int, err error) {
	entry := Logger.WithFields(logrus.Fields{
		"source": "gin",
	})
	entry.Info(string(p))
	return len(p), nil
}

// GetGormLogger retourne un logger pour GORM compatible avec le format global
func GetGormLogger() logger.Interface {
	return &gormLogger{
		LogLevel: logger.Info,
	}
}

type gormLogger struct {
	LogLevel logger.LogLevel
}

// LogMode implémente logger.Interface
func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info implémente logger.Interface
func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	Logger.WithFields(logrus.Fields{
		"source": "gorm",
		"data":   data,
	}).Info(msg)
}

// Warn implémente logger.Interface
func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	Logger.WithFields(logrus.Fields{
		"source": "gorm",
		"data":   data,
	}).Warn(msg)
}

// Error implémente logger.Interface
func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	Logger.WithFields(logrus.Fields{
		"source": "gorm",
		"data":   data,
	}).Error(msg)
}

// Trace implémente logger.Interface
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	
	fields := logrus.Fields{
		"source":   "gorm",
		"elapsed":  elapsed.String(),
		"sql":      sql,
		"rows":     rows,
	}
	
	if err != nil {
		fields["error"] = err.Error()
		Logger.WithFields(fields).Error("SQL query error")
	} else {
		Logger.WithFields(fields).Debug("SQL query executed")
	}
}

func LogSuccess(message string) {
	Logger.WithFields(logrus.Fields{
		"function": getCaller(),
		"status": "success",
		"source": "app",
	}).Info(message)
}

func LogInfo(message string) {
	Logger.WithFields(logrus.Fields{
		"function": getCaller(),
		"source": "app",
	}).Info(message)
}

func LogError(err error, message string) {
	entry := Logger.WithFields(logrus.Fields{
		"function": getCaller(),
		"status": "error",
		"source": "app",
	})
	if err != nil {
		entry = entry.WithField("error", err.Error())
	}
	entry.Error(message)
}

func LogSuccessWithUser(userID interface{}, message string) {
	if userID == nil || userID == "" {
		userID = "0"
	}
	Logger.WithFields(logrus.Fields{
		"function": getCaller(),
		"status": "success",
		"source": "app",
		"user_id": userID,
	}).Info(message)
}

func LogErrorWithUser(userID interface{}, err error, message string) {
	if userID == nil || userID == "" {
		userID = "0"
	}
	entry := Logger.WithFields(logrus.Fields{
		"function": getCaller(),
		"status": "error",
		"source": "app",
		"user_id": userID,
	})
	if err != nil {
		entry = entry.WithField("error", err.Error())
	}
	entry.Error(message)
}

func getCaller() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}
