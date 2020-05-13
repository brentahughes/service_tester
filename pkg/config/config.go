package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           int
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

	discoveryURL := os.Getenv("DISCOVERY_NAME")
	if discoveryURL == "" {
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
		Discovery:      discoveryURL,
		CheckInterval:  checkInterval,
		InternalPDNS:   internalIP,
		PublicIPDNS:    publicIP,
		ParallelChecks: parallelChecks,
	}, nil
}
