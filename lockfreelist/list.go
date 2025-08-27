// Package lockfreelist 实现无锁双向链表，支持插件扩展和原子操作
package lockfreelist

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
)

// PluginCap 表示插件能力标志
type PluginCap int

const (
	CapConcurrentSafe PluginCap = 1 << iota
	CapIndexed
	CapLRU
)

// Plugin 定义链表行为扩展接口
type Plugin[K comparable, V any] interface {
	Attach(list *List[K, V])          // Attach 将插件加到链上
	OnInsert(node *Node[K, V])        // OnInsert 插入节点时触发
	OnRemove(node *Node[K, V])        // OnRemove 移除节点时触发
	OnUpdate(node *Node[K, V], old V) // OnUpdate 更新节点键/值时触发
	Capabilities() PluginCap          // Capabilities 插件能力标识
}

// element 是内部节点结构
type element[K comparable, V any] struct {
	key   K
	value atomic.Value
	next  atomic.Pointer[element[K, V]]
	prev  atomic.Pointer[element[K, V]]
	list  *List[K, V]
}

// Node 是外部使用的节点
type Node[K comparable, V any] struct {
	element *element[K, V]
}

// Key 返回节点键值
func (n *Node[K, V]) Key() K {
	if n == nil || n.element == nil {
		var zero K
		return zero
	}
	return n.element.key
}

// Value 返回节点值
func (n *Node[K, V]) Value() V {
	if n == nil || n.element == nil {
		var zero V
		return zero
	}
	val := n.element.value.Load()
	if val == nil {
		var zero V
		return zero
	}
	return val.(V)
}

// SetValue 设置节点值并触发更新事件
func (n *Node[K, V]) SetValue(v V) bool {
	if n == nil || n.element == nil || n.element.list == nil {
		return false
	}

	oldVal := n.Value()
	if reflect.DeepEqual(oldVal, v) {
		return false
	}

	n.element.value.Store(v)
	list := n.element.list

	// 如果链表有 LRU 能力，直接移动节点到头部
	if list.HasCapability(CapLRU) {
		list.MoveToFront(n)
	}

	// 触发插件和回调
	for _, p := range list.plugins {
		p.OnUpdate(n, oldVal)
	}
	for _, cb := range list.onUpdate {
		cb(n, oldVal)
	}
	return true
}

// Prev 返回前驱节点
func (n *Node[K, V]) Prev() *Node[K, V] {
	if n == nil || n.element == nil {
		return nil
	}
	prev := n.element.prev.Load()
	if prev == nil || prev == n.element.list.sentinel {
		return nil
	}
	return &Node[K, V]{element: prev}
}

// Next 返回后继节点
func (n *Node[K, V]) Next() *Node[K, V] {
	if n == nil || n.element == nil {
		return nil
	}
	next := n.element.next.Load()
	if next == nil || next == n.element.list.sentinel {
		return nil
	}
	return &Node[K, V]{element: next}
}

// IsValid 检查节点是否有效
func (n *Node[K, V]) IsValid() bool {
	return n != nil && n.element != nil && n.element.list != nil
}

// List 是无锁双向链表实现
type List[K comparable, V any] struct {
	head     atomic.Pointer[element[K, V]]
	tail     atomic.Pointer[element[K, V]]
	sentinel *element[K, V]
	len      atomic.Int32

	plugins  []Plugin[K, V]
	onInsert []func(*Node[K, V])
	onRemove []func(*Node[K, V])
	onUpdate []func(*Node[K, V], V)

	index     map[K]*element[K, V]
	indexLock sync.RWMutex
}

// ListBuilder 提供链式配置接口
type ListBuilder[K comparable, V any] struct {
	plugins []Plugin[K, V]
}

// NewBuilder 创建新构建器
func NewBuilder[K comparable, V any]() *ListBuilder[K, V] {
	return &ListBuilder[K, V]{}
}

// WithPlugin 添加插件
func (b *ListBuilder[K, V]) WithPlugin(p Plugin[K, V]) *ListBuilder[K, V] {
	b.plugins = append(b.plugins, p)
	return b
}

