package handler

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"

	"portfolio/internal/store"
)

// mailer sends one notification email per inquiry over SMTP.
// It is nil when the SMTP_* env vars are unset — the contact handler then
// just skips notifying, so local dev and tests need no mail server.
type mailer struct {
	addr string // host:port
	auth smtp.Auth
	from string // must be on an SMTP-verified domain
	to   string // inbox that receives the notification
}

// mailerFromEnv builds a mailer from SMTP_* env vars, or returns nil if any
// required var is missing. SMTP_PORT defaults to 587 (STARTTLS submission).
func mailerFromEnv() *mailer {
	host := os.Getenv("SMTP_HOST")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")
	to := os.Getenv("SMTP_TO")
	if host == "" || user == "" || pass == "" || from == "" || to == "" {
		return nil
	}
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	return &mailer{
		addr: host + ":" + port,
		auth: smtp.PlainAuth("", user, pass, host),
		from: from,
		to:   to,
	}
}

func (m *mailer) send(q store.Inquiry) error {
	return smtp.SendMail(m.addr, m.auth, m.from, []string{m.to}, m.compose(q))
}

// compose builds the RFC 5322 message. Reply-To is the visitor so a reply from
// your inbox reaches them directly. Header values are CRLF-stripped to prevent
// header injection from free-text fields; the message body is safe (net/smtp's
// DataWriter dot-stuffs it, and it sits below the header/body separator).
func (m *mailer) compose(q store.Inquiry) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", m.from)
	fmt.Fprintf(&b, "To: %s\r\n", m.to)
	fmt.Fprintf(&b, "Reply-To: %s\r\n", headerSafe(q.Email))
	fmt.Fprintf(&b, "Subject: Portfolio inquiry (%s) from %s\r\n", headerSafe(q.Kind), headerSafe(q.Name))
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	b.WriteString("\r\n")
	fmt.Fprintf(&b, "Name:    %s\r\n", q.Name)
	fmt.Fprintf(&b, "Email:   %s\r\n", q.Email)
	fmt.Fprintf(&b, "Company: %s\r\n", q.Company)
	fmt.Fprintf(&b, "Kind:    %s\r\n", q.Kind)
	b.WriteString("\r\n")
	b.WriteString(q.Message)
	b.WriteString("\r\n")
	return []byte(b.String())
}

func headerSafe(s string) string {
	return strings.NewReplacer("\r", " ", "\n", " ").Replace(s)
}
