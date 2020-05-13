package servicecheck

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/brentahughes/service_tester/pkg/models"
)

type checkHostResponse struct {
	models.Host

	statusCode   int
	responseBody string
	responseTime time.Duration
	errorMessage error
}

func (c *Checker) checkHost(checkType models.CheckType, endpoint string) {
	c.pool.Submit(func() { c.runHostCheck(checkType, endpoint) })
}

func (c *Checker) runHostCheck(checkType models.CheckType, endpoint string) {
	host, err := models.GetHostByIP(c.db, endpoint)
	if err != nil && err != storm.ErrNotFound {
		c.logger.Errorf("error checking for existing host on ip %s: %v", endpoint, err)
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
	if host.CurrentHost {
		return
	}

	resp := c.checkEndpoint(endpoint)
	if host.ID == 0 && resp.errorMessage != nil {
		return
	}

	check := &models.Check{
		CheckType:    checkType,
		Status:       models.StatusSuccess,
		StatusCode:   resp.statusCode,
		ResponseBody: resp.responseBody,
		ResponseTime: resp.responseTime,
	}

	if resp.errorMessage != nil {
		check.Status = models.StatusError
		check.CheckErrorMessage = resp.errorMessage.Error()
	} else {
		existingHost, _ := models.GetHostByHostname(c.db, resp.Hostname)
		if existingHost != nil {
			host = existingHost
		}

		host.Hostname = resp.Hostname

		if resp.PublicIP != "" {
			host.PublicIP = resp.PublicIP
		}
		if resp.InternalIP != "" {
			host.InternalIP = resp.InternalIP
		}
	}

	if err := host.Save(c.db); err != nil {
		c.logger.Errorf("error saving host (%s): %v", host.Hostname, err)
		return
	}

	if err := host.AddCheck(c.db, check); err != nil {
		c.logger.Errorf("error adding check: ", err)
		return
	}
}

func (c *Checker) checkEndpoint(host string) (checkResp checkHostResponse) {
	client := http.DefaultClient
	client.Timeout = 1 * time.Second
	defer client.CloseIdleConnections()

	timer := time.Now()
	resp, err := client.Get(fmt.Sprintf("http://%s:%d/api/health", host, c.servicePort))
	checkResp.responseTime = time.Since(timer)

	if err != nil {
		checkResp.statusCode = 408
		checkResp.errorMessage = err
		c.logger.Errorf("error checking %s health: %v", host, err)
		return
	}
	defer resp.Body.Close()

	checkResp.statusCode = resp.StatusCode
	if resp.StatusCode > 399 {
		checkResp.errorMessage = fmt.Errorf("error bad status response: %d", resp.StatusCode)
		c.logger.Errorf("error bad status response: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		checkResp.errorMessage = err
		c.logger.Errorf("error reading body from %s health: %v", host, err)
		return
	}
	checkResp.responseBody = string(body)

	if err := json.Unmarshal(body, &checkResp); err != nil {
		checkResp.errorMessage = err
		c.logger.Errorf("error unmarshaling body into struct from %s health: %v", host, err)
	}

	return
}
