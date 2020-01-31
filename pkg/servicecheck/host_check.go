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

func (c *Checker) checkHost(host string, port uint16) {
	client := http.DefaultClient
	client.Timeout = 10 * time.Second

	schema := "http"
	if port == 443 {
		schema = "https"
	}

	timer := time.Now()
	resp, err := client.Get(fmt.Sprintf("%s://%s:%d/check", schema, host, port))
	if err != nil {
		log.Print(err)
		return
	}
	defer resp.Body.Close()

	status := db.StatusSuccess
	if resp.StatusCode > 399 {
		status = db.StatusError
	}

	checkResp := checkHostResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		return
	}

	if err := json.Unmarshal(body, &checkResp); err != nil {
		status = db.StatusError
	}

	internal := checkResp.Addresses["internal"]
	public := checkResp.Addresses["public"]

	hostCheck := db.HostCheck{
		Name:         host,
		Status:       status,
		InternalIP:   net.ParseIP(internal),
		PublicIP:     net.ParseIP(public),
		ResponseTime: time.Since(timer),
	}
	if err := c.db.Create(hostCheck); err != nil {
		log.Print(err)
		return
	}
}
