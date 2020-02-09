package servicecheck

import (
	"log"
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

	currentHost, err := models.GetCurrentHost(c.db)
	if err != nil {
		log.Fatal("error getting host info: ", err)
	}
	c.currentHostname = currentHost.Hostname

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

	for _, host := range hostnames {
		checkType := models.CheckPublic
		if strings.HasPrefix(host, "10.") || strings.HasPrefix(host, "192.") || strings.HasPrefix(host, "172.") {
			checkType = models.CheckInternal
		}

		if _, err := models.GetHostByIP(c.db, host); err != nil {
			c.checkHost(checkType, host)
		}
	}

	hosts, err := models.GetAllHosts(c.db)
	if err != nil {
		return err
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
