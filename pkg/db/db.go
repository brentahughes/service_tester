package db

import (
	"time"
)

const (
	StatusSuccess Status = "success"
	StatusError   Status = "error"
)

type Status string

type HostCheck struct {
	ID         string
	Name       string
	InternalIP string
	PublicIP   string
	Latest     CheckData
	Checks     []CheckData
}

type CheckData struct {
	InternalStatus Status
	PublicStatus   Status
	ResponseTime   time.Duration
	ResponseTimeMS int64
	CheckTime      time.Time
	CheckTimeMS    int64
}

type DB interface {
	Create(host HostCheck) error
	GetAllHosts() ([]HostCheck, error)
	GetHost(id string) (*HostCheck, error)
}
