package models

import (
	"time"

	"github.com/asdine/storm/q"
	"github.com/asdine/storm/v3"
)

const (
	checkLimit = 100

	StatusSuccess Status = "success"
	StatusError   Status = "error"

	CheckInternal CheckType = "internal"
	CheckPublic   CheckType = "public"
)

type CheckType string
type Status string

type Check struct {
	ID           int    `storm:"id,increment"`
	HostID       int    `storm:"index"`
	Status       Status `storm:"index"`
	ResponseTime time.Duration
	CheckType    CheckType `storm:"index"`
	CheckedAt    time.Time `storm:"index"`
}

func (h *Host) AddCheck(db *storm.DB, check *Check) error {
	check.HostID = h.ID
	check.CheckedAt = time.Now().UTC()
	if err := db.Save(check); err != nil {
		return err
	}

	// Check If more than 100 checks exist and delete older ones
	count, err := db.Select(q.Eq("HostID", h.ID), q.Eq("CheckType", check.CheckType)).Count(&Check{})
	if err != nil {
		return err
	}

	if count > checkLimit {
		query := db.Select(
			q.Eq("HostID", h.ID),
			q.Eq("CheckType", check.CheckType),
		).OrderBy("CheckedAt").
			Limit(count - checkLimit)
		if err := query.Delete(&Check{}); err != nil {
			return err
		}
	}

	return nil
}

func (h *Host) addChecks(db *storm.DB) {
	var checks []Check

	if err := h.getHostCheckTypeQuery(db, CheckInternal).Find(&checks); err == nil {
		h.Checks.Internal = checks
	}

	if err := h.getHostCheckTypeQuery(db, CheckPublic).Find(&checks); err == nil {
		h.Checks.Public = checks
	}
}

func (h *Host) addLastChecks(db *storm.DB) {
	var check Check

	if err := h.getHostCheckTypeQuery(db, CheckInternal).First(&check); err == nil {
		h.LatestChecks.Internal = check
	}

	if err := h.getHostCheckTypeQuery(db, CheckPublic).First(&check); err == nil {
		h.LatestChecks.Public = check
	}
}

func (h *Host) getHostCheckTypeQuery(db *storm.DB, checkType CheckType) storm.Query {
	return db.Select(q.Eq("HostID", h.ID), q.Eq("CheckType", checkType)).OrderBy("CheckedAt").Reverse()
}
