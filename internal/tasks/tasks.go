package tasks

import (
	"bytes"
	"errors"
	"log"
	"net/smtp"
	"text/template"
)

type emailData struct {
	From    string
	To      string
	Payload any
}

func sendEmailWithData(configEmail string, configPassword string, email, emailTemplate string, data emailData) error {
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	smtpUsername := configEmail
	smtpPassword := configPassword

	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)

	data.From = smtpUsername

	var body bytes.Buffer
	t := template.Must(template.New("email").Parse(emailTemplate))
	if err := t.Execute(&body, data); err != nil {
		log.Printf("error generating email: %v", err)
		return errors.New("error generating email")
	}

	err := smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		smtpUsername,
		[]string{email},
		body.Bytes(),
	)
	if err != nil {
		log.Printf("error sending email: %v", err)
		return errors.New("error sending email")
	}

	return nil
}

func SendVerificationEmail(configEmail string, configPassword string) func(email string, code string) error {
	return func(email string, code string) error {
		emailTemplate := `From: {{.From}}
		To: {{.To}}
		Subject: Shortener Email Verification

		Hello,

		Please use the following code to verify your email address: {{.Payload}}

		If you didn't request this, please ignore this email.

		Best regards,
		Shortener Team
	`
		data := emailData{
			To:      email,
			Payload: code,
		}
		return sendEmailWithData(configEmail, configPassword, email, emailTemplate, data)
	}
}

func SendResetPasswordEmail(configEmail string, configPassword string) func(email string, code string) error {
	return func(email string, link string) error {
		emailTemplate := `From: {{.From}}
		To: {{.To}}
		Subject: Shortener Password Reset

		Hello,

		Please use the following link to reset your password: {{.Payload}}

		If you didn't request this, please ignore this email.

		Best regards,
		Shortener Team
	`
		data := emailData{
			To:      email,
			Payload: link,
		}
		return sendEmailWithData(configEmail, configPassword, email, emailTemplate, data)
	}
}
