package models

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger"
)

type Uptime interface {
	incrementTotalChecks()
	incrementTotalSuccess()
	setPercent()
}

type CheckUptime struct {
	Percent      float64            `json:"percent"`
	TotalSuccess uint64             `json:"totalSuccess"`
	TotalChecks  uint64             `json:"totalChecks"`
	Internal     CheckNetworkUptime `json:"internal"`
	Public       CheckNetworkUptime `json:"public"`
}

type CheckNetworkUptime struct {
	Percent      float64           `json:"percent"`
	TotalSuccess uint64            `json:"totalSuccess"`
	TotalChecks  uint64            `json:"totalChecks"`
	HTTP         CheckUptimeByType `json:"http"`
	TCP          CheckUptimeByType `json:"tcp"`
	UDP          CheckUptimeByType `json:"udp"`
	ICMP         CheckUptimeByType `json:"icmp"`
}

type CheckUptimeByType struct {
	Percent      float64 `json:"percent"`
	TotalSuccess uint64  `json:"totalSuccess"`
	TotalChecks  uint64  `json:"totalChecks"`
}

func (h *Host) updateUptime(db *badger.DB, check *Check) error {
	return db.Update(func(txn *badger.Txn) error {
		key := fmt.Sprintf("uptime.%s.%s.%s", h.ID, check.Network, check.CheckType)

		// Get the previous uptime data if it exists
		item, err := txn.Get([]byte(key))
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}

		var uptime CheckUptimeByType

		// If a record was found then get the existing uptime metrics from it
		if item != nil {
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &uptime)
			})
			if err != nil {
				return err
			}
		}

		// Increment the counters and calculate the percent
		uptime.TotalChecks++
		if check.Status == StatusSuccess {
			uptime.TotalSuccess++
		}
		uptime.Percent = (float64(uptime.TotalSuccess) / float64(uptime.TotalChecks)) * 100

		// Store updated uptime metrics
		data, err := json.Marshal(uptime)
		if err != nil {
			return err
		}
		return txn.Set([]byte(key), data)
	})
}

func (h *Host) setUptimes(db *badger.DB) error {
	return db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		uptime := &CheckUptime{}
		key := fmt.Sprintf("uptime.%s", h.ID)
		for it.Seek([]byte(key)); it.ValidForPrefix([]byte(key)); it.Next() {
			fullKey := string(it.Item().Key())
			keyParts := strings.Split(fullKey, ".")

			var metrics CheckUptimeByType
			it.Item().Value(func(val []byte) error {
				return json.Unmarshal(val, &metrics)
			})

			uptime.TotalChecks += metrics.TotalChecks
			uptime.TotalSuccess += metrics.TotalSuccess
			uptime.Percent = (float64(uptime.TotalSuccess) / float64(uptime.TotalChecks) * 100)

			switch keyParts[2] {
			case string(NetworkInternal):
				uptime.Internal.TotalChecks += metrics.TotalChecks
				uptime.Internal.TotalSuccess += metrics.TotalSuccess
				uptime.Internal.Percent = (float64(uptime.Internal.TotalSuccess) / float64(uptime.Internal.TotalChecks) * 100)

				switch keyParts[3] {
				case string(CheckHTTP):
					uptime.Internal.HTTP = metrics
				case string(CheckICMP):
					uptime.Internal.ICMP = metrics
				case string(CheckTCP):
					uptime.Internal.TCP = metrics
				case string(CheckUDP):
					uptime.Internal.UDP = metrics
				}
			case string(NetworkPublic):
				uptime.Public.TotalChecks += metrics.TotalChecks
				uptime.Public.TotalSuccess += metrics.TotalSuccess
				uptime.Public.Percent = (float64(uptime.Public.TotalSuccess) / float64(uptime.Public.TotalChecks) * 100)

				switch keyParts[3] {
				case string(CheckHTTP):
					uptime.Public.HTTP = metrics
				case string(CheckICMP):
					uptime.Public.ICMP = metrics
				case string(CheckTCP):
					uptime.Public.TCP = metrics
				case string(CheckUDP):
					uptime.Public.UDP = metrics
				}
			}
		}

		h.CheckUptime = uptime
		return nil
	})
}
