package list

import (
	"fmt"
	"sync"
	"testing"
)

type mockPlugin[K comparable, V any] struct {
	inserted []*Node[K, V]
	removed  []*Node[K, V]
	updated  []*Node[K, V]
	updates  []V
	attached bool
}

func (m *mockPlugin[K, V]) Attach(l *List[K, V]) {
	m.attached = true
}

func (m *mockPlugin[K, V]) OnInsert(n *Node[K, V]) {
	m.inserted = append(m.inserted, n)
}

func (m *mockPlugin[K, V]) OnRemove(n *Node[K, V]) {
	m.removed = append(m.removed, n)
}

func (m *mockPlugin[K, V]) OnUpdate(n *Node[K, V], old V) {
	m.updated = append(m.updated, n)
	m.updates = append(m.updates, old)
}

func TestListBasicOperations(t *testing.T) {
	l := New[int, string](false)
	l.OnInsert(func(n *Node[int, string]) { t.Log("Insert:", n.Key) })
	l.OnRemove(func(n *Node[int, string]) { t.Log("Remove:", n.Key) })
	l.OnUpdate(func(n *Node[int, string], old string) { t.Log("Update:", n.Key, old, "->", n.Value) })

	n1 := l.AddFront(1, "a")
	n2 := l.AddBack(2, "b")
	n3 := l.InsertAfter(n1, 3, "c")

	if l.Len() != 3 {
		t.Errorf("Expected length 3, got %d", l.Len())
	}

	if l.Front() != "a" || l.Back() != "b" {
		t.Errorf("Front or Back mismatch")
	}

	if s := l.String(); s != "[1:a 3:c 2:b]" {
		t.Errorf("String output mismatch: %s", s)
	}

	if !l.Contains(2) || !l.Contains(3) || !l.Contains(1) {
		t.Errorf("Contains failed")
	}

	if found := l.Find(func(k int, v string) bool { return k == 3 }); found != n3 {
		t.Errorf("Find failed")
	}

	ok := l.Remove(n2)
	if !ok || l.Len() != 2 {
		t.Errorf("Remove failed or length mismatch")
	}

	if !n1.SetValue("aa") || n1.Value != "aa" {
		t.Errorf("SetValue did not update")
	}
}

func TestListTraversal(t *testing.T) {
	l := New[int, string](false)
	l.AddBack(1, "a")
	l.AddBack(2, "b")
	l.AddBack(3, "c")

	sum := ""
	l.Traversal(func(k int, v string) bool {
		sum += v
		return true
	})
	if sum != "abc" {
		t.Errorf("Traversal sum mismatch: %s", sum)
	}
}

func TestPluginCallbacks(t *testing.T) {
	p := &mockPlugin[int, string]{}
	l := New[int, string](false, p)

	n1 := l.AddBack(1, "x")
	n2 := l.AddBack(2, "y")
	n2.SetValue("yy")
	l.Remove(n2)

	if !p.attached {
		t.Errorf("Plugin not attached")
	}

	if len(p.inserted) != 2 || len(p.updated) != 1 || len(p.removed) != 1 {
		t.Errorf("Plugin not triggered correctly")
	}
	if p.updates[0] != "y" {
		t.Errorf("Old value in OnUpdate incorrect")
	}

	if p.updated[0] != n2 {
		t.Errorf("Updated node incorrect")
	}
	if p.inserted[0] != n1 || p.inserted[1] != n2 {
		t.Errorf("Inserted nodes incorrect")
	}
	if p.removed[0] != n2 {
		t.Errorf("Removed node incorrect")
	}
}

func TestConcurrencySafeList(t *testing.T) {
	l := New[int, string](true)
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()
			l.AddBack(k, fmt.Sprintf("v%d", k))
		}(i)
	}
	wg.Wait()
	if l.Len() != 100 {
		t.Errorf("Expected 100 elements, got %d", l.Len())
	}
}

func TestInvalidNodeOperations(t *testing.T) {
	l := New[int, string](false)
	var n *Node[int, string]
	if n.Prev() != nil || n.Next() != nil || n.IsValid() {
		t.Errorf("Nil node operations failed")
	}

	n1 := l.AddBack(1, "a")
	fake := &Node[int, string]{Key: 2, Value: "b"}
	if l.InsertAfter(fake, 3, "c") != nil {
		t.Errorf("InsertAfter with fake node should return nil")
	}
	if l.Remove(fake) {
		t.Errorf("Remove with fake node should return false")
	}

	if n1.SetValue("a") {
		t.Errorf("SetValue with same value should return false")
	}
}

func TestClear(t *testing.T) {
	l := New[int, string](false)
	l.AddBack(1, "a")
	l.AddBack(2, "b")
	l.Clear()

	if l.Len() != 0 || l.FrontNode() != nil || l.BackNode() != nil {
		t.Errorf("Clear did not reset list properly")
	}

	// Check internal root connections restored
	if l.root.next != l.root || l.root.prev != l.root {
		t.Errorf("Root node not reset correctly")
	}
}
