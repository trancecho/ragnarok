package heap

import (
	"reflect"
	"testing"
)

func intLess(a, b int) bool { return a < b }

func TestHeap_DownAdjustMin(t *testing.T) {
	tests := []struct {
		name        string
		initialHeap []int
		parentIndex int
		wantHeap    []int
		cmpType     string // "min" or "max"
	}{
		{
			name:        "小根堆-左子节点更小",
			initialHeap: []int{5, 3, 4},
			parentIndex: 0,
			wantHeap:    []int{3, 5, 4},
			cmpType:     "min",
		},
		{
			name:        "小根堆-右子节点更小",
			initialHeap: []int{5, 6, 3},
			parentIndex: 0,
			wantHeap:    []int{3, 6, 5},
			cmpType:     "min",
		},
		{
			name:        "小根堆-无需调整",
			initialHeap: []int{1, 3, 2},
			parentIndex: 0,
			wantHeap:    []int{1, 3, 2},
			cmpType:     "min",
		},
		{
			name:        "大根堆-左子节点更大",
			initialHeap: []int{2, 5, 4},
			parentIndex: 0,
			wantHeap:    []int{5, 2, 4},
			cmpType:     "max",
		},
		{
			name:        "大根堆-右子节点更大",
			initialHeap: []int{2, 1, 4},
			parentIndex: 0,
			wantHeap:    []int{4, 1, 2},
			cmpType:     "max",
		},
		{
			name:        "大根堆-无需调整",
			initialHeap: []int{5, 3, 2},
			parentIndex: 0,
			wantHeap:    []int{5, 3, 2},
			cmpType:     "max",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cmp Comparable[int]
			if tt.cmpType == "min" {
				cmp = MinHeap(intLess)
			} else {
				cmp = MaxHeap(intLess)
			}
			h := NewHeap[int](len(tt.initialHeap), cmp)
			copy(h.HPDate, tt.initialHeap)
			h.heapSize = len(tt.initialHeap)
			h.DownAdjust(&tt.parentIndex)
			if !reflect.DeepEqual(h.HPDate, tt.wantHeap) {
				t.Errorf("DownAdjust() = %v, want %v", h.HPDate, tt.wantHeap)
			}
		})
	}
}

func TestHeap_UpAdjust(t *testing.T) {
	tests := []struct {
		name        string
		initialHeap []int
		childIndex  int
		wantHeap    []int
		cmpType     string // "min" or "max"
	}{
		{
			name:        "小根堆-插入后需上浮到根节点",
			initialHeap: []int{3, 5, 4, 7, 6, 8, 9, 10, 2},
			childIndex:  8,
			wantHeap:    []int{2, 3, 4, 5, 6, 8, 9, 10, 7},
			cmpType:     "min",
		},
		{
			name:        "小根堆-插入后无需调整",
			initialHeap: []int{1, 3, 2, 7, 6, 4, 5, 8, 9},
			childIndex:  8,
			wantHeap:    []int{1, 3, 2, 7, 6, 4, 5, 8, 9},
			cmpType:     "min",
		},
		{
			name:        "大根堆-插入后需上浮到根节点",
			initialHeap: []int{3, 2, 4, 1, 5, 0, 6, 7, 9},
			childIndex:  8,
			wantHeap:    []int{9, 3, 4, 2, 5, 0, 6, 7, 1},
			cmpType:     "max",
		},
		{
			name:        "大根堆-插入后无需调整",
			initialHeap: []int{9, 7, 6, 3, 5, 0, 4, 1, 2},
			childIndex:  8,
			wantHeap:    []int{9, 7, 6, 3, 5, 0, 4, 1, 2},
			cmpType:     "max",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cmp Comparable[int]
			if tt.cmpType == "min" {
				cmp = MinHeap(intLess)
			} else {
				cmp = MaxHeap(intLess)
			}
			h := NewHeap[int](len(tt.initialHeap), cmp)
			copy(h.HPDate, tt.initialHeap)
			h.heapSize = len(tt.initialHeap)
			h.UpAdjust(&tt.childIndex)
			if !reflect.DeepEqual(h.HPDate, tt.wantHeap) {
				t.Errorf("UpAdjust() = %v, want %v", h.HPDate, tt.wantHeap)
			}
		})
	}
}

