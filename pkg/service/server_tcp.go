package service

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/brentahughes/service_tester/pkg/models"
)

type tcpServer struct {
	logger *models.Logger

	port   int
	server net.Listener
}

func (s *tcpServer) run() {
	var err error
	s.server, err = net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		s.logger.Errorf("could not listen on tcp: %v", err)
		return
	}
	defer s.close()

	s.logger.Infof("tcp server listening on :%d", s.port)

	for {
		conn, err := s.server.Accept()
		if err != nil {
			s.logger.Errorf("error accepting tcp connection: %v", err)
			break
		}

		if conn == nil {
			s.logger.Errorf("could not create tcp connection: %v", err)
			break
		}

		s.handleConnections()
	}
}

func (s *tcpServer) close() {
	s.server.Close()
}

func (s *tcpServer) handleConnections() {
	for {
		conn, err := s.server.Accept()
		if err != nil || conn == nil {
			s.logger.Errorf("could not accept new tcp connection: %v", err)
			break
		}

		go s.handleConnection(conn)
	}
}

func (s *tcpServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	for {
		req, err := rw.ReadString('\n')
		if err != nil {
			if strings.HasSuffix(err.Error(), "connection reset by peer") {
				return
			}

			s.logger.Errorf("error reading tcp input: %v", err)
			rw.Write(marshalResponse("error", "", "failed to read input: "+err.Error()))
			rw.Flush()
			return
		}
		req = strings.TrimSuffix(req, "\n")

		rw.Write(marshalResponse("success", req, ""))
		rw.Flush()
	}
}
