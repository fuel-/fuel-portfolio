package handler

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"portfolio"
	"portfolio/internal/store"
	"portfolio/internal/terminal"
)

// newTestServer builds a Server against the real embedded templates and
// static assets, with a throwaway SQLite store.
func newTestServer(t *testing.T) (*Server, *store.Store) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv, err := New(portfolio.Templates, portfolio.Static, st, terminal.New(), log)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return srv, st
}

func get(t *testing.T, h http.Handler, path string, hdr map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("GET", path, nil)
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHomePage(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := get(t, srv, "/", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"Jasen Nicely", `id="projects"`, `id="contact"`, `id="resume"`, "Property Tax Pipeline", "Announcement System", "Redline"} {
		if !strings.Contains(body, want) {
			t.Errorf("home page missing %q", want)
		}
	}
}

func TestStaticAssetServed(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := get(t, srv, "/static/js/htmx.min.js", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "max-age") {
		t.Errorf("Cache-Control = %q, want max-age", cc)
	}
}

func TestResumePDFServed(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := get(t, srv, "/static/resume/jasen-nicely-resume.pdf", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}
