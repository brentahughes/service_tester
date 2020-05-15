package service

import (
	"bytes"
	"fmt"
	"net"

	"github.com/brentahughes/service_tester/pkg/models"
)

type udpServer struct {
	logger *models.Logger
	port int
	server *net.UDPConn
}

func (s *udpServer) run(){
	laddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		s.logger.Errorf("could not resolved UDP addr: %v", err)
		return
	}

	s.server, err = net.ListenUDP("udp", laddr)
	if err != nil {
		s.logger.Errorf("error listing for UDP: %v", err)
		return
	}
	defer s.close()

	s.logger.Infof("udp server listening on :%d", s.port)

	s.handleConnections()
}

func (s *udpServer) close() {
	s.server.Close()
}

func (s *udpServer) handleConnections() {
	for {
		buf := make([]byte, 2048)
		n, conn, err := s.server.ReadFromUDP(buf)
		if err != nil {
			s.logger.Errorf("error reading from udp: %v", err)
			break
		}
		if conn == nil {
			continue
		}

		go s.handleConnection(conn, buf[:n])
	}
}

func (s *udpServer) handleConnection(addr *net.UDPAddr, cmd []byte) {
	cmd = bytes.TrimSuffix(cmd, []byte("\n"))
	s.server.WriteToUDP(marshalResponse("success", string(cmd), ""), addr)
}