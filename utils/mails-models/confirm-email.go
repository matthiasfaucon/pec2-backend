package mailsmodels

import (
	"fmt"
	"pec2-backend/utils"
)

func ConfirmEmail(email string, token string) {
	link := "http://localhost:8080/valid-email/" + token
	subject := "Subject: Inscription à OnlyFlick \r\n"
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	body := fmt.Sprintf(`
	<div style="background-color: #722ED1; width: 100%%; min-height: 300px; padding: 30px; box-sizing:border-box">
		<table style="background-color: #ffffff; width: 100%%;  min-height: 300px;">
			<tbody>
				<tr>
					<td><h1 style="text-align:center">Merci d'avoir rejoint OnlyFlick</h1></td>
				</tr>
				<tr>
					<td style="text-align:center; padding-bottom: 30px;">Pour finaliser l'inscription, merci de valider votre email en cliquant sur le lien ci-dessous</td>
				</tr>
				<tr>
					<td style="text-align:center; padding-bottom: 30px;">
						<a href="%s" style="background-color: #722ED1; color: #ffffff; padding:20px; border-radius: 10px; font-weight: bold;">Confirmer mon email</a>
					</td>
				</tr>
			</tbody>
		</table>
	</div>
`, link)

	message := []byte(subject + mime + body)

	utils.SendMail(email, message)
}
