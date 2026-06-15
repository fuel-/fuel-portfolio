// Package handler wires HTTP routes to templates, the content package,
// the inquiry store, and the terminal registry.
package handler

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"portfolio/internal/content"
	"portfolio/internal/store"
	"portfolio/internal/terminal"
)

type Server struct {
	pages    map[string]*template.Template
	partials *template.Template
	st       *store.Store
	reg      *terminal.Registry
	log      *slog.Logger
	h        http.Handler
}

// pageData is the single context type passed to full-page templates.
type pageData struct {
	Title    string
	Path     string
	Year     int
	Profile  content.Profile
	Stats    []content.Stat
	Projects []content.Project
	Resume   []content.ResumeEntry
	Skills   []content.Skill
	School   content.Education
	Project  *content.Project // set on the standalone project page
	Form     contactForm
}

func New(templatesFS, staticFS fs.FS, st *store.Store, reg *terminal.Registry, log *slog.Logger) (*Server, error) {
	s := &Server{pages: map[string]*template.Template{}, st: st, reg: reg, log: log}

	for _, page := range []string{"home.html", "project_page.html", "404.html", "500.html"} {
		t, err := template.ParseFS(templatesFS,
			"templates/layout.html", "templates/partials/*.html", "templates/pages/"+page)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", page, err)
		}
		s.pages[page] = t
	}
	partials, err := template.ParseFS(templatesFS, "templates/partials/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse partials: %w", err)
	}
	s.partials = partials

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.home)
	mux.HandleFunc("GET /projects/{slug}", s.project)
	mux.HandleFunc("GET /projects/{slug}/card", s.projectCard)
	mux.HandleFunc("POST /contact", s.contact)
	mux.HandleFunc("POST /terminal", s.terminal)
	mux.Handle("GET /static/", staticHandler(staticFS))
	mux.HandleFunc("/", s.notFound)

	s.h = recoverPanics(log, s.serverError, mux)
	return s, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.h.ServeHTTP(w, r) }

func (s *Server) data() pageData {
	return pageData{
		Year:     time.Now().Year(),
		Profile:  content.Me,
		Stats:    content.Stats,
		Projects: content.Projects,
		Resume:   content.Resume,
		Skills:   content.Skills,
		School:   content.School,
	}
}

// renderPage buffers template execution so a render error can still
// produce a clean 500 instead of a half-written page.
func (s *Server) renderPage(w http.ResponseWriter, status int, page string, d pageData) {
	t, ok := s.pages[page]
	if !ok {
		s.log.Error("unknown page template", "page", page)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout", d); err != nil {
		s.log.Error("render page", "page", page, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (s *Server) renderPartial(w http.ResponseWriter, name string, d any) {
	var buf bytes.Buffer
	if err := s.partials.ExecuteTemplate(&buf, name, d); err != nil {
		s.log.Error("render partial", "partial", name, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

func staticHandler(fsys fs.FS) http.Handler {
	fileServer := http.FileServerFS(fsys)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=86400")
		fileServer.ServeHTTP(w, r)
	})
}
