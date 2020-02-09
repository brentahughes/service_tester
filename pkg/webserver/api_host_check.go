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

func (s *Server) handleApiCheck(w http.ResponseWriter, req *http.Request) {
	internal, public, err := s.getIPs()
	if err != nil {
		s.writeErr(w, 500, err)
		return
	}

	r := checkResponse{
		Status: "success",
		Addresses: map[string]string{
			"internal": internal,
			"public":   public,
		},
	}

	response, _ := json.Marshal(r)
	w.Write(response)
}

func (s *Server) getIPs() (string, string, error) {
	var internalIP, publicIP string

	// Attempt to get the ip from the SP dns first
	internal, _ := net.LookupIP(s.config.InternalPDNS)
	if len(internal) > 0 {
		internalIP = internal[0].String()
	}

	public, _ := net.LookupIP(s.config.PublicIPDNS)
	if len(public) > 0 {
		publicIP = public[0].String()
	}

	if publicIP != "" && internalIP != "" {
		return internalIP, publicIP, nil
	}

	nets, err := net.Interfaces()
	if err != nil {
		return "", "", err
	}

	for _, hostInterface := range nets {
		ips, err := hostInterface.Addrs()
		if err != nil {
			return "", "", err
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

			if strings.HasPrefix(addr, "10.") || strings.HasPrefix(addr, "192.") || strings.HasPrefix(addr, "172") {
				if internalIP == "" {
					internalIP = addr
				}
			} else {
				if publicIP == "" {
					publicIP = addr
				}
			}
		}
	}

	return internalIP, publicIP, nil
}
