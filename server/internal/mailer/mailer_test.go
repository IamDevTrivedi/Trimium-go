package mailer

import (
	"bufio"
	"context"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	gomail "github.com/wneessen/go-mail"
)

func resetState() {
	instance = nil
	once = sync.Once{}
}

// startMockSMTPServer starts an in-memory SMTP server on 127.0.0.1:0 for testing.
func startMockSMTPServer(t *testing.T) (string, int, func()) {
	t.Helper()

	var lc net.ListenConfig

	ln, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock SMTP listener: %v", err)
	}

	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("expected TCP listener address, got %T", ln.Addr())
	}
	stop := make(chan struct{})

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-stop:
					return
				default:
					return
				}
			}

			go handleMockSMTPConn(conn)
		}
	}()

	cleanup := func() {
		close(stop)
		_ = ln.Close()
	}

	return addr.IP.String(), addr.Port, cleanup
}

func handleMockSMTPConn(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Initial greeting
	_, _ = writer.WriteString("220 127.0.0.1 Simple Mail Server\r\n")
	_ = writer.Flush()

	inData := false
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		line = strings.TrimRight(line, "\r\n")

		if inData {
			if line == "." {
				inData = false
				_, _ = writer.WriteString("250 2.0.0 OK: message queued\r\n")
				_ = writer.Flush()
			}
			continue
		}

		upper := strings.ToUpper(line)
		switch {
		case strings.HasPrefix(upper, "EHLO") || strings.HasPrefix(upper, "HELO"):
			_, _ = writer.WriteString("250-127.0.0.1\r\n250 8BITMIME\r\n")
			_ = writer.Flush()
		case strings.HasPrefix(upper, "MAIL FROM:"):
			_, _ = writer.WriteString("250 2.1.0 Sender OK\r\n")
			_ = writer.Flush()
		case strings.HasPrefix(upper, "RCPT TO:"):
			_, _ = writer.WriteString("250 2.1.5 Recipient OK\r\n")
			_ = writer.Flush()
		case upper == "DATA":
			inData = true
			_, _ = writer.WriteString("354 Start mail input; end with <CRLF>.<CRLF>\r\n")
			_ = writer.Flush()
		case upper == "QUIT":
			_, _ = writer.WriteString("221 2.0.0 Bye\r\n")
			_ = writer.Flush()
			return
		case upper == "RSET" || upper == "NOOP":
			_, _ = writer.WriteString("250 2.0.0 OK\r\n")
			_ = writer.Flush()
		default:
			_, _ = writer.WriteString("500 5.5.2 Command unrecognized\r\n")
			_ = writer.Flush()
		}
	}
}

func TestSend_Validation(t *testing.T) {
	// Create a dummy client without dialing
	gc, err := gomail.NewClient("127.0.0.1", gomail.WithPort(2525), gomail.WithTLSPolicy(gomail.TLSOpportunistic))
	if err != nil {
		t.Fatalf("unexpected error creating gomail client: %v", err)
	}
	client := &Client{client: gc}

	tests := []struct {
		name      string
		from      string
		to        []string
		subject   string
		body      string
		errSubstr string
	}{
		{
			name:      "no recipients",
			from:      "sender@example.com",
			to:        []string{},
			subject:   "Test",
			body:      "<p>body</p>",
			errSubstr: "no recipients",
		},
		{
			name:      "invalid from address",
			from:      "invalid-email-address",
			to:        []string{"recipient@example.com"},
			subject:   "Test",
			body:      "<p>body</p>",
			errSubstr: "invalid from address",
		},
		{
			name:      "invalid to address",
			from:      "sender@example.com",
			to:        []string{"valid@example.com", "not-an-email"},
			subject:   "Test",
			body:      "<p>body</p>",
			errSubstr: "invalid to address",
		},
		{
			name:      "subject with newline",
			from:      "sender@example.com",
			to:        []string{"recipient@example.com"},
			subject:   "Line1\nLine2",
			body:      "<p>body</p>",
			errSubstr: "must not contain line breaks",
		},
		{
			name:      "subject with carriage return",
			from:      "sender@example.com",
			to:        []string{"recipient@example.com"},
			subject:   "Line1\rLine2",
			body:      "<p>body</p>",
			errSubstr: "must not contain line breaks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.Send(tt.from, tt.to, tt.subject, tt.body)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.errSubstr)
			}
			if !strings.Contains(err.Error(), tt.errSubstr) {
				t.Errorf("expected error containing %q, got %q", tt.errSubstr, err.Error())
			}
		})
	}
}

func TestVerify_Success(t *testing.T) {
	host, port, cleanup := startMockSMTPServer(t)
	defer cleanup()

	gc, err := gomail.NewClient(host, gomail.WithPort(port), gomail.WithTLSPolicy(gomail.TLSOpportunistic))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client := &Client{client: gc}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Verify(ctx); err != nil {
		t.Fatalf("expected Verify() to succeed, got error: %v", err)
	}
}

func TestVerify_Failure(t *testing.T) {
	// Point to an unavailable port
	gc, err := gomail.NewClient("127.0.0.1", gomail.WithPort(64999), gomail.WithTLSPolicy(gomail.TLSOpportunistic))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client := &Client{client: gc}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := client.Verify(ctx); err == nil {
		t.Fatal("expected Verify() to fail with unavailable server, got nil")
	}
}

func TestSend_Success(t *testing.T) {
	host, port, cleanup := startMockSMTPServer(t)
	defer cleanup()

	gc, err := gomail.NewClient(host, gomail.WithPort(port), gomail.WithTLSPolicy(gomail.TLSOpportunistic))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client := &Client{client: gc}

	err = client.Send(
		"sender@example.com",
		[]string{"recipient1@example.com", "Recipient Two <recipient2@example.com>"},
		"Welcome to Trimium",
		"<h1>Welcome</h1><p>Your account is ready.</p>",
	)
	if err != nil {
		t.Fatalf("expected Send() to succeed, got error: %v", err)
	}
}

func TestGet_ReturnsCachedInstance(t *testing.T) {
	t.Cleanup(resetState)
	resetState()

	gc, err := gomail.NewClient("127.0.0.1", gomail.WithPort(2525), gomail.WithTLSPolicy(gomail.TLSOpportunistic))
	if err != nil {
		t.Fatalf("unexpected error creating gomail client: %v", err)
	}

	want := &Client{client: gc}
	instance = want
	once.Do(func() {})

	got := Get()
	if got == nil {
		t.Fatal("expected non-nil Client from Get()")
	}
	if got != want {
		t.Errorf("expected cached client %p, got %p", want, got)
	}
	if got.client != gc {
		t.Errorf("expected underlying gomail client %p, got %p", gc, got.client)
	}
}
