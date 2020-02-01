package webserver

import (
	"html/template"
	"net/http"
)

func (s *Server) handleIndex(w http.ResponseWriter, req *http.Request) {
	hosts, err := s.db.GetAllHosts()
	if err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}

	t, err := template.ParseFiles("templates/index.tmpl")
	if err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}

	if err := t.Execute(w, hosts); err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}
}
