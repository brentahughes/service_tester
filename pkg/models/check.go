package models

import (
	"net"
	"time"
)

const (
	StatusSuccess Status = "success"
	StatusError   Status = "error"
)

type Status string

type Check struct {
	ID           int    `storm:"id,increment"`
	HostID       int    `storm:"index"`
	Status       Status `storm:"index"`
	IP           net.IP
	ResponseTime time.Duration
	CheckedAt    time.Time
}
