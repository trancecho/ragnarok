package heap

import "errors"

// Comparable 定义了一个比较函数类型
type Comparable[T any] func(a, b T) int

// Heap 使用泛型T代替具体类型
type Heap[T any] struct {
	HPDate    []T
	heapSize  int           // 堆的大小
	capacity  int           // 堆的容量
	compareFn Comparable[T] // 比较函数
}

// NewHeap 创建一个新的堆
func NewHeap[T any](capacity int, compareFn Comparable[T]) *Heap[T] {
	return &Heap[T]{
		HPDate:    make([]T, capacity),
		heapSize:  0,
		capacity:  capacity,
		compareFn: compareFn,
	}
}

// MinHeap 创建小根堆的比较函数
func MinHeap[T any](less func(a, b T) bool) Comparable[T] {
	return func(a, b T) int {
		if less(a, b) {
			return -1
		} else if less(b, a) {
			return 1
		}
		return 0
	}
}

// MaxHeap 创建大根堆的比较函数
func MaxHeap[T any](less func(a, b T) bool) Comparable[T] {
	return func(a, b T) int {
		if less(a, b) {
			return 1
		} else if less(b, a) {
			return -1
		}
		return 0
	}
}

// DownAdjust 向下调整算法
func (h *Heap[T]) DownAdjust(parentIndex *int) {
	parent := *parentIndex
	child := 2*parent + 1

	for child < h.heapSize {
		// 选择较小/较大的子节点(取决于比较函数)
		if child+1 < h.heapSize && h.compareFn(h.HPDate[child+1], h.HPDate[child]) < 0 {
			child++
		}
		// 如果父节点满足堆特性，则调整结束
		if h.compareFn(h.HPDate[parent], h.HPDate[child]) <= 0 {
			break
		}
		// 交换节点
		h.HPDate[parent], h.HPDate[child] = h.HPDate[child], h.HPDate[parent]
		parent = child
		child = 2*parent + 1
	}
}

// UpAdjust 向上调整算法
func (h *Heap[T]) UpAdjust(childIndex *int) {
	child := *childIndex
	for child > 0 {
		parent := (child - 1) / 2
		if h.compareFn(h.HPDate[parent], h.HPDate[child]) <= 0 {
			break
		}
		// 交换父子节点
		h.HPDate[parent], h.HPDate[child] = h.HPDate[child], h.HPDate[parent]
		child = parent
	}
}

// Insert 插入元素
func (h *Heap[T]) Insert(value T) {
	if h.heapSize >= h.capacity {
		return // 堆已满
	}
	h.HPDate[h.heapSize] = value
	childIndex := h.heapSize
	h.heapSize++
	h.UpAdjust(&childIndex)
}

// Pop 弹出堆顶元素
func (h *Heap[T]) Pop() (T, error) {
	var zero T
	if h.heapSize == 0 {
		return zero, errors.New("栈为空") // 堆为空
	}
	top := h.HPDate[0]
	h.HPDate[0] = h.HPDate[h.heapSize-1]
	h.heapSize--
	parentIndex := 0
	h.DownAdjust(&parentIndex)
	return top, nil
}

// Peek 获取堆顶元素
func (h *Heap[T]) Peek() (T, error) {
	var zero T
	if h.heapSize == 0 {
		return zero, errors.New("栈为空") // 堆为空
	}
	return h.HPDate[0], nil // 返回堆顶元素
}

// Size 获取堆的大小
func (h *Heap[T]) Size() int {
	return h.heapSize
}

// IsEmpty 检查堆是否为空
func (h *Heap[T]) IsEmpty() bool {
	return h.heapSize == 0
}

// Clear 清空堆
func (h *Heap[T]) Clear() {
	h.HPDate = make([]T, h.capacity)
	h.heapSize = 0
}

// AdjustHeap 自下而上建堆法
func (h *Heap[T]) AdjustHeap() {
	if h.heapSize <= 1 {
		return // 如果堆中只有一个元素或为空，直接返回
	}
	for i := h.heapSize/2 - 1; i >= 0; i-- {
		idx := i
		h.DownAdjust(&idx)
	}
}

// HeapifyUp 自上而下建堆法
func HeapifyUp[T any](array []T, compareFn Comparable[T]) Heap[T] {
	h := NewHeap[T](len(array), compareFn)
	if len(array) == 0 {
		return *h // 如果数组为空，直接返回
	}
	for i := 0; i < len(array); i++ {
		h.Insert(array[i])
	}
	return *h
}

