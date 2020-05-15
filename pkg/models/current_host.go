package models

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	"github.com/brentahughes/service_tester/pkg/config"
	servicehost "github.com/shirou/gopsutil/host"
)

func GetCurrentHost(db *storm.DB) (*Host, error) {
	var host Host
	if err := db.From("host").Select(q.Eq("CurrentHost", true)).First(&host); err != nil {
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

	hosts, err := GetRecentHosts(db)
	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	host.DiscoveredHosts = make([]string, 0)
	for _, h := range hosts {
		host.DiscoveredHosts = append(host.DiscoveredHosts, h.Hostname)
	}

	return &host, nil
}

func UpdateCurrentHost(db *storm.DB, conf *config.Config) error {
	host, err := GetCurrentHost(db)
	if err != nil && err != storm.ErrNotFound {
		return err
	}

	if host == nil {
		host = &Host{
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

	host.Hostname = hostname
	host.InternalIP = internal
	host.PublicIP = public
	host.ServiceLastStart = time.Now().UTC()
	host.Port = conf.ServicePort
	host.ServiceRestarts++
	return db.From("host").Save(host)
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
