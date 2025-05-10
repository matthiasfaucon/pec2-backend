package mailsmodels

import (
	"fmt"
	"pec2-backend/utils"
)

type ContentCreatorUpdateData struct {
	FirstName     string
	LastName      string
	Email         string
	CompanyName   string
	CompanyType   string
	SiretNumber   string
	VatNumber     string
	StreetAddress string
	PostalCode    string
	City          string
	Country       string
	Iban          string
	Bic           string
}

func ContentCreatorUpdate(data ContentCreatorUpdateData) {
	subject := "Subject: Mise à jour de votre compte créateur - OnlyFlick \r\n"
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	body := fmt.Sprintf(`
	<div style="background-color: #722ED1; width: 100%%; min-height: 300px; padding: 30px; box-sizing:border-box">
		<table style="background-color: #ffffff; width: 100%%; min-height: 300px; border-radius: 10px;">
			<tbody>
				<tr>
					<td style="padding: 20px;">
						<h1 style="text-align:center; color: #333; margin-bottom: 30px;">Mise à jour de votre compte créateur</h1>
						
						<div style="text-align:center; margin-bottom: 30px;">
							<p style="font-size: 16px; color: #444;">Bonjour %s %s,</p>
							<p style="font-size: 16px; color: #444;">Nous avons bien reçu les modifications apportées à votre compte créateur.</p>
						</div>

						<div style="text-align:center; margin-bottom: 20px;">
							<p style="font-size: 16px; color: #444; font-weight: bold;">Récapitulatif de vos modifications :</p>
						</div>

						<div style="background-color: #f5f5f5; padding: 25px; border-radius: 10px; width: 80%%; max-width: 500px; margin: 0 auto 30px auto;">
							<table style="width: 100%%; text-align: left;">
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
								<tr>
									<td style="padding: 10px;">
										<strong style="color: #722ED1;">Numéro TVA</strong>
										<p style="margin: 5px 0; color: #444;">%s</p>
									</td>
								</tr>
								<tr>
									<td style="padding: 10px;">
										<strong style="color: #722ED1;">Adresse</strong>
										<p style="margin: 5px 0; color: #444;">%s</p>
										<p style="margin: 5px 0; color: #444;">%s %s</p>
										<p style="margin: 5px 0; color: #444;">%s</p>
									</td>
								</tr>
								<tr>
									<td style="padding: 10px;">
										<strong style="color: #722ED1;">Coordonnées bancaires</strong>
										<p style="margin: 5px 0; color: #444;">IBAN : %s</p>
										<p style="margin: 5px 0; color: #444;">BIC : %s</p>
									</td>
								</tr>
							</table>
						</div>

						<div style="text-align:center; margin-bottom: 20px;">
							<p style="font-size: 16px; color: #444;">Vos modifications ont été enregistrées et votre compte est maintenant en attente de validation.</p>
							<p style="font-size: 16px; color: #444;">Notre équipe va examiner ces changements dans les plus brefs délais.</p>
							<p style="font-size: 16px; color: #444; margin-top: 30px; font-weight: bold;">Merci de votre confiance !</p>
						</div>
					</td>
				</tr>
			</tbody>
		</table>
	</div>
`, data.FirstName, data.LastName, data.CompanyName, data.CompanyType, data.SiretNumber, data.VatNumber,
		data.StreetAddress, data.PostalCode, data.City, data.Country, data.Iban, data.Bic)

	message := []byte(subject + mime + body)
	utils.SendMail(data.Email, message)
}
