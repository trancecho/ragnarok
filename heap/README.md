## 关于Ragnarok的Heap数据结构

**Heap**（堆）是一种特殊的树形数据结构，满足堆性质：每个节点的值都大于或等于（最大堆）或小于或等于（最小堆）其子节点的值。Ragnarok 的堆实现支持以下操作：
- `Insert(value T)`: 将元素插入堆中，保持堆的性质。
- `Pop() (T, error)`: 弹出堆顶元素，返回值和错误。
- `Peek() (T, error)`: 查看堆顶元素，但不弹出。
- `Size() int`: 获取堆的当前大小。
- `IsEmpty() bool`: 检查堆是否为空。
- `Clear()`: 清空堆。
- `AdjustHeap()`: 自下而上建堆(调整)

**注意！！！！！**
**NewHeap的比较函数默认使用**
```go
package heap

func intLess(a, b int) bool { return a < b }
```
**如果想要大根堆就使用函数MaxHeap,如果想要小根堆就使用MinHeap,然后将函数初始化给compareFn**
**由于泛型化后无法进行比较所以比较麻烦，如果感觉这套流程麻烦，欢迎大佬来优化**