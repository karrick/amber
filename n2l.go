package main

import (
	"sync"
)

type lockUrnDb struct {
	db   map[string][]string
	lock sync.RWMutex
}

func (this *lockUrnDb) get(key string) ([]string, bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	v, ok := this.db[key]
	return v, ok
}

func (this *lockUrnDb) append(key, value string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.db == nil {
		this.db = make(map[string][]string)
	}
	this.db[key] = append(this.db[key], value)
}

func (this *lockUrnDb) keys() (keys []string) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	keys = make([]string, 0, len(this.db))
	for k, _ := range this.db {
		keys = append(keys, k)
	}
	return
}

// TODO: need delete, but don't know whether to delete all values, or a given value
