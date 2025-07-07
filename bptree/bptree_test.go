package bptree

import (
	"math/rand"
	"sort"
	"testing"
)

func intCompare(a, b int) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}
	return 0
}

func TestBPTreeInsertAndFind(t *testing.T) {
	tree := NewBPTree[int, string](4, intCompare)

	tests := []struct {
		key   int
		value string
	}{
		{5, "five"},
		{3, "three"},
		{7, "seven"},
		{1, "one"},
		{9, "nine"},
		{2, "two"},
		{8, "eight"},
		{6, "six"},
		{4, "four"},
	}

	// 测试插入
	for _, tt := range tests {
		tree.Insert(tt.key, tt.value)
	}

	// 测试查找
	for _, tt := range tests {
		value, found := tree.Find(tt.key)
		if !found {
			t.Errorf("Find(%d) = not found, want %s", tt.key, tt.value)
		}
		if value != tt.value {
			t.Errorf("Find(%d) = %s, want %s", tt.key, value, tt.value)
		}
	}

	// 测试不存在的键
	if _, found := tree.Find(100); found {
		t.Error("Find(100) = found, want not found")
	}
}

func TestBPTreeDelete(t *testing.T) {
	tree := NewBPTree[int, string](4, intCompare)

	// 准备测试数据
	keys := []int{5, 3, 7, 1, 9, 2, 8, 6, 4}
	for _, key := range keys {
		tree.Insert(key, "value")
	}

	// 测试删除
	deleteOrder := []int{5, 3, 7, 1, 9, 2, 8, 6, 4}
	for _, key := range deleteOrder {
		if !tree.Delete(key) {
			t.Errorf("Delete(%d) = false, want true", key)
		}
		if _, found := tree.Find(key); found {
			t.Errorf("After Delete(%d), key still exists", key)
		}
	}

	// 测试删除不存在的键
	if tree.Delete(100) {
		t.Error("Delete(100) = true, want false")
	}
}

func TestBPTreeRangeQuery(t *testing.T) {
	tree := NewBPTree[int, string](4, intCompare)

	// 准备测试数据
	data := []struct {
		key   int
		value string
	}{
		{10, "ten"},
		{20, "twenty"},
		{30, "thirty"},
		{40, "forty"},
		{50, "fifty"},
		{60, "sixty"},
		{70, "seventy"},
		{80, "eighty"},
		{90, "ninety"},
	}
	for _, d := range data {
		tree.Insert(d.key, d.value)
	}

	tests := []struct {
		name     string
		start    int
		end      int
		expected []string
	}{
		{"完全包含", 25, 75, []string{"thirty", "forty", "fifty", "sixty", "seventy"}},
		{"下限边界", 10, 30, []string{"ten", "twenty", "thirty"}},
		{"上限边界", 70, 90, []string{"seventy", "eighty", "ninety"}},
		{"单元素", 50, 50, []string{"fifty"}},
		{"无结果", 95, 100, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := tree.RangeQuery(tt.start, tt.end)
			if len(results) != len(tt.expected) {
				t.Errorf("RangeQuery(%d, %d) returned %d results, want %d",
					tt.start, tt.end, len(results), len(tt.expected))
				return
			}
			for i := range results {
				if results[i] != tt.expected[i] {
					t.Errorf("RangeQuery(%d, %d) result[%d] = %s, want %s",
						tt.start, tt.end, i, results[i], tt.expected[i])
				}
			}
		})
	}
}

func TestBPTreeRandomOperations(t *testing.T) {
	tree := NewBPTree[int, int](4, intCompare)
	refMap := make(map[int]int)

	// 随机种子
	rand.Seed(42)

	// 执行1000次随机操作
	for i := 0; i < 1000; i++ {
		op := rand.Intn(3)
		key := rand.Intn(100)

		switch op {
		case 0: // 插入
			value := rand.Int()
			tree.Insert(key, value)
			refMap[key] = value
		case 1: // 查找
			value, found := tree.Find(key)
			refValue, refFound := refMap[key]

			if found != refFound {
				t.Errorf("Find(%d) consistency failed: tree=%v, map=%v", key, found, refFound)
			}
			if found && value != refValue {
				t.Errorf("Find(%d) value mismatch: tree=%v, map=%v", key, value, refValue)
			}
		case 2: // 删除
			deleted := tree.Delete(key)
			_, refExists := refMap[key]

			if deleted != refExists {
				t.Errorf("Delete(%d) consistency failed: tree=%v, map=%v", key, deleted, refExists)
			}
			if deleted {
				delete(refMap, key)
			}
		}
	}

	// 最终一致性检查
	for key, refValue := range refMap {
		value, found := tree.Find(key)
		if !found {
			t.Errorf("Final check: key %d not found in tree but exists in map", key)
		}
		if value != refValue {
			t.Errorf("Final check: value mismatch for key %d: tree=%v, map=%v", key, value, refValue)
		}
	}
}

func TestBPTreeEmpty(t *testing.T) {
	tree := NewBPTree[int, string](4, intCompare)

	// 测试空树查找
	if _, found := tree.Find(1); found {
		t.Error("Find on empty tree returned true, want false")
	}

	// 测试空树删除
	if tree.Delete(1) {
		t.Error("Delete on empty tree returned true, want false")
	}

	// 测试空树范围查询
	if results := tree.RangeQuery(1, 10); len(results) != 0 {
		t.Errorf("RangeQuery on empty tree returned %d results, want 0", len(results))
	}
}

func TestBPTreeOrder(t *testing.T) {
	tests := []struct {
		order int
		valid bool
	}{
		{3, true},
		{4, true},
		{10, true},
		{2, false},
		{0, false},
		{-1, false},
	}

	for _, tt := range tests {
		tree := NewBPTree[int, int](tt.order, intCompare)
		if tt.valid && tree.order != tt.order {
			t.Errorf("With order=%d, got tree.order=%d, want %d", tt.order, tree.order, tt.order)
		}
		if !tt.valid && tree.order != defaultOrder {
			t.Errorf("With invalid order=%d, got tree.order=%d, want default %d",
				tt.order, tree.order, defaultOrder)
		}
	}
}

func TestBPTreeLeafLink(t *testing.T) {
	tree := NewBPTree[int, int](4, intCompare)
	keys := []int{5, 3, 7, 1, 9, 2, 8, 6, 4}
	for _, key := range keys {
		tree.Insert(key, key*10)
	}

	// 检查叶子节点链表
	var leafKeys []int
	current := tree.leafHeader
	for current != nil {
		leafKeys = append(leafKeys, current.keys...)
		current = current.next
	}

	// 链表中的键应该是排序的
	if !sort.IntsAreSorted(leafKeys) {
		t.Errorf("Leaf keys are not sorted: %v", leafKeys)
	}

	// 应该包含所有插入的键
	if len(leafKeys) != len(keys) {
		t.Errorf("Leaf keys count mismatch: got %d, want %d", len(leafKeys), len(keys))
	}

	// 检查每个键是否存在
	keySet := make(map[int]bool)
	for _, key := range leafKeys {
		keySet[key] = true
	}
	for _, key := range keys {
		if !keySet[key] {
			t.Errorf("Key %d missing from leaf links", key)
		}
	}
}

func TestBPTreePrint(t *testing.T) {
	// 主要用于调试，不进行实际测试
	tree := NewBPTree[int, string](4, intCompare)
	for i := 0; i < 20; i++ {
		tree.Insert(i, "value")
	}
	tree.Print() // 检查输出是否正确
}
