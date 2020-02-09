package models

import (
	"time"

	"github.com/asdine/storm/q"
	"github.com/asdine/storm/v3"
)

type Host struct {
	ID          int    `storm:"id,increment"`
	Hostname    string `storm:"unique"`
	InternalIP  string `storm:"unique"`
	PublicIP    string `storm:"unique"`
	Port        int
	FirstSeenAt time.Time
	LastSeenAt  time.Time

	LatestChecks LatestChecks `json:"-"`
	Checks       Checks       `json:"-"`
}

type LatestChecks struct {
	Internal Check
	Public   Check
}

type Checks struct {
	Internal []Check
	Public   []Check
}

func GetHostByHostname(db *storm.DB, hostname string) (*Host, error) {
	var host Host
	if err := db.One("Hostname", hostname, &host); err != nil {
		return nil, err
	}

	return &host, nil
}

func GetHostByID(db *storm.DB, id int) (*Host, error) {
	var host Host
	if err := db.One("ID", id, &host); err != nil {
		return nil, err
	}

	h := &host
	h.addChecks(db)

	return h, nil
}

func GetHostByIP(db *storm.DB, ip string) (*Host, error) {
	var host Host
	if err := db.Select(q.Or(q.Eq("InternalIP", ip), q.Eq("PublicIP", ip))).First(&host); err != nil {
		return nil, err
	}

	h := &host
	h.addChecks(db)

	return h, nil
}

func GetAllHosts(db *storm.DB) ([]Host, error) {
	currentHost, err := GetCurrentHost(db)
	if err != nil {
		return nil, err
	}

	var hosts []Host
	if err := db.Select(q.Not(q.Eq("Hostname", currentHost.Hostname))).OrderBy("Hostname").Find(&hosts); err != nil {
		return nil, err
	}

	for i, host := range hosts {
		host.addLastChecks(db)
		hosts[i] = host
	}

	return hosts, nil
}

func GetHostsWithPublicIPs(db *storm.DB) ([]Host, error) {
	currentHost, err := GetCurrentHost(db)
	if err != nil {
		return nil, err
	}

	var hosts []Host
	if err := db.Select(
		q.Not(
			q.Eq("PublicIP", ""),
			q.Eq("Hostname", currentHost.Hostname),
		),
	).OrderBy("Hostname").Find(&hosts); err != nil {
		return nil, err
	}
	return hosts, nil
}

func (h *Host) Save(db *storm.DB) error {
	h.LastSeenAt = time.Now().UTC()
	if h.FirstSeenAt.IsZero() {
		h.FirstSeenAt = time.Now().UTC()
	}

	return db.Save(h)
}

func (h *Host) Delete(db *storm.DB) error {
	// Delete all the checks for the host
	if err := db.Select(q.Eq("HostID", h.ID)).Delete(&Check{}); err != nil {
		return err
	}

	// Delete the host
	return db.DeleteStruct(h)
}
