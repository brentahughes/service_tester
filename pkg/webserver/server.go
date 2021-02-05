package webserver

import (
	"fmt"
	"log"

	"github.com/brentahughes/service_tester/pkg/config"
	"github.com/dgraph-io/badger"
	"github.com/gin-gonic/gin"
)

type Server struct {
	config config.Config
	db     *badger.DB
	port   int
	router *gin.Engine
}

type errResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func NewServer(config config.Config, db *badger.DB, port int) *Server {
	return &Server{
		db:     db,
		port:   port,
		config: config,
	}
}

func (s *Server) Start() error {
	gin.SetMode(gin.ReleaseMode)

	s.router = gin.Default()
	s.router.Use(gin.Recovery(), gin.Logger())
	s.setupInterfaceEndpoints()
	s.setupAPIEndpoints()

	log.Printf("web interface listening on :%d", s.port)
	return s.router.Run(fmt.Sprintf(":%d", s.port))
}

func (s *Server) Stop() {
	log.Printf("Stopping webserver")
}

func (s *Server) writeErr(c *gin.Context, code int, err error) {
	c.JSON(code, &errResponse{
		Type:    "error",
		Message: err.Error(),
	})
}
