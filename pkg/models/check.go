package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dgraph-io/badger"
)

const (
	checkTTL = 1 * time.Hour

	checkLimit = 100

	StatusSuccess Status = "success"
	StatusError   Status = "error"
	StatusUnknown Status = "unknown"

	NetworkInternal Network = "internal"
	NetworkPublic   Network = "public"

	CheckTCP  CheckType = "TCP"
	CheckUDP  CheckType = "UDP"
	CheckICMP CheckType = "ICMP"
	CheckHTTP CheckType = "HTTP"
)

type Network string
type CheckType string
type Status string

type Check struct {
	ID                string        `json:"id"`
	HostID            string        `json:"hostId"`
	Status            Status        `json:"status"`
	ResponseTime      time.Duration `json:"responseTime"`
	StatusCode        int           `json:"statusCode"`
	ResponseBody      string        `json:"responseBody"`
	CheckErrorMessage string        `json:"checkErrorMessage"`
	Network           Network       `json:"network"`
	CheckType         CheckType     `json:"checkType"`
	CheckedAt         time.Time     `json:"checkedAt"`
}

func (h *Host) AddCheck(db *badger.DB, check *Check) error {
	check.ID = getID()
	check.HostID = h.ID
	check.CheckedAt = time.Now().UTC()
	return db.Update(func(txn *badger.Txn) error {
		checkJSON, _ := json.Marshal(check)
		key := fmt.Sprintf("checks.%s.%s.%s.%d", h.ID, check.Network, check.CheckType, check.CheckedAt.Unix())
		entry := badger.NewEntry([]byte(key), checkJSON).WithTTL(checkTTL)
		if err := txn.SetEntry(entry); err != nil {
			return err
		}

		// Add the latest
		key = fmt.Sprintf("checks.%s.latest.%s.%s", h.ID, check.Network, check.CheckType)
		if err := txn.Set([]byte(key), checkJSON); err != nil {
			return err
		}

		return h.updateUptime(db, check)
	})
}

func (h *Host) addChecks(db *badger.DB) error {
	h.Checks = &ServiceChecks{}
	return db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{})
		defer it.Close()

		for it.Seek([]byte("checks." + h.ID + ".")); it.ValidForPrefix([]byte("checks." + h.ID + ".")); it.Next() {
			item := it.Item()

			var check Check
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &check)
			})
			if err != nil {
				return err
			}

			keyParts := strings.Split(string(item.Key()), ".")
			switch fmt.Sprintf("%s.%s", keyParts[2], keyParts[3]) {
			case "internal.HTTP":
				h.Checks.Internal.HTTP = append(h.Checks.Internal.HTTP, check)
			case "internal.ICMP":
				h.Checks.Internal.ICMP = append(h.Checks.Internal.ICMP, check)
			case "internal.TCP":
				h.Checks.Internal.TCP = append(h.Checks.Internal.TCP, check)
			case "internal.UDP":
				h.Checks.Internal.UDP = append(h.Checks.Internal.UDP, check)
			case "public.HTTP":
				h.Checks.Public.HTTP = append(h.Checks.Public.HTTP, check)
			case "public.ICMP":
				h.Checks.Public.ICMP = append(h.Checks.Public.ICMP, check)
			case "public.TCP":
				h.Checks.Public.TCP = append(h.Checks.Public.TCP, check)
			case "public.UDP":
				h.Checks.Public.UDP = append(h.Checks.Public.UDP, check)
			}
		}
		return nil
	})
}

func (h *Host) addLatestStatuses(db *badger.DB) error {
	h.LatestChecks = &ServiceChecks{}
	return db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte("checks." + h.ID + ".latest.")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			var check Check
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &check)
			})
			if err != nil {
				return err
			}

			keyParts := strings.Split(string(item.Key()), ".")
			switch fmt.Sprintf("%s.%s", keyParts[3], keyParts[4]) {
			case "internal.HTTP":
				h.LatestChecks.Internal.HTTP = []Check{check}
			case "internal.ICMP":
				h.LatestChecks.Internal.ICMP = []Check{check}
			case "internal.TCP":
				h.LatestChecks.Internal.TCP = []Check{check}
			case "internal.UDP":
				h.LatestChecks.Internal.UDP = []Check{check}
			case "public.HTTP":
				h.LatestChecks.Public.HTTP = []Check{check}
			case "public.ICMP":
				h.LatestChecks.Public.ICMP = []Check{check}
			case "public.TCP":
				h.LatestChecks.Public.TCP = []Check{check}
			case "public.UDP":
				h.LatestChecks.Public.UDP = []Check{check}
			}
		}
		return nil
	})
}
