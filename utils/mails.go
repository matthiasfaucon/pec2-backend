package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendMail(email string, message []byte) {
	from := "onlyflick.eemi@gmail.com"
	password := os.Getenv("GOOGLE_SMTP_MDP")
	to := email

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Email envoyé avec succès.")
}
