// Package list 提供一个带插件机制的通用双向链表实现。
package list

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// Plugin 定义了链表插件的接口。
type Plugin[K comparable, V any] interface {
	Attach(list *List[K, V])             // 插件附加到链表时调用
	OnInsert(node *Node[K, V])           // 插入节点时调用
	OnRemove(node *Node[K, V])           // 删除节点时调用
	OnUpdate(node *Node[K, V], oldVal V) // 节点更新时调用
}

// Node 表示链表中的一个节点。
type Node[K comparable, V any] struct {
	prev  *Node[K, V] // 前一个节点
	next  *Node[K, V] // 后一个节点
	Key   K           // 节点的键
	Value V           // 节点的值
	list  *List[K, V] // 所属链表
}

// Prev 返回前一个节点，若为首节点则返回 nil。
func (n *Node[K, V]) Prev() *Node[K, V] {
	if n == nil || n.list == nil || n.prev == n.list.root {
		return nil
	}
	return n.prev
}

// Next 返回后一个节点，若为尾节点则返回 nil。
func (n *Node[K, V]) Next() *Node[K, V] {
	if n == nil || n.list == nil || n.next == n.list.root {
		return nil
	}
	return n.next
}

// IsValid 判断该节点是否有效并属于某个链表。
func (n *Node[K, V]) IsValid() bool {
	return n != nil && n.list != nil
}

// SetValue 设置节点的新值，并触发更新回调。
func (n *Node[K, V]) SetValue(val V) bool {
	if n.list == nil {
		return false
	}
	if reflect.DeepEqual(n.Value, val) {
		return false
	}
	old := n.Value
	n.Value = val
	for _, pl := range n.list.plugins {
		pl.OnUpdate(n, old)
	}
	for _, cb := range n.list.onUpdate {
		cb(n, old)
	}
	return true
}

// List 表示一个带插件支持的双向链表。
type List[K comparable, V any] struct {
	root       *Node[K, V]            // 哨兵节点，形成循环链表结构
	len        int                    // 链表长度
	mu         sync.RWMutex           // 并发读写锁
	concurrent bool                   // 是否启用并发安全
	plugins    []Plugin[K, V]         // 插件列表
	onInsert   []func(*Node[K, V])    // 插入回调列表
	onRemove   []func(*Node[K, V])    // 删除回调列表
	onUpdate   []func(*Node[K, V], V) // 更新回调列表
}

func New[K comparable, V any](concurrent bool, plugins ...Plugin[K, V]) *List[K, V] {
	root := &Node[K, V]{}
	root.prev, root.next = root, root

	l := &List[K, V]{
		root:       root,
		len:        0,
		concurrent: concurrent,
		plugins:    plugins,
	}

	for _, p := range plugins {
		p.Attach(l)
	}

	return l
}

// Clear 删除所有节点，但保留根节点和插件设置
func (l *List[K, V]) Clear() {
	if l.concurrent {
		l.mu.Lock()
		defer l.mu.Unlock()
	}
	for n := l.FrontNode(); n != nil; {
		next := n.Next()
		n.list = nil
		n.prev, n.next = nil, nil
		n = next
	}
	l.root.prev, l.root.next = l.root, l.root
	l.len = 0
}

// OnInsert 注册插入事件的回调函数。
func (l *List[K, V]) OnInsert(f func(*Node[K, V])) {
	l.onInsert = append(l.onInsert, f)
}

// OnRemove 注册删除事件的回调函数。
func (l *List[K, V]) OnRemove(f func(*Node[K, V])) {
	l.onRemove = append(l.onRemove, f)
}

// OnUpdate 注册更新事件的回调函数。
func (l *List[K, V]) OnUpdate(f func(*Node[K, V], V)) {
	l.onUpdate = append(l.onUpdate, f)
}