// HeapSorted 对数组进行堆排序
func HeapSorted[T any](arr []T, compareFn Comparable[T]) []T {
	if len(arr) <= 1 {
		return arr // 如果数组为空或只有一个元素，直接返回
	}

	// 创建反向比较函数的堆（若要升序则创建大根堆，若要降序则创建小根堆）
	reverseCompare := func(a, b T) int {
		return -compareFn(a, b)
	}
	h := NewHeap[T](len(arr), reverseCompare)

	// 复制数组到堆中
	copy(h.HPDate, arr)
	h.heapSize = len(arr)
	h.AdjustHeap()

	size := h.heapSize
	// 排序过程
	for i := size - 1; i > 0; i-- {
		// 交换堆顶与当前末尾元素
		h.HPDate[0], h.HPDate[i] = h.HPDate[i], h.HPDate[0]
		// 缩小堆的有效范围
		h.heapSize--
		// 调整堆顶
		parentIndex := 0
		h.DownAdjust(&parentIndex)
	}

	return h.HPDate[:size]
}

// PriorityQueue 优先队列
type PriorityQueue[T any] struct {
	heap *Heap[T] // 使用堆来实现优先队列
}

// NewPriorityQueue 创建新的优先队列
func NewPriorityQueue[T any](capacity int, compareFn Comparable[T]) *PriorityQueue[T] {
	return &PriorityQueue[T]{
		heap: NewHeap[T](capacity, compareFn),
	}
}

// Enqueue 入队
func (pq *PriorityQueue[T]) Enqueue(value T) {
	pq.heap.Insert(value)
}

// Dequeue 出队
func (pq *PriorityQueue[T]) Dequeue() (T, error) {
	return pq.heap.Pop()
}

