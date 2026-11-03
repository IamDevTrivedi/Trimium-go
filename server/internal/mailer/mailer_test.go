package mailer

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/quotedprintable"
	"net"
	"strings"
	"sync"
	"testing"

	gomail "github.com/wneessen/go-mail"
)

const (
	crlf = "\r\n"

	testAuthUser = "mailer-user"
	testAuthPass = "mailer-secret"
)

type serverConfig struct {
	advertiseAuth bool

	rejectRcpt bool
}

type testSMTPServer struct {
	ln net.Listener

	advertiseAuth bool
	rejectRcpt    bool

	mu   sync.Mutex
	wg   sync.WaitGroup
	done chan struct{}

	mailFrom     string
	rcpts        []string
	data         string
	authAttempts int
	authUserGot  string
	authPassGot  string
}

func newTestSMTPServer(t *testing.T, cfg serverConfig) *testSMTPServer {
	t.Helper()

	var lc net.ListenConfig
	ln, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("start test SMTP server: %v", err)
	}

	s := &testSMTPServer{
		ln:            ln,
		advertiseAuth: cfg.advertiseAuth,
		rejectRcpt:    cfg.rejectRcpt,
		done:          make(chan struct{}),
	}

	s.wg.Add(1)
	go s.acceptLoop()

	t.Cleanup(func() {
		close(s.done)
		_ = ln.Close()
		s.wg.Wait()
	})

	return s
}

func (s *testSMTPServer) port() int {
	tcpAddr, ok := s.ln.Addr().(*net.TCPAddr)
	if !ok {
		return 0
	}
	return tcpAddr.Port
}

func (s *testSMTPServer) acceptLoop() {
	defer s.wg.Done()
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		s.wg.Add(1)
		go s.handle(conn)
	}
}

func (s *testSMTPServer) handle(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	stop := make(chan struct{})
	defer close(stop)
	go func() {
		select {
		case <-s.done:
			_ = conn.Close()
		case <-stop:
		}
	}()

	if _, err := conn.Write([]byte("220 test.local ESMTP mailer test service ready" + crlf)); err != nil {
		return
	}

	br := bufio.NewReader(conn)
	reply := func(status int, text string) bool {
		_, err := conn.Write(fmt.Appendf(nil, "%d %s%s", status, text, crlf))
		return err == nil
	}

	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")

		switch cmd := strings.ToUpper(line); {
		case strings.HasPrefix(cmd, "EHLO"):
			if !s.ehloReply(conn) {
				return
			}
		case strings.HasPrefix(cmd, "HELO"):
			if !reply(250, "test.local") {
				return
			}
		case strings.HasPrefix(cmd, "MAIL FROM:"):
			s.mu.Lock()
			s.mailFrom = line
			s.mu.Unlock()
			if !reply(250, "OK") {
				return
			}
		case strings.HasPrefix(cmd, "RCPT TO:"):
			s.mu.Lock()
			s.rcpts = append(s.rcpts, line)
			s.mu.Unlock()
			if s.rejectRcpt {
				if !reply(550, "5.1.1 mailbox unavailable") {
					return
				}
				continue
			}
			if !reply(250, "OK") {
				return
			}
		case strings.HasPrefix(cmd, "AUTH"):
			if !s.handleAuthPlain(conn, br, line) {
				return
			}
		case strings.HasPrefix(cmd, "DATA"):
			if !reply(354, "End data with <CR><LF>.<CR><LF>") {
				return
			}
			if !s.readData(br) {
				return
			}
			if !reply(250, "OK: message accepted for delivery") {
				return
			}
		case strings.HasPrefix(cmd, "RSET"), strings.HasPrefix(cmd, "NOOP"):
			if !reply(250, "OK") {
				return
			}
		case strings.HasPrefix(cmd, "QUIT"):
			_ = reply(221, "test.local closing transmission channel")
			return
		default:
			if !reply(500, "5.5.2 command not recognized") {
				return
			}
		}
	}
}

