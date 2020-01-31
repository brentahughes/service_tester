package webserver

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
)

type checkResponse struct {
	Status    string            `json:"status"`
	Addresses map[string]string `json:"addresses"`
}

func (s *Server) handleCheck(w http.ResponseWriter, req *http.Request) {
	nets, err := net.Interfaces()
	if err != nil {
		s.writeErr(w, 500, err)
		return
	}

	addresses := make(map[string]string, 0)
	for _, hostInterface := range nets {
		ips, err := hostInterface.Addrs()
		if err != nil {
			s.writeErr(w, 500, err)
			return
		}
		for _, ip := range ips {
			if strings.HasPrefix(ip.String(), "127.") {
				continue
			}

			addrs := strings.Split(ip.String(), "/")
			addr := addrs[0]
			parsed := net.ParseIP(addr)
			if parsed == nil {
				continue
			}

			parsed = parsed.To4()
			if parsed == nil {
				continue
			}

			key := "public"
			if strings.HasPrefix(addr, "10.") || strings.HasPrefix(addr, "192.") || strings.HasPrefix(addr, "172") {
				key = "internal"
			}
			addresses[key] = addr
		}
	}

	r := checkResponse{
		Status:    "success",
		Addresses: addresses,
	}

	response, _ := json.Marshal(r)
	w.Write(response)
}
