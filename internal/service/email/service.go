package email

import (
	"EduSync/internal/service"
	"context"
	"fmt"
	"gopkg.in/gomail.v2"
)

type smtpEmailService struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewSMTPEmailService(host string, port int, username, password, from string) service.EmailService {
	return &smtpEmailService{host, port, username, password, from}
}

func (s *smtpEmailService) SendCode(ctx context.Context, toEmail, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)
	d := gomail.NewDialer(s.host, s.port, s.username, s.password)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}
