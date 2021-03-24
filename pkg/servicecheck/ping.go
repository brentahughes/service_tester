package servicecheck

import (
	"errors"
	"net"
	"time"
)

var errPingDisabled = errors.New("ping disabled")

type pinger interface {
	Close()
	Ping(*net.IPAddr, time.Duration) (time.Duration, error)
}

type pingNoOp struct{}

func (p *pingNoOp) Close() {}

func (p *pingNoOp) Ping(ip *net.IPAddr, timeout time.Duration) (time.Duration, error) {
	return 0, errPingDisabled
}
