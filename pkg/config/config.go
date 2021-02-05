package config

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port           int
	ServicePort    int
	Hosts          []string
	Discovery      string
	PublicIPDNS    string
	InternalPDNS   string
	CheckInterval  time.Duration
	ParallelChecks int
}

func LoadEnvConfig() (*Config, error) {
	var err error

	portStr := os.Getenv("WEB_INTERFACE_PORT")
	port := 80
	if portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return nil, err
		}
	}

	servicePortStr := os.Getenv("SERVICE_PORT")
	servicePort := 5500
	if servicePortStr != "" {
		servicePort, err = strconv.Atoi(servicePortStr)
		if err != nil {
			return nil, err
		}
	}

	host := os.Getenv("SERVICE_HOSTS")
	hosts := strings.Split(host, ",")

	// If SERVICE_HOSTS is a file load the ips from the file
	// Should be a plain text file of ips
	if _, err := os.Stat(host); err == nil {
		f, err := os.Open(host)
		if err != nil {
			return nil, err
		}

		hosts = make([]string, 0)
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			hosts = append(hosts, scanner.Text())
		}
	}

	discoveryURL := os.Getenv("DISCOVERY_NAME")
	if discoveryURL == "" && len(hosts) == 0 {
		return nil, errors.New("no DISCOVERY_NAME defined")
	}

	parallelChecksStr := os.Getenv("PARALLEL_CHECKS")
	parallelChecks := 20
	if parallelChecksStr != "" {
		parallelChecks, err = strconv.Atoi(parallelChecksStr)
		if err != nil {
			return nil, err
		}
	}

	checkIntervalStr := os.Getenv("CHECK_INTERVAL")
	checkInterval := 10 * time.Second
	if checkIntervalStr != "" {
		checkInterval, err = time.ParseDuration(checkIntervalStr)
		if err != nil {
			return nil, err
		}
	}

	internalIP := "self.metadata.edgeengine.internal"
	publicIP := "self.metadata.compute.edgeengine.io"

	return &Config{
		Port:           port,
		ServicePort:    servicePort,
		Discovery:      discoveryURL,
		CheckInterval:  checkInterval,
		InternalPDNS:   internalIP,
		PublicIPDNS:    publicIP,
		ParallelChecks: parallelChecks,
		Hosts:          hosts,
	}, nil
}
