package bloom_filter

import (
	"hash/fnv"
	"math"
	"sync"
)

type BloomFilter struct {
	bitset   []bool
	m        int // 位数组大小
	k        int // 哈希函数数量
	mu       sync.RWMutex
}

// n: 预期插入的元素数量
// p: 假阳率，表示误判的概率
func NewBloomFilter(n int, p float64) *BloomFilter {
	if n <= 0 || p <= 0 || p >= 1 {
		return &BloomFilter{
			bitset: make([]bool, 1),
			m:      1,
			k:      1,
		}
	}
	m := int(-float64(n)*math.Log(p)/(math.Log(2)*math.Log(2))) + 1
	k := int(float64(m)/float64(n)*math.Log(2)) + 1
	if m < 1 {
		m = 1
	}
	if k < 1 {
		k = 1
	}
	return &BloomFilter{
		bitset: make([]bool, m),
		m:      m,
		k:      k,
	}
}

// fnvHash64 计算 FNV-1 64位哈希
func fnvHash64(data []byte) uint64 {
	h := fnv.New64()
	h.Write(data)
	return h.Sum64()
}

// fnvHash64a 计算 FNV-1a 64位哈希
func fnvHash64a(data []byte) uint64 {
	h := fnv.New64a()
	h.Write(data)
	return h.Sum64()
}

// hash 使用双哈希技术计算索引
// index = (h1 + i * h2) % m
func (bf *BloomFilter) hash(item []byte, i int) int {
	h1 := fnvHash64(item)
	h2 := fnvHash64a(item)
	// 使用 h2 的绝对值避免负数
	if h2 == 0 {
		h2 = 1
	}
	index := (h1 + uint64(i)*h2) % uint64(bf.m)
	return int(index)
}

func (bf *BloomFilter) Add(item []byte) {
	if len(item) == 0 {
		return
	}
	bf.mu.Lock()
	defer bf.mu.Unlock()
	for i := 0; i < bf.k; i++ {
		index := bf.hash(item, i)
		bf.bitset[index] = true
	}
}

func (bf *BloomFilter) Contains(item []byte) bool {
	if len(item) == 0 {
		return false
	}
	bf.mu.RLock()
	defer bf.mu.RUnlock()
	for i := 0; i < bf.k; i++ {
		index := bf.hash(item, i)
		if !bf.bitset[index] {
			return false // 如果有一个位为false，则说明不包含
		}
	}
	return true // 如果所有位都为true，则可能包含
}

func (bf *BloomFilter) Reset() {
	bf.mu.Lock()
	defer bf.mu.Unlock()
	for i := range bf.bitset {
		bf.bitset[i] = false // 重置位数组
	}
}
