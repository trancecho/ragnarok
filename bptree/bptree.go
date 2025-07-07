package bptree

import (
	"fmt"
)

const (
	defaultOrder = 4
)

// Comparable 定义了比较函数的类型
type Comparable[K any] func(a, b K) int

// BPTree 是 B+ 树的实现
type BPTree[K any, V any] struct {
	root       node[K, V]
	order      int
	compare    Comparable[K]
	minKeys    int
	maxKeys    int
	leafHeader *leafNode[K, V]
}

// node 接口定义了B+树内部节点和叶子节点的共同行为
type node[K any, V any] interface {
	isLeaf() bool
	find(key K, compare Comparable[K]) (V, bool)
	insert(key K, value V, compare Comparable[K], tree *BPTree[K, V]) (K, node[K, V], bool)
	delete(key K, compare Comparable[K], tree *BPTree[K, V]) bool
	smallestKey() K
	size() int
	print(indent string)
}

// internalNode 表示B+树的内部节点
type internalNode[K any, V any] struct {
	keys     []K
	children []node[K, V]
}

// leafNode 表示B+树的叶子节点
type leafNode[K any, V any] struct {
	keys   []K
	values []V
	next   *leafNode[K, V]
}

// NewBPTree 创建一个新的B+树实例
// 参数:
//
//	order - 树的阶数(每个节点最多包含2*order-1个键)
//	compare - 键比较函数
//
// 返回值:
//
//	*BPTree[K, V] - 初始化好的B+树指针
func NewBPTree[K any, V any](order int, compare Comparable[K]) *BPTree[K, V] {
	if order < 3 {
		order = defaultOrder
	}

	return &BPTree[K, V]{
		order:   order,
		compare: compare,
		minKeys: order - 1,
		maxKeys: 2*order - 1,
	}
}

// isLeaf 内部节点实现 - 总是返回false
func (n *internalNode[K, V]) isLeaf() bool {
	return false
}

// isLeaf 叶子节点实现 - 总是返回true
func (n *leafNode[K, V]) isLeaf() bool {
	return true
}

// Insert 向B+树中插入键值对
// 参数:
//
//	key - 要插入的键
//	value - 要插入的值
func (t *BPTree[K, V]) Insert(key K, value V) {
	if t.root == nil {
		t.root = &leafNode[K, V]{
			keys:   make([]K, 0, t.maxKeys+1),
			values: make([]V, 0, t.maxKeys+1),
		}
		t.leafHeader = t.root.(*leafNode[K, V])
	}

	newKey, newNode, split := t.root.insert(key, value, t.compare, t)
	if split {
		newRoot := &internalNode[K, V]{
			keys:     make([]K, 0, t.maxKeys+1),
			children: make([]node[K, V], 0, t.maxKeys+2),
		}
		newRoot.keys = append(newRoot.keys, newKey)
		newRoot.children = append(newRoot.children, t.root, newNode)
		t.root = newRoot
	}
}

// Find 在B+树中查找指定键
// 参数:
//
//	key - 要查找的键
//
// 返回值:
//
//	V - 找到的值
//	bool - 是否找到
func (t *BPTree[K, V]) Find(key K) (V, bool) {
	if t.root == nil {
		var zero V
		return zero, false
	}
	return t.root.find(key, t.compare)
}

// Delete 从B+树中删除指定键
// 参数:
//
//	key - 要删除的键
//
// 返回值:
//
//	bool - 是否删除成功
func (t *BPTree[K, V]) Delete(key K) bool {
	if t.root == nil {
		return false
	}

	deleted := t.root.delete(key, t.compare, t)
	if !t.root.isLeaf() && t.root.(*internalNode[K, V]).size() == 1 {
		t.root = t.root.(*internalNode[K, V]).children[0]
	}

	return deleted
}