func (s *testSMTPServer) ehloReply(conn net.Conn) bool {
	var b strings.Builder
	b.WriteString("250-test.local" + crlf)
	if s.advertiseAuth {
		b.WriteString("250-AUTH PLAIN" + crlf)
	}
	b.WriteString("250 OK" + crlf)
	_, err := conn.Write([]byte(b.String()))
	return err == nil
}

func (s *testSMTPServer) handleAuthPlain(conn net.Conn, br *bufio.Reader, line string) bool {
	payload := ""
	if fields := strings.SplitN(line, " ", 3); len(fields) == 3 {
		payload = fields[2]
	} else {
		if _, err := conn.Write([]byte("334 " + crlf)); err != nil {
			return false
		}
		next, err := br.ReadString('\n')
		if err != nil {
			return false
		}
		payload = strings.TrimSpace(next)
	}

	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		_, _ = conn.Write([]byte("501 5.5.2 invalid base64 data" + crlf))
		return true
	}

	user, pass := "", ""
	if parts := strings.Split(string(raw), "\x00"); len(parts) == 3 {
		user, pass = parts[1], parts[2]
	}

	s.mu.Lock()
	s.authAttempts++
	s.authUserGot = user
	s.authPassGot = pass
	s.mu.Unlock()

	if user == testAuthUser && pass == testAuthPass {
		_, err = conn.Write([]byte("235 2.7.0 authentication successful" + crlf))
	} else {
		_, err = conn.Write([]byte("535 5.7.8 authentication credentials rejected" + crlf))
	}
	return err == nil
}

func (s *testSMTPServer) readData(br *bufio.Reader) bool {
	var b strings.Builder
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return false
		}
		b.WriteString(line)
		if line == "."+crlf || line == ".\n" {
			break
		}
	}
	s.mu.Lock()
	s.data += b.String()
	s.mu.Unlock()
	return true
}

func (s *testSMTPServer) snapshot() (mailFrom string, rcpts []string, data string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mailFrom, append([]string(nil), s.rcpts...), s.data
}

func (s *testSMTPServer) authSnapshot() (attempts int, user, pass string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.authAttempts, s.authUserGot, s.authPassGot
}

func closedTCPPort(t *testing.T) int {
	t.Helper()
	var lc net.ListenConfig
	ln, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("pick closed port: %v", err)
	}
	tcpAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("unexpected addr type: %T", ln.Addr())
	}
	port := tcpAddr.Port
	_ = ln.Close()
	return port
}

func newTestClient(t *testing.T, host string, port int, tls TLSPolicy) *Client {
	t.Helper()
	c, err := New(Options{Host: host, Port: port, TLS: tls})
	if err != nil {
		t.Fatalf("New() = %v, want nil error", err)
	}
	return c
}

func cancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func decodedBody(t *testing.T, data string) (headers, body string) {
	t.Helper()

	idx := strings.Index(data, crlf+crlf)
	if idx < 0 {
		t.Fatalf("no header/body separator found in DATA payload:\n%s", data)
	}
	headers, body = data[:idx], data[idx+len(crlf+crlf):]

	body = strings.TrimSuffix(body, crlf+".")

	switch enc := strings.ToLower(headers); {
	case strings.Contains(enc, "content-transfer-encoding: base64"):
		clean := strings.Map(func(r rune) rune {
			if r == '\r' || r == '\n' {
				return -1
			}
			return r
		}, body)
		dec, err := base64.StdEncoding.DecodeString(clean)
		if err != nil {
			t.Fatalf("decode base64 message body: %v", err)
		}
		return headers, string(dec)
	case strings.Contains(enc, "content-transfer-encoding: quoted-printable"):
		dec, err := io.ReadAll(quotedprintable.NewReader(strings.NewReader(body)))
		if err != nil {
			t.Fatalf("decode quoted-printable message body: %v", err)
		}
		return headers, string(dec)
	default:
		return headers, body
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr string
	}{
		{
			name: "valid minimal options",
			opts: Options{Host: "smtp.example.com", Port: 587, TLS: TLSMandatory},
		},
		{
			name: "valid with authentication",
			opts: Options{Host: "smtp.example.com", Port: 2525, Username: "user", Password: "pass", TLS: TLSOpportunistic},
		},
		{
			name: "valid lowest port",
			opts: Options{Host: "smtp.example.com", Port: 1, TLS: TLSNone},
		},
		{
			name: "valid highest port",
			opts: Options{Host: "smtp.example.com", Port: 65535, TLS: TLSNone},
		},
		{
			name:    "missing host",
			opts:    Options{Port: 587, TLS: TLSMandatory},
			wantErr: "host is required",
		},
		{
			name:    "zero port",
			opts:    Options{Host: "smtp.example.com", TLS: TLSMandatory},
			wantErr: "invalid port 0",
		},
		{
			name:    "negative port",
			opts:    Options{Host: "smtp.example.com", Port: -1, TLS: TLSMandatory},
			wantErr: "invalid port -1",
		},
		{
			name:    "port above 65535",
			opts:    Options{Host: "smtp.example.com", Port: 65536, TLS: TLSMandatory},
			wantErr: "invalid port 65536",
		},
		{
			name:    "unknown TLS policy",
			opts:    Options{Host: "smtp.example.com", Port: 587, TLS: TLSPolicy(99)},
			wantErr: "invalid TLS policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := New(tt.opts)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("New() error = nil, want error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("New() error = %q, want it to contain %q", err, tt.wantErr)
				}
				if c != nil {
					t.Fatalf("New() client = %v, want nil on error", c)
				}
				return
			}

			if err != nil {
				t.Fatalf("New() unexpected error: %v", err)
			}
			if c == nil || c.client == nil {
				t.Fatal("New() client = nil, want a usable client")
			}
		})
	}
}

