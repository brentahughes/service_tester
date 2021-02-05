package service

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type tcpServer struct {
	port   int
	server net.Listener
}

func (s *tcpServer) run() {
	var err error
	s.server, err = net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Printf("could not listen on tcp: %v", err)
		return
	}
	defer s.close()

	log.Printf("tcp server listening on :%d", s.port)

	for {
		conn, err := s.server.Accept()
		if err != nil {
			log.Printf("error accepting tcp connection: %v", err)
			break
		}

		if conn == nil {
			log.Printf("could not create tcp connection: %v", err)
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
			log.Printf("could not accept new tcp connection: %v", err)
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

			log.Printf("error reading tcp input: %v", err)
			rw.Write(marshalResponse("error", "", "failed to read input: "+err.Error()))
			rw.Flush()
			return
		}
		req = strings.TrimSuffix(req, "\n")

		rw.Write(marshalResponse("success", req, ""))
		rw.Flush()
	}
}
