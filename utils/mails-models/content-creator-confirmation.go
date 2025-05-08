package mailsmodels

import (
	"fmt"
	"pec2-backend/utils"
)

type ContentCreatorConfirmationData struct {
	FirstName   string
	LastName    string
	Email       string
	CompanyName string
	CompanyType string
	SiretNumber string
}

func ContentCreatorConfirmation(data ContentCreatorConfirmationData) {
	subject := "Subject: Confirmation de votre demande de compte créateur - OnlyFlick \r\n"
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	body := fmt.Sprintf(`
	<div style="background-color: #722ED1; width: 100%%; min-height: 300px; padding: 30px; box-sizing:border-box">
		<table style="background-color: #ffffff; width: 100%%; min-height: 300px; border-radius: 10px;">
			<tbody>
				<tr>
					<td style="padding: 20px;">
						<h1 style="text-align:center; color: #333; margin-bottom: 30px;">Demande de compte créateur reçue !</h1>
						
						<div style="text-align:center; margin-bottom: 30px;">
							<p style="font-size: 16px; color: #444;">Bonjour %s %s,</p>
							<p style="font-size: 16px; color: #444;">Nous avons bien reçu votre demande de compte créateur pour votre entreprise.</p>
						</div>

						<div style="text-align:center; margin-bottom: 20px;">
							<p style="font-size: 16px; color: #444; font-weight: bold;">Récapitulatif de votre demande :</p>
						</div>

						<div style="background-color: #f5f5f5; padding: 25px; border-radius: 10px; width: 80%%; max-width: 500px; margin: 0 auto 30px auto;">
							<table style="width: 100%%; text-align: center;">
								<tr>
									<td style="padding: 10px;">
										<strong style="color: #722ED1;">Entreprise</strong>
										<p style="margin: 5px 0; color: #444;">%s</p>
									</td>
								</tr>
								<tr>
									<td style="padding: 10px;">
										<strong style="color: #722ED1;">Type d'entreprise</strong>
										<p style="margin: 5px 0; color: #444;">%s</p>
									</td>
								</tr>
								<tr>
									<td style="padding: 10px;">
										<strong style="color: #722ED1;">Numéro SIRET</strong>
										<p style="margin: 5px 0; color: #444;">%s</p>
									</td>
								</tr>
							</table>
						</div>

						<div style="text-align:center; margin-bottom: 20px;">
							<p style="font-size: 16px; color: #444;">Notre équipe va examiner votre demande dans les plus brefs délais.</p>
							<p style="font-size: 16px; color: #444;">Vous recevrez un email dès que votre demande aura été traitée.</p>
							<p style="font-size: 16px; color: #444; margin-top: 30px; font-weight: bold;">Merci de votre confiance !</p>
						</div>
					</td>
				</tr>
			</tbody>
		</table>
	</div>
`, data.FirstName, data.LastName, data.CompanyName, data.CompanyType, data.SiretNumber)

	message := []byte(subject + mime + body)
	utils.SendMail(data.Email, message)
}
