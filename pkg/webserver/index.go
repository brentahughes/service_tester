package webserver

import (
	"encoding/json"
	"github.com/brentahughes/service_tester/pkg/db"
	"net/http"
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

	response, _ := json.Marshal(r)
	w.Write(response)
}
