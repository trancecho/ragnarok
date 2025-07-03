package bloom_filter

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBloomFilter_CachePenetrationProtection(t *testing.T) {
	// 模拟数据库中存在的数据ID
	existingKeys := []string{
		"product-1001",
		"product-1002",
		"product-1003",
		"user-2001",
		"user-2002",
	}

	// 创建布隆过滤器，预期10000个元素，假阳率1%
	bf := NewBloomFilter(10000, 0.01)

	// 1. 初始状态下，所有key都应该返回false
	for _, key := range existingKeys {
		assert.False(t, bf.Contains([]byte(key)), "Key should not exist before adding")
	}

	// 2. 加载数据库中存在的数据到布隆过滤器
	// 通常在系统启动时执行，这里模拟这个过程
	for _, key := range existingKeys {
		bf.Add([]byte(key))
	}

	// 3. 测试存在的key应该返回true
	for _, key := range existingKeys {
		assert.True(t, bf.Contains([]byte(key)), "Existing key should be found")
	}

	// 4. 测试不存在的key
	nonExistingKeys := []string{
		"product-9999",
		"user-9999",
		"invalid-key",
		"",
	}
	for _, key := range nonExistingKeys {
		assert.False(t, bf.Contains([]byte(key)), "Non-existing key should return false")
	}

	// 5. 模拟缓存击穿场景
	t.Run("CachePenetrationScenario", func(t *testing.T) {
		// 模拟一个热点key过期后，大量并发请求涌入
		hotKey := "hot-product-123"
		concurrentRequests := 1000
		foundCount := 0
		notFoundCount := 0

		// 初始状态下key不存在
		assert.False(t, bf.Contains([]byte(hotKey)), "Hot key should not exist initially")

		// 模拟并发请求
		var wg sync.WaitGroup
		wg.Add(concurrentRequests)

		start := time.Now()
		for i := 0; i < concurrentRequests; i++ {
			go func() {
				defer wg.Done()
				if bf.Contains([]byte(hotKey)) {
					foundCount++
				} else {
					notFoundCount++
					// 在实际应用中，这里会查询数据库
					// 查询到数据后，会添加到缓存和布隆过滤器
				}
			}()
		}
		wg.Wait()
		duration := time.Since(start)

		t.Logf("Processed %d concurrent requests in %v", concurrentRequests, duration)
		t.Logf("Found: %d, Not Found: %d", foundCount, notFoundCount)

		// 所有请求都应该得到key不存在的结论
		assert.Equal(t, concurrentRequests, notFoundCount, "All requests should get 'not found'")
		assert.Equal(t, 0, foundCount, "No requests should get false positive")

		// 模拟数据库查询到数据后，添加到布隆过滤器
		bf.Add([]byte(hotKey))
		assert.True(t, bf.Contains([]byte(hotKey)), "Hot key should exist after adding")

		// 再次模拟并发请求
		foundCount = 0
		notFoundCount = 0
		wg.Add(concurrentRequests)
		start = time.Now()
		for i := 0; i < concurrentRequests; i++ {
			go func() {
				defer wg.Done()
				if bf.Contains([]byte(hotKey)) {
					foundCount++
				} else {
					notFoundCount++
				}
			}()
		}
		wg.Wait()
		duration = time.Since(start)

		t.Logf("Processed %d concurrent requests in %v (after adding)", concurrentRequests, duration)
		t.Logf("Found: %d, Not Found: %d", foundCount, notFoundCount)

		// 现在所有请求都应该得到key存在的结论
		assert.Equal(t, concurrentRequests, foundCount, "All requests should get 'found'")
		assert.Equal(t, 0, notFoundCount, "No requests should get false negative")
	})

	// 6. 测试性能：布隆过滤器查询 vs 模拟数据库查询
	t.Run("PerformanceComparison", func(t *testing.T) {
		testKey := []byte("product-1001")
		iterations := 100000

		// 测试布隆过滤器查询性能
		start := time.Now()
		for i := 0; i < iterations; i++ {
			bf.Contains(testKey)
		}
		bfDuration := time.Since(start)
		t.Logf("Bloom filter: %d queries in %v", iterations, bfDuration)

		// 模拟数据库查询性能（假设每次查询耗时1ms）
		dbQueryTime := time.Duration(iterations) * time.Millisecond
		t.Logf("Simulated DB: %d queries would take %v", iterations, dbQueryTime)

		// 验证布隆过滤器确实快得多
		assert.True(t, bfDuration < dbQueryTime/100, "Bloom filter should be much faster than DB query")
	})

	// 7. 测试误判率
	t.Run("FalsePositiveRate", func(t *testing.T) {
		// 添加所有存在的key
		for _, key := range existingKeys {
			bf.Add([]byte(key))
		}

		// 测试10000个不存在的key
		tests := 10000
		falsePositives := 0
		for i := 0; i < tests; i++ {
			key := []byte(fmt.Sprintf("nonexistent-key-%d", i))
			if bf.Contains(key) {
				falsePositives++
			}
		}

		fpRate := float64(falsePositives) / float64(tests)
		t.Logf("False positive rate: %.4f (expected <0.01)", fpRate)

		// 验证误判率在预期范围内
		assert.Less(t, fpRate, 0.02, "False positive rate should be less than 2%")
	})
}
