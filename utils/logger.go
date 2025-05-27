package utils

import (
	"io"
	"log"
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func init() {
	Logger.SetFormatter(&logrus.JSONFormatter{})
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
		Logger.SetOutput(io.MultiWriter(os.Stdout, file))
	}
}

func LogSuccess(message string) {
	Logger.WithField("function", getCaller()).WithField("status", "success").Info(message)
}

func LogError(err error, message string) {
	entry := Logger.WithField("function", getCaller()).WithField("status", "error")
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
