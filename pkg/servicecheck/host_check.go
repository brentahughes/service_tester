package servicecheck

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/brentahughes/service_tester/pkg/models"
	"github.com/dgraph-io/badger"
)

const checkTimeout = 3 * time.Second

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
		log.Printf("error getting health of new host: %s", resp.errorMessage)
		return
	}

	var discoveredIP string
	if ip != resp.PublicIP || ip != resp.InternalIP {
		discoveredIP = ip
	}

	host := resp.Host
	host.DiscoveredIP = discoveredIP

	if err := host.Save(c.db); err != nil {
		log.Printf("error saving host (%s): %v", host.Hostname, err)
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

	// Get list of hosts known by the current checked host and add them if not known
	if err := c.checkForNewHosts(host.PublicIP); err != nil {
		log.Printf("error getting new hosts from %s", host)
		return
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
		ResponseTime: checkTimeout,
	}

	duration, err := c.pinger.Ping(parsedIP, checkTimeout)
	if err != nil {
		if err == errPingDisabled {
			check.Status = models.StatusUnknown
			check.StatusCode = http.StatusNotImplemented
		} else {
			check.Status = models.StatusError
			check.StatusCode = http.StatusInternalServerError
		}
	} else {
		check.ResponseTime = duration
	}

	if err := host.AddCheck(c.db, check); err != nil {
		log.Printf("error adding check: %v", err)
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
		log.Printf("error adding check: %v", err)
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
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, c.cfg.ServicePort), checkTimeout)
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
				log.Printf("error unmarshaling tcp response %s:%d %v", ip, c.cfg.ServicePort, err)
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
		log.Printf("error adding check: %v", err)
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

	raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, c.cfg.ServicePort))
	if err != nil {
		log.Printf("error resolving udp addr %s:%d %v", ip, c.cfg.ServicePort, err)
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
		conn.SetDeadline(time.Now().Add(checkTimeout))

		// Try UDP up to 3 times before considering it a failure
		for attempt := 0; attempt < 3; attempt++ {
			fmt.Fprintln(conn, "ping")
			message, err := bufio.NewReader(conn).ReadBytes('\n')

			if err != nil {
				check.CheckErrorMessage = err.Error()
				check.Status = models.StatusError
				check.StatusCode = 500
			} else if len(message) > 0 && string(message) != "pong\n" {
				check.Status = models.StatusError
				check.StatusCode = 500
				check.CheckErrorMessage = "wrong response"
			} else {
				break
			}
		}
	}

	check.ResponseTime = time.Since(start)
	if err := host.AddCheck(c.db, check); err != nil {
		log.Printf("error adding check: %v", err)
		return
	}
}

func (c *Checker) checkHealth(host string) (checkResp healthResponse) {
	timer := time.Now()
	resp, err := c.httpClient.Get(fmt.Sprintf("http://%s/api/health", host))
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
		log.Printf("error bad status response: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		checkResp.errorMessage = err
		log.Printf("error reading body from %s health: %v", host, err)
		return
	}
	checkResp.responseBody = string(body)

	// Backwards compatibility to handle remote service still using an int for the ID
	// Change the ID in the response to a string value
	r := regexp.MustCompile(`"ID":(\d+),`)
	body = r.ReplaceAll(body, []byte(`"ID":"${1}",`))

	if err := json.Unmarshal(body, &checkResp); err != nil {
		checkResp.errorMessage = err
		log.Printf("error unmarshaling body into struct from %s health: %v", host, err)
	}

	return
}

// checkForNewHosts will call /api/hosts on the target host and add any hosts that are not currently known
func (c *Checker) checkForNewHosts(host string) error {
	resp, err := c.httpClient.Get(fmt.Sprintf("http://%s/api/hosts", host))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var hosts []models.Host
	if err := json.Unmarshal(body, &hosts); err != nil {
		return err
	}

	for _, h := range hosts {
		if _, err := models.GetHostByIP(c.db, h.PublicIP); err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}

			log.Printf("adding new host %s", h.PublicIP)
			c.newHost(h.PublicIP)
		}
	}

	return nil
}
