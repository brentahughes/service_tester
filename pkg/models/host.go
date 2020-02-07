package models

import (
	"net"
	"time"

	"github.com/asdine/storm"
)

type Host struct {
	ID          int    `storm:"id,increment"`
	Hostname    string `storm:"unique"`
	InternalIP  net.IP `storm:"unique"`
	PublicIP    net.IP `storm:"unique"`
	FirstSeenAt time.Time
	LastSeenAt  time.Time
}

func GetHostByHostname(db *storm.DB, hostname string) (*Host, error) {
	var host Host
	if err := db.One("Hostname", hostname, &host); err != nil {
		return nil, err
	}

	return &host, nil
}

func (h *Host) Save(db *storm.DB) error {
	if h.FirstSeenAt.IsZero() {
		h.FirstSeenAt = time.Now().UTC()
	}

	h.LastSeenAt = time.Now().UTC()

	return db.Save(h)
}

func (h *Host) AddCheck(db *storm.DB, check *Check) error {
	check.HostID = h.ID
	check.CheckedAt = time.Now().UTC()
	return db.Save(check)
}
