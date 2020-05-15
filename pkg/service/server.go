package service

import (
	"encoding/json"

	"github.com/brentahughes/service_tester/pkg/models"
)

type Service struct {
	logger *models.Logger

	udpServer server
	tcpServer server
}

type server interface {
	run()
	close()
}

type serviceResponse struct {
	Status        string `json:"status"`
	ReceivedInput string `json:"receivedInput,omitempty"`
	Error         string `json:"error,omitempty"`
}

func NewService(logger *models.Logger, port int) *Service {
	return &Service{
		udpServer: &udpServer{
			logger: logger,
			port:   port,
		},
		tcpServer: &tcpServer{
			logger: logger,
			port:   port,
		},
	}
}

func (s *Service) Start() {
	go s.tcpServer.run()
	go s.udpServer.run()
}

func marshalResponse(status, input, err string) []byte {
	data, _ := json.Marshal(serviceResponse{
		Status:        status,
		ReceivedInput: input,
		Error:         err,
	})

	return append(data, []byte("\n")...)
}
