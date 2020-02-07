package servicecheck

import (
	"log"
	"net"
	"strings"
	"time"

	"github.com/asdine/storm/v3"
)

type Checker struct {
	db            *storm.DB
	serviceName   string
	checkInterval time.Duration
}

func NewChecker(db *storm.DB, serviceName string, checkInterval time.Duration) *Checker {
	return &Checker{
		db:            db,
		serviceName:   serviceName,
		checkInterval: checkInterval,
	}
}

func (c *Checker) Start() {
	c.runCheck()

	tick := time.NewTicker(c.checkInterval)
	for range tick.C {
		c.runCheck()
	}
}

func (c *Checker) Stop() {}

func (c *Checker) runCheck() {
	parts := strings.SplitN(c.serviceName, ".", 3)
	if len(parts) != 3 {
		log.Printf("%s is not a valid srv", c.serviceName)
		return
	}

	_, addrs, err := net.LookupSRV(strings.TrimPrefix(parts[0], "_"), strings.TrimPrefix(parts[1], "_"), parts[2])
	if err != nil {
		log.Print(err)
		return
	}

	for _, addr := range addrs {
		go c.checkHost(addr.Target, addr.Port)
	}
}
