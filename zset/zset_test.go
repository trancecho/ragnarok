package zset

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestZSetBasicOperations(t *testing.T) {
	zset := NewZSet()

	// 测试空集合
	if zset.skiplist.length != 0 {
		t.Error("New ZSet should be empty")
	}
	// 添加元素
	zset.ZAdd("Alice", 85.5)
	zset.ZAdd("Bob", 72.0)
	zset.ZAdd("Charlie", 92.5)
	zset.Print() // 打印跳表结构

	// 测试长度
	if zset.skiplist.length != 3 {
		t.Errorf("Expected length 3, got %d", zset.skiplist.length)
	}

	// 测试分数查询
	if score, ok := zset.ZScore("Alice"); !ok || score != 85.5 {
		t.Errorf("Expected Alice score 85.5, got %.1f", score)
	}

	// 测试不存在的元素
	if _, ok := zset.ZScore("David"); ok {
		t.Error("David should not exist")
	}

	// 测试排名
	if rank, ok := zset.ZRank("Bob"); !ok || rank != 2 {
		t.Log(ok)
		t.Errorf("Expected Bob rank 2, got %d", rank)
	}

	if rank, ok := zset.ZRank("Alice"); !ok || rank != 1 {
		t.Log(ok)
		t.Errorf("Expected Alice rank 1, got %d", rank)
	}

	if rank, ok := zset.ZRank("Charlie"); !ok || rank != 0 {
		t.Log(ok)
		t.Errorf("Expected Charlie rank 0, got %d", rank)
	}

	// 测试逆序排名
	if revRank, ok := zset.ZRevRank("Charlie"); !ok || revRank != 2 {
		t.Errorf("Expected Charlie revRank 0, got %d", revRank)
	}

	// 测试范围查询
	rangeResult := zset.ZRange(0, 1)
	expected := []string{"Charlie:92.50", "Alice:85.50"}
	if len(rangeResult) != len(expected) {
		t.Errorf("Expected range result %v, got %v", expected, rangeResult)
	} else {
		for i, s := range expected {
			if rangeResult[i] != s {
				t.Errorf("At index %d: expected %s, got %s", i, s, rangeResult[i])
			}
		}
	}

	// 测试逆序范围查询
	revRangeResult := zset.ZRevRange(0, 1)
	expectedRev := []string{"Bob:72.00", "Alice:85.00"}
	if len(revRangeResult) != len(expectedRev) {
		t.Errorf("Expected rev range result %v, got %v", expectedRev, revRangeResult)
	}

	// 测试删除
	if !zset.ZRem("Bob") {
		t.Error("Failed to remove Bob")
	}

	if zset.skiplist.length != 2 {
		t.Errorf("Expected length 2 after removal, got %d", zset.skiplist.length)
	}

	if _, ok := zset.ZScore("Bob"); ok {
		t.Error("Bob should be removed")
	}
}

func TestZSetUpdateScore(t *testing.T) {
	zset := NewZSet()
	zset.ZAdd("Alice", 85.5)
	zset.ZAdd("Bob", 72.0)
	log.Println("aaa")

	// 更新分数
	zset.ZAdd("Bob", 90.0)
	log.Println("aaa")

	zset.Print()

	// 验证新分数
	if score, ok := zset.ZScore("Bob"); !ok || score != 90.0 {
		t.Errorf("Expected Bob score 90.0, got %.1f", score)
	}

	// 验证新排名
	if rank, ok := zset.ZRank("Bob"); !ok || rank != 0 {
		t.Errorf("Expected Bob rank 1 after update, got %d", rank)
	}

	if rank, ok := zset.ZRank("Alice"); !ok || rank != 1 {
		t.Errorf("Expected Alice rank 0 after update, got %d", rank)
	}
}

func TestZSetEdgeCases(t *testing.T) {
	zset := NewZSet()

	// 测试相同分数的元素
	zset.ZAdd("A", 100.0)
	zset.ZAdd("B", 100.0)
	zset.ZAdd("C", 100.0)
	zset.Print()

	// 验证排名 (按字典序)
	if rank, ok := zset.ZRank("A"); !ok || rank != 2 {
		t.Errorf("Expected A rank 0, got %d", rank)
	}
	if rank, ok := zset.ZRank("B"); !ok || rank != 1 {
		t.Errorf("Expected B rank 1, got %d", rank)
	}
	if rank, ok := zset.ZRank("C"); !ok || rank != 0 {
		t.Errorf("Expected C rank 2, got %d", rank)
	}

	// 测试范围查询边界
	all := zset.ZRange(0, 2)
	if len(all) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(all))
	}

	// 超出范围的查询
	outOfRange := zset.ZRange(5, 10)
	if len(outOfRange) != 0 {
		t.Errorf("Expected empty slice, got %v", outOfRange)
	}

	// 无效范围
	invalidRange := zset.ZRange(2, 1)
	if len(invalidRange) != 0 {
		t.Errorf("Expected empty slice for invalid range, got %v", invalidRange)
	}
}

