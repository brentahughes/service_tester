package webserver

import (
	"errors"
	"html/template"
	"net/http"

	"github.com/asdine/storm"
	"github.com/brentahughes/service_tester/pkg/models"
	"github.com/gin-gonic/gin"
)

type dashboardData struct {
	CurrentHost models.Host
	Hosts       []models.Host
	Logs        []models.Log
}

func (s *Server) dashboard(c *gin.Context) {
	currentHost, err := models.GetCurrentHost(s.db)
	if err != nil && err != storm.ErrNotFound {
		s.writeErr(c, http.StatusInternalServerError, err)
		return
	}

	hosts, err := models.GetHostsWithStatuses(s.db)
	if err != nil {
		if err != storm.ErrNotFound {
			s.writeErr(c, http.StatusInternalServerError, err)
			return
		}
	}

	t := template.New("layout").Funcs(template.FuncMap{
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	})

	t, err = t.ParseFiles("templates/layout.tmpl.html", "templates/overview.tmpl.html")
	if err != nil {
		s.writeErr(c, http.StatusInternalServerError, err)
		return
	}

	logs, err := models.GetLogs(s.db, 150)
	if err != nil && err != storm.ErrNotFound {
		s.writeErr(c, http.StatusInternalServerError, err)
		return
	}

	o := dashboardData{
		CurrentHost: *currentHost,
		Hosts:       hosts,
		Logs:        logs,
	}
	if err := t.ExecuteTemplate(c.Writer, "layout", o); err != nil {
		s.writeErr(c, http.StatusInternalServerError, err)
		return
	}
}

type detailsData struct {
	Hosts       []models.Host
	CurrentHost models.Host
	Host        models.Host
}

func (s *Server) hostDetails(c *gin.Context) {
	currentHost, err := models.GetCurrentHost(s.db)
	if err != nil {
		s.writeErr(c, http.StatusInternalServerError, err)
		return
	}

	hosts, err := models.GetHosts(s.db)
	if err != nil {
		s.writeErr(c, http.StatusInternalServerError, err)
		return
	}

	host, err := models.GetHostByID(s.db, c.Param("id"))
	if err != nil {
		s.writeErr(c, http.StatusInternalServerError, err)
		return
	}

	t, err := template.ParseFiles("templates/layout.tmpl.html", "templates/details.tmpl.html")
	if err != nil {
		s.writeErr(c, http.StatusInternalServerError, err)
		return
	}

	d := detailsData{
		CurrentHost: *currentHost,
		Hosts:       hosts,
		Host:        *host,
	}

	if err := t.ExecuteTemplate(c.Writer, "layout", d); err != nil {
		s.writeErr(c, http.StatusInternalServerError, err)
		return
	}
}
