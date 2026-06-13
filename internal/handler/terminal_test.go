package handler

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestTerminalWhoami(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := postForm(t, srv, "/terminal", url.Values{"cmd": {"whoami"}})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "term-echo") || !strings.Contains(body, "whoami") {
		t.Error("response missing echo line")
	}
	if !strings.Contains(body, "Jasen Nicely") {
		t.Error("response missing whoami output")
	}
}

func TestTerminalActionHeader(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := postForm(t, srv, "/terminal", url.Values{"cmd": {"clear"}})
	trigger := rec.Header().Get("HX-Trigger")
	if !strings.Contains(trigger, "term-action") || !strings.Contains(trigger, "clear") {
		t.Errorf("HX-Trigger = %q, want term-action clear", trigger)
	}
}

func TestTerminalEchoEscapesInput(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := postForm(t, srv, "/terminal", url.Values{"cmd": {"<img src=x onerror=alert(1)>"}})
	if strings.Contains(rec.Body.String(), "<img") {
		t.Error("unescaped input echoed back")
	}
}

func TestTerminalUnknownCommand(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := postForm(t, srv, "/terminal", url.Values{"cmd": {"frobnicate"}})
	if !strings.Contains(rec.Body.String(), "command not found") {
		t.Error("missing command-not-found output")
	}
}
