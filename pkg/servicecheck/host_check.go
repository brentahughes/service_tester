package servicecheck

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/brentahughes/service_tester/pkg/models"
)

type checkHostResponse struct {
	Hostname  string `json:"hostname"`
	Addresses struct {
		Public   string `json:"public"`
		Internal string `json:"internal"`
	} `json:"addresses"`
}

func (c *Checker) checkHost(checkType models.CheckType, endpoint string) {
	status := models.StatusSuccess

	host, err := models.GetHostByIP(c.db, endpoint)
	if err != nil && err != storm.ErrNotFound {
		log.Println(err)
		return
	}

	if host == nil {
		host = &models.Host{
			Port: c.servicePort,
		}
	}

	resp, responseTime, err := c.checkEndpoint(endpoint)
	if err != nil {
		log.Println(err)
		status = models.StatusError
	}
	if resp == nil && host == nil {
		return
	}

	check := &models.Check{
		Status:       status,
		CheckType:    checkType,
		ResponseTime: responseTime,
	}

	if resp != nil {
		host.Hostname = resp.Hostname

		if resp.Addresses.Public != "" {
			host.PublicIP = resp.Addresses.Public
		}
		if resp.Addresses.Internal != "" {
			host.InternalIP = resp.Addresses.Internal
		}
	}

	switch checkType {
	case models.CheckInternal:
		host.InternalIP = endpoint
	case models.CheckPublic:
		host.PublicIP = endpoint
	}

	if err := host.AddCheck(c.db, check); err != nil {
		log.Println("error adding check: ", err)
		return
	}

	if err := host.Save(c.db); err != nil {
		log.Printf("error saving host (%s): %v", host.Hostname, err)
		return
	}
}

func (c *Checker) checkEndpoint(host string) (*checkHostResponse, time.Duration, error) {
	client := http.DefaultClient
	client.Timeout = 1 * time.Second
	defer client.CloseIdleConnections()

	timer := time.Now()
	resp, err := client.Get(fmt.Sprintf("http://%s:%d/api/check", host, c.servicePort))
	responseTime := time.Since(timer)

	if err != nil {
		return nil, responseTime, err
	}

	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		return nil, responseTime, fmt.Errorf("error bad status response: %d", resp.StatusCode)
	}

	checkResp := checkHostResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, responseTime, err
	}

	if err := json.Unmarshal(body, &checkResp); err != nil {
		return nil, responseTime, err
	}

	return &checkResp, responseTime, nil
}
