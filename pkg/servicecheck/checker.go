package servicecheck

import (
	"net"
	"net/http"
	"time"

	"github.com/brentahughes/service_tester/pkg/config"
	"github.com/brentahughes/service_tester/pkg/models"
	"github.com/dgraph-io/badger"
	"github.com/digineo/go-ping"
	"github.com/panjf2000/ants"
)

type Checker struct {
	db            *badger.DB
	servicePort   int
	serviceName   string
	checkInterval time.Duration
	logger        *models.Logger
	pool          *ants.PoolWithFunc
	pinger        *ping.Pinger
	httpClient    *http.Client
}

func NewChecker(
	db *badger.DB,
	logger *models.Logger,
	conf *config.Config,
) (*Checker, error) {
	c := &Checker{
		db:            db,
		servicePort:   conf.ServicePort,
		serviceName:   conf.Discovery,
		checkInterval: conf.CheckInterval,
		logger:        logger,
		httpClient: &http.Client{
			Timeout: checkTimeout,
		},
	}

	pool, err := ants.NewPoolWithFunc(conf.ParallelChecks, c.checkHost)
	if err != nil {
		return nil, err
	}
	c.pool = pool

	pinger, err := ping.New("0.0.0.0", "")
	if err != nil {
		return nil, err
	}
	c.pinger = pinger

	return c, nil
}

func (c *Checker) Start() {
	c.runCheck()

	tick := time.NewTicker(c.checkInterval)
	for range tick.C {
		c.runCheck()
	}
}

func (c *Checker) Stop() {
	c.logger.Infof("Shutting down checker")
	c.httpClient.CloseIdleConnections()
	c.pinger.Close()
}

func (c *Checker) runCheck() {
	c.discoverNewHosts()

	hosts, err := models.GetHosts(c.db)
	if err != nil {
		c.logger.Errorf("error getting recent hosts: %v", err)
		return
	}

	for _, host := range hosts {
		c.pool.Invoke(host)
	}
}

func (c *Checker) discoverNewHosts() {
	ips, err := c.getHostnamesFromSRV()
	if err != nil {
		c.logger.Errorf("error checking discovery endpoint (%s) %v", c.serviceName, err)
		return
	}

	currentHost, err := models.GetCurrentHost(c.db)
	if err != nil {
		c.logger.Errorf("error getting current host %v", err)
		return
	}

	for _, ip := range ips {
		if currentHost.PublicIP == ip || currentHost.InternalIP == ip {
			continue
		}

		host, err := models.GetHostByIP(c.db, ip)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				c.logger.Errorf("error looking up host by ip (%s) %v", ip, err)
				continue
			}

			// Call health endpoint and save host information
			c.newHost(ip)
		} else {
			// Update the last seen
			if err := host.Save(c.db); err != nil {
				c.logger.Errorf("error updating host (%s): %v", host.Hostname, err)
				continue
			}
		}
	}
}

func (c *Checker) getHostnamesFromSRV() ([]string, error) {
	addrs, err := net.LookupIP(c.serviceName)
	if err != nil {
		return nil, err
	}

	var hosts []string
	for _, addr := range addrs {
		hosts = append(hosts, addr.String())
	}
	return hosts, nil
}
