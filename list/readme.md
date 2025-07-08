# list

`list` 是一个支持插件机制的泛型双向链表实现，支持线程安全选项以及自定义事件回调，可用于构建灵活、可扩展的链表结构。

## 特性

- 泛型支持（Go 1.18+）
- 插件机制（`Plugin` 接口）
- 插入、删除、更新事件回调
- 并发安全可选
- 提供完整的遍历、查找、清空等操作

---

## 安装

你可以将该包拷贝到你的项目中使用，或将其封装为模块。

---

## 核心结构体

### Node[K, V]

链表中的一个节点，含键值对及前后指针：

```go
type Node[K comparable, V any] struct {
    Key   K
    Value V
}

func (n *Node[K, V]) Prev() *Node[K, V]
func (n *Node[K, V]) Next() *Node[K, V]
func (n *Node[K, V]) IsValid() bool
func (n *Node[K, V]) SetValue(val V) bool
```

### List[K, V]

带插件支持的双向链表：

```go
type List[K comparable, V any] struct {
    // ...内部字段省略...
}

func New[K comparable, V any](concurrent bool, plugins ...Plugin[K, V]) *List[K, V]
func (l *List[K, V]) AddBack(key K, value V) *Node[K, V]
func (l *List[K, V]) AddFront(key K, value V) *Node[K, V]
func (l *List[K, V]) InsertAfter(at *Node[K, V], key K, value V) *Node[K, V]
func (l *List[K, V]) Remove(n *Node[K, V]) bool
func (l *List[K, V]) Traversal(visitor func(key K, value V) bool)
func (l *List[K, V]) Find(pred func(key K, value V) bool) *Node[K, V]
func (l *List[K, V]) Contains(key K) bool
func (l *List[K, V]) Clear()
func (l *List[K, V]) Len() int
func (l *List[K, V]) FrontNode() *Node[K, V]
func (l *List[K, V]) BackNode() *Node[K, V]
func (l *List[K, V]) String() string
```

### Plugin[K, V]

插件接口，支持自定义索引、统计等功能：

```go
type Plugin[K comparable, V any] interface {
    Attach(list *List[K, V])
    OnInsert(node *Node[K, V])
    OnRemove(node *Node[K, V])
    OnUpdate(node *Node[K, V], oldVal V)
}
```

---

## 使用示例

```go
package main

import (
    "fmt"
    "./list"
)

// 自定义插件：统计插入、删除、更新次数
type statPlugin[K comparable, V any] struct {
    inserted, removed, updated int
}
func (s *statPlugin[K, V]) Attach(l *list.List[K, V]) {}
func (s *statPlugin[K, V]) OnInsert(n *list.Node[K, V]) { s.inserted++ }
func (s *statPlugin[K, V]) OnRemove(n *list.Node[K, V]) { s.removed++ }
func (s *statPlugin[K, V]) OnUpdate(n *list.Node[K, V], old V) { s.updated++ }

func main() {
    // 基本用法
    l := list.New[int, string](false)
    l.AddBack(1, "one")
    n2 := l.AddBack(2, "two")
    l.AddFront(0, "zero")
    fmt.Println("遍历:")
    l.Traversal(func(k int, v string) bool {
        fmt.Printf("%d: %s ", k, v)
        return true
    })
    fmt.Println("\n查找key=2:", l.Find(func(k, v int) bool { return k == 2 }))
    fmt.Println("包含key=1:", l.Contains(1))
    l.Remove(n2)
    fmt.Println("删除key=2后:", l.String())

    // 事件回调
    l.OnInsert(func(n *list.Node[int, string]) { fmt.Println("插入节点:", n.Key) })
    l.OnRemove(func(n *list.Node[int, string]) { fmt.Println("删除节点:", n.Key) })
    l.OnUpdate(func(n *list.Node[int, string], old string) { fmt.Printf("更新节点:%d %s->%s\n", n.Key, old, n.Value) })
    n := l.AddBack(3, "three")
    n.SetValue("THREE")
    l.Remove(n)

    // 自定义插件
    stat := &statPlugin[int, string]{}
    l2 := list.New[int, string](false, stat)
    l2.AddBack(1, "a")
    l2.AddBack(2, "b")
    l2.FrontNode().SetValue("A")
    l2.Remove(l2.BackNode())
    fmt.Printf("插件统计 插入:%d 删除:%d 更新:%d\n", stat.inserted, stat.removed, stat.updated)

    // 并发安全链表
    l3 := list.New[int, string](true)
    ch := make(chan struct{})
    for i := 0; i < 10; i++ {
        go func(k int) {
            l3.AddBack(k, fmt.Sprintf("v%d", k))
            ch <- struct{}{}
        }(i)
    }
    for i := 0; i < 10; i++ {
        <-ch
    }
    fmt.Println("并发链表长度:", l3.Len())

    // 清空链表
    l3.Clear()
    fmt.Println("清空后长度:", l3.Len())
}
```

---

## 说明

- 插件机制允许你为链表添加索引、统计、自动排序等功能，只需实现 `Plugin` 接口并传入 `New` 构造函数即可。
- 支持事件回调（OnInsert/OnRemove/OnUpdate），可用于业务逻辑钩子。
- 支持并发安全（传入 `true` 开启）。
- 支持链表遍历、查找、清空、字符串输出等常用操作。
