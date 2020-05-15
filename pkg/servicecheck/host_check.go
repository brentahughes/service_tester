package servicecheck

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/brentahughes/service_tester/pkg/models"
	ping "github.com/digineo/go-ping"
)

type healthResponse struct {
	models.Host

	statusCode   int
	responseBody string
	responseTime time.Duration
	errorMessage error
}

type serviceResponse struct {
	Status        string `json:"status"`
	ReceivedInput string `json:"receivedInput,omitempty"`
	Error         string `json:"error,omitempty"`
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

	if host.PublicIP != "" {
		c.checkNetworkHTTP(host, models.NetworkPublic)
		c.checkNetworkICMP(host, models.NetworkPublic)
		c.checkNetworkTCP(host, models.NetworkPublic)
		c.checkNetworkUDP(host, models.NetworkPublic)
	}

	if host.InternalIP != "" {
		c.checkNetworkHTTP(host, models.NetworkInternal)
		c.checkNetworkICMP(host, models.NetworkInternal)
		c.checkNetworkTCP(host, models.NetworkInternal)
		c.checkNetworkUDP(host, models.NetworkInternal)
	}
}

func (c *Checker) checkNetworkICMP(host models.Host, network models.Network) {
	ip := host.PublicIP
	if network == models.NetworkInternal {
		ip = host.InternalIP
	}

	parsedIP := &net.IPAddr{
		IP: net.ParseIP(ip),
	}

	check := &models.Check{
		CheckType:    models.CheckICMP,
		Status:       models.StatusSuccess,
		StatusCode:   200,
		Network:      network,
		ResponseTime: 3 * time.Second,
	}

	p, err := ping.New("0.0.0.0", "")
	if err != nil {
		c.logger.Errorf("error setting up pinger %v", err)
	}

	duration, err := p.Ping(parsedIP, 3*time.Second)
	if err != nil {
		check.Status = models.StatusError
		check.StatusCode = 500
	} else {
		check.ResponseTime = duration
	}

	if err := host.AddCheck(c.db, check); err != nil {
		c.logger.Errorf("error adding check: %v", err)
		return
	}
}

func (c *Checker) checkNetworkHTTP(host models.Host, network models.Network) {
	ip := host.PublicIP
	if network == models.NetworkInternal {
		ip = host.InternalIP
	}

	check := &models.Check{
		CheckType:  models.CheckHTTP,
		Status:     models.StatusSuccess,
		StatusCode: 200,
		Network:    network,
	}

	resp := c.checkHealth(ip)
	if resp.errorMessage != nil {
		check.CheckErrorMessage = resp.errorMessage.Error()
		check.Status = models.StatusError
	} else {
		check.ResponseBody = resp.responseBody
	}
	check.StatusCode = resp.statusCode
	check.ResponseTime = resp.responseTime

	if err := host.AddCheck(c.db, check); err != nil {
		c.logger.Errorf("error adding check: %v", err)
		return
	}
}

func (c *Checker) checkNetworkTCP(host models.Host, network models.Network) {
	ip := host.PublicIP
	if network == models.NetworkInternal {
		ip = host.InternalIP
	}

	check := &models.Check{
		CheckType:  models.CheckTCP,
		Status:     models.StatusSuccess,
		StatusCode: 200,
		Network:    network,
	}

	start := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, c.servicePort), 3*time.Second)
	if err != nil {
		check.CheckErrorMessage = err.Error()
		check.Status = models.StatusError
		check.StatusCode = 500
	} else {
		defer conn.Close()
		fmt.Fprintln(conn, host.Hostname)
		message, err := bufio.NewReader(conn).ReadBytes('\n')

		if err != nil {
			check.CheckErrorMessage = err.Error()
			check.Status = models.StatusError
			check.StatusCode = 500
		}

		if len(message) > 0 {
			var resp serviceResponse
			if err := json.Unmarshal(message, &resp); err != nil {
				c.logger.Errorf("error unmarshaling tcp response %s:%d %v", ip, c.servicePort, err)
				return
			}

			check.ResponseBody = string(message)
			if resp.Status == "error" {
				check.Status = models.StatusError
				check.StatusCode = 500
				check.CheckErrorMessage = resp.Error
			}
		}
	}

	check.ResponseTime = time.Since(start)
	if err := host.AddCheck(c.db, check); err != nil {
		c.logger.Errorf("error adding check: %v", err)
		return
	}
}

func (c *Checker) checkNetworkUDP(host models.Host, network models.Network) {
	ip := host.PublicIP
	if network == models.NetworkInternal {
		ip = host.InternalIP
	}

	check := &models.Check{
		CheckType:  models.CheckUDP,
		Status:     models.StatusSuccess,
		StatusCode: 200,
		Network:    network,
	}

	raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, c.servicePort))
	if err != nil {
		c.logger.Errorf("error resolvind udp addr %s:%d %v", ip, c.servicePort, err)
		return
	}

	start := time.Now()
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		check.CheckErrorMessage = err.Error()
		check.Status = models.StatusError
		check.StatusCode = 500
	} else {
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(3 * time.Second))

		fmt.Fprintln(conn, host.Hostname)
		message, err := bufio.NewReader(conn).ReadBytes('\n')

		if err != nil {
			check.CheckErrorMessage = err.Error()
			check.Status = models.StatusError
			check.StatusCode = 500
		}

		if len(message) > 0 {
			var resp serviceResponse
			if err := json.Unmarshal(message, &resp); err != nil {
				c.logger.Errorf("error unmarshaling tcp response %s:%d %v", ip, c.servicePort, err)
				return
			}

			check.ResponseBody = string(message)
			if resp.Status == "error" {
				check.Status = models.StatusError
				check.StatusCode = 500
				check.CheckErrorMessage = resp.Error
			}
		}
	}

	check.ResponseTime = time.Since(start)
	if err := host.AddCheck(c.db, check); err != nil {
		c.logger.Errorf("error adding check: %v", err)
		return
	}
}

func (c *Checker) checkHealth(host string) (checkResp healthResponse) {
	client := http.DefaultClient
	client.Timeout = 3 * time.Second
	defer client.CloseIdleConnections()

	timer := time.Now()
	resp, err := client.Get(fmt.Sprintf("http://%s/api/health", host))
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
