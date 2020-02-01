package webserver

import (
	"html/template"
	"net/http"
	"sort"
)

func (s *Server) handleDetails(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get("id")

	host, err := s.db.GetHost(id)
	if err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}

	sort.Slice(host.Checks, func(a, b int) bool {
		return host.Checks[a].CheckTimeMS < host.Checks[b].CheckTimeMS
	})

	t, err := template.ParseFiles("templates/details.tmpl")
	if err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}

	if err := t.Execute(w, host); err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}
}