// RangeQuery 范围查询，返回键在[start,end]闭区间内的所有键值对
// 参数:
//
//	start - 范围起始键
//	end - 范围结束键
//
// 返回值:
//
//	[]V - 符合范围条件的值切片
func (t *BPTree[K, V]) RangeQuery(start, end K) []V {
	var results []V
	if t.root == nil {
		return results
	}

	// Find the starting leaf node
	current := t.findLeaf(start)
	for current != nil {
		for i, key := range current.keys {
			cmpStart := t.compare(key, start)
			cmpEnd := t.compare(key, end)

			if cmpStart >= 0 && cmpEnd <= 0 {
				results = append(results, current.values[i])
			} else if cmpEnd > 0 {
				return results
			}
		}
		current = current.next
	}

	return results
}

// findLeaf 查找包含指定键的叶子节点
// 参数:
//
//	key - 要查找的键
//
// 返回值:
//
//	*leafNode - 找到的叶子节点指针
func (t *BPTree[K, V]) findLeaf(key K) *leafNode[K, V] {
	if t.root == nil {
		return nil
	}

	current := t.root
	for !current.isLeaf() {
		internal := current.(*internalNode[K, V])
		idx := t.findInsertPosition(internal.keys, key)
		current = internal.children[idx]
	}

	return current.(*leafNode[K, V])
}

// findInsertPosition 二分查找键在切片中的插入位置
// 参数:
//
//	keys - 已排序的键切片
//	key - 要查找的键
//
// 返回值:
//
//	int - 键应该插入的位置索引
func (t *BPTree[K, V]) findInsertPosition(keys []K, key K) int {
	low, high := 0, len(keys)
	for low < high {
		mid := (low + high) / 2
		cmp := t.compare(key, keys[mid])
		if cmp < 0 {
			high = mid
		} else {
			low = mid + 1
		}
	}
	return low
}

// Print 打印B+树结构(用于调试)
func (t *BPTree[K, V]) Print() {
	if t.root == nil {
		fmt.Println("Empty tree")
		return
	}
	t.root.print("")
}

// ========== 内部节点方法实现 ==========

func (n *internalNode[K, V]) find(key K, compare Comparable[K]) (V, bool) {
	idx := 0
	for i, k := range n.keys {
		if compare(key, k) < 0 {
			break
		}
		idx = i + 1
	}
	return n.children[idx].find(key, compare)
}

func (n *internalNode[K, V]) insert(key K, value V, compare Comparable[K], tree *BPTree[K, V]) (K, node[K, V], bool) {
	idx := 0
	for i, k := range n.keys {
		if compare(key, k) < 0 {
			break
		}
		idx = i + 1
	}

	newKey, newNode, split := n.children[idx].insert(key, value, compare, tree)
	if !split {
		return *new(K), nil, false
	}

	// 插入新键和子节点
	n.keys = insertAt(n.keys, idx, newKey)
	n.children = insertAtNode(n.children, idx+1, newNode)

	if len(n.keys) <= tree.maxKeys {
		return *new(K), nil, false
	}

	// 分裂内部节点
	splitIdx := len(n.keys) / 2
	promotedKey := n.keys[splitIdx]

	right := &internalNode[K, V]{
		keys:     make([]K, len(n.keys[splitIdx+1:])),
		children: make([]node[K, V], len(n.children[splitIdx+1:])),
	}
	copy(right.keys, n.keys[splitIdx+1:])
	copy(right.children, n.children[splitIdx+1:])

	// 更新原节点
	n.keys = n.keys[:splitIdx]
	n.children = n.children[:splitIdx+1]

	return promotedKey, right, true
}

func (n *internalNode[K, V]) delete(key K, compare Comparable[K], tree *BPTree[K, V]) bool {
	idx := 0
	for i, k := range n.keys {
		if compare(key, k) < 0 {
			break
		}
		idx = i + 1
	}

	deleted := n.children[idx].delete(key, compare, tree)
	if !deleted {
		return false
	}

	// 处理节点下溢
	if n.children[idx].size() < tree.minKeys {
		tree.handleUnderflow(n, idx)
	}

	return true
}

func (n *internalNode[K, V]) smallestKey() K {
	return n.children[0].smallestKey()
}

func (n *internalNode[K, V]) size() int {
	return len(n.children)
}

func (n *internalNode[K, V]) print(indent string) {
	fmt.Printf("%sInternalNode: %v\n", indent, n.keys)
	for _, child := range n.children {
		child.print(indent + "  ")
	}
}

