package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dgraph-io/badger"
)

const (
	checkLimit = 100

	StatusSuccess Status = "success"
	StatusError   Status = "error"

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

type Checks []Check

func (c Checks) getByType(t CheckType) Checks {
	var checks Checks
	for _, check := range c {
		if check.CheckType == t {
			checks = append(checks, check)
		}
	}
	return checks
}

func (c Checks) getByNetwork(n Network) Checks {
	var checks Checks
	for _, check := range c {
		if check.Network == n {
			checks = append(checks, check)
		}
	}
	return checks
}

func (h *Host) AddCheck(db *badger.DB, check *Check) error {
	check.ID = getID()
	check.HostID = h.ID
	check.CheckedAt = time.Now().UTC()
	return db.Update(func(txn *badger.Txn) error {
		checkJSON, _ := json.Marshal(check)
		key := fmt.Sprintf("checks.%s.%s.%s.%d", h.ID, check.Network, check.CheckType, check.CheckedAt.Unix())
		entry := badger.NewEntry([]byte(key), checkJSON).WithTTL(3 * time.Hour)
		if err := txn.SetEntry(entry); err != nil {
			return err
		}

		// Add the latest
		key = fmt.Sprintf("checks.%s.latest.%s.%s", h.ID, check.Network, check.CheckType)
		entry = badger.NewEntry([]byte(key), checkJSON).WithTTL(3 * time.Hour)
		return txn.SetEntry(entry)
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
	h.LatestStatuses = &LatestStatuses{}
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
				h.LatestStatuses.Internal.HTTP = check.Status
			case "internal.ICMP":
				h.LatestStatuses.Internal.ICMP = check.Status
			case "internal.TCP":
				h.LatestStatuses.Internal.TCP = check.Status
			case "internal.UDP":
				h.LatestStatuses.Internal.UDP = check.Status
			case "public.HTTP":
				h.LatestStatuses.Public.HTTP = check.Status
			case "public.ICMP":
				h.LatestStatuses.Public.ICMP = check.Status
			case "public.TCP":
				h.LatestStatuses.Public.TCP = check.Status
			case "public.UDP":
				h.LatestStatuses.Public.UDP = check.Status
			}
		}
		return nil
	})
}
