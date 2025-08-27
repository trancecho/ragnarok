package lockfreelist_test

import (
	"fmt"
	"github.com/trancecho/ragnarok/lockfreelist"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicOperations(t *testing.T) {
	list := lockfreelist.NewConcurrentList[string, int]()

	assert.True(t, list.IsEmpty())
	assert.Equal(t, 0, list.Len())

	n1 := list.AddFront("a", 1)
	n2 := list.AddBack("b", 2)

	assert.False(t, list.IsEmpty())
	assert.Equal(t, 2, list.Len())
	assert.Equal(t, "a", list.FrontNode().Key())
	assert.Equal(t, "b", list.BackNode().Key())
	assert.Equal(t, n2.Prev().Key(), "a")
	assert.Equal(t, n1.Next().Key(), "b")
	assert.Nil(t, n1.Prev())
	assert.Nil(t, n2.Next())

	assert.True(t, list.Contains("a"))
	assert.Equal(t, 1, list.Find("a").Value())

	assert.True(t, n1.SetValue(11))
	assert.Equal(t, 11, n1.Value())
	assert.False(t, n1.SetValue(11)) // no change

	assert.True(t, list.Remove(n1))
	assert.False(t, list.Contains("a"))
	assert.Equal(t, 1, list.Len())
	assert.False(t, list.Remove(n1)) // already removed

	list.Clear()
	assert.Equal(t, 0, list.Len())
	assert.Nil(t, list.FrontNode())
}

func TestIndexPlugin(t *testing.T) {
	list := lockfreelist.NewIndexedList[string, string]()

	// 测试正常添加和查找
	n := list.AddFront("k1", "v1")
	assert.Equal(t, "v1", list.Find("k1").Value())

	// 测试删除后查找
	list.Remove(n)
	assert.Nil(t, list.Find("k1"))

	// 测试重复删除
	assert.False(t, list.Remove(n))

	// 测试空列表操作
	emptyList := lockfreelist.NewIndexedList[int, int]()
	assert.Nil(t, emptyList.Find(123))
}

func TestIndexPluginConcurrency(t *testing.T) {
	list := lockfreelist.NewIndexedList[int, string]()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			list.AddFront(i, fmt.Sprintf("val%d", i))
			if list.Find(i) == nil {
				t.Errorf("Failed to find key %d", i)
			}
		}(i)
	}
	wg.Wait()
	assert.Equal(t, 100, list.Len())
}

func TestLRUPlugin(t *testing.T) {
	list := lockfreelist.NewLRUCache[string, string](2)

	list.AddBack("a", "1")
	list.AddBack("b", "2")
	assert.Equal(t, 2, list.Len())

	list.AddBack("c", "3")
	//fmt.Println(list.String())
	//fmt.Println(list.Len())
	//fmt.Println(list.Find("a"))
	//fmt.Println(list.Find("b"))
	//fmt.Println(list.Find("c"))
	assert.False(t, list.Contains("a"))
	assert.True(t, list.Contains("b"))
	assert.True(t, list.Contains("c"))

	nb := list.Find("b")
	nb.SetValue("2-updated") // triggers MoveToFront
	assert.Equal(t, "2-updated", list.FrontNode().Value())
}

func TestTraversalAndString(t *testing.T) {
	list := lockfreelist.NewIndexedList[int, string]()
	list.AddBack(1, "one")
	list.AddBack(2, "two")
	list.AddBack(3, "three")

	collected := make(map[int]string)
	list.Traversal(func(k int, v string) bool {
		collected[k] = v
		return true
	})
	assert.Len(t, collected, 3)

	str := list.String()
	assert.Contains(t, str, "1:one")
	assert.Contains(t, str, "2:two")
	assert.Contains(t, str, "3:three")
}

func TestMoveToFront(t *testing.T) {
	list := lockfreelist.NewIndexedList[int, string]()
	_ = list.AddBack(1, "one")
	n2 := list.AddBack(2, "two")
	assert.Equal(t, 1, list.FrontNode().Key())

	moved := list.MoveToFront(n2)
	assert.True(t, moved)
	assert.Equal(t, 2, list.FrontNode().Key())

	invalid := lockfreelist.Node[int, string]{}
	assert.False(t, list.MoveToFront(&invalid))
}

func TestCallbacks(t *testing.T) {
	list := lockfreelist.NewIndexedList[string, int]()
	inserted := false
	removed := false
	updated := false

	list.OnInsert(func(n *lockfreelist.Node[string, int]) {
		inserted = true
	})
	list.OnRemove(func(n *lockfreelist.Node[string, int]) {
		removed = true
	})
	list.OnUpdate(func(n *lockfreelist.Node[string, int], old int) {
		updated = true
	})

	n := list.AddFront("x", 100)
	assert.True(t, inserted)
	n.SetValue(200)
	assert.True(t, updated)
	list.Remove(n)
	assert.True(t, removed)
}

func TestConcurrency(t *testing.T) {
	list := lockfreelist.NewLRUCache[int, string](100)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			list.AddFront(i, fmt.Sprintf("val%d", i))
		}(i)
	}
	wg.Wait()

	assert.LessOrEqual(t, list.Len(), 50)
}

func TestNodeValidity(t *testing.T) {
	var zero *lockfreelist.Node[string, int]
	assert.False(t, zero.IsValid())
	assert.Equal(t, "", zero.Key())
	assert.Equal(t, 0, zero.Value())

	list := lockfreelist.NewConcurrentList[string, int]()
	n := list.AddBack("k", 1)
	assert.True(t, n.IsValid())
	list.Remove(n)
	assert.False(t, n.IsValid())

}
