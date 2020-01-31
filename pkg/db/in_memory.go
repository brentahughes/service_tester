package db

import (
	"errors"
	"sync"
	"time"
)

type InMemory struct {
	latest sync.Map
	data   sync.Map
}

func NewInMemoryStore() *InMemory {
	return &InMemory{}
}

func (db *InMemory) Create(host HostCheck) error {
	host.CheckTime = time.Now().UTC()

	checksI, ok := db.data.Load(host.Name)
	var checks []HostCheck
	if ok {
		checks = checksI.([]HostCheck)
	}
	checks = append(checks, host)
	db.data.Store(host.Name, checks)

	db.latest.Store(host.Name, host)
	return nil
}

func (db *InMemory) GetLastForAllHosts() ([]HostCheck, error) {
	var hosts []HostCheck
	db.latest.Range(func(key, value interface{}) bool {
		hosts = append(hosts, value.(HostCheck))
		return true
	})
	return hosts, nil
}

func (db *InMemory) GetLastByHost(host string) (*HostCheck, error) {
	val, ok := db.latest.Load(host)
	if !ok {
		return nil, errors.New("not found")
	}
	storedHost := val.(HostCheck)
	return &storedHost, nil
}

func (db *InMemory) GetAllForHost(host string) ([]HostCheck, error) {
	checks, ok := db.data.Load(host)
	if !ok {
		return nil, errors.New("not found")
	}
	return checks.([]HostCheck), nil
}
