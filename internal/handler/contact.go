package handler

import (
	"net/http"
	"net/mail"
	"strings"

	"portfolio/internal/content"
	"portfolio/internal/store"
)

// contactForm carries submitted values and per-field validation errors
// back into the form partial.
type contactForm struct {
	Name    string
	Email   string
	Company string
	Kind    string
	Message string
	Errors  map[string]string
}

func (s *Server) contact(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10) // 64 KB; message cap is 5000 chars
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	f := contactForm{
		Name:    strings.TrimSpace(r.PostFormValue("name")),
		Email:   strings.TrimSpace(r.PostFormValue("email")),
		Company: strings.TrimSpace(r.PostFormValue("company")),
		Kind:    r.PostFormValue("kind"),
		Message: strings.TrimSpace(r.PostFormValue("message")),
	}
	f.Errors = validate(f)
	if len(f.Errors) > 0 {
		// 200, not 422: this is a UX state and htmx should swap it in.
		s.renderPartial(w, "contact_form", f)
		return
	}
	_, err := s.st.SaveInquiry(store.Inquiry{
		Name: f.Name, Email: f.Email, Company: f.Company, Kind: f.Kind, Message: f.Message,
	})
	if err != nil {
		s.log.Error("save inquiry", "err", err)
		s.renderPartial(w, "contact_error", content.Me)
		return
	}
	s.renderPartial(w, "contact_success", f)
}

func validate(f contactForm) map[string]string {
	errs := map[string]string{}
	if f.Name == "" {
		errs["name"] = "tell me who you are"
	}
	switch {
	case f.Email == "":
		errs["email"] = "an email is required — it's how I reply"
	default:
		if _, err := mail.ParseAddress(f.Email); err != nil {
			errs["email"] = "that doesn't look like a valid email"
		}
	}
	switch f.Kind {
	case "hiring", "contract", "other":
	default:
		errs["kind"] = "pick a topic"
	}
	switch {
	case f.Message == "":
		errs["message"] = "say something — even just hi"
	case len(f.Message) > 5000:
		errs["message"] = "keep it under 5000 characters"
	}
	return errs
}