func TestHeapSorted(t *testing.T) {
	tests := []struct {
		name    string
		arr     []int
		cmpType string // "min" or "max"
		expect  []int
	}{
		{
			name:    "空数组排序",
			arr:     []int{},
			cmpType: "max",
			expect:  []int{},
		},
		{
			name:    "单个元素数组排序",
			arr:     []int{5},
			cmpType: "max",
			expect:  []int{5},
		},
		{
			name:    "升序排序",
			arr:     []int{5, 2, 9, 1, 7, 6, 8, 3, 4},
			cmpType: "min",
			expect:  []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			name:    "降序排序",
			arr:     []int{5, 2, 9, 1, 7, 6, 8, 3, 4},
			cmpType: "max",
			expect:  []int{9, 8, 7, 6, 5, 4, 3, 2, 1},
		},
		{
			name:    "包含重复元素",
			arr:     []int{3, 1, 4, 1, 5, 9, 2, 6},
			cmpType: "min",
			expect:  []int{1, 1, 2, 3, 4, 5, 6, 9},
		},
		{
			name:    "已排序数组（升序）",
			arr:     []int{1, 2, 3, 4, 5},
			cmpType: "min",
			expect:  []int{1, 2, 3, 4, 5},
		},
		{
			name:    "已排序数组（降序）",
			arr:     []int{5, 4, 3, 2, 1},
			cmpType: "max",
			expect:  []int{5, 4, 3, 2, 1},
		},
		{
			name:    "负数元素测试",
			arr:     []int{-3, 0, -5, 2, -1, 4},
			cmpType: "min",
			expect:  []int{-5, -3, -1, 0, 2, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := make([]int, len(tt.arr))
			copy(input, tt.arr)
			var cmp Comparable[int]
			if tt.cmpType == "min" {
				cmp = MinHeap(intLess)
			} else {
				cmp = MaxHeap(intLess)
			}
			result := HeapSorted(input, cmp)
			if !reflect.DeepEqual(result, tt.expect) {
				t.Errorf("HeapSortedWithCmp(%v, %v) = %v, want %v", tt.arr, tt.cmpType, result, tt.expect)
			}
		})
	}
}

func TestPriorityQueue(t *testing.T) {
	tests := []struct {
		name          string
		operations    []string
		values        []int
		expectedQueue []int
		expectedPops  []int
		cmpType       string // "min" 或 "max"
	}{
		{
			name:          "小根堆优先队列-基本操作",
			operations:    []string{"enqueue", "enqueue", "enqueue", "dequeue", "enqueue", "dequeue", "dequeue", "dequeue"},
			values:        []int{5, 3, 8, -1, 2, -1, -1, -1},
			expectedQueue: []int{},
			expectedPops:  []int{3, 2, 5, 8},
			cmpType:       "min",
		},
		{
			name:          "大根堆优先队列-基本操作",
			operations:    []string{"enqueue", "enqueue", "enqueue", "dequeue", "enqueue", "dequeue", "dequeue", "dequeue"},
			values:        []int{5, 3, 8, -1, 2, -1, -1, -1},
			expectedQueue: []int{},
			expectedPops:  []int{8, 5, 3, 2},
			cmpType:       "max",
		},
		{
			name:          "空队列出队",
			operations:    []string{"dequeue"},
			values:        []int{-1},
			expectedQueue: []int{},
			expectedPops:  []int{0}, // 出队失败，返回零值
			cmpType:       "min",
		},
		{
			name:          "队列满后入队",
			operations:    []string{"enqueue", "enqueue", "enqueue", "enqueue", "enqueue", "dequeue", "dequeue", "dequeue", "dequeue"},
			values:        []int{5, 3, 8, 1, 10, -1, -1, -1, -1},
			expectedQueue: []int{10},
			expectedPops:  []int{1, 3, 5, 8},
			cmpType:       "min",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cmp Comparable[int]
			if tt.cmpType == "min" {
				cmp = MinHeap(intLess)
			} else {
				cmp = MaxHeap(intLess)
			}

			// 创建优先队列，容量为5
			pq := NewPriorityQueue[int](5, cmp)

			actualPops := []int{}

			for i, op := range tt.operations {
				if op == "enqueue" {
					pq.Enqueue(tt.values[i])
				} else if op == "dequeue" {
					val, ok := pq.Dequeue()
					if ok == nil {
						actualPops = append(actualPops, val)
					} else {
						actualPops = append(actualPops, 0) // 出队失败
					}
				}
			}

			// 验证出队的元素是否符合预期
			if !reflect.DeepEqual(actualPops, tt.expectedPops) {
				t.Errorf("出队元素 = %v, 期望 = %v", actualPops, tt.expectedPops)
			}

			// 额外测试：获取队列中剩余元素
			remainingElements := []int{}
			for !pq.heap.IsEmpty() {
				val, _ := pq.Dequeue()
				remainingElements = append(remainingElements, val)
			}

			if !reflect.DeepEqual(remainingElements, tt.expectedQueue) {
				t.Errorf("剩余元素 = %v, 期望 = %v", remainingElements, tt.expectedQueue)
			}
		})
	}
}
