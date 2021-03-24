package servicecheck

import (
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/brentahughes/service_tester/pkg/config"
	"github.com/brentahughes/service_tester/pkg/models"
	"github.com/dgraph-io/badger"
	"github.com/digineo/go-ping"
	"github.com/panjf2000/ants"
)

type Checker struct {
	db         *badger.DB
	cfg        *config.Config
	pool       *ants.PoolWithFunc
	pinger     pinger
	httpClient *http.Client
}

func NewChecker(
	db *badger.DB,
	conf *config.Config,
) (*Checker, error) {
	c := &Checker{
		db:  db,
		cfg: conf,
		httpClient: &http.Client{
			Timeout: checkTimeout,
		},
	}

	pool, err := ants.NewPoolWithFunc(conf.ParallelChecks, c.checkHost)
	if err != nil {
		return nil, err
	}
	c.pool = pool

	var p pinger
	p, err = ping.New("0.0.0.0", "")
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok && strings.Contains(opErr.Err.Error(), "operation not permitted") {
			p = &pingNoOp{}
		} else {
			return nil, err
		}
	}
	c.pinger = p

	return c, nil
}

func (c *Checker) Start() {
	c.runCheck()

	tick := time.NewTicker(c.cfg.CheckInterval)
	for range tick.C {
		c.runCheck()
	}
}

func (c *Checker) Stop() {
	log.Printf("Shutting down checker")
	c.httpClient.CloseIdleConnections()
	c.pinger.Close()
}

func (c *Checker) runCheck() {
	c.discoverNewHosts()

	hosts, err := models.GetHosts(c.db)
	if err != nil {
		log.Printf("error getting recent hosts: %v", err)
		return
	}

	for _, host := range hosts {
		c.pool.Invoke(host)
	}
}

func (c *Checker) discoverNewHosts() {
	var err error

	ips := c.cfg.Hosts
	if c.cfg.Discovery != "" {
		ips, err = c.discoverHosts()
		if err != nil {
			log.Printf("error checking discovery endpoint (%s) %v", c.cfg.Discovery, err)
			return
		}
	}

	currentHost, err := models.GetCurrentHost(c.db)
	if err != nil {
		log.Printf("error getting current host %v", err)
		return
	}

	for _, ip := range ips {
		if currentHost.PublicIP == ip || currentHost.InternalIP == ip {
			continue
		}

		host, err := models.GetHostByIP(c.db, ip)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				log.Printf("error looking up host by ip (%s) %v", ip, err)
				continue
			}

			// Call health endpoint and save host information
			c.newHost(ip)
		} else {
			// Update the last seen
			if err := host.Save(c.db); err != nil {
				log.Printf("error updating host (%s): %v", host.Hostname, err)
				continue
			}
		}
	}
}

func (c *Checker) discoverHosts() ([]string, error) {
	addrs, err := net.LookupIP(c.cfg.Discovery)
	if err != nil {
		return nil, err
	}

	var hosts []string
	for _, addr := range addrs {
		hosts = append(hosts, addr.String())
	}
	return hosts, nil
}
