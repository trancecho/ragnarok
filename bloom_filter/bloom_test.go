package bloom_filter

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBloomFilter_CachePenetrationProtection(t *testing.T) {
	// 模拟数据库中存在的数据ID
	existingIDs := []string{
		"user-1001",
		"product-2002",
		"order-3003",
	}

	// 创建布隆过滤器（存储所有存在的键）
	bf := NewBloomFilter(1000000, 7)

	// 初始化过滤器（通常在系统启动时加载数据库中存在的数据）
	for _, id := range existingIDs {
		bf.Add([]byte(id))
	}

	// 测试存在的键
	for _, id := range existingIDs {
		assert.True(t, bf.Contains([]byte(id)), "Existing key should be found")
	}

	// 测试不存在的键
	nonExistingIDs := []string{
		"user-9999",
		"product-0000",
		"order-XXXX",
	}
	for _, id := range nonExistingIDs {
		found := bf.Contains([]byte(id))
		if found {
			t.Logf("False positive for key: %s", id)
		}
		// 注意：这里不能assert，因为可能有合理的误判
	}

	// 性能对比：布隆过滤器 vs 直接查询数据库（模拟）
	// 布隆过滤器检查
	start := time.Now()
	for i := 0; i < 100000; i++ {
		bf.Contains([]byte(fmt.Sprintf("key-%d", i)))
	}
	bfTime := time.Since(start)
	t.Logf("Bloom filter checked 100,000 keys in %v", bfTime)

	// 模拟数据库查询（假设每次查询耗时1ms）
	dbQueryTime := time.Duration(100000) * time.Millisecond
	t.Logf("Simulated DB would take %v for 100,000 queries", dbQueryTime)

	// 验证布隆过滤器确实快得多
	assert.True(t, bfTime < dbQueryTime/100, "Bloom filter should be much faster than DB query")
}
