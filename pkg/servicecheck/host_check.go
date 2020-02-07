package servicecheck

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/brentahughes/service_tester/pkg/db"
)

type checkHostResponse struct {
	Addresses map[string]string `json:"addresses"`
}

func (c *Checker) checkHost(hostname string, port uint16) {

	var public, internal string
	internalStatus := db.StatusError
	publicStatus := db.StatusError

	timer := time.Now()
	r, ok := c.checkEndpoint(hostname, port)
	if ok {
		internalStatus = db.StatusSuccess
	}

	responseTime := time.Since(timer)

	if r != nil {
		if publicIP, ok := r.Addresses["public"]; ok && publicIP != "" {
			if _, ok := c.checkEndpoint(publicIP, port); ok {
				publicStatus = db.StatusSuccess
			}
		}

		internal = r.Addresses["internal"]
		public = r.Addresses["public"]
	}

	host := &db.Host{
		Hostname:   hostname,
		InternalIP: net.ParseIP(internal),
		PublicIP:   net.ParseIP(public),
	}
	if err := host.Save(c.db); err != nil {
		log.Println(err)
		return
	}

	hostCheck := db.HostCheck{
		Name:       host,
		InternalIP: internal,
		PublicIP:   public,

		Checks: []db.CheckData{
			db.CheckData{
				InternalStatus: internalStatus,
				PublicStatus:   publicStatus,
				ResponseTime:   responseTime,
				ResponseTimeMS: responseTime.Milliseconds(),
			},
		},
	}
	if err := c.db.Create(hostCheck); err != nil {
		log.Print(err)
		return
	}
}

func (c *Checker) checkEndpoint(host string, port uint16) (*checkHostResponse, bool) {
	client := http.DefaultClient
	client.Timeout = 10 * time.Second
	defer client.CloseIdleConnections()

	schema := "http"
	if port == 443 {
		schema = "https"
	}

	resp, err := client.Get(fmt.Sprintf("%s://%s:%d/check", schema, host, port))
	if err != nil {
		log.Print(err)
		return nil, false
	}
	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		return nil, false
	}

	checkResp := checkHostResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		return nil, false
	}

	if err := json.Unmarshal(body, &checkResp); err != nil {
		return nil, false
	}

	return &checkResp, true
}
