package webserver

import (
	"encoding/json"
	"net/http"

	"github.com/brentahughes/service_tester/pkg/models"
)

func (s *Server) handleApiHealth(w http.ResponseWriter, req *http.Request) {
	currentHost, err := models.GetCurrentHost(s.db)
	if err != nil {
		s.writeErr(w, 500, err)
		return
	}

	response, _ := json.Marshal(currentHost)
	w.Write(response)
}
