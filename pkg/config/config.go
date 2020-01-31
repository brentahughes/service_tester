package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port          int
	Discovery     string
	CheckInterval time.Duration
}

func LoadEnvConfig() (*Config, error) {
	portStr := os.Getenv("WEB_INTERFACE_PORT")

	var err error

	port := 8080
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

	checkIntervalStr := os.Getenv("CHECK_INTERVAL")
	checkInterval := 10 * time.Second
	if checkIntervalStr != "" {
		checkInterval, err = time.ParseDuration(checkIntervalStr)
		if err != nil {
			return nil, err
		}
	}

	return &Config{
		Port:          port,
		Discovery:     discoveryURL,
		CheckInterval: checkInterval,
	}, nil
}