//package heap
//
//type Heap struct {
//	HPDate   []int
//	heapSize int // 堆的大小
//	capacity int // 堆的容量
//}
//
//func NewHeap(capacity int) *Heap {
//	return &Heap{
//		HPDate:   make([]int, capacity),
//		heapSize: 0,
//		capacity: capacity,
//	}
//}
//
//// DownAdjustMin 向下调整算法-小根堆（调整父节点）
//func (h *Heap) DownAdjustMin(parentIndex *int) {
//	// 保存父节点索引
//	parent := *parentIndex
//	// 先找到左子节点
//	child := 2*parent + 1
//
//	// 如果有子节点
//	for child < h.heapSize {
//		// 如果右子节点存在且比左子节点小，则定位到右子节点
//		if child+1 < h.heapSize && h.HPDate[child+1] < h.HPDate[child] {
//			child++
//		}
//		// 如果父节点小于等于子节点，则调整结束
//		if h.HPDate[parent] <= h.HPDate[child] {
//			break
//		}
//		// 将子节点的值赋给父节点
//		h.HPDate[parent], h.HPDate[child] = h.HPDate[child], h.HPDate[parent]
//		// 父节点索引指向子节点
//		parent = child
//		// 找到新的左子节点
//		child = 2*parent + 1
//	}
//}
//
//// DownAdjustMax 向下调整算法-大根堆（调整父节点）
//func (h *Heap) DownAdjustMax(parentIndex *int) {
//	// 保存父节点索引
//	parent := *parentIndex
//	// 先找到左子节点
//	child := 2*parent + 1
//
//	// 如果有子节点
//	for child < h.heapSize {
//		// 如果右子节点存在且比左子节点大，则定位到右子节点
//		if child+1 < h.heapSize && h.HPDate[child+1] > h.HPDate[child] {
//			child++
//		}
//		// 如果父节点大于等于子节点，则调整结束
//		if h.HPDate[parent] >= h.HPDate[child] {
//			break
//		}
//		// 将子节点的值赋给父节点
//		h.HPDate[parent], h.HPDate[child] = h.HPDate[child], h.HPDate[parent]
//		// 父节点索引指向子节点
//		parent = child
//		// 找到新的左子节点
//		child = 2*parent + 1
//	}
//}
//
//// UpAdjustMin 向上调整算法-小根堆（调整子节点）
//func (h *Heap) UpAdjustMin(childIndex *int) {
//	child := *childIndex
//	for child > 0 {
//		parent := (child - 1) / 2
//		if h.HPDate[parent] <= h.HPDate[child] {
//			break
//		}
//		// 交换父节点与子节点的值
//		h.HPDate[parent], h.HPDate[child] = h.HPDate[child], h.HPDate[parent]
//		child = parent
//	}
//}
//
//// UpAdjustMax 向上调整算法-大根堆（调整子节点）
//func (h *Heap) UpAdjustMax(childIndex *int) {
//	child := *childIndex
//	for child > 0 {
//		parent := (child - 1) / 2
//		if h.HPDate[parent] >= h.HPDate[child] {
//			break
//		}
//		// 交换父节点与子节点的值
//		h.HPDate[parent], h.HPDate[child] = h.HPDate[child], h.HPDate[parent]
//		child = parent
//	}
//}
//
//// Insert 插入元素
//func (h *Heap) Insert(value int, isMin bool) {
//	if h.heapSize >= h.capacity {
//		return // 堆已满
//	}
//	h.HPDate[h.heapSize] = value
//	h.heapSize++
//	if isMin {
//		h.UpAdjustMin(&h.heapSize)
//	} else {
//		h.UpAdjustMax(&h.heapSize)
//	}
//}
//
//// Pop 弹出堆顶元素
//func (h *Heap) Pop(isMin bool) (int, bool) {
//	if h.heapSize == 0 {
//		return 0, false // 堆为空
//	}
//	top := h.HPDate[0]
//	h.HPDate[0] = h.HPDate[h.heapSize-1]
//	h.heapSize--
//	if isMin {
//		h.DownAdjustMin(&h.heapSize)
//	} else {
//		h.DownAdjustMax(&h.heapSize)
//	}
//	return top, true
//}
//
//// Peek 获取堆顶元素
//func (h *Heap) Peek() (int, bool) {
//	if h.heapSize == 0 {
//		return 0, false // 堆为空
//	}
//	return h.HPDate[0], true // 返回堆顶元素
//}
//
//// Size 获取堆的大小
//func (h *Heap) Size() int {
//	return h.heapSize
//}
//
//// IsEmpty 检查堆是否为空
//func (h *Heap) IsEmpty() bool {
//	return h.heapSize == 0
//}
//
//// Clear 清空堆
//func (h *Heap) Clear() {
//	h.HPDate = make([]int, h.capacity)
//	h.heapSize = 0
//}
//
//// AdjustHeap 自下而上建堆法
//func (h *Heap) AdjustHeap(isMin bool) {
//	if h.heapSize <= 1 {
//		return // 如果堆中只有一个元素或为空，直接返回
//	}
//	for i := h.heapSize/2 - 1; i >= 0; i-- {
//		if isMin {
//			h.DownAdjustMin(&i)
//		} else {
//			h.DownAdjustMax(&i)
//		}
//	}
//}
//
//// HeapifyUp 自上而下建堆法
//func HeapifyUp(array []int, isMin bool) Heap {
//	h := NewHeap(len(array))
//	if len(array) == 0 {
//		return *h // 如果数组为空，直接返回
//	}
//	for i := 0; i < len(array); i++ {
//		h.Insert(array[i], isMin)
//	}
//	return *h
//}
//
//// HeapSorted 对数组进行堆排序 升序是大根堆，降序是小根堆
//func HeapSorted(arr []int, isUP bool) []int {
//	if len(arr) <= 1 {
//		return arr // 如果数组为空或只有一个元素，直接返回
//	}
//
//	// 创建足够大的堆
//	h := NewHeap(len(arr))
//
//	// 复制数组到堆中
//	copy(h.HPDate, arr)
//	h.heapSize = len(arr)
//
//	h.AdjustHeap(!isUP)
//
//	size := h.heapSize
//	// 排序过程
//	for i := size - 1; i > 0; i-- {
//		// 交换堆顶与当前末尾元素
//		h.HPDate[0], h.HPDate[i] = h.HPDate[i], h.HPDate[0]
//		// 缩小堆的有效范围
//		h.heapSize--
//		// 调整堆顶
//		parentIndex := 0
//		if isUP {
//			h.DownAdjustMax(&parentIndex)
//		} else {
//			h.DownAdjustMin(&parentIndex)
//		}
//	}
//
//	return h.HPDate[:size]
//}
//
//// PriorityQueue 优先队列
//type PriorityQueue struct {
//	heap *Heap // 使用堆来实现优先队列
//}
