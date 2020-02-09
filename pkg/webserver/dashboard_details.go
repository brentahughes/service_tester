package webserver

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/brentahughes/service_tester/pkg/models"
	"github.com/gorilla/mux"
)

type details struct {
	CurrentHost models.CurrentHost
	View        string
	Host        models.Host
}

func (s *Server) handleDashboardHostDetails(w http.ResponseWriter, req *http.Request) {
	currentHost, err := models.GetCurrentHost(s.db)
	if err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}

	vars := mux.Vars(req)
	id := vars["id"]
	idInt, err := strconv.Atoi(id)
	if err != nil {
		s.writeErr(w, http.StatusBadRequest, err)
		return
	}

	host, err := models.GetHostByID(s.db, idInt)
	if err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}

	t, err := template.ParseFiles("templates/layout.tmpl.html", "templates/details.tmpl.html")
	if err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}

	if err := t.ExecuteTemplate(w, "layout", details{View: "details", CurrentHost: *currentHost, Host: *host}); err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}
}
