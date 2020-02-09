package models

import (
	"fmt"
	"os"
	"time"

	"github.com/asdine/storm/v3"
	servicehost "github.com/shirou/gopsutil/host"
)

type CurrentHost struct {
	Hostname          string `storm:"id"`
	ServiceRestarts   int
	ServiceFirstStart time.Time
	ServiceLastStart  time.Time

	ServiceUptime time.Duration `json:"-"`
	HostUptime    time.Duration `json:"-"`
}

func GetCurrentHost(db *storm.DB) (*CurrentHost, error) {
	var host CurrentHost
	if err := db.Select().First(&host); err != nil {
		return nil, err
	}

	uptime, err := servicehost.Uptime()
	if err != nil {
		return nil, err
	}

	uptimeDur, err := time.ParseDuration(fmt.Sprintf("%ds", uptime))
	if err != nil {
		return nil, err
	}

	host.ServiceUptime = time.Since(host.ServiceLastStart)
	host.HostUptime = uptimeDur
	return &host, nil
}

func UpdateCurrentHost(db *storm.DB) error {
	host, err := GetCurrentHost(db)
	if err != nil && err != storm.ErrNotFound {
		return err
	}

	if host == nil {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}

		host = &CurrentHost{
			Hostname:          hostname,
			ServiceRestarts:   -1,
			ServiceFirstStart: time.Now().UTC(),
		}
	}

	host.ServiceLastStart = time.Now().UTC()
	host.ServiceRestarts++
	return db.Save(host)
}
