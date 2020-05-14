package servicecheck

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/brentahughes/service_tester/pkg/models"
)

type checkHostResponse struct {
	models.Host

	statusCode   int
	responseBody string
	responseTime time.Duration
	errorMessage error
}

func (c *Checker) newHost(ip string) {
	resp := c.checkHealth(ip)
	if resp.errorMessage != nil {
		c.logger.Errorf("error getting health of new host: %s", resp.errorMessage)
		return
	}

	host := &models.Host{
		Hostname:   resp.Hostname,
		PublicIP:   resp.PublicIP,
		InternalIP: resp.InternalIP,
	}

	if err := host.Save(c.db); err != nil {
		c.logger.Errorf("error saving host (%s): %v", host.Hostname, err)
		return
	}
}

func (c *Checker) checkHost(input interface{}) {
	host := input.(models.Host)

	resp := c.checkHealth(host.PublicIP)

	check := &models.Check{
		CheckType:    models.CheckTCP,
		Status:       models.StatusSuccess,
		StatusCode:   resp.statusCode,
		ResponseBody: resp.responseBody,
		ResponseTime: resp.responseTime,
		Network:      models.NetworkPublic,
	}

	if resp.errorMessage != nil {
		check.CheckErrorMessage = resp.errorMessage.Error()
		check.Status = models.StatusError
	}

	if err := host.AddCheck(c.db, check); err != nil {
		c.logger.Errorf("error adding check: %v", err)
		return
	}
}

func (c *Checker) checkHealth(host string) (checkResp checkHostResponse) {
	client := http.DefaultClient
	client.Timeout = 3 * time.Second
	defer client.CloseIdleConnections()

	timer := time.Now()
	resp, err := client.Get(fmt.Sprintf("http://%s:%d/api/health", host, c.servicePort))
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