// Build 构建链表实例
func (b *ListBuilder[K, V]) Build() *List[K, V] {
	return New(b.plugins...)
}

// New 创建新链表实例
func New[K comparable, V any](plugins ...Plugin[K, V]) *List[K, V] {
	s := &element[K, V]{}
	l := &List[K, V]{
		sentinel: s,
		index:    make(map[K]*element[K, V]),
	}
	l.head.Store(nil)
	l.tail.Store(nil)

	for _, p := range plugins {
		p.Attach(l)
		l.plugins = append(l.plugins, p)
	}
	return l
}

// HasCapability 检查是否支持特定能力
func (l *List[K, V]) HasCapability(cap PluginCap) bool {
	for _, p := range l.plugins {
		if p.Capabilities()&cap != 0 {
			return true
		}
	}
	return false
}

// AddFront 在链表头部添加节点
func (l *List[K, V]) AddFront(key K, val V) *Node[K, V] {
	e := &element[K, V]{key: key, list: l}
	e.value.Store(val)

	if l.HasCapability(CapIndexed) {
		l.indexLock.Lock()
		l.index[key] = e
		l.indexLock.Unlock()
	}

	for {
		head := l.head.Load()
		e.next.Store(head)
		if head != nil {
			head.prev.Store(e)
		}
		if l.head.CompareAndSwap(head, e) {
			if l.tail.Load() == nil {
				l.tail.Store(e)
			}
			l.len.Add(1)
			n := &Node[K, V]{element: e}
			for _, p := range l.plugins {
				p.OnInsert(n)
			}
			for _, cb := range l.onInsert {
				cb(n)
			}
			return n
		}
	}
}

// AddBack 在链表尾部添加节点
func (l *List[K, V]) AddBack(key K, val V) *Node[K, V] {
	e := &element[K, V]{key: key, list: l}
	e.value.Store(val)

	if l.HasCapability(CapIndexed) {
		l.indexLock.Lock()
		l.index[key] = e
		l.indexLock.Unlock()
	}

	for {
		tail := l.tail.Load()
		e.prev.Store(tail)
		if tail != nil {
			tail.next.Store(e)
		}
		if l.tail.CompareAndSwap(tail, e) {
			if l.head.Load() == nil {
				l.head.Store(e)
			}
			l.len.Add(1)
			n := &Node[K, V]{element: e}
			for _, p := range l.plugins {
				p.OnInsert(n)
			}
			for _, cb := range l.onInsert {
				cb(n)
			}
			return n
		}
	}
}

// Remove 移除指定节点（乐观无锁并发）
func (l *List[K, V]) Remove(n *Node[K, V]) bool {
	if n == nil || n.element == nil || n.element.list != l {
		return false
	}

	elem := n.element

	for {
		prev := elem.prev.Load()
		next := elem.next.Load()

		// 尝试断开前驱和当前节点的连接
		if prev != nil {
			if !prev.next.CompareAndSwap(elem, next) {
				continue // 失败重试
			}
		} else {
			// 说明是头部节点
			if !l.head.CompareAndSwap(elem, next) {
				continue // 失败重试
			}
		}

		// 尝试断开后继和当前节点的连接
		if next != nil {
			if !next.prev.CompareAndSwap(elem, prev) {
				continue // 失败重试
			}
		} else {
			// 是尾部节点
			if !l.tail.CompareAndSwap(elem, prev) {
				continue // 失败重试
			}
		}

		break // 成功退出
	}

	if l.HasCapability(CapIndexed) {
		l.indexLock.Lock()
		delete(l.index, elem.key)
		l.indexLock.Unlock()
	}

	l.len.Add(-1)
	elem.list = nil

	for _, p := range l.plugins {
		p.OnRemove(n)
	}
	for _, cb := range l.onRemove {
		cb(n)
	}
	return true
}

// OnInsert 注册插入回调
func (l *List[K, V]) OnInsert(f func(*Node[K, V])) {
	l.onInsert = append(l.onInsert, f)
}

