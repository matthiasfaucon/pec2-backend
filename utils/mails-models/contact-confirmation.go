package mailsmodels

import (
	"fmt"
	"pec2-backend/utils"
)

func ContactConfirmation(contact ContactEmailData) {
	subject := "Subject: Confirmation de votre demande de contact - OnlyFlick \r\n"
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	body := fmt.Sprintf(`
	<div style="background-color: #722ED1; width: 100%%; min-height: 300px; padding: 30px; box-sizing:border-box">
		<table style="background-color: #ffffff; width: 100%%;  min-height: 300px;">
			<tbody>
				<tr>
					<td><h1 style="text-align:center">Merci pour votre message !</h1></td>
				</tr>
				<tr>
					<td style="text-align:center; padding-bottom: 30px;">
						<p>Bonjour %s %s,</p>
						<p>Nous avons bien reçu votre demande de contact concernant : "%s"</p>
						<p>Nous vous répondrons dans les plus brefs délais.</p>
						<p>Votre message :</p>
						<blockquote style="background-color: #f5f5f5; padding: 15px; border-left: 5px solid #722ED1;">
							%s
						</blockquote>
					</td>
				</tr>
			</tbody>
		</table>
	</div>
`, contact.FirstName, contact.LastName, contact.Subject, contact.Message)

	message := []byte(subject + mime + body)
	utils.SendMail(contact.Email, message)
}

type ContactEmailData struct {
	FirstName string
	LastName  string
	Email     string
	Subject   string
	Message   string
}

type ContactStatusUpdateData struct {
	FirstName string
	LastName  string
	Email     string
	Subject   string
	Status    string
}

func GetStatusFrench(status string) string {
	switch status {
	case "open":
		return "Ouvert"
	case "processing":
		return "En cours de traitement"
	case "closed":
		return "Fermé"
	case "rejected":
		return "Rejeté"
	default:
		return status
	}
}

func ContactStatusUpdate(data ContactStatusUpdateData) {
	statusFrench := GetStatusFrench(string(data.Status))

	subject := "Subject: Mise à jour de votre demande de contact - OnlyFlick \r\n"
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	body := fmt.Sprintf(`
	<div style="background-color: #722ED1; width: 100%%; min-height: 300px; padding: 30px; box-sizing:border-box">
		<table style="background-color: #ffffff; width: 100%%;  min-height: 300px;">
			<tbody>
				<tr>
					<td><h1 style="text-align:center">Mise à jour de votre demande</h1></td>
				</tr>
				<tr>
					<td style="text-align:center; padding-bottom: 30px;">
						<p>Bonjour %s %s,</p>
						<p>Nous vous informons que le statut de votre demande concernant : "%s" a été mis à jour.</p>
						<p><strong>Nouveau statut : %s</strong></p>
						<p>N'hésitez pas à nous contacter si vous avez des questions.</p>
						<p>Merci de votre confiance.</p>
					</td>
				</tr>
			</tbody>
		</table>
	</div>
`, data.FirstName, data.LastName, data.Subject, statusFrench)

	message := []byte(subject + mime + body)
	utils.SendMail(data.Email, message)
}
