package servicecheck

import (
	"net"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/brentahughes/service_tester/pkg/models"
	"github.com/panjf2000/ants"
)

type Checker struct {
	db            *storm.DB
	servicePort   int
	serviceName   string
	checkInterval time.Duration
	logger        *models.Logger
	pool          *ants.PoolWithFunc
}

func NewChecker(
	db *storm.DB,
	serviceName string,
	servicePort int,
	logger *models.Logger,
	checkInterval time.Duration,
	parallelChecks int,
) (*Checker, error) {
	c := &Checker{
		db:            db,
		servicePort:   servicePort,
		serviceName:   serviceName,
		checkInterval: checkInterval,
		logger:        logger,
	}

	pool, err := ants.NewPoolWithFunc(parallelChecks, c.checkHost)
	if err != nil {
		return nil, err
	}
	c.pool = pool

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
}

func (c *Checker) runCheck() {
	c.discoverNewHosts()

	hosts, err := models.GetRecentHosts(c.db)
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
	}

	for _, ip := range ips {
		host, err := models.GetHostByIP(c.db, ip)
		if err != nil {
			if err != storm.ErrNotFound {
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