// OnRemove 注册移除回调
func (l *List[K, V]) OnRemove(f func(*Node[K, V])) {
	l.onRemove = append(l.onRemove, f)
}

// OnUpdate 注册更新回调
func (l *List[K, V]) OnUpdate(f func(*Node[K, V], V)) {
	l.onUpdate = append(l.onUpdate, f)
}

// Len 返回链表长度
func (l *List[K, V]) Len() int {
	return int(l.len.Load())
}

// IsEmpty 检查链表是否为空
func (l *List[K, V]) IsEmpty() bool {
	return l.Len() == 0
}

// FrontNode 获取链表头节点
func (l *List[K, V]) FrontNode() *Node[K, V] {
	e := l.head.Load()
	if e == nil || e == l.sentinel {
		return nil
	}
	return &Node[K, V]{element: e}
}

// BackNode 获取链表尾节点
func (l *List[K, V]) BackNode() *Node[K, V] {
	e := l.tail.Load()
	if e == nil || e == l.sentinel {
		return nil
	}
	return &Node[K, V]{element: e}
}

// Traversal 遍历链表(非线程安全)
func (l *List[K, V]) Traversal(visitor func(K, V) bool) {
	for e := l.head.Load(); e != nil && e != l.sentinel; e = e.next.Load() {
		val := e.value.Load()
		if val != nil {
			if !visitor(e.key, val.(V)) {
				break
			}
		}
	}
}

// Find 查找指定键的节点
func (l *List[K, V]) Find(key K) *Node[K, V] {
	if l.HasCapability(CapIndexed) {
		l.indexLock.RLock()
		e, ok := l.index[key]
		l.indexLock.RUnlock()
		if ok && e.list != nil {
			return &Node[K, V]{element: e}
		}
		return nil
	}

	for e := l.head.Load(); e != nil && e != l.sentinel; e = e.next.Load() {
		if e.key == key {
			return &Node[K, V]{element: e}
		}
	}
	return nil
}

// Contains 检查键是否存在
func (l *List[K, V]) Contains(key K) bool {
	return l.Find(key) != nil
}

// Clear 清空链表
func (l *List[K, V]) Clear() {
	for {
		e := l.head.Load()
		if e == nil || e == l.sentinel {
			break
		}
		l.Remove(&Node[K, V]{element: e})
	}
}

// MoveToFront 移动节点到链表头部
func (l *List[K, V]) MoveToFront(n *Node[K, V]) bool {
	if n == nil || n.element == nil || n.element.list != l {
		return false
	}

	// 如果已经是头部节点，不需要移动
	if l.head.Load() == n.element {
		return true
	}

	if l.Remove(n) {
		newNode := l.AddFront(n.Key(), n.Value())
		*n = *newNode
		return true
	}
	return false
}

// String 返回链表字符串表示
func (l *List[K, V]) String() string {
	var sb strings.Builder
	sb.WriteString("[")
	first := true
	l.Traversal(func(k K, v V) bool {
		if !first {
			sb.WriteString(", ")
		} else {
			first = false
		}
		sb.WriteString(fmt.Sprintf("%v:%v", k, v))
		return true
	})
	sb.WriteString("]")
	return sb.String()
}

// ConcurrentPlugin 并发安全插件
type ConcurrentPlugin[K comparable, V any] struct{}

func (p *ConcurrentPlugin[K, V]) Attach(l *List[K, V])    {}
func (p *ConcurrentPlugin[K, V]) OnInsert(*Node[K, V])    {}
func (p *ConcurrentPlugin[K, V]) OnRemove(*Node[K, V])    {}
func (p *ConcurrentPlugin[K, V]) OnUpdate(*Node[K, V], V) {}
func (p *ConcurrentPlugin[K, V]) Capabilities() PluginCap {
	return CapConcurrentSafe
}

// IndexPlugin 索引加速插件
type IndexPlugin[K comparable, V any] struct{}

// Attach 初始化索引插件，在链表上创建索引映射表
func (p *IndexPlugin[K, V]) Attach(l *List[K, V]) {
	l.index = make(map[K]*element[K, V])
}

