package webserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/brentahughes/service_tester/pkg/db"
)

type Server struct {
	db   db.DB
	port int
}

type errResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func NewServer(db db.DB, port int) *Server {
	return &Server{
		db:   db,
		port: port,
	}
}

func (s *Server) Start() {
	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/check", s.handleCheck)

	log.Printf("Listing on :%d", s.port)
	http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}

func (s *Server) Stop() {}

func (s *Server) writeErr(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)

	e := errResponse{
		Type:    "error",
		Message: err.Error(),
	}

	response, _ := json.Marshal(e)
	w.Write(response)
}
