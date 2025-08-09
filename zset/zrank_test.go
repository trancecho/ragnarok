package zset

//
//import (
//	"testing"
//)
//
//func TestZRank(t *testing.T) {
//	// 创建一个新的 ZSet
//	zset := NewZSet()
//
//	// 测试空集合的情况
//	if rank, exists := zset.ZRank("nonexistent"); exists {
//		t.Errorf("Expected non-existent element to return false, got true with rank %d", rank)
//	}
//
//	// 添加测试数据
//	testData := []struct {
//		ele   string
//		score float64
//	}{
//		{"one", 1.0},
//		{"two", 2.0},
//		{"three", 3.0},
//		{"four", 4.0},
//		{"five", 5.0},
//	}
//
//	for _, td := range testData {
//		zset.ZAdd(td.ele, td.score)
//	}
//
//	// 测试用例
//	testCases := []struct {
//		name     string
//		ele      string
//		expected int
//		exists   bool
//	}{
//		{"first element", "one", 4, true},
//		{"middle element", "three", 2, true},
//		{"last element", "five", 0, true},
//		{"non-existent element", "six", -1, false},
//	}
//	zset.printSkipList()
//
//	for _, tc := range testCases {
//		t.Run(tc.name, func(t *testing.T) {
//			rank, exists := zset.ZRank(tc.ele)
//			if exists != tc.exists {
//				t.Errorf("Expected exists=%v, got %v", tc.exists, exists)
//			}
//			if exists && rank != tc.expected {
//				t.Errorf("Expected rank %d for element %s, got %d", tc.expected, tc.ele, rank)
//			}
//		})
//	}
//
//	// 测试相同分数的情况
//	zset.ZAdd("another_three", 3.0)
//	rank, exists := zset.ZRank("another_three")
//	if !exists {
//		t.Error("Expected element 'another_three' to exist")
//	}
//	if rank < 2 || rank > 3 {
//		t.Errorf("Expected rank for 'another_three' to be 2 or 3, got %d", rank)
//	}
//
//	// 测试添加重复元素
//	zset.ZAdd("three", 3.5) // 修改分数
//	rank, exists = zset.ZRank("three")
//	if !exists {
//		t.Error("Expected element 'three' to exist after update")
//	}
//	if rank != 3 {
//		t.Errorf("Expected rank 3 for 'three' after update, got %d", rank)
//	}
//}