func TestZSetLargeDataset(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	zset := NewZSet()
	const numElements = 20

	// 添加随机元素
	for i := 0; i < numElements; i++ {
		ele := fmt.Sprintf("element-%d", i)
		score := rand.Float64() * 1000
		zset.ZAdd(ele, score)
		zset.printSkipList()
	}

	// 验证长度
	if zset.skiplist.length != numElements {
		t.Errorf("Expected %d elements, got %d", numElements, zset.skiplist.length)
	}
	zset.Print()

	// 验证排名一致性
	for i := 0; i < 50; i++ { // 随机检查50个元素
		ele := fmt.Sprintf("element-%d", rand.Intn(numElements))
		rank, ok := zset.ZRank(ele)
		if !ok {
			t.Errorf("Element %s not found", ele)
			continue
		}

		// 通过范围查询验证排名
		elements := zset.ZRange(rank, rank)
		if len(elements) != 1 || !strings.Contains(elements[0], ele) {
			t.Errorf("Rank mismatch for %s. Expected at rank %d, got: %v", ele, rank, elements)
		}
	}

	// 测试逆序排名
	for i := 0; i < 50; i++ { // 随机检查50个元素
		ele := fmt.Sprintf("element-%d", rand.Intn(numElements))
		revRank, ok := zset.ZRevRank(ele)
		if !ok {
			t.Errorf("Element %s not found for reverse rank", ele)
			continue
		}
		// 通过逆序范围查询验证逆序排名
		revElements := zset.ZRevRange(revRank, revRank)
		if len(revElements) != 1 || !strings.Contains(revElements[0], ele) {
			t.Errorf("Reverse rank mismatch for %s. Expected at rev rank %d, got: %v", ele, revRank, revElements)
		}
	}
}

func TestZSetConcurrentOperations(t *testing.T) {
	zset := NewZSet()
	const numWorkers = 10
	const opsPerWorker = 100

	done := make(chan bool)

	// 并发添加元素
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			for j := 0; j < opsPerWorker; j++ {
				ele := fmt.Sprintf("worker-%d-element-%d", workerID, j)
				score := float64(workerID*100 + j)
				zset.ZAdd(ele, score)
			}
			done <- true
		}(i)
	}

	// 等待所有添加完成
	for i := 0; i < numWorkers; i++ {
		<-done
	}

	// 验证总数
	expectedTotal := numWorkers * opsPerWorker
	if zset.skiplist.length != int(expectedTotal) {
		t.Errorf("Expected %d elements, got %d", expectedTotal, zset.skiplist.length)
	}

	// 并发读取和删除
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			for j := 0; j < opsPerWorker; j++ {
				ele := fmt.Sprintf("worker-%d-element-%d", workerID, j)

				// 读取操作
				if score, ok := zset.ZScore(ele); ok {
					expected := float64(workerID*100 + j)
					if score != expected {
						t.Errorf("Expected %.1f, got %.1f for %s", expected, score, ele)
					}
				}

				// 删除操作
				zset.ZRem(ele)
			}
			done <- true
		}(i)
	}

	// 等待所有操作完成
	for i := 0; i < numWorkers; i++ {
		<-done
	}

	// 验证所有元素已被删除
	if zset.skiplist.length != 0 {
		t.Errorf("Expected empty set, got %d elements", zset.skiplist.length)
	}
}

//todo 有问题

func TestZSetRankAccuracy(t *testing.T) {
	zset := NewZSet()

	// 创建有序测试数据
	type testData struct {
		ele   string
		score float64
	}

	var data []testData
	for i := 0; i < 100; i++ {
		data = append(data, testData{ele: fmt.Sprintf("item-%d", i), score: float64(i)})
	}

	// 打乱顺序插入
	rand.Shuffle(len(data), func(i, j int) {
		data[i], data[j] = data[j], data[i]
	})

	for _, d := range data {
		zset.ZAdd(d.ele, d.score)
	}
	zset.printSkipList()
	// 验证排名顺序
	sort.Slice(data, func(i, j int) bool {
		if data[i].score == data[j].score {
			return data[i].ele > data[j].ele
		}
		return data[i].score > data[j].score
	})

	for expectedRank, d := range data {
		actualRank, ok := zset.ZRank(d.ele)
		if !ok {
			t.Errorf("Element %s not found", d.ele)
			continue
		}

		if int(expectedRank) != actualRank {
			t.Errorf("For element %s (score %.1f): expected rank %d, got %d",
				d.ele, d.score, expectedRank, actualRank)
		}

		// 验证逆序排名
		expectedRevRank := len(data) - 1 - expectedRank
		actualRevRank, ok := zset.ZRevRank(d.ele)
		if !ok {
			t.Errorf("Element %s not found for rev rank", d.ele)
			continue
		}

		if int(expectedRevRank) != actualRevRank {
			t.Errorf("For element %s (score %.1f): expected rev rank %d, got %d",
				d.ele, d.score, expectedRevRank, actualRevRank)
		}
	}
}

