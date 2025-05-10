package mailsmodels

import (
	"fmt"
	"pec2-backend/models"
	"pec2-backend/utils"
)

type ContentCreatorStatusUpdateData struct {
	FirstName   string
	LastName    string
	Email       string
	CompanyName string
	Status      models.ContentCreatorStatusType
}

func getStatusFrenchForContentCreator(status models.ContentCreatorStatusType) string {
	switch status {
	case models.ContentCreatorStatusPending:
		return "En attente"
	case models.ContentCreatorStatusApproved:
		return "Approuvée"
	case models.ContentCreatorStatusRejected:
		return "Rejetée"
	default:
		return string(status)
	}
}

func ContentCreatorStatusUpdate(data ContentCreatorStatusUpdateData) {
	statusFrench := getStatusFrenchForContentCreator(data.Status)

	subject := "Subject: Mise à jour du statut de votre compte créateur - OnlyFlick \r\n"
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	body := fmt.Sprintf(`
	<div style="background-color: #722ED1; width: 100%%; min-height: 300px; padding: 30px; box-sizing:border-box">
		<table style="background-color: #ffffff; width: 100%%; min-height: 300px; border-radius: 10px;">
			<tbody>
				<tr>
					<td style="padding: 20px;">
						<h1 style="text-align:center; color: #333; margin-bottom: 30px;">Mise à jour du statut de votre compte créateur</h1>
						
						<div style="text-align:center; margin-bottom: 30px;">
							<p style="font-size: 16px; color: #444;">Bonjour %s %s,</p>
							<p style="font-size: 16px; color: #444;">Le statut de votre compte créateur pour %s a été mis à jour.</p>
						</div>

						<div style="text-align:center; margin-bottom: 20px;">
							<p style="font-size: 18px; color: #444; font-weight: bold;">Nouveau statut : %s</p>
						</div>

						<div style="text-align:center; margin-bottom: 20px;">
							%s
						</div>

						<div style="text-align:center; margin-bottom: 20px;">
							<p style="font-size: 16px; color: #444; margin-top: 30px;">L'équipe OnlyFlick</p>
						</div>
					</td>
				</tr>
			</tbody>
		</table>
	</div>
`, data.FirstName, data.LastName, data.CompanyName, statusFrench,
		getStatusMessage(data.Status))

	message := []byte(subject + mime + body)
	utils.SendMail(data.Email, message)
}

func getStatusMessage(status models.ContentCreatorStatusType) string {
	switch status {
	case models.ContentCreatorStatusApproved:
		return `<p style="font-size: 16px; color: #444;">Félicitations ! Votre compte créateur a été approuvé. Vous pouvez maintenant commencer à publier du contenu sur OnlyFlick.</p>`
	case models.ContentCreatorStatusRejected:
		return `<p style="font-size: 16px; color: #444;">Malheureusement, votre demande a été rejetée. Vous pouvez mettre à jour votre demande avec les informations nécessaires et la soumettre à nouveau.</p>`
	default:
		return `<p style="font-size: 16px; color: #444;">Notre équipe examine actuellement votre demande. Nous vous tiendrons informé de l'avancement.</p>`
	}
}
