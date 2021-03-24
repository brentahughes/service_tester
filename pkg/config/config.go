package config

import (
	"bufio"
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	webPort        = flag.Int("web.port", 80, "Port to use for the web and api interface")
	servicePort    = flag.Int("service.port", 5500, "Port to use for the service endpoint")
	serviceHosts   = flag.String("service.hosts", "", "Comma serparated list of hosts or file with hosts listed one per line to use for testing")
	discoveryName  = flag.String("discovery.name", "", "DNS name for A record containing list of host ips")
	parallelChecks = flag.Int("check.parallel", 20, "Number of checks to run in parallel at any time")
	checkInterval  = flag.Duration("check.interval", 10*time.Second, "Time between checking each host")
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

func Init() {
	flag.Parse()
}

func LoadEnvConfig() (*Config, error) {
	var err error

	portStr := os.Getenv("WEB_INTERFACE_PORT")
	port := *webPort
	if portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return nil, err
		}
	}

	servicePortStr := os.Getenv("SERVICE_PORT")
	servicePort := *servicePort
	if servicePortStr != "" {
		servicePort, err = strconv.Atoi(servicePortStr)
		if err != nil {
			return nil, err
		}
	}

	host := os.Getenv("SERVICE_HOSTS")
	if host == "" {
		host = *serviceHosts
	}
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
	if discoveryURL == "" {
		discoveryURL = *discoveryName
	}
	if discoveryURL == "" && len(hosts) == 0 {
		return nil, errors.New("no DISCOVERY_NAME defined")
	}

	parallelChecksStr := os.Getenv("PARALLEL_CHECKS")
	parallelChecks := *parallelChecks
	if parallelChecksStr != "" {
		parallelChecks, err = strconv.Atoi(parallelChecksStr)
		if err != nil {
			return nil, err
		}
	}

	checkIntervalStr := os.Getenv("CHECK_INTERVAL")
	checkInterval := *checkInterval
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
