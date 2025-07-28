package services

import (
	"fmt"
	"net/smtp"

	"2fa-system/config"
)

type EmailService struct {
	Config *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{Config: cfg}
}

func (e *EmailService) SendCodeEmail(to, code string) error {
	subject := "Seu código 2FA"
	body := fmt.Sprintf("Seu código de verificação é: %s", code)
	msg := "From: " + e.Config.FromEmail + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	auth := smtp.PlainAuth("", e.Config.SMTPUser, e.Config.SMTPPassword, e.Config.SMTPHost)
	return smtp.SendMail(
		e.Config.SMTPHost+":"+e.Config.SMTPPort,
		auth,
		e.Config.FromEmail,
		[]string{to},
		[]byte(msg),
	)
}
