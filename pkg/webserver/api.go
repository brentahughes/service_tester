package webserver

import (
	"net/http"

	"github.com/brentahughes/service_tester/pkg/models"
	"github.com/gin-gonic/gin"
)

func (s *Server) setupAPIEndpoints() {
	api := s.router.Group("/api")
	api.GET("/health", s.getHealth)
	api.GET("/hosts", s.getHosts)
	api.GET("/hosts/:id", s.getHost)
}

func (s *Server) getHealth(c *gin.Context) {
	currentHost, err := models.GetCurrentHost(s.db)
	if err != nil {
		s.writeErr(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, currentHost)
}

func (s *Server) getHosts(c *gin.Context) {
	hosts, err := models.GetHostsWithStatuses(s.db)
	if err != nil {
		s.writeErr(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, hosts)
}

func (s *Server) getHost(c *gin.Context) {
	host, err := models.GetHostByID(s.db, c.Param("id"))
	if err != nil {
		s.writeErr(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, host)
}
