package servicecheck

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/brentahughes/service_tester/pkg/models"
)

type Checker struct {
	db              *storm.DB
	servicePort     int
	serviceName     string
	checkInterval   time.Duration
	currentHostname string
	logger          *models.Logger
}

func NewChecker(db *storm.DB, serviceName string, servicePort int, logger *models.Logger, checkInterval time.Duration) *Checker {
	return &Checker{
		db:            db,
		servicePort:   servicePort,
		serviceName:   serviceName,
		checkInterval: checkInterval,
		logger:        logger,
	}
}

func (c *Checker) Start() {
	if err := c.runCheck(); err != nil {
		c.logger.Errorf("error on first check run: %v", err)
	}

	tick := time.NewTicker(c.checkInterval)
	for range tick.C {
		if err := c.runCheck(); err != nil {
			c.logger.Errorf("%v", err)
			return
		}
	}
}

func (c *Checker) Stop() {
	c.logger.Infof("Shutting down checker")
}

func (c *Checker) runCheck() error {
	ips, err := c.getHostnamesFromSRV()
	if err != nil {
		return err
	}

	for _, ip := range ips {
		checkType := models.CheckPublic
		if strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "192.") || strings.HasPrefix(ip, "172.") {
			checkType = models.CheckInternal
		}

		if host, err := models.GetHostByIP(c.db, ip); err != nil {
			// Run the first check if this host does not already exist
			c.checkHost(checkType, ip)
		} else {
			// If the host is available then update it's last seen
			host.Save(c.db)
		}
	}

	hosts, err := models.GetRecentHostsWithChecks(c.db)
	if err != nil {
		return fmt.Errorf("error getting recent hosts: %v", err)
	}

	for _, host := range hosts {
		if host.Hostname == c.currentHostname {
			continue
		}

		if host.InternalIP != "" {
			c.checkHost(models.CheckInternal, host.InternalIP)
		}

		if host.PublicIP != "" {
			c.checkHost(models.CheckPublic, host.PublicIP)
		}
	}

	return nil
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
