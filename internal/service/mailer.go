package service

import (
	"crypto/tls"

	"github.com/LevOrlov5404/matcha/internal/config"
	"github.com/pkg/errors"
	goMail "gopkg.in/mail.v2"
)

type (
	MailerService struct {
		cfg    config.Mailer
		dialer *goMail.Dialer
	}
)

func NewMailerService(cfg config.Mailer) *MailerService {
	d := goMail.NewDialer(
		cfg.ServerAddress.Host, cfg.ServerAddress.Port, cfg.Username, cfg.Password,
	)
	d.Timeout = cfg.Timeout.Duration()
	d.TLSConfig = &tls.Config{
		ServerName:         cfg.ServerAddress.Host,
		InsecureSkipVerify: false,
	}

	return &MailerService{
		cfg:    cfg,
		dialer: d,
	}
}

func (s *MailerService) SendEmailConfirm(toEmail, emailConfirmToken string) error {
	m := goMail.NewMessage()

	m.SetHeader("From", s.cfg.Username)
	m.SetHeader("To", toEmail)
	m.SetHeader("To", "lev.orlov.5404@gmail.com")
	m.SetHeader("Subject", "Matcha registration")
	m.SetBody("text/plain",
		"We greet you.\nTo complete the registration go by this link.\n"+
			"localhost:8080/confirm-email?token="+emailConfirmToken+
			"\nThank you for choosing us :)")

	if err := s.dialer.DialAndSend(m); err != nil {
		return errors.Wrap(err, "failed to send email confirm")
	}

	//if err := s.dialer.DialAndSend(m); err != nil {
	//	return errors.Wrap(err, "failed to send email confirm")
	//}

	return nil
}
