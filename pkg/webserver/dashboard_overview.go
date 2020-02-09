package webserver

import (
	"html/template"
	"net/http"

	"github.com/brentahughes/service_tester/pkg/models"
)

type overview struct {
	PublicHosts []models.Host
	CurrentHost models.CurrentHost
	Hosts       []models.Host
}

func (s *Server) handleDashboardOverview(w http.ResponseWriter, req *http.Request) {
	currentHost, err := models.GetCurrentHost(s.db)
	if err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}

	publicHosts, err := models.GetHostsWithPublicIPs(s.db)
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

	o := overview{
		PublicHosts: publicHosts,
		CurrentHost: *currentHost,
		Hosts:       hosts,
	}
	if err := t.ExecuteTemplate(w, "layout", o); err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}
}
