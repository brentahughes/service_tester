package servicecheck

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/brentahughes/service_tester/pkg/models"
)

type Checker struct {
	db            *storm.DB
	servicePort   int
	serviceName   string
	checkInterval time.Duration
}

func NewChecker(db *storm.DB, serviceName string, servicePort int, checkInterval time.Duration) *Checker {
	return &Checker{
		db:            db,
		servicePort:   servicePort,
		serviceName:   serviceName,
		checkInterval: checkInterval,
	}
}

func (c *Checker) Start() {
	if err := c.runCheck(); err != nil {
		log.Fatal("error on first check run: ", err)
	}

	tick := time.NewTicker(c.checkInterval)
	for range tick.C {
		if err := c.runCheck(); err != nil {
			log.Println(err)
			return
		}
	}
}

func (c *Checker) Stop() {}

func (c *Checker) runCheck() error {
	hostnames, err := c.getHostnamesFromSRV()
	if err != nil {
		return err
	}

	for _, hostname := range hostnames {
		if _, err := models.GetHostByHostname(c.db, hostname); err != nil && err == storm.ErrNotFound {
			c.checkHost(models.CheckHostname, hostname, hostname)
		}
	}

	hosts, err := models.GetAllHosts(c.db)
	if err != nil {
		return err
	}

	for _, host := range hosts {
		c.checkHost(models.CheckHostname, host.Hostname, host.Hostname)

		if host.InternalIP != "" {
			c.checkHost(models.CheckInternal, host.Hostname, host.InternalIP)
		}

		if host.PublicIP != "" {
			c.checkHost(models.CheckPublic, host.Hostname, host.PublicIP)
		}
	}

	return nil
}

func (c *Checker) getHostnamesFromSRV() ([]string, error) {
	parts := strings.SplitN(c.serviceName, ".", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("%s is not a valid srv", c.serviceName)
	}

	_, addrs, err := net.LookupSRV(strings.TrimPrefix(parts[0], "_"), strings.TrimPrefix(parts[1], "_"), parts[2])
	if err != nil {
		return nil, err
	}

	var hosts []string
	for _, addr := range addrs {
		hosts = append(hosts, strings.TrimSuffix(addr.Target, "."))
	}
	return hosts, nil
}
