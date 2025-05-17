package mailsmodels

import (
	"fmt"
	"pec2-backend/utils"
)

func ConfirmEmail(email string, code string) {
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
					<td style="text-align:center; padding-bottom: 30px;">Pour finaliser l'inscription, merci de valider votre email saisissant le code suivant sur l'application :</td>
				</tr>
				<tr>
					<td style="text-align:center; padding-bottom: 30px;">
						<p style="font-weight: bold; color: #722ED1; text-align:center;">%s</p>
					</td>
				</tr>
			</tbody>
		</table>
	</div>
`, code)

	message := []byte(subject + mime + body)

	utils.SendMail(email, message)
}
