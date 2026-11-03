package mailer

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	gomail "github.com/wneessen/go-mail"
)

var (
	ErrNoRecipients     = errors.New("mailer: no recipients")
	ErrInvalidFrom      = errors.New("mailer: invalid from address")
	ErrInvalidRecipient = errors.New("mailer: invalid recipient address")
	ErrInvalidSubject   = errors.New("mailer: subject must not contain line breaks")
	ErrDialAndSend      = errors.New("mailer: dial/send failed")
	ErrAlreadyClosed    = errors.New("mailer: client already closed")
)

type TLSPolicy int

const (
	TLSMandatory TLSPolicy = iota
	TLSOpportunistic
	TLSNone
)

func (p TLSPolicy) toGomail() (gomail.TLSPolicy, error) {
	switch p {
	case TLSMandatory:
		return gomail.TLSMandatory, nil
	case TLSOpportunistic:
		return gomail.TLSOpportunistic, nil
	case TLSNone:
		return gomail.NoTLS, nil
	default:
		return 0, fmt.Errorf("mailer: invalid TLS policy %d", p)
	}
}

type Options struct {
	Host     string
	Port     int
	Username string
	Password string
	TLS      TLSPolicy
}

type Client struct {
	client *gomail.Client
}

func New(opts Options) (*Client, error) {
	if opts.Host == "" {
		return nil, errors.New("mailer: host is required")
	}

	if opts.Port <= 0 || opts.Port > 65535 {
		return nil, fmt.Errorf("mailer: invalid port %d", opts.Port)
	}

	policy, err := opts.TLS.toGomail()
	if err != nil {
		return nil, err
	}

	mailOpts := []gomail.Option{
		gomail.WithPort(opts.Port),
		gomail.WithTLSPolicy(policy),
	}

	if opts.Username != "" {
		mailOpts = append(mailOpts,
			gomail.WithSMTPAuth(gomail.SMTPAuthPlain),
			gomail.WithUsername(opts.Username),
			gomail.WithPassword(opts.Password),
		)
	}

	c, err := gomail.NewClient(opts.Host, mailOpts...)

	if err != nil {
		return nil, fmt.Errorf("mailer: create client: %w", err)
	}

	return &Client{client: c}, nil
}

func (c *Client) Close() error {
	if c == nil || c.client == nil {
		return ErrAlreadyClosed
	}

	if err := c.client.Close(); err != nil {
		return fmt.Errorf("mailer: close: %w", err)
	}

	return nil
}

func (c *Client) Send(
	ctx context.Context,
	from string,
	to []string,
	subject, htmlBody string,
) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("mailer: send aborted: %w", err)
	}

	if len(to) == 0 {
		return ErrNoRecipients
	}

	if strings.ContainsAny(subject, "\r\n") {
		return ErrInvalidSubject
	}

	sender, err := mail.ParseAddress(from)
	if err != nil {
		return fmt.Errorf("%w: %q: %w", ErrInvalidFrom, from, err)
	}

	rcpts := make([]string, 0, len(to))
	for _, t := range to {
		a, err := mail.ParseAddress(t)
		if err != nil {
			return fmt.Errorf("%w: %q: %w", ErrInvalidRecipient, t, err)
		}
		rcpts = append(rcpts, a.Address)
	}

	msg := gomail.NewMsg()
	if err := msg.From(sender.Address); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidFrom, err)
	}

	if err := msg.To(rcpts...); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidRecipient, err)
	}

	msg.Subject(subject)
	msg.SetBodyString(gomail.TypeTextHTML, htmlBody)

	if err := c.client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("%w: %w", ErrDialAndSend, err)
	}

	return nil
}
