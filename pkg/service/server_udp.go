package service

import (
	"bytes"
	"fmt"
	"log"
	"net"
)

type udpServer struct {
	port   int
	server *net.UDPConn
}

func (s *udpServer) run() {
	laddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Printf("could not resolved UDP addr: %v", err)
		return
	}

	s.server, err = net.ListenUDP("udp", laddr)
	if err != nil {
		log.Printf("error listing for UDP: %v", err)
		return
	}
	defer s.close()

	log.Printf("udp server listening on :%d", s.port)

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
			log.Printf("error reading from udp: %v", err)
			break
		}
		if conn == nil {
			continue
		}

		s.handleConnection(conn, buf[:n])
	}
}

func (s *udpServer) handleConnection(addr *net.UDPAddr, cmd []byte) {
	cmd = bytes.TrimSuffix(cmd, []byte("\n"))
	if string(cmd) == "ping" {
		s.server.WriteToUDP([]byte("pong\n"), addr)
	}
}
