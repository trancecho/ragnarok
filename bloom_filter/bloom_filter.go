package bloom_filter

type BloomFilter struct {
	bitset      []uint8 // 位数组
	size        int     // 位数组大小
	hashFuncNum int     // 哈希函数数量
}
