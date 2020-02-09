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

	statusCode   int
	responseBody []byte
	responseTime time.Duration
	errorMessage error
}

func (c *Checker) checkHost(checkType models.CheckType, endpoint string) {
	host, err := models.GetHostByIP(c.db, endpoint)
	if err != nil && err != storm.ErrNotFound {
		log.Println(err)
		return
	}

	if host == nil {
		host = &models.Host{
			Port: c.servicePort,
		}

		switch checkType {
		case models.CheckInternal:
			host.InternalIP = endpoint
		case models.CheckPublic:
			host.PublicIP = endpoint
		}
	}

	resp := c.checkEndpoint(endpoint)

	check := &models.Check{
		CheckType:    checkType,
		Status:       models.StatusSuccess,
		StatusCode:   resp.statusCode,
		ResponseBody: resp.responseBody,
		ResponseTime: resp.responseTime,
	}

	if resp.errorMessage != nil {
		check.Status = models.StatusError
	} else {
		host.Hostname = resp.Hostname

		if resp.Addresses.Public != "" {
			host.PublicIP = resp.Addresses.Public
		}
		if resp.Addresses.Internal != "" {
			host.InternalIP = resp.Addresses.Internal
		}
	}

	if err := host.Save(c.db); err != nil {
		log.Printf("error saving host (%s): %v", host.Hostname, err)
		return
	}

	if err := host.AddCheck(c.db, check); err != nil {
		log.Println("error adding check: ", err)
		return
	}
}

func (c *Checker) checkEndpoint(host string) (checkResp checkHostResponse) {
	client := http.DefaultClient
	client.Timeout = 1 * time.Second
	defer client.CloseIdleConnections()

	timer := time.Now()
	resp, err := client.Get(fmt.Sprintf("http://%s:%d/api/check", host, c.servicePort))
	checkResp.responseTime = time.Since(timer)

	if err != nil {
		checkResp.statusCode = 408
		checkResp.errorMessage = err
		return
	}
	defer resp.Body.Close()

	checkResp.statusCode = resp.StatusCode
	if resp.StatusCode > 399 {
		checkResp.errorMessage = fmt.Errorf("error bad status response: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		checkResp.errorMessage = err
		return
	}
	checkResp.responseBody = body

	if err := json.Unmarshal(body, &checkResp); err != nil {
		checkResp.errorMessage = err
	}

	return
}
