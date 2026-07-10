package handler

import (
	"strings"
	"testing"

	"portfolio/internal/store"
)

func TestComposeReplyToAndHeaderInjection(t *testing.T) {
	m := &mailer{from: "contact@example.com", to: "me@example.com"}
	msg := string(m.compose(store.Inquiry{
		Name:    "Ada\r\nBcc: evil@example.com",
		Email:   "ada@example.com",
		Kind:    "contract",
		Message: "line1\r\n.\r\nline2",
	}))

	if !strings.Contains(msg, "Reply-To: ada@example.com\r\n") {
		t.Error("Reply-To not set to visitor email")
	}
	// CRLF smuggled through the name must not become its own header line.
	// Only the header block (above the blank-line separator) can be injected;
	// anything below it is body text a mail server never parses as a header.
	headers, _, _ := strings.Cut(msg, "\r\n\r\n")
	if strings.Contains(headers, "\r\nBcc:") {
		t.Errorf("header injection not neutralized:\n%s", headers)
	}
	if !strings.Contains(msg, "line1") || !strings.Contains(msg, "line2") {
		t.Error("message body missing")
	}
}

func TestMailerFromEnv(t *testing.T) {
	for _, k := range []string{"SMTP_HOST", "SMTP_USER", "SMTP_PASS", "SMTP_FROM", "SMTP_TO", "SMTP_PORT"} {
		t.Setenv(k, "")
	}
	if mailerFromEnv() != nil {
		t.Error("expected nil mailer when SMTP_* unset")
	}
	t.Setenv("SMTP_HOST", "smtp.resend.com")
	t.Setenv("SMTP_USER", "resend")
	t.Setenv("SMTP_PASS", "re_x")
	t.Setenv("SMTP_FROM", "contact@example.com")
	t.Setenv("SMTP_TO", "me@example.com")
	m := mailerFromEnv()
	if m == nil || m.addr != "smtp.resend.com:587" {
		t.Fatalf("mailer = %+v, want addr smtp.resend.com:587", m)
	}
}