// AddBack 在链表尾部添加一个新节点。
func (l *List[K, V]) AddBack(key K, value V) *Node[K, V] {
	if l.concurrent {
		l.mu.Lock()
		defer l.mu.Unlock()
	}
	n := &Node[K, V]{Key: key, Value: value}
	l.insertNode(l.root.prev, n)
	for _, pl := range l.plugins {
		pl.OnInsert(n)
	}
	for _, cb := range l.onInsert {
		cb(n)
	}
	return n
}

// AddFront 在链表头部添加一个新节点。
func (l *List[K, V]) AddFront(key K, value V) *Node[K, V] {
	if l.concurrent {
		l.mu.Lock()
		defer l.mu.Unlock()
	}
	n := &Node[K, V]{Key: key, Value: value}
	l.insertNode(l.root, n)
	for _, pl := range l.plugins {
		pl.OnInsert(n)
	}
	for _, cb := range l.onInsert {
		cb(n)
	}
	return n
}

// InsertAfter 在指定节点之后插入新节点。
func (l *List[K, V]) InsertAfter(at *Node[K, V], key K, value V) *Node[K, V] {
	if at == nil || at.list != l {
		return nil
	}
	if l.concurrent {
		l.mu.Lock()
		defer l.mu.Unlock()
	}
	n := &Node[K, V]{Key: key, Value: value}
	l.insertNode(at, n)
	for _, pl := range l.plugins {
		pl.OnInsert(n)
	}
	for _, cb := range l.onInsert {
		cb(n)
	}
	return n
}

// Remove 从链表中删除指定节点。
func (l *List[K, V]) Remove(n *Node[K, V]) bool {
	if n == nil || n.list != l {
		return false
	}
	if l.concurrent {
		l.mu.Lock()
		defer l.mu.Unlock()
	}
	n.prev.next = n.next
	n.next.prev = n.prev
	n.list = nil
	l.len--
	for _, pl := range l.plugins {
		pl.OnRemove(n)
	}
	for _, cb := range l.onRemove {
		cb(n)
	}
	n.prev, n.next = nil, nil
	return true
}

// insertNode 将节点插入到指定节点之后。
func (l *List[K, V]) insertNode(at, n *Node[K, V]) {
	n.next = at.next
	at.next.prev = n
	at.next = n
	n.prev = at
	n.list = l
	l.len++
}

// Len 返回链表中节点数量。
func (l *List[K, V]) Len() int {
	return l.len
}

// FrontNode 返回第一个有效节点。
func (l *List[K, V]) FrontNode() *Node[K, V] {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// BackNode 返回最后一个有效节点。
func (l *List[K, V]) BackNode() *Node[K, V] {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// Front 返回第一个节点的值（为空时会 panic）。
func (l *List[K, V]) Front() V {
	if l.Len() == 0 {
		panic("list is empty")
	}
	return l.root.next.Value
}

// Back 返回最后一个节点的值（为空时会 panic）。
func (l *List[K, V]) Back() V {
	if l.Len() == 0 {
		panic("list is empty")
	}
	return l.root.prev.Value
}

// Traversal 遍历链表中所有节点，visitor 返回 false 时中断遍历。
func (l *List[K, V]) Traversal(visitor func(key K, value V) bool) {
	for n := l.FrontNode(); n != nil; n = n.Next() {
		if !visitor(n.Key, n.Value) {
			break
		}
	}
}

// Find 返回第一个满足 predicate 的节点
func (l *List[K, V]) Find(pred func(key K, value V) bool) *Node[K, V] {
	for n := l.FrontNode(); n != nil; n = n.Next() {
		if pred(n.Key, n.Value) {
			return n
		}
	}
	return nil
}

// Contains 判断是否存在某个键
func (l *List[K, V]) Contains(key K) bool {
	return l.Find(func(k K, _ V) bool { return k == key }) != nil
}

// String 返回链表的字符串表示形式。
func (l *List[K, V]) String() string {
	out := "["
	for n := l.FrontNode(); n != nil; n = n.Next() {
		out += fmt.Sprintf("%v:%v ", n.Key, n.Value)
	}
	out = strings.TrimSpace(out) + "]"
	return out
}
