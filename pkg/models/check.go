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
	CheckHTTP CheckType = "HTTP"
)

type Network string
type CheckType string
type Status string

type Check struct {
	ID                int `storm:"id,increment"`
	Status            Status
	ResponseTime      time.Duration
	StatusCode        int
	ResponseBody      string
	CheckErrorMessage string
	Network           Network   `storm:"index"`
	CheckType         CheckType `storm:"index"`
	CheckedAt         time.Time `storm:"index"`
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

func (h *Host) getCheckDB(db *storm.DB) storm.Node {
	return db.From(fmt.Sprintf("host.%d.check", h.ID))
}

func (h *Host) AddCheck(db *storm.DB, check *Check) error {
	check.CheckedAt = time.Now().UTC()
	if err := h.getCheckDB(db).Save(check); err != nil {
		return err
	}

	// Check If more than 100 checks exist and delete older ones
	count, err := h.getCheckDB(db).
		Select(q.Eq("CheckType", check.CheckType), q.Eq("Network", check.Network)).
		Count(&Check{})
	if err != nil {
		return err
	}

	if count > checkLimit {
		query := h.getCheckDB(db).
			Select(q.Eq("CheckType", check.CheckType), q.Eq("Network", check.Network)).
			OrderBy("CheckedAt").
			Limit(count - checkLimit)
		if err := query.Delete(&Check{}); err != nil && err != storm.ErrNotFound {
			return err
		}
	}

	return nil
}

func (h *Host) addChecks(db *storm.DB) {
	var checks Checks
	if err := h.getCheckDB(db).AllByIndex("CheckedAt", &checks, storm.Reverse()); err != nil {
		logger.Errorf("error getting internal checks %v", err)
		return
	}

	h.Checks.Internal.HTTP = checks.getByType(CheckHTTP).getByNetwork(NetworkInternal)
	h.Checks.Internal.TCP = checks.getByType(CheckTCP).getByNetwork(NetworkInternal)
	h.Checks.Internal.UDP = checks.getByType(CheckUDP).getByNetwork(NetworkInternal)
	h.Checks.Internal.ICMP = checks.getByType(CheckICMP).getByNetwork(NetworkInternal)
	h.Checks.Public.HTTP = checks.getByType(CheckHTTP).getByNetwork(NetworkPublic)
	h.Checks.Public.TCP = checks.getByType(CheckTCP).getByNetwork(NetworkPublic)
	h.Checks.Public.UDP = checks.getByType(CheckUDP).getByNetwork(NetworkPublic)
	h.Checks.Public.ICMP = checks.getByType(CheckICMP).getByNetwork(NetworkPublic)
}

func (h *Host) addLastHTTPChecks(db *storm.DB) {
	var check Check

	err := h.getCheckDB(db).
		Select(q.Eq("Network", NetworkInternal), q.Eq("CheckType", CheckHTTP)).
		OrderBy("CheckedAt").
		Reverse().
		First(&check)
	if err == nil {
		h.LatestHTTPChecks.Internal = check
	} else {
		logger.Errorf("error getting last http check %v", err)
	}

	err = h.getCheckDB(db).
		Select(q.Eq("Network", NetworkPublic), q.Eq("CheckType", CheckHTTP)).
		OrderBy("CheckedAt").
		Reverse().
		First(&check)
	if err == nil {
		h.LatestHTTPChecks.Public = check
	} else {
		logger.Errorf("error getting last http check %v", err)
	}
}

func (h *Host) getHostCheckByNetwork(db *storm.DB, network Network) storm.Query {
	return h.getCheckDB(db).Select(q.Eq("Network", network)).OrderBy("CheckedAt").Reverse()
}
