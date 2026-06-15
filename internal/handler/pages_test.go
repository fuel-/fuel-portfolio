package handler

import (
	"net/http"
	"strings"
	"testing"
)

func TestProjectFragmentForHTMX(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := get(t, srv, "/projects/redline", map[string]string{"HX-Request": "true"})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `id="proj-redline"`) {
		t.Error("fragment missing project article")
	}
	if strings.Contains(body, "<!doctype") {
		t.Error("fragment contains full page doctype")
	}
}

func TestProjectStandalonePage(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := get(t, srv, "/projects/property-tax-pipeline", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "<!doctype") {
		t.Error("standalone page missing doctype")
	}
	if !strings.Contains(body, "Property Tax Pipeline") {
		t.Error("standalone page missing project name")
	}
}

func TestProjectUnknownSlugIs404(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := get(t, srv, "/projects/nope", nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestProjectCardFragment(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := get(t, srv, "/projects/redline/card", map[string]string{"HX-Request": "true"})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `data-project-link="redline"`) {
		t.Error("card fragment missing expand button")
	}
	if strings.Contains(body, "<!doctype") {
		t.Error("card fragment contains full page doctype")
	}
}