// ========== 叶子节点方法实现 ==========

func (n *leafNode[K, V]) find(key K, compare Comparable[K]) (V, bool) {
	for i, k := range n.keys {
		cmp := compare(key, k)
		if cmp == 0 {
			return n.values[i], true
		}
		if cmp < 0 {
			break
		}

	}
	var zero V
	return zero, false
}

func (n *leafNode[K, V]) insert(key K, value V, compare Comparable[K], tree *BPTree[K, V]) (K, node[K, V], bool) {
	idx := 0
	for i, k := range n.keys {
		cmp := compare(key, k)
		if cmp == 0 {
			// 更新现有值
			n.values[i] = value
			return *new(K), nil, false
		}
		if cmp < 0 {
			break
		}
		idx = i + 1
	}

	// 插入新键值对
	n.keys = insertAt(n.keys, idx, key)
	n.values = insertAtValue(n.values, idx, value)

	if len(n.keys) <= tree.maxKeys {
		return *new(K), nil, false
	}

	// 分裂叶子节点
	splitIdx := len(n.keys) / 2
	right := &leafNode[K, V]{
		keys:   make([]K, len(n.keys[splitIdx:])),
		values: make([]V, len(n.values[splitIdx:])),
		next:   n.next,
	}
	copy(right.keys, n.keys[splitIdx:])
	copy(right.values, n.values[splitIdx:])

	// 更新原节点
	n.keys = n.keys[:splitIdx]
	n.values = n.values[:splitIdx]
	n.next = right

	return right.keys[0], right, true
}

func (n *leafNode[K, V]) delete(key K, compare Comparable[K], tree *BPTree[K, V]) bool {
	idx := -1
	for i, k := range n.keys {
		if compare(key, k) == 0 {
			idx = i
			break
		}
	}

	if idx == -1 {
		return false
	}

	// 删除键值对
	n.keys = append(n.keys[:idx], n.keys[idx+1:]...)
	n.values = append(n.values[:idx], n.values[idx+1:]...)
	return true
}

func (n *leafNode[K, V]) smallestKey() K {
	if len(n.keys) == 0 {
		var zero K
		return zero
	}
	return n.keys[0]
}

func (n *leafNode[K, V]) size() int {
	return len(n.keys)
}

func (n *leafNode[K, V]) print(indent string) {
	fmt.Printf("%sLeafNode: ", indent)
	for i := range n.keys {
		fmt.Printf("%v:%v ", n.keys[i], n.values[i])
	}
	fmt.Println()
}

// ========== 辅助函数实现 ==========

func (t *BPTree[K, V]) handleUnderflow(parent *internalNode[K, V], idx int) {
	if idx > 0 && parent.children[idx-1].size() > t.minKeys {
		t.borrowFromLeft(parent, idx)
		return
	}

	if idx < len(parent.children)-1 && parent.children[idx+1].size() > t.minKeys {
		t.borrowFromRight(parent, idx)
		return
	}

	if idx > 0 {
		t.mergeNodes(parent, idx-1)
	} else {
		t.mergeNodes(parent, idx)
	}
}

func (t *BPTree[K, V]) borrowFromLeft(parent *internalNode[K, V], idx int) {
	child := parent.children[idx]
	leftSibling := parent.children[idx-1]

	if child.isLeaf() {
		// 叶子节点处理
		leaf := child.(*leafNode[K, V])
		leftLeaf := leftSibling.(*leafNode[K, V])

		// 从左兄弟移动最后一个元素到当前节点头部
		lastIdx := leftLeaf.size() - 1
		leaf.keys = insertAt(leaf.keys, 0, leftLeaf.keys[lastIdx])
		leaf.values = insertAtValue(leaf.values, 0, leftLeaf.values[lastIdx])

		leftLeaf.keys = leftLeaf.keys[:lastIdx]
		leftLeaf.values = leftLeaf.values[:lastIdx]

		// 更新父节点分隔键
		parent.keys[idx-1] = leaf.keys[0]
	} else {
		// 内部节点处理
		internal := child.(*internalNode[K, V])
		leftInternal := leftSibling.(*internalNode[K, V])

		// 将左兄弟最后一个key提升到父节点
		lastKeyIdx := len(leftInternal.keys) - 1
		parentKey := parent.keys[idx-1]
		parent.keys[idx-1] = leftInternal.keys[lastKeyIdx]

		// 将左兄弟最后一个child移动到当前节点
		lastChildIdx := len(leftInternal.children) - 1
		internal.keys = insertAt(internal.keys, 0, parentKey)
		internal.children = insertAtNode(internal.children, 0, leftInternal.children[lastChildIdx])

		leftInternal.keys = leftInternal.keys[:lastKeyIdx]
		leftInternal.children = leftInternal.children[:lastChildIdx]
	}
}

