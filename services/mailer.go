package services

import (
	"crypto/tls"
	"fmt"
	"os"
	"strconv"

	"github.com/wneessen/go-mail"
)

type mailer struct{}

var Mailer = mailer{}

func (mailer) SendMail(to string, subject string, content string) error {
	sender := os.Getenv("IMGU2_SMTP_SENDER")
	if sender == "" {
		return fmt.Errorf("sendmail: empty sender")
	}

	msg := mail.NewMsg()

	if err := msg.From(sender); err != nil {
		return fmt.Errorf("sendmail: %w", err)
	}

	if err := msg.To(to); err != nil {
		return fmt.Errorf("sendmail: %w", err)
	}

	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, content)

	username := os.Getenv("IMGU2_SMTP_USERNAME")
	password := os.Getenv("IMGU2_SMTP_PASSWORD")
	host := os.Getenv("IMGU2_SMTP_HOST")
	auth_tls, err := strconv.ParseBool(os.Getenv("IMGU2_SMTP_AUTH_TLS"))
	if err != nil {
		return fmt.Errorf("sendmail: invalid IMGU2_SMTP_AUTH_TLS: %s", os.Getenv("IMGU2_SMTP_AUTH_TLS"))
	}
	port, err := strconv.Atoi(os.Getenv("IMGU2_SMTP_PORT"))
	if err != nil {
		return fmt.Errorf("sendmail: invalid IMGU2_SMTP_PORT: %s", os.Getenv("IMGU2_SMTP_PORT"))
	}

	var tlsConfig mail.Option
	tls_ := tls.Config{InsecureSkipVerify: false, ServerName: host}
	if auth_tls {
		tlsConfig = mail.WithTLSConfig(&tls_)
	} else {
		tlsConfig = nil
	}

	client, err := mail.NewClient(
		host,
		mail.WithPort(port),
		tlsConfig,
		mail.WithSMTPAuth(mail.SMTPAuthLogin),
		mail.WithUsername(username),
		mail.WithPassword(password),
	)
	if err != nil {
		return fmt.Errorf("sendmail: %w", err)
	}

	err = client.DialAndSend(msg)
	if err != nil {
		return fmt.Errorf("sendmail: %w", err)
	}

	return nil
}
