package handler

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNotFoundPage(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := get(t, srv, "/no/such/path", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "command not found: 404") {
		t.Error("404 page missing terminal gag")
	}
	if !strings.Contains(body, "/no/such/path") {
		t.Error("404 page missing requested path")
	}
}

func TestRecoverPanics(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") })
	on500 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "kernel panic")
	}
	h := recoverPanics(log, on500, next)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "kernel panic") {
		t.Error("500 body not rendered")
	}
}
