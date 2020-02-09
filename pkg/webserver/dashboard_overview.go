package webserver

import (
	"html/template"
	"net/http"

	"github.com/brentahughes/service_tester/pkg/models"
)

type overview struct {
	CurrentHost models.CurrentHost
	View        string
	Hosts       []models.Host
}

func (s *Server) handleDashboardOverview(w http.ResponseWriter, req *http.Request) {
	currentHost, err := models.GetCurrentHost(s.db)
	if err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}

	hosts, err := models.GetAllHosts(s.db)
	if err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}

	t, err := template.ParseFiles("templates/layout.tmpl.html", "templates/overview.tmpl.html")
	if err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}

	if err := t.ExecuteTemplate(w, "layout", overview{CurrentHost: *currentHost, Hosts: hosts}); err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}
}
