package zset

import (
	"fmt"
	"math/rand/v2"
	"sync"
)

const (
	maxLevel    = 32
	probability = 0.25 // 跳跃表的概率因子，决定新节点的层数
)

type zskiplistNode struct {
	ele      string
	score    float64
	backward *zskiplistNode // 直接指向前一个原始链表节点
	level    []struct {
		forward *zskiplistNode //每层一个forward
		span    int
	}
}

type zskiplist struct {
	header *zskiplistNode
	tail   *zskiplistNode
	length int
	level  int // 跳跃表的最大层数
}

type ZSet struct {
	dict     map[string]float64 // 元素到分数的映射
	skiplist *zskiplist         // 跳跃表
	mu       sync.RWMutex
}

func (this *ZSet) ZRem(ele string) bool {
	this.mu.Lock()
	defer this.mu.Unlock()
	score, ok := this.dict[ele]
	if !ok {
		return false
	}
	updatePosNodes := make([]*zskiplistNode, maxLevel)
	x := this.skiplist.header
	for i := this.skiplist.level - 1; i >= 0; i-- {
		for nxt := x.level[i].forward; nxt != nil; {
			if nxt.score > score || (nxt.score == score && nxt.ele > ele) {
				x = nxt
				nxt = x.level[i].forward // todo 为什么forward指向整个node
			} else {
				break
			}
		}
		updatePosNodes[i] = x
	}
	x = x.level[0].forward // 找到要删除的节点
	if x != nil && x.score == score && x.ele == ele {
		zslDeleteNode(this.skiplist, x, updatePosNodes)
		delete(this.dict, ele) // 从字典中删除
		return true
	}
	return false
}

func zslDeleteNode(zsl *zskiplist, x *zskiplistNode, updatePosNodes []*zskiplistNode) {
	// 更新前节点
	for i := 0; i < len(x.level); i++ {
		if updatePosNodes[i].level[i].forward == x {
			updatePosNodes[i].level[i].span += x.level[i].span - 1 // rank为什么不是一个一直维持的值？因为删除会影响所有排名。而用span就可以很好计算排名
			updatePosNodes[i].level[i].forward = x.level[i].forward
		} else {
			updatePosNodes[i].level[i].span-- // 如果不是直接指向x，说明x在这个层级上并不存在
		}
	}
	//更新后节点
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x.backward
	} else {
		zsl.tail = x.backward
	}
	//可能最高节点被删，那么
	for zsl.level > 1 && zsl.header.level[zsl.level-1].forward == nil {
		zsl.level--
	}
	zsl.length-- // 跳跃表长度减一
}

func (this *ZSet) ZScore(ele string) (float64, bool) {
	this.mu.RLock()
	defer this.mu.RUnlock()
	score, ok := this.dict[ele]
	return score, ok
}

func (this *ZSet) ZRank(ele string) (int, bool) {
	this.mu.RLock()
	defer this.mu.RUnlock()
	score, ok := this.dict[ele]
	if !ok {
		return -1, false
	}
	rank := int(0)
	x := this.skiplist.header
	for i := this.skiplist.level - 1; i >= 0; i-- {
		for nxt := x.level[i].forward; nxt != nil; {
			if nxt.score > score || nxt.score == score && nxt.ele > ele {
				rank += x.level[i].span
				x = nxt
				nxt = x.level[i].forward
			} else {
				break
			}
		}
	}
	if x.level[0].forward.ele == ele {
		return rank, true
	}
	return -1, false // 如果没有找到，返回-1和false
}

func (this *ZSet) ZRevRank(ele string) (int, bool) {
	rank, ok := this.ZRank(ele)
	if !ok {
		return 0, false
	}
	return this.skiplist.length - 1 - rank, true
}

func (this *ZSet) ZRange(start, stop int) []string {
	this.mu.RLock()
	defer this.mu.RUnlock()
	if start > stop || start >= this.skiplist.length {
		return nil
	}
	if stop >= this.skiplist.length {
		stop = this.skiplist.length - 1
	}
	res := make([]string, stop-start+1)
	curSpan := 0
	x := this.skiplist.header
	for i := this.skiplist.level - 1; i >= 0; i-- {
		for nxt := x.level[i].forward; nxt != nil; {
			if curSpan+x.level[i].span > start {
				break
			}
			curSpan += x.level[i].span
			x = nxt
			nxt = x.level[i].forward
		}
	}
	for i := 0; i <= stop-start && x != nil; i++ {
		x = x.level[0].forward
		res[i] = fmt.Sprintf("%s:%.2f", x.ele, x.score)
	}
	return res
}

