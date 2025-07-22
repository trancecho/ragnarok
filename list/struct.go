package list

import (
	"sync"
)

type IndexedList[K comparable, T any] struct {
	Len      int
	List     []*Node[T]
	EntryMap map[K]*Node[T]
	lock     sync.RWMutex
}

type Node[T any] struct {
	Values []T
}

func (this *IndexedList[K, T]) Append(key K, item T) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	node, ok := this.EntryMap[key]
	if !ok {
		node = &Node[T]{Values: []T{item}}
		this.List = append(this.List, node)
	} else {
		node.Values = append(node.Values, item)
	}
	this.Len++
	this.EntryMap[key] = node
	return true
}

func (this *IndexedList[K, T]) Get(key K) (*Node[T], bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	node, exists := this.EntryMap[key]
	if !exists {
		return nil, false
	}
	return node, true
}

func (this *IndexedList[K, T]) Remove(key K) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	node, exists := this.EntryMap[key]
	if !exists {
		return false
	}
	delete(this.EntryMap, key)
	for i, n := range this.List {
		if n == node {
			this.List = append(this.List[:i], this.List[i+1:]...)
			break
		}
	}
	this.Len--
	return true
}
