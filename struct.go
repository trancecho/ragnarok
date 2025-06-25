package model

import "sync"

type IndexedList[K comparable, T any] struct {
	List     []T
	EntryMap map[K]T
	lock     sync.RWMutex
}

func (this *IndexedList[K, T]) Append(key K, item T) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.List = append(this.List, item)
	this.EntryMap[key] = item
}

func (this *IndexedList[K, T]) Get(key K) (T, bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	item, exists := this.EntryMap[key]
	if !exists {
		var zero T
		return zero, false
	}
	return item, true
}