// OnInsert 在插入节点时，将节点添加到索引映射表中
func (p *IndexPlugin[K, V]) OnInsert(n *Node[K, V]) {
	if n == nil || n.element == nil || n.element.list == nil {
		return
	}
	list := n.element.list
	list.indexLock.Lock()
	defer list.indexLock.Unlock()
	if list.index == nil {
		list.index = make(map[K]*element[K, V])
	}
	list.index[n.Key()] = n.element
}

// OnRemove 在移除节点时，从索引中删除该节点
func (p *IndexPlugin[K, V]) OnRemove(n *Node[K, V]) {
	if n == nil || n.element == nil || n.element.list == nil {
		return
	}
	list := n.element.list
	list.indexLock.Lock()
	defer list.indexLock.Unlock()
	if list.index != nil {
		delete(list.index, n.Key())
	}
}

// OnUpdate 索引插件不处理更新事件
func (p *IndexPlugin[K, V]) OnUpdate(*Node[K, V], V) {}

// Capabilities 声明该插件支持索引功能
func (p *IndexPlugin[K, V]) Capabilities() PluginCap {
	return CapIndexed
}

// LRUPlugin LRU 缓存插件：基于链表尾部驱逐超出容量的数据
type LRUPlugin[K comparable, V any] struct {
	capacity int
}

// NewLRUPlugin 创建新的LRU插件实例
func NewLRUPlugin[K comparable, V any](capacity int) *LRUPlugin[K, V] {
	return &LRUPlugin[K, V]{capacity: capacity}
}

// Attach LRU 插件不需要特殊初始化逻辑
func (p *LRUPlugin[K, V]) Attach(l *List[K, V]) {}

// OnInsert LRU 插件在插入节点时操作
func (p *LRUPlugin[K, V]) OnInsert(n *Node[K, V]) {
	if n == nil || n.element == nil {
		return
	}
	list := n.element.list

	// 检查节点是否仍然有效（未被其他协程移除）
	if !n.IsValid() {
		return
	}

	// 移动节点到头部
	if moved := list.MoveToFront(n); !moved {
		return
	}

	// 再次检查节点有效性
	if !n.IsValid() {
		return
	}

	// 连续驱逐直到不超出容量
	for list.Len() > p.capacity {
		tail := list.BackNode()
		// 检查尾部节点有效性
		if tail == nil || !tail.IsValid() {
			break
		}
		// 避免误删刚插入的自己
		if tail == n {
			break
		}

		// 尝试移除尾部节点
		if !list.Remove(tail) {
			break // 移除失败，可能是并发修改
		}
	}
}

// OnRemove LRU 插件对移除操作不做处理
func (p *LRUPlugin[K, V]) OnRemove(*Node[K, V]) {}

// OnUpdate 节点被更新（访问）时，移动到链表头部以表示最近使用
func (p *LRUPlugin[K, V]) OnUpdate(n *Node[K, V], old V) {
	// 不要在这里调用 MoveToFront，因为这会导致递归事件
}

// Capabilities 声明该插件支持 LRU 功能
func (p *LRUPlugin[K, V]) Capabilities() PluginCap {
	return CapLRU
}

// NewConcurrentList 创建并发安全链表
func NewConcurrentList[K comparable, V any]() *List[K, V] {
	return NewBuilder[K, V]().
		WithPlugin(&ConcurrentPlugin[K, V]{}).
		Build()
}

// NewIndexedList 创建带索引链表
func NewIndexedList[K comparable, V any]() *List[K, V] {
	return NewBuilder[K, V]().
		WithPlugin(&ConcurrentPlugin[K, V]{}).
		WithPlugin(&IndexPlugin[K, V]{}).
		Build()
}

// NewLRUCache 创建LRU缓存
func NewLRUCache[K comparable, V any](capacity int) *List[K, V] {
	return NewBuilder[K, V]().
		WithPlugin(&ConcurrentPlugin[K, V]{}).
		WithPlugin(&IndexPlugin[K, V]{}).
		WithPlugin(NewLRUPlugin[K, V](capacity)).
		Build()
}
