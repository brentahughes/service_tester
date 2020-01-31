package webserver

import (
	"html/template"
	"net/http"

	"github.com/brentahughes/service_tester/pkg/db"
)

type indexResponse struct {
	Checks []db.HostCheck
}

func (s *Server) handleIndex(w http.ResponseWriter, req *http.Request) {
	checks, err := s.db.GetLastForAllHosts()
	if err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}

	r := indexResponse{
		Checks: checks,
	}

	t, err := template.ParseFiles("templates/index.tmpl")
	if err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}

	// b, _ := json.MarshalIndent(r, "", "  ")
	// w.Write(b)

	if err := t.Execute(w, r); err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}
}
