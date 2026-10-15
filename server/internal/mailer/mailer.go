package mailer

import (
	"errors"
	"fmt"
	"mime"
	"net/mail"
	"net/smtp"
	"server/internal/config"
	"strconv"
	"strings"
	"sync"
)

type Client struct {
	addr string
	auth smtp.Auth
}

var (
	instance *Client
	once     sync.Once
)

func Get() *Client {
	once.Do(func() {

		cfg, _ := config.Get()

		SMTPUsername := cfg.SMTPUsername
		SMTPPassword := cfg.SMTPUsername
		SMTPPort := cfg.SMTPPort
		SMTPHost := cfg.SMTPHost

		var auth smtp.Auth
		if SMTPUsername != "" {
			auth = smtp.PlainAuth("", SMTPUsername, SMTPPassword, SMTPHost)
		}
		instance = &Client{
			addr: SMTPHost + ":" + strconv.Itoa(SMTPPort),
			auth: auth,
		}
	})
	return instance
}

func (c *Client) Send(from string, to []string, subject, htmlBody string) error {
	if len(to) == 0 {
		return errors.New("mailer: no recipients")
	}

	sender, err := mail.ParseAddress(from)
	if err != nil {
		return fmt.Errorf("mailer: invalid from address: %w", err)
	}

	rcpts := make([]string, 0, len(to))
	for _, t := range to {
		a, err := mail.ParseAddress(t)
		if err != nil {
			return fmt.Errorf("mailer: invalid to address %q: %w", t, err)
		}
		rcpts = append(rcpts, a.Address)
	}

	if strings.ContainsAny(subject, "\r\n") {
		return errors.New("mailer: subject must not contain line breaks")
	}

	var msg strings.Builder
	fmt.Fprintf(&msg, "From: %s\r\n", sender.String())
	fmt.Fprintf(&msg, "To: %s\r\n", strings.Join(rcpts, ", "))
	fmt.Fprintf(&msg, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	return smtp.SendMail(c.addr, c.auth, sender.Address, rcpts, []byte(msg.String()))
}
