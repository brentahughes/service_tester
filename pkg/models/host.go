package models

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/dgraph-io/badger"
)

const (
	hostsPrefix    = "hosts.id."
	hostnamePrefix = "hosts.hostname."
	ipPrefix       = "hosts.ip."
)

type Host struct {
	ID                string         `json:"id" badgerhold:"key"`
	CurrentHost       bool           `json:"-"`
	Hostname          string         `json:"hostname" badgerhold:"unique"`
	ServiceRestarts   int            `json:"serviceRestarts,omitempty"`
	ServiceFirstStart time.Time      `json:"serviceFirstStart"`
	ServiceLastStart  time.Time      `json:"serviceLastStart"`
	InternalIP        string         `json:"internalIp" badgerhold:"unique"`
	PublicIP          string         `json:"publicIp" badgerhold:"unique"`
	DiscoveredIP      string         `json:"-"`
	ServiceUptime     time.Duration  `json:"serviceUptime,omitempty"`
	HostUptime        time.Duration  `json:"hostUptime,omitempty"`
	FirstSeenAt       time.Time      `json:"firstSeenAt"`
	LastSeenAt        time.Time      `json:"lastSeenAt" badgerhold:"index"`
	LatestChecks      *ServiceChecks `json:"latestChecks,omitempty"`
	Checks            *ServiceChecks `json:"checks,omitempty"`
	CheckUptime       *CheckUptime   `json:"checkUptime"`
	CityCode          string         `json:"cityCode,omitempty"`
	Longitude         string         `json:"longitude,omitempty"`
	Latitude          string         `json:"latitude,omitempty"`
}

type ServiceChecks struct {
	Internal CheckTypes `json:"internal"`
	Public   CheckTypes `json:"public"`
}

type CheckTypes struct {
	HTTP []Check `json:"http"`
	TCP  []Check `json:"tcp"`
	UDP  []Check `json:"udp"`
	ICMP []Check `json:"icmp"`
}

func GetHostByHostname(db *badger.DB, hostname string) (*Host, error) {
	var host Host
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(hostnamePrefix + hostname))
		if err != nil {
			return err
		}

		var id string
		item.Value(func(val []byte) error {
			id = string(val)
			return nil
		})

		item, err = txn.Get([]byte(hostsPrefix + id))
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

	return &host, nil
}

func GetHostByID(db *badger.DB, id string) (*Host, error) {
	var host Host
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(hostsPrefix + id))
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

	if err := host.addLatestStatuses(db); err != nil {
		return nil, err
	}

	if err := host.addChecks(db); err != nil {
		return nil, err
	}

	if err := host.setUptimes(db); err != nil {
		return nil, err
	}

	return &host, nil
}

func GetHostByIP(db *badger.DB, ip string) (*Host, error) {
	var host Host
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(ipPrefix + ip))
		if err != nil {
			return err
		}

		var id string
		item.Value(func(val []byte) error {
			id = string(val)
			return nil
		})

		item, err = txn.Get([]byte(hostsPrefix + id))
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

	return &host, nil
}

func GetHostsWithStatuses(db *badger.DB) ([]Host, error) {
	hosts, err := GetHosts(db)
	if err != nil {
		return nil, err
	}

	for i, host := range hosts {
		if err := host.addLatestStatuses(db); err != nil {
			return nil, err
		}

		if err := host.setUptimes(db); err != nil {
			return nil, err
		}

		hosts[i] = host
	}

	return hosts, nil
}

func GetHosts(db *badger.DB) ([]Host, error) {
	var hosts []Host
	err := db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek([]byte(hostsPrefix)); it.ValidForPrefix([]byte(hostsPrefix)); it.Next() {
			var host Host

			item := it.Item()

			if strings.HasSuffix(string(item.Key()), currentHostID) {
				continue
			}

			err := item.Value(func(val []byte) error {
				if err := json.Unmarshal(val, &host); err != nil {
					return err
				}

				hosts = append(hosts, host)
				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(hosts, func(a, b int) bool {
		return sort.StringsAreSorted([]string{hosts[a].Hostname, hosts[b].Hostname})
	})

	return hosts, nil
}

func (h *Host) Save(db *badger.DB) error {
	h.LastSeenAt = time.Now().UTC()
	if h.FirstSeenAt.IsZero() {
		h.FirstSeenAt = time.Now().UTC()
	}

	if h.ID == "" {
		h.ID = getID()
		if h.CurrentHost {
			h.ID = currentHostID
		}
	}
	h.Checks = nil
	h.LatestChecks = nil
	h.CheckUptime = nil

	return db.Update(func(txn *badger.Txn) error {
		jsonHost, _ := json.Marshal(h)
		if err := txn.Set([]byte(hostsPrefix+h.ID), jsonHost); err != nil {
			return err
		}

		if err := txn.Set([]byte(hostnamePrefix+h.Hostname), []byte(h.ID)); err != nil {
			return err
		}

		if h.PublicIP != "" {
			if err := txn.Set([]byte(ipPrefix+h.PublicIP), []byte(h.ID)); err != nil {
				return err
			}
		}

		if h.InternalIP != "" {
			if err := txn.Set([]byte(ipPrefix+h.InternalIP), []byte(h.ID)); err != nil {
				return err
			}
		}

		if h.DiscoveredIP != "" {
			if err := txn.Set([]byte(ipPrefix+h.DiscoveredIP), []byte(h.ID)); err != nil {
				return err
			}
		}
		return nil
	})
}
