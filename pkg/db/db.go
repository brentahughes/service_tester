package db

import (
	"net"
	"time"
)

const (
	StatusSuccess Status = "success"
	StatusError   Status = "error"
)

type Status string

type HostCheck struct {
	Name         string
	InternalIP   net.IP
	PublicIP     net.IP
	Status       Status
	ResponseTime time.Duration
	CheckTime    time.Time
}

type DB interface {
	Create(host HostCheck) error
	GetLastForAllHosts() ([]HostCheck, error)
	GetLastByHost(host string) (*HostCheck, error)
	GetAllForHost(host string) ([]HostCheck, error)
}
