package webserver

import (
	"net/http"
	"strconv"

	"github.com/brentahughes/service_tester/pkg/models"
	"github.com/gorilla/mux"
)

func (s *Server) handleApiHostDelete(w http.ResponseWriter, req *http.Request) {
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

	if err := host.Delete(s.db); err != nil {
		s.writeErr(w, http.StatusInternalServerError, err)
		return
	}

	return
}
