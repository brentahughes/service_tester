package webserver

import (
	"html/template"
	"net/http"

	"github.com/asdine/storm/v3"
	"github.com/brentahughes/service_tester/pkg/models"
)

type overview struct {
	CurrentHost models.Host
	Hosts       []models.Host
	Logs        []models.Log
}

func (s *Server) handleDashboardOverview(w http.ResponseWriter, req *http.Request) {
	currentHost, err := models.GetCurrentHost(s.db)
	if err != nil && err != storm.ErrNotFound {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}

	hosts, err := models.GetRecentHostsWithChecks(s.db)
	if err != nil {
		if err != storm.ErrNotFound {
			s.writeErr(w, http.StatusInternalServerError, err)
			return
		}
	}

	t, err := template.ParseFiles("templates/layout.tmpl.html", "templates/overview.tmpl.html")
	if err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}

	logs, err := models.GetLogs(s.db, 150)
	if err != nil && err != storm.ErrNotFound {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}

	o := overview{
		CurrentHost: *currentHost,
		Hosts:       hosts,
		Logs:        logs,
	}
	if err := t.ExecuteTemplate(w, "layout", o); err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}
}
