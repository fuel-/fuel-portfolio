package handler

import (
	"net/http"

	"portfolio/internal/content"
)

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	s.renderPage(w, http.StatusOK, "home.html", s.data())
}

func (s *Server) project(w http.ResponseWriter, r *http.Request) {
	p, ok := content.ProjectBySlug(r.PathValue("slug"))
	if !ok {
		s.notFound(w, r)
		return
	}
	if r.Header.Get("HX-Request") == "true" {
		s.renderPartial(w, "project_detail", p)
		return
	}
	d := s.data()
	d.Title = p.Name
	d.Project = &p
	s.renderPage(w, http.StatusOK, "project_page.html", d)
}

func (s *Server) projectCard(w http.ResponseWriter, r *http.Request) {
	p, ok := content.ProjectBySlug(r.PathValue("slug"))
	if !ok {
		s.notFound(w, r)
		return
	}
	s.renderPartial(w, "project_card", p)
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	d := s.data()
	d.Title = "404"
	d.Path = r.URL.Path
	s.renderPage(w, http.StatusNotFound, "404.html", d)
}

func (s *Server) serverError(w http.ResponseWriter, r *http.Request) {
	d := s.data()
	d.Title = "500"
	s.renderPage(w, http.StatusInternalServerError, "500.html", d)
}
