package webserver

import (
	"fmt"
	"net/http"

	"github.com/brentahughes/service_tester/pkg/config"
	"github.com/brentahughes/service_tester/pkg/models"
	"github.com/dgraph-io/badger"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

type Server struct {
	config config.Config
	db     *badger.DB
	logger *models.Logger
	port   int
}

type errResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func NewServer(config config.Config, db *badger.DB, logger *models.Logger, port int) *Server {
	return &Server{
		db:     db,
		port:   port,
		config: config,
		logger: logger,
	}
}

func (s *Server) Start() error {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(gin.Recovery(), gin.Logger())

	// React Frontend
	r.Use(static.Serve("/", static.LocalFile("./frontend/build", true)))
	// This is a hack to make the react frontend is used for any route that wasn't defined elsewhere
	r.NoRoute(func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "./frontend/build/index.html")
	})

	// API Endpoints
	api := r.Group("/api")
	api.GET("/health", s.getHealth)
	api.GET("/hosts", s.getHosts)
	api.GET("/hosts/:id", s.getHost)

	s.logger.Infof("web interface listening on :%d", s.port)
	return r.Run(fmt.Sprintf(":%d", s.port))
}

func (s *Server) Stop() {
	s.logger.Infof("Stopping webserver")
}

func (s *Server) writeErr(c *gin.Context, code int, err error) {
	c.JSON(code, &errResponse{
		Type:    "error",
		Message: err.Error(),
	})
}
