package mailer

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"sync"

	gomail "github.com/wneessen/go-mail"

	"server/internal/config"
)

type Client struct {
	client *gomail.Client
}

var (
	instance *Client
	once     sync.Once
)

func Get() *Client {
	once.Do(func() {
		cfg, _ := config.Get()

		host := cfg.SMTPHost
		port := cfg.SMTPPort
		username := cfg.SMTPUsername
		password := cfg.SMTPPassword

		opts := []gomail.Option{
			gomail.WithPort(port),
			gomail.WithTLSPolicy(gomail.TLSOpportunistic),
		}

		if username != "" {
			opts = append(opts,
				gomail.WithSMTPAuth(gomail.SMTPAuthPlain),
				gomail.WithUsername(username),
				gomail.WithPassword(password),
			)
		}

		c, err := gomail.NewClient(host, opts...)
		if err != nil {
			panic(fmt.Sprintf("mailer: failed to create client: %v", err))
		}

		instance = &Client{client: c}
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

	msg := gomail.NewMsg()
	if err := msg.From(sender.Address); err != nil {
		return fmt.Errorf("mailer: invalid from address: %w", err)
	}

	if err := msg.To(rcpts...); err != nil {
		return fmt.Errorf("mailer: invalid to address: %w", err)
	}

	msg.Subject(subject)
	msg.SetBodyString(gomail.TypeTextHTML, htmlBody)

	if err := c.client.DialAndSend(msg); err != nil {
		return fmt.Errorf("mailer: send failed: %w", err)
	}

	return nil
}

func (c *Client) Verify(ctx context.Context) error {
	if err := c.client.DialWithContext(ctx); err != nil {
		return fmt.Errorf("mailer: dial/auth failed: %w", err)
	}
	defer func() {
		_ = c.client.Close()
	}()
	return nil
}
