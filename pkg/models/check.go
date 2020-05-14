package models

import (
	"fmt"
	"time"

	"github.com/asdine/storm/q"
	"github.com/asdine/storm/v3"
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
)

type Network string
type CheckType string
type Status string

type Check struct {
	ID                int    `storm:"id,increment"`
	HostID            int    `storm:"index"`
	Status            Status `storm:"index"`
	ResponseTime      time.Duration
	StatusCode        int
	ResponseBody      string
	CheckErrorMessage string
	Network           Network   `storm:"index"`
	CheckType         CheckType `storm:"index"`
	CheckedAt         time.Time `storm:"index"`
}

func (h *Host) getCheckDB(db *storm.DB) storm.Node {
	return db.From(fmt.Sprintf("host.%d.check", h.ID))
}

func (h *Host) AddCheck(db *storm.DB, check *Check) error {
	check.HostID = h.ID
	check.CheckedAt = time.Now().UTC()
	if err := h.getCheckDB(db).Save(check); err != nil {
		return err
	}

	// Check If more than 100 checks exist and delete older ones
	count, err := h.getCheckDB(db).
		Select(q.Eq("HostID", h.ID), q.Eq("CheckType", check.CheckType), q.Eq("Network", check.Network)).
		Count(&Check{})
	if err != nil {
		return err
	}

	if count > checkLimit {
		query := h.getCheckDB(db).
			Select(q.Eq("HostID", h.ID), q.Eq("CheckType", check.CheckType), q.Eq("Network", check.Network)).
			OrderBy("CheckedAt").
			Limit(count - checkLimit)
		if err := query.Delete(&Check{}); err != nil && err != storm.ErrNotFound {
			return err
		}
	}

	return nil
}

func (h *Host) addChecks(db *storm.DB) {
	var checks []Check

	if err := h.getHostCheckByNetwork(db, NetworkInternal).Find(&checks); err == nil {
		h.Checks.Internal = checks
	}

	if err := h.getHostCheckByNetwork(db, NetworkPublic).Find(&checks); err == nil {
		h.Checks.Public = checks
	}
}

func (h *Host) addLastChecks(db *storm.DB) {
	var check Check

	if err := h.getHostCheckByNetwork(db, NetworkInternal).First(&check); err == nil {
		h.LatestChecks.Internal = check
	}

	if err := h.getHostCheckByNetwork(db, NetworkPublic).First(&check); err == nil {
		h.LatestChecks.Public = check
	}
}

func (h *Host) getHostCheckByNetwork(db *storm.DB, network Network) storm.Query {
	return h.getCheckDB(db).Select(q.Eq("HostID", h.ID), q.Eq("Network", network)).OrderBy("CheckedAt").Reverse()
}
