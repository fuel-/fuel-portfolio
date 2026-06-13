package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func postForm(t *testing.T, h http.Handler, path string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("POST", path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestContactValidSubmission(t *testing.T) {
	srv, st := newTestServer(t)
	rec := postForm(t, srv, "/contact", url.Values{
		"name": {"Ada Lovelace"}, "email": {"ada@example.com"},
		"company": {"Analytical Engines"}, "kind": {"contract"}, "message": {"Need a compiler."},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "message queued") {
		t.Error("success fragment missing confirmation")
	}
	got, err := st.ListInquiries()
	if err != nil || len(got) != 1 {
		t.Fatalf("inquiries = %d (%v), want 1", len(got), err)
	}
	if got[0].Name != "Ada Lovelace" || got[0].Kind != "contract" {
		t.Errorf("saved inquiry mismatch: %+v", got[0])
	}
}

func TestContactValidationErrors(t *testing.T) {
	srv, st := newTestServer(t)
	rec := postForm(t, srv, "/contact", url.Values{"kind": {"hiring"}})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (form re-render)", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"tell me who you are", "an email is required", "say something"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing validation message %q", want)
		}
	}
	if got, _ := st.ListInquiries(); len(got) != 0 {
		t.Errorf("invalid submission was saved: %d rows", len(got))
	}
}

func TestContactPreservesValuesOnError(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := postForm(t, srv, "/contact", url.Values{
		"name": {"Ada Lovelace"}, "email": {"not-an-email"}, "kind": {"other"}, "message": {"hi"},
	})
	body := rec.Body.String()
	if !strings.Contains(body, `value="Ada Lovelace"`) {
		t.Error("name value not preserved on re-render")
	}
	if !strings.Contains(body, "valid email") {
		t.Error("missing email format error")
	}
}

func TestValidate(t *testing.T) {
	valid := contactForm{Name: "n", Email: "e@x.com", Kind: "hiring", Message: "m"}
	tests := []struct {
		name   string
		mutate func(*contactForm)
		field  string
	}{
		{"missing name", func(f *contactForm) { f.Name = "" }, "name"},
		{"missing email", func(f *contactForm) { f.Email = "" }, "email"},
		{"bad email", func(f *contactForm) { f.Email = "nope" }, "email"},
		{"bad kind", func(f *contactForm) { f.Kind = "spam" }, "kind"},
		{"missing message", func(f *contactForm) { f.Message = "" }, "message"},
		{"huge message", func(f *contactForm) { f.Message = strings.Repeat("x", 5001) }, "message"},
	}
	if errs := validate(valid); len(errs) != 0 {
		t.Errorf("valid form has errors: %v", errs)
	}
	for _, tc := range tests {
		f := valid
		tc.mutate(&f)
		if errs := validate(f); errs[tc.field] == "" {
			t.Errorf("%s: no error on field %q (got %v)", tc.name, tc.field, errs)
		}
	}
}
