package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"portfolio/internal/store"
)

// mailer sends one notification email per inquiry via Resend's HTTPS API.
// It is nil when the required env vars are unset — the contact handler then
// skips notifying, so local dev and tests need no mail config.
//
// We use the HTTP API (port 443) rather than SMTP because VPS hosts commonly
// block outbound SMTP ports (Linode does), which silently blackholes mail —
// the TCP connection completes but no data flows. 443 is always open.
type mailer struct {
	apiKey string // Resend API key (re_...); read from SMTP_PASS for env back-compat
	from   string // must be on a Resend-verified domain
	to     string // inbox that receives the notification
	http   *http.Client
}

// mailerFromEnv builds a mailer from env vars, or returns nil if any is missing.
// Var names keep the SMTP_ prefix so an existing deploy's .env keeps working:
// SMTP_PASS holds the Resend API key; SMTP_HOST/PORT/USER are no longer used.
func mailerFromEnv() *mailer {
	key := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")
	to := os.Getenv("SMTP_TO")
	if key == "" || from == "" || to == "" {
		return nil
	}
	return &mailer{
		apiKey: key,
		from:   from,
		to:     to,
		http:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (m *mailer) send(q store.Inquiry) error {
	req, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(m.payload(q)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := m.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<10))
		return fmt.Errorf("resend %s: %s", resp.Status, body)
	}
	return nil
}

// payload builds the Resend API request body. Reply-To is the visitor so a
// reply from your inbox reaches them directly. JSON encoding escapes every
// field, so free-text input can't inject headers or break the request.
func (m *mailer) payload(q store.Inquiry) []byte {
	b, _ := json.Marshal(map[string]any{
		"from":     m.from,
		"to":       []string{m.to},
		"reply_to": q.Email,
		"subject":  fmt.Sprintf("Portfolio inquiry (%s) from %s", q.Kind, q.Name),
		"text": fmt.Sprintf("Name:    %s\nEmail:   %s\nCompany: %s\nKind:    %s\n\n%s\n",
			q.Name, q.Email, q.Company, q.Kind, q.Message),
	})
	return b
}
