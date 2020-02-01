package db

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
	"time"
)

const checkLimit = 1000

type InMemory struct {
	lock sync.Mutex
	data sync.Map
}

func NewInMemoryStore() *InMemory {
	return &InMemory{}
}

func (db *InMemory) Create(host HostCheck) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	host.ID = fmt.Sprintf("%x", sha256.Sum256([]byte(host.Name)))

	data := host.Checks[0]
	data.CheckTime = time.Now().UTC()
	data.CheckTimeMS = data.CheckTime.Unix() * 1000

	hostI, ok := db.data.Load(host.ID)
	if ok {
		host = hostI.(HostCheck)
	}

	// Prepend the check to the begginning
	if ok {
		host.Checks = append([]CheckData{data}, host.Checks...)
	} else {
		host.Checks = []CheckData{data}
	}

	// Make sure the list doesn't go over N checks
	if len(host.Checks) > checkLimit {
		host.Checks = host.Checks[:checkLimit]
	}

	db.data.Store(host.ID, host)
	return nil
}

func (db *InMemory) GetAllHosts() ([]HostCheck, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	var hosts []HostCheck
	db.data.Range(func(key, value interface{}) bool {
		host := value.(HostCheck)
		host.Latest = host.Checks[0]
		hosts = append(hosts, host)
		return true
	})
	return hosts, nil
}

func (db *InMemory) GetHost(id string) (*HostCheck, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	hostI, ok := db.data.Load(id)
	if !ok {
		return nil, errors.New("not found")
	}

	host := hostI.(HostCheck)
	return &host, nil
}