func (t *BPTree[K, V]) borrowFromRight(parent *internalNode[K, V], idx int) {
	child := parent.children[idx]
	rightSibling := parent.children[idx+1]

	if child.isLeaf() {
		// 叶子节点处理
		leaf := child.(*leafNode[K, V])
		rightLeaf := rightSibling.(*leafNode[K, V])

		// 从右兄弟移动第一个元素到当前节点尾部
		leaf.keys = append(leaf.keys, rightLeaf.keys[0])
		leaf.values = append(leaf.values, rightLeaf.values[0])

		rightLeaf.keys = rightLeaf.keys[1:]
		rightLeaf.values = rightLeaf.values[1:]

		// 更新父节点分隔键
		parent.keys[idx] = rightLeaf.keys[0]
	} else {
		// 内部节点处理
		internal := child.(*internalNode[K, V])
		rightInternal := rightSibling.(*internalNode[K, V])

		// 将右兄弟第一个key提升到父节点
		parentKey := parent.keys[idx]
		parent.keys[idx] = rightInternal.keys[0]

		// 将右兄弟第一个child移动到当前节点
		internal.keys = append(internal.keys, parentKey)
		internal.children = append(internal.children, rightInternal.children[0])

		rightInternal.keys = rightInternal.keys[1:]
		rightInternal.children = rightInternal.children[1:]
	}
}

func (t *BPTree[K, V]) mergeNodes(parent *internalNode[K, V], idx int) {
	left := parent.children[idx]
	right := parent.children[idx+1]

	if left.isLeaf() {
		// 合并叶子节点
		leftLeaf := left.(*leafNode[K, V])
		rightLeaf := right.(*leafNode[K, V])

		leftLeaf.keys = append(leftLeaf.keys, rightLeaf.keys...)
		leftLeaf.values = append(leftLeaf.values, rightLeaf.values...)
		leftLeaf.next = rightLeaf.next
	} else {
		// 合并内部节点
		leftInternal := left.(*internalNode[K, V])
		rightInternal := right.(*internalNode[K, V])

		// 合并键和子节点
		leftInternal.keys = append(leftInternal.keys, parent.keys[idx])
		leftInternal.keys = append(leftInternal.keys, rightInternal.keys...)
		leftInternal.children = append(leftInternal.children, rightInternal.children...)
	}

	// 从父节点移除已合并的键和子节点
	parent.keys = append(parent.keys[:idx], parent.keys[idx+1:]...)
	parent.children = append(parent.children[:idx+1], parent.children[idx+2:]...)
}

// ========== 切片操作辅助函数 ==========

func insertAt[K any](slice []K, index int, value K) []K {
	if index >= len(slice) {
		return append(slice, value)
	}
	slice = append(slice[:index+1], slice[index:]...)
	slice[index] = value
	return slice
}

func insertAtNode[K any, V any](slice []node[K, V], index int, value node[K, V]) []node[K, V] {
	if index >= len(slice) {
		return append(slice, value)
	}
	slice = append(slice[:index+1], slice[index:]...)
	slice[index] = value
	return slice
}

func insertAtValue[V any](slice []V, index int, value V) []V {
	if index >= len(slice) {
		return append(slice, value)
	}
	slice = append(slice[:index+1], slice[index:]...)
	slice[index] = value
	return slice
}
