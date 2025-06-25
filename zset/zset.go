package zset

import "math/rand/v2"

const (
	maxLevel    = 32
	probability = 0.25 // 跳跃表的概率因子，决定新节点的层数
)

type zskiplistNode struct {
	ele      string
	score    float64
	backward *zskiplistNode
	level    []struct {
		forward *zskiplistNode
		span    uint
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
}

func (this *ZSet) ZRem(ele string) bool {
	score, ok := this.dict[ele]
	if !ok {
		return false
	}
	updatePosNodes := make([]*zskiplistNode, maxLevel)
	x := this.skiplist.header
	for i := this.skiplist.level - 1; i >= 0; i-- {
		for nxt := x.level[i].forward; nxt != nil; {
			if nxt.score < score || (nxt.score == score && nxt.ele < ele) {
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
	//TODO implement me
	panic("implement me")
}

func (this *ZSet) ZRank(ele string) (uint, bool) {
	//TODO implement me
	panic("implement me")
}

func (this *ZSet) ZRevRank(ele string) (uint, bool) {
	//TODO implement me
	panic("implement me")
}

func (this *ZSet) ZRange(start, stop int) []string {
	//TODO implement me
	panic("implement me")
}

func (this *ZSet) ZRevRange(start, stop int) []string {
	//TODO implement me
	panic("implement me")
}

type IZSet interface {
	ZAdd(ele string, score float64) bool // 添加元素到有序集合
	ZRem(ele string) bool
	ZScore(ele string) (float64, bool)  // 获取元素的分数
	ZRank(ele string) (uint, bool)      // 获取元素的排名
	ZRevRank(ele string) (uint, bool)   // 获取元素的逆序排名
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
			span    uint // 前向指针和跨度
		}, level),
	}
	return node
}

func newSkipList() *zskiplist {
	zsl := &zskiplist{
		length: 0,
		header: newSkipListNode(maxLevel, 0, ""),
		tail:   nil,
		level:  maxLevel,
	}
	for i := 0; i < maxLevel; i++ {
		zsl.header.level[i].forward = nil
		zsl.header.level[i].span = 0
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

// 是否更新
func (this *ZSet) ZAdd(ele string, score float64) bool {
	if old, ok := this.dict[ele]; ok {
		if old == score {
			return false
		}
		this.ZRem(ele)
	}
	this.dict[ele] = score

	updatePosNodes := make([]*zskiplistNode, maxLevel) // 更新路径，用于确定每层要插入的位置（的前驱节点）

	rank := make([]uint, maxLevel) // 在链表中的排名

	// 顶层开始遍历（记录一下插入需要的数据，如前驱节点）
	x := this.skiplist.header
	for i := this.skiplist.level - 1; i >= 0; i-- {
		if i == this.skiplist.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		for nxt := x.level[i].forward; nxt != nil; {
			if nxt.score < score || (nxt.score == score && nxt.ele < ele) { // 分数主导，元素次导
				rank[i] += x.level[i].span // todo 干嘛
				x = nxt
				nxt = x.level[i].forward
			}
		}
		updatePosNodes[i] = x // 记录更新点
	}
	// 现在已经找到数据链表的位置
	curLevel := randomLevel() // 随机层数
	if curLevel > this.skiplist.level {
		for i := this.skiplist.level; i < curLevel; i++ {
			updatePosNodes[i] = this.skiplist.header
			this.skiplist.header.level[i].span = 0
		}
		this.skiplist.level = curLevel
	}

	// 创建新节点
	x = newSkipListNode(curLevel, score, ele)
	for i := 0; i < curLevel; i++ { // 说明层数是随机，并不是直接到顶的。
		x.level[i].forward = updatePosNodes[i].level[i].forward
		updatePosNodes[i].level[i].forward = x

		x.level[i].span = updatePosNodes[i].level[i].span - (rank[0] - rank[i])
		updatePosNodes[i].level[i].span = rank[0] + 1 - rank[i]
	}
	for i := curLevel; i < this.skiplist.level; i++ {
		updatePosNodes[i].level[i].span++
	}
	// 更新后向指针
	if updatePosNodes[0] != this.skiplist.header {
		x.backward = updatePosNodes[0] // backward指向原始链表的前驱节点
	} else {
		x.backward = nil
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		this.skiplist.tail = x
	}
	this.skiplist.length++
	return true
}
