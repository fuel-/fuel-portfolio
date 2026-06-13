package handler

import (
	"fmt"
	"html"
	"net/http"
)

func (s *Server) terminal(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	cmd := r.PostFormValue("cmd")
	if len(cmd) > 200 {
		cmd = cmd[:200]
	}
	res := s.reg.Execute(cmd)
	if res.Action != "" {
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"term-action": %q}`, res.Action))
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<div class="term-line term-echo"><span class="term-prompt">$</span> %s</div>%s`,
		html.EscapeString(cmd), res.HTML)
}