func (this *ZSet) ZRevRange(start, stop int) []string {
	this.mu.RLock()
	defer this.mu.RUnlock()
	if start > stop || start >= this.skiplist.length {
		return nil
	}
	if stop >= this.skiplist.length {
		stop = this.skiplist.length - 1
	}
	res := make([]string, stop-start+1)
	x := this.skiplist.tail
	for i := int(0); i < start; i++ {
		if x.backward == nil {
			return nil // 如果没有足够的元素
		}
		x = x.backward
	}
	for i := int(0); i <= stop-start && x != nil; i++ {
		res[i] = fmt.Sprintf("%s:%.2f", x.ele, x.score)
		x = x.backward
	}
	return res
}

type IZSet interface {
	ZAdd(ele string, score float64) bool // 添加元素到有序集合
	ZRem(ele string) bool
	ZScore(ele string) (float64, bool)  // 获取元素的分数
	ZRank(ele string) (int, bool)       // 获取元素的排名
	ZRevRank(ele string) (int, bool)    // 获取元素的逆序排名
	ZRange(start, stop int) []string    // 获取指定范围内的元素
	ZRevRange(start, stop int) []string // 获取指定范围内的元素（逆序）
}

var _ IZSet = (*ZSet)(nil) // 确保 ZSet 实现了 IZSet 接口

func newSkipListNode(level int, score float64, ele string) *zskiplistNode {
	node := &zskiplistNode{
		ele:   ele,
		score: score,
		level: make([]struct {
			forward *zskiplistNode
			span    int // 前向指针和跨度
		}, level),
	}
	return node
}

func newSkipList() *zskiplist {
	zsl := &zskiplist{
		length: 0,
		header: newSkipListNode(maxLevel, 0, ""),
		tail:   nil,
		level:  0,
	}
	for i := 0; i < maxLevel; i++ {
		zsl.header.level[i].forward = nil
		zsl.header.level[i].span = 1
	}
	zsl.header.backward = nil
	return zsl
}

func NewZSet() *ZSet {
	return &ZSet{
		dict:     make(map[string]float64),
		skiplist: newSkipList(),
	}
}

func randomLevel() int {
	level := 1
	for level < maxLevel && rand.Float64() < probability {
		level++
	}
	return level
}

// todo 复盘一下span
func (this *ZSet) ZAdd(ele string, score float64) bool {
	this.mu.Lock()
	if old, ok := this.dict[ele]; ok {
		if old == score {
			return false
		}
		this.mu.Unlock()
		this.ZRem(ele)
		this.mu.Lock()
	}
	this.dict[ele] = score

	updatePosNodes := make([]*zskiplistNode, maxLevel)
	rank := make([]int, maxLevel) // 记录 header 到每层 update 节点的跨度

	x := this.skiplist.header
	for i := this.skiplist.level - 1; i >= 0; i-- {
		if i == this.skiplist.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}
		for x.level[i].forward != nil &&
			(x.level[i].forward.score > score ||
				(x.level[i].forward.score == score && x.level[i].forward.ele > ele)) {
			rank[i] += x.level[i].span
			x = x.level[i].forward
		}
		updatePosNodes[i] = x
	}

	curLevel := randomLevel()
	if curLevel > this.skiplist.level {
		for i := this.skiplist.level; i < curLevel; i++ {
			updatePosNodes[i] = this.skiplist.header
			updatePosNodes[i].level[i].span = this.skiplist.length + 1
			rank[i] = 0
		}
		this.skiplist.level = curLevel
	}

	x = newSkipListNode(curLevel, score, ele)
	for i := 0; i < curLevel; i++ {
		x.level[i].forward = updatePosNodes[i].level[i].forward
		updatePosNodes[i].level[i].forward = x

		x.level[i].span = updatePosNodes[i].level[i].span - (rank[0] - rank[i])
		updatePosNodes[i].level[i].span = (rank[0] - rank[i]) + 1
	}
	for i := curLevel; i < this.skiplist.level; i++ {
		updatePosNodes[i].level[i].span++
	}

	if updatePosNodes[0] != this.skiplist.header {
		x.backward = updatePosNodes[0]
	} else {
		x.backward = nil
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		this.skiplist.tail = x
	}
	this.skiplist.length++
	this.mu.Unlock()
	return true
}

func (this *ZSet) Print() {
	this.mu.RLock()
	defer this.mu.RUnlock()
	fmt.Println("==== Skip List ====")
	for i := this.skiplist.level - 1; i >= 0; i-- {
		fmt.Printf("Level %d: ", i)
		p := this.skiplist.header.level[i].forward
		fmt.Printf("%v -> ", this.skiplist.header.level[i].span)
		for p != nil {
			fmt.Printf("%s:%.2f:%v -> ", p.ele, p.score, p.level[i].span)
			p = p.level[i].forward
		}
		fmt.Println("nil")
	}
	fmt.Println("===================")
}