func TestZSetRangeQueries(t *testing.T) {
	zset := NewZSet()

	// 添加元素
	for i := 0; i < 10; i++ {
		ele := fmt.Sprintf("elem-%d", i)
		zset.ZAdd(ele, float64(i*10))
	}

	// 测试各种范围查询
	testCases := []struct {
		start, end int
		expected   []string
	}{
		{0, 0, []string{"elem-9:90.00"}},
		{0, 4, []string{"elem-9:90.00", "elem-8:80.00", "elem-7:70.00", "elem-6:60.00", "elem-5:50.00"}},
		{5, 9, []string{"elem-4:40.00", "elem-3:30.00", "elem-2:20.00", "elem-1:10.00", "elem-0:0.00"}},
	}

	for i, tc := range testCases {
		result := zset.ZRange(tc.start, tc.end)
		if len(result) != len(tc.expected) {
			t.Errorf("Test case %d: expected %d elements, got %d", i, len(tc.expected), len(result))
			continue
		}

		for j, s := range tc.expected {
			if result[j] != s {
				t.Errorf("Test case %d at index %d: expected %s, got %s", i, j, s, result[j])
			}
		}
	}

	// 测试逆序范围查询
	revTestCases := []struct {
		start, end int
		expected   []string
	}{
		{0, 0, []string{"elem-0:0.00"}},
		{0, 4, []string{"elem-0:0.00", "elem-1:10.00", "elem-2:20.00", "elem-3:30.00", "elem-4:40.00"}},
		{5, 9, []string{"elem-5:50.00", "elem-6:60.00", "elem-7:70.00", "elem-8:80.00", "elem-9:90.00"}},
		{8, 12, []string{"elem-8:80.00", "elem-9:90.00"}}, // 超出范围
		{3, 3, []string{"elem-3:30.00"}},
	}

	for i, tc := range revTestCases {
		result := zset.ZRevRange(tc.start, tc.end)
		if len(result) != len(tc.expected) {
			t.Errorf("Rev test case %d: expected %d elements, got %d", i, len(tc.expected), len(result))
			continue
		}

		for j, s := range tc.expected {
			if result[j] != s {
				t.Errorf("Rev test case %d at index %d: expected %s, got %s", i, j, s, result[j])
			}
		}
	}
}

// 辅助函数：打印跳表结构（用于调试）
func (zset *ZSet) printSkipList() {
	zset.mu.RLock()
	defer zset.mu.RUnlock()
	fmt.Println("\nSkip List Structure:")
	fmt.Printf("Level: %d, Length: %d\n", zset.skiplist.level, zset.skiplist.length)

	for i := zset.skiplist.level; i >= 0; i-- {
		fmt.Printf("Level %d: ", i)
		x := zset.skiplist.header
		for x != nil {
			if x == zset.skiplist.header {
				fmt.Print("HEAD")
				fmt.Printf("[span:%d]", x.level[i].span)
			} else {
				fmt.Printf("[%s(%.1f)|span:%d]", x.ele, x.score, x.level[i].span)
			}

			if x.level[i].forward != nil {
				fmt.Print(" → ")
			}
			x = x.level[i].forward
		}
		fmt.Println()
	}

	// 打印原始链表
	fmt.Print("Original List: ")
	x := zset.skiplist.header.level[0].forward
	for x != nil {
		fmt.Printf("%s(%.1f)", x.ele, x.score)
		if x.level[0].forward != nil {
			fmt.Print(" → ")
		}
		x = x.level[0].forward
	}
	fmt.Println("")
}

func TestZSetVisualization(t *testing.T) {
	zset := NewZSet()
	//zset.ZAdd("element-0", 914.77)
	//zset.ZAdd("element-6", 833.04)
	//zset.ZAdd("element-9", 505.13)
	//zset.ZAdd("element-3", 303.39)
	//zset.ZAdd("element-2", 302.53)
	//zset.ZAdd("element-8", 164.90)
	//zset.ZAdd("element-7", 126.54)
	//zset.ZAdd("element-5", 108.67)
	//zset.ZAdd("element-4", 40.17)
	//zset.ZAdd("element-1", 8.30)

	for i := 0; i < 10; i++ {
		ele := fmt.Sprintf("element-%d", i)
		score := rand.Float64() * 1000
		zset.ZAdd(ele, score)
		zset.printSkipList()

	}
	//res := zset.ZRange(0, 8)
	//log.Println("ZRange(0,8):", res)
}
