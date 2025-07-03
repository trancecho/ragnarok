package bloom_filter

import (
	"hash"
	"hash/fnv"
	"math"
	"sync"
)

type BloomFilter struct {
	bitset      []bool      // 位数组
	m           int         // 位数组大小
	hashFuncs   []hash.Hash // 哈希函数集合
	hashFuncNum int         // 哈希函数数量
	mu          sync.RWMutex
}

// n: 预期插入的元素数量
// p: 假阳率，表示误判的概率
func NewBloomFilter(n int, p float64) *BloomFilter {
	m := int(-float64(n) * math.Log(p) / (math.Log(2) * math.Log(2)))
	k := int(float64(m) / float64(n) * math.Log(2))
	res := &BloomFilter{
		bitset:      make([]bool, m),
		m:           m,
		hashFuncs:   make([]hash.Hash, k),
		hashFuncNum: k,
	}
	// 初始化哈希函数
	for i := 0; i < k; i++ {
		res.hashFuncs[i] = fnv.New64a()
	}
	return res
}

func (this *BloomFilter) Add(item []byte) {
	this.mu.Lock()
	defer this.mu.Unlock()
	for _, h := range this.hashFuncs {
		h.Write(item)
		index := h.Sum(nil)[0] % uint8(this.m)
		this.bitset[index] = true
		h.Reset() // 重置哈希函数以便下次使用
	}
}

func (this *BloomFilter) Contains(item []byte) bool {
	this.mu.RLock()
	defer this.mu.RUnlock()
	for _, h := range this.hashFuncs {
		h.Write(item)
		index := h.Sum(nil)[0] % uint8(this.m)
		if !this.bitset[index] {
			return false // 如果有一个位为false，则说明不包含
		}
		h.Reset() // 重置哈希函数以便下次使用
	}
	return true // 如果所有位都为true，则可能包含
}

func (this *BloomFilter) Reset() {
	this.mu.Lock()
	defer this.mu.Unlock()
	for i := range this.bitset {
		this.bitset[i] = false // 重置位数组
	}
	for _, h := range this.hashFuncs {
		h.Reset() // 重置所有哈希函数
	}
}
