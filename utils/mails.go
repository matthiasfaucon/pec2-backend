package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendMail(email string, message []byte) {
	// Infos de connexion
	from := "onlyflick.eemi@gmail.com"
	password := os.Getenv("GOOGLE_MDP")
	to := email

	// SMTP server configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Message

	// Auth
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Envoi
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Email envoyé avec succès.")
}