func TestTLSPolicyToGomail(t *testing.T) {
	tests := []struct {
		name    string
		policy  TLSPolicy
		want    gomail.TLSPolicy
		wantErr bool
	}{
		{name: "TLSMandatory maps to go-mail mandatory", policy: TLSMandatory, want: gomail.TLSMandatory},
		{name: "TLSOpportunistic maps to go-mail opportunistic", policy: TLSOpportunistic, want: gomail.TLSOpportunistic},
		{name: "TLSNone maps to go-mail NoTLS", policy: TLSNone, want: gomail.NoTLS},
		{name: "unknown policy is rejected", policy: TLSPolicy(42), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.policy.toGomail()

			if tt.wantErr {
				if err == nil {
					t.Fatal("toGomail() error = nil, want error")
				}
				if !strings.Contains(err.Error(), "invalid TLS policy") {
					t.Fatalf("toGomail() error = %q, want it to mention the invalid policy", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("toGomail() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("toGomail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClose(t *testing.T) {
	t.Run("nil receiver returns ErrAlreadyClosed", func(t *testing.T) {
		var c *Client
		if err := c.Close(); !errors.Is(err, ErrAlreadyClosed) {
			t.Fatalf("Close() on nil client = %v, want ErrAlreadyClosed", err)
		}
	})

	t.Run("zero value client returns ErrAlreadyClosed", func(t *testing.T) {
		c := &Client{}
		if err := c.Close(); !errors.Is(err, ErrAlreadyClosed) {
			t.Fatalf("Close() on zero value client = %v, want ErrAlreadyClosed", err)
		}
	})

	t.Run("fresh client closes without error", func(t *testing.T) {
		c := newTestClient(t, "smtp.example.com", 587, TLSMandatory)
		if err := c.Close(); err != nil {
			t.Fatalf("Close() = %v, want nil", err)
		}
	})
}

func TestSendValidationErrors(t *testing.T) {
	c := newTestClient(t, "127.0.0.1", closedTCPPort(t), TLSNone)

	tests := []struct {
		name        string
		ctx         func() context.Context
		from        string
		to          []string
		subject     string
		wantErr     error
		wantContain string
	}{
		{
			name: "cancelled context wins over other validation",
			ctx:  cancelledContext,
			to:   []string{"rcpt@example.com"},

			wantErr: context.Canceled,
		},
		{
			name:    "no recipients",
			ctx:     context.Background,
			from:    "sender@example.com",
			subject: "no recipients",
			wantErr: ErrNoRecipients,
		},
		{
			name:    "empty recipient address",
			ctx:     context.Background,
			from:    "sender@example.com",
			to:      []string{""},
			subject: "empty rcpt",
			wantErr: ErrInvalidRecipient,
		},
		{
			name:        "malformed recipient address",
			ctx:         context.Background,
			from:        "sender@example.com",
			to:          []string{"rcpt1@example.com", "not-an-email"},
			subject:     "bad rcpt",
			wantErr:     ErrInvalidRecipient,
			wantContain: `"not-an-email"`,
		},
		{
			name:    "missing from address",
			ctx:     context.Background,
			to:      []string{"rcpt@example.com"},
			subject: "missing from",
			wantErr: ErrInvalidFrom,
		},
		{
			name:        "malformed from address",
			ctx:         context.Background,
			from:        "not-an-email",
			to:          []string{"rcpt@example.com"},
			subject:     "bad from",
			wantErr:     ErrInvalidFrom,
			wantContain: `"not-an-email"`,
		},
		{
			name:    "subject with line feed",
			ctx:     context.Background,
			from:    "sender@example.com",
			to:      []string{"rcpt@example.com"},
			subject: "hello\nBcc: victim@example.com",
			wantErr: ErrInvalidSubject,
		},
		{
			name:    "subject with carriage return",
			ctx:     context.Background,
			from:    "sender@example.com",
			to:      []string{"rcpt@example.com"},
			subject: "hello\rBcc: victim@example.com",
			wantErr: ErrInvalidSubject,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.Send(tt.ctx(), tt.from, tt.to, tt.subject, "<p>body</p>")

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Send() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantContain != "" && !strings.Contains(err.Error(), tt.wantContain) {
				t.Fatalf("Send() error = %q, want it to contain %q", err, tt.wantContain)
			}
		})
	}
}

func TestSendDialFailure(t *testing.T) {
	c := newTestClient(t, "127.0.0.1", closedTCPPort(t), TLSNone)

	err := c.Send(context.Background(), "sender@example.com",
		[]string{"rcpt@example.com"}, "dial failure", "<p>hello</p>")

	if !errors.Is(err, ErrDialAndSend) {
		t.Fatalf("Send() to unreachable server = %v, want ErrDialAndSend", err)
	}
}

func TestSendTLSMandatoryFailsWithoutSTARTTLS(t *testing.T) {
	srv := newTestSMTPServer(t, serverConfig{})
	c := newTestClient(t, "127.0.0.1", srv.port(), TLSMandatory)

	err := c.Send(context.Background(), "sender@example.com",
		[]string{"rcpt@example.com"}, "tls mandatory", "<p>hello</p>")

	if !errors.Is(err, ErrDialAndSend) {
		t.Fatalf("Send() with TLSMandatory against plain server = %v, want ErrDialAndSend", err)
	}
}

func TestSendServerRejectsRecipient(t *testing.T) {
	srv := newTestSMTPServer(t, serverConfig{rejectRcpt: true})
	c := newTestClient(t, "127.0.0.1", srv.port(), TLSNone)

	err := c.Send(context.Background(), "sender@example.com",
		[]string{"rcpt@example.com"}, "rejected rcpt", "<p>hello</p>")

	if !errors.Is(err, ErrDialAndSend) {
		t.Fatalf("Send() with server-side RCPT rejection = %v, want ErrDialAndSend", err)
	}
}

func TestSendAuthRejected(t *testing.T) {
	srv := newTestSMTPServer(t, serverConfig{advertiseAuth: true})
	c, err := New(Options{
		Host:     "127.0.0.1",
		Port:     srv.port(),
		Username: testAuthUser,
		Password: "wrong-password",
		TLS:      TLSNone,
	})
	if err != nil {
		t.Fatalf("New() = %v, want nil error", err)
	}

	err = c.Send(context.Background(), "sender@example.com",
		[]string{"rcpt@example.com"}, "bad credentials", "<p>hello</p>")

	if !errors.Is(err, ErrDialAndSend) {
		t.Fatalf("Send() with rejected credentials = %v, want ErrDialAndSend", err)
	}

	attempts, user, _ := srv.authSnapshot()
	if attempts == 0 {
		t.Fatal("server never received an AUTH attempt")
	}
	if user != testAuthUser {
		t.Fatalf("server authenticated as %q, want %q", user, testAuthUser)
	}
}

func TestSendSuccess(t *testing.T) {
	srv := newTestSMTPServer(t, serverConfig{})
	c := newTestClient(t, "127.0.0.1", srv.port(), TLSNone)

	const (
		from    = "Mailer Sender <sender@example.com>"
		subject = "Test Subject"
		body    = "<html><body><p>Hello from the mailer test suite</p></body></html>"
	)

	to := []string{"rcpt1@example.com", "Rcpt Two <rcpt2@example.com>"}

	if err := c.Send(context.Background(), from, to, subject, body); err != nil {
		t.Fatalf("Send() = %v, want nil", err)
	}

	mailFrom, rcpts, data := srv.snapshot()

	if !strings.Contains(mailFrom, "<sender@example.com>") {
		t.Errorf("envelope MAIL FROM = %q, want it to contain <sender@example.com>", mailFrom)
	}
	if len(rcpts) != 2 {
		t.Fatalf("server received %d RCPT commands (%v), want 2", len(rcpts), rcpts)
	}
	for i, want := range []string{"<rcpt1@example.com>", "<rcpt2@example.com>"} {
		if !strings.Contains(rcpts[i], want) {
			t.Errorf("RCPT #%d = %q, want it to contain %s", i+1, rcpts[i], want)
		}
	}

	headers, gotBody := decodedBody(t, data)
	if !strings.Contains(headers, "Subject: Test Subject") {
		t.Errorf("message headers = %q, want them to contain %q", headers, "Subject: Test Subject")
	}
	if !strings.Contains(headers, "text/html") {
		t.Errorf("message headers = %q, want an HTML content type", headers)
	}
	if !strings.Contains(gotBody, body) {
		t.Errorf("decoded body = %q, want it to contain %q", gotBody, body)
	}

	if err := c.Close(); err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
}

func TestSendWithAuth(t *testing.T) {
	srv := newTestSMTPServer(t, serverConfig{advertiseAuth: true})
	c, err := New(Options{
		Host:     "127.0.0.1",
		Port:     srv.port(),
		Username: testAuthUser,
		Password: testAuthPass,
		TLS:      TLSNone,
	})
	if err != nil {
		t.Fatalf("New() = %v, want nil error", err)
	}

	if err := c.Send(context.Background(), "sender@example.com",
		[]string{"rcpt@example.com"}, "authenticated send", "<p>hello</p>"); err != nil {
		t.Fatalf("Send() with valid credentials = %v, want nil", err)
	}

	attempts, user, pass := srv.authSnapshot()
	if attempts == 0 {
		t.Fatal("server never received an AUTH attempt")
	}
	if user != testAuthUser || pass != testAuthPass {
		t.Fatalf("server got credentials %q/%q, want %q/%q", user, pass, testAuthUser, testAuthPass)
	}
}

func TestSendOpportunisticTLSFallsBackToPlaintext(t *testing.T) {
	srv := newTestSMTPServer(t, serverConfig{})
	c := newTestClient(t, "127.0.0.1", srv.port(), TLSOpportunistic)

	err := c.Send(context.Background(), "sender@example.com",
		[]string{"rcpt@example.com"}, "opportunistic", "<p>hello</p>")

	if err != nil {
		t.Fatalf("Send() with TLSOpportunistic against plain server = %v, want nil", err)
	}
	if _, rcpts, _ := srv.snapshot(); len(rcpts) != 1 {
		t.Fatalf("server received %d RCPT commands, want 1", len(rcpts))
	}
}
