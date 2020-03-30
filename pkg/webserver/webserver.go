package webserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/asdine/storm/v3"
	"github.com/brentahughes/service_tester/pkg/config"
	"github.com/brentahughes/service_tester/pkg/models"
	"github.com/gorilla/mux"
)

type Server struct {
	config config.Config
	db     *storm.DB
	logger *models.Logger
	port   int
}

type errResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func NewServer(config config.Config, db *storm.DB, logger *models.Logger, port int) *Server {
	return &Server{
		db:     db,
		port:   port,
		config: config,
		logger: logger,
	}
}

func (s *Server) Start() {
	// Root resources and redirect
	rootRouter := mux.NewRouter()
	rootRouter.PathPrefix("/resources/").Handler(http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))
	rootRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard", http.StatusPermanentRedirect)
	})

	// Dashboard pages
	dashboardRouter := rootRouter.PathPrefix("/dashboard").Subrouter()
	dashboardRouter.HandleFunc("", s.handleDashboardOverview).Methods("GET")
	dashboardRouter.HandleFunc("/", s.handleDashboardOverview).Methods("GET")
	dashboardRouter.HandleFunc("/host/{id:[0-9]+}/details", s.handleDashboardHostDetails).Name("host").Methods("GET")

	// Api endpoints
	apiRouter := rootRouter.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/health", s.handleApiHealth).Methods("GET")
	apiRouter.HandleFunc("/host/{id:[0-9]+", s.handleApiHostDelete).Methods("DELETE")

	http.Handle("/", rootRouter)

	s.logger.Infof("Listing on :%d", s.port)
	http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}

func (s *Server) Stop() {
	s.logger.Infof("Stopping webserver")
}

func (s *Server) writeErr(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)

	e := errResponse{
		Type:    "error",
		Message: err.Error(),
	}

	response, _ := json.Marshal(e)
	w.Write(response)
}
