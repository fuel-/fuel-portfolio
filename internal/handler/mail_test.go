package handler

import (
	"encoding/json"
	"strings"
	"testing"

	"portfolio/internal/store"
)

func TestPayload(t *testing.T) {
	m := &mailer{from: "contact@contact.gideon.gg", to: "me@example.com"}
	var got map[string]any
	if err := json.Unmarshal(m.payload(store.Inquiry{
		Name: "Ada", Email: "ada@example.com", Kind: "contract", Message: "hi",
	}), &got); err != nil {
		t.Fatalf("payload is not valid json: %v", err)
	}
	if got["reply_to"] != "ada@example.com" {
		t.Errorf("reply_to = %v, want visitor email", got["reply_to"])
	}
	if got["from"] != "contact@contact.gideon.gg" {
		t.Errorf("from = %v", got["from"])
	}
	to, _ := got["to"].([]any)
	if len(to) != 1 || to[0] != "me@example.com" {
		t.Errorf("to = %v, want [me@example.com]", got["to"])
	}
	if s, _ := got["subject"].(string); !strings.Contains(s, "Ada") {
		t.Errorf("subject = %q, want visitor name", got["subject"])
	}
}

func TestMailerFromEnv(t *testing.T) {
	t.Setenv("SMTP_PASS", "")
	t.Setenv("SMTP_FROM", "")
	t.Setenv("SMTP_TO", "")
	if mailerFromEnv() != nil {
		t.Error("expected nil mailer when creds unset")
	}
	t.Setenv("SMTP_PASS", "re_x")
	t.Setenv("SMTP_FROM", "contact@contact.gideon.gg")
	t.Setenv("SMTP_TO", "me@example.com")
	m := mailerFromEnv()
	if m == nil || m.apiKey != "re_x" || m.to != "me@example.com" {
		t.Fatalf("mailer = %+v, want configured", m)
	}
}
