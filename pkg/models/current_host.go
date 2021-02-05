package models

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/brentahughes/service_tester/pkg/config"
	"github.com/dgraph-io/badger"
	servicehost "github.com/shirou/gopsutil/host"
)

func GetCurrentHost(db *badger.DB) (*Host, error) {
	var host Host
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(hostsPrefix + currentHostID))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &host)
		})
	})
	if err != nil {
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

	host.ServiceUptime = time.Since(host.ServiceLastStart).Truncate(time.Second)
	host.HostUptime = uptimeDur

	return &host, nil
}

func UpdateCurrentHost(db *badger.DB, conf *config.Config, init bool) error {
	host, err := GetCurrentHost(db)
	if err != nil && err != badger.ErrKeyNotFound {
		return err
	}

	if host == nil {
		host = &Host{
			ID:                currentHostID,
			CurrentHost:       true,
			ServiceRestarts:   -1,
			ServiceFirstStart: time.Now().UTC(),
		}
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	internal, public, err := getLocalHostIPs(conf)
	if err != nil {
		return err
	}

	if host.Hostname != hostname {
		host.Hostname = hostname
	}
	if host.InternalIP != internal {
		host.InternalIP = internal
	}
	if host.PublicIP != public {
		host.PublicIP = public
	}

	if init {
		host.ServiceLastStart = time.Now().UTC()
		host.ServiceRestarts++
	}

	return db.Update(func(txn *badger.Txn) error {
		hostJSON, _ := json.Marshal(host)
		return txn.Set([]byte(hostsPrefix+host.ID), hostJSON)
	})
}

func getLocalHostIPs(conf *config.Config) (string, string, error) {
	var internalIP, publicIP string

	// Attempt to get the ip from the SP dns first
	internal, _ := net.LookupIP(conf.InternalPDNS)
	if len(internal) > 0 {
		internalIP = internal[0].String()
	}

	public, _ := net.LookupIP(conf.PublicIPDNS)
	if len(public) > 0 {
		publicIP = public[0].String()
	}

	if publicIP != "" && internalIP != "" {
		return internalIP, publicIP, nil
	}

	nets, err := net.Interfaces()
	if err != nil {
		return "", "", err
	}

	for _, hostInterface := range nets {
		ips, err := hostInterface.Addrs()
		if err != nil {
			return "", "", err
		}
		for _, ip := range ips {
			if strings.HasPrefix(ip.String(), "127.") {
				continue
			}

			addrs := strings.Split(ip.String(), "/")
			addr := addrs[0]
			parsed := net.ParseIP(addr)
			if parsed == nil {
				continue
			}

			parsed = parsed.To4()
			if parsed == nil {
				continue
			}

			if strings.HasPrefix(addr, "10.") || strings.HasPrefix(addr, "192.") || strings.HasPrefix(addr, "172") {
				if internalIP == "" {
					internalIP = addr
				}
			} else {
				if publicIP == "" {
					publicIP = addr
				}
			}
		}
	}

	return internalIP, publicIP, nil
}
