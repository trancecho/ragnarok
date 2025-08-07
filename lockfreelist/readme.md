# lockfreelist

`lockfreelist` 是一个支持插件机制的无锁（lock-free）双向链表实现，支持原子操作、并发安全、索引、LRU缓存等扩展能力。

## 特性

- 无锁并发安全，基于 atomic.Pointer 和 atomic.Value
- 插件机制（实现 Plugin 接口即可扩展功能）
- 支持索引、LRU、事件回调等
- 泛型支持（Go 1.18+）
- 提供遍历、查找、清空等常用操作

---

## 结构体定义

### Node[K, V]

链表节点，外部操作接口：

```go
type Node[K comparable, V any] struct {
    // ...内部字段省略...
}
func (n *Node[K, V]) Key() K
func (n *Node[K, V]) Value() V
func (n *Node[K, V]) SetValue(v V) bool
func (n *Node[K, V]) Prev() *Node[K, V]
func (n *Node[K, V]) Next() *Node[K, V]
func (n *Node[K, V]) IsValid() bool
```

### List[K, V]

无锁双向链表：

```go
type List[K comparable, V any] struct {
    // ...内部字段省略...
}
func New[K comparable, V any](plugins ...Plugin[K, V]) *List[K, V]
func (l *List[K, V]) AddFront(key K, val V) *Node[K, V]
func (l *List[K, V]) AddBack(key K, val V) *Node[K, V]
func (l *List[K, V]) Remove(n *Node[K, V]) bool
func (l *List[K, V]) Len() int
func (l *List[K, V]) IsEmpty() bool
func (l *List[K, V]) FrontNode() *Node[K, V]
func (l *List[K, V]) BackNode() *Node[K, V]
func (l *List[K, V]) Traversal(visitor func(K, V) bool)
func (l *List[K, V]) Find(key K) *Node[K, V]
func (l *List[K, V]) Contains(key K) bool
func (l *List[K, V]) Clear()
func (l *List[K, V]) String() string
```

### Plugin[K, V]

插件接口，支持自定义索引、LRU、并发等能力：

```go
type Plugin[K comparable, V any] interface {
    Attach(list *List[K, V])
    OnInsert(node *Node[K, V])
    OnRemove(node *Node[K, V])
    OnUpdate(node *Node[K, V], old V)
    Capabilities() PluginCap
}
```

---

## 使用示例

```go
package main

import (
    "fmt"
    "lockfreelist"
)

func main() {
    // 创建无锁链表
    l := lockfreelist.New[int, string]()
    l.AddBack(1, "one")
    n2 := l.AddBack(2, "two")
    l.AddFront(0, "zero")
    fmt.Println("遍历:")
    l.Traversal(func(k int, v string) bool {
        fmt.Printf("%d: %s ", k, v)
        return true
    })
    fmt.Println("\n查找key=2:", l.Find(2))
    fmt.Println("包含key=1:", l.Contains(1))
    l.Remove(n2)
    fmt.Println("删除key=2后:", l.String())

    // 支持事件回调
    l.OnInsert(func(n *lockfreelist.Node[int, string]) { fmt.Println("插入节点:", n.Key()) })
    l.OnRemove(func(n *lockfreelist.Node[int, string]) { fmt.Println("删除节点:", n.Key()) })
    l.OnUpdate(func(n *lockfreelist.Node[int, string], old string) { fmt.Printf("更新节点:%d %s->%s\n", n.Key(), old, n.Value()) })
    n := l.AddBack(3, "three")
    n.SetValue("THREE")
    l.Remove(n)

    // 支持插件扩展（如索引、LRU）
    l2 := lockfreelist.NewIndexedList[int, string]()
    l2.AddBack(1, "a")
    l2.AddBack(2, "b")
    fmt.Println("带索引链表查找key=2:", l2.Find(2))

    lru := lockfreelist.NewLRUCache[int, string](2)
    lru.AddBack(1, "A")
    lru.AddBack(2, "B")
    lru.AddBack(3, "C") // 超出容量自动驱逐最久未用节点
    fmt.Println("LRU缓存内容:", lru.String())
}
```

---

## 说明

- 插件机制允许你为链表添加并发安全、索引、LRU等功能，只需实现 `Plugin` 接口并传入 `New` 或 `NewBuilder` 构造函数即可。
- 支持事件回调（OnInsert/OnRemove/OnUpdate），可用于业务逻辑钩子。
- 支持链表遍历、查找、清空、字符串输出等常用操作。

