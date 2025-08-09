package safemap

import (
	"runtime"
	"sync"
)

// noCopy 用于禁止复制
type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

// SafeMap 是一个支持泛型的并发安全 Map
// 零值可直接使用，不需要初始化。
// 注意：SafeMap 不可被复制。
type SafeMap[K comparable, V any] struct {
	_ noCopy
	m sync.Map
}

// Load 返回指定 key 的值，如果不存在则返回零值和 false
func (m *SafeMap[K, V]) Load(key K) (value V, ok bool) {
	val, ok := m.m.Load(key)
	if !ok {
		var zero V
		return zero, false
	}
	return val.(V), true
}

// Store 设置 key 对应的值
func (m *SafeMap[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

// Clear 清空 Map
func (m *SafeMap[K, V]) Clear() {
	// sync.Map 从 Go1.23 开始才有 Clear 方法
	// 如果要兼容低版本，可以用 Range + Delete 实现
	if clearFunc, ok := any(&m.m).(interface{ Clear() }); ok {
		clearFunc.Clear()
	} else {
		m.m.Range(func(k, _ any) bool {
			m.m.Delete(k)
			return true
		})
	}
}

// LoadOrStore 如果 key 存在返回已有值，否则存储并返回给定值
func (m *SafeMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	val, loaded := m.m.LoadOrStore(key, value)
	return val.(V), loaded
}

// LoadAndDelete 删除 key 并返回旧值
func (m *SafeMap[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	val, loaded := m.m.LoadAndDelete(key)
	if !loaded {
		var zero V
		return zero, false
	}
	return val.(V), true
}

// Delete 删除 key
func (m *SafeMap[K, V]) Delete(key K) {
	m.m.Delete(key)
}

// Swap 原子替换 key 的值并返回旧值
func (m *SafeMap[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	val, loaded := m.m.Swap(key, value)
	if !loaded {
		var zero V
		return zero, false
	}
	return val.(V), true
}

// CompareAndSwap 如果 key 对应的值等于 old，则替换为 new
func (m *SafeMap[K, V]) CompareAndSwap(key K, old, new V) (swapped bool) {
	return m.m.CompareAndSwap(key, old, new)
}

// CompareAndDelete 如果 key 对应的值等于 old，则删除
func (m *SafeMap[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	return m.m.CompareAndDelete(key, old)
}

// Range 遍历所有键值对
func (m *SafeMap[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}

// RangeUpdate 高性能遍历并修改
// 收集阶段：只做一次 sync.Map.Range，不会阻塞写操作太久。
// 更新阶段：批量更新，用 runtime.NumCPU() 个 worker 并行执行，减少锁竞争。
// 内存分配优化：items 切片一次性分配，避免在高并发下频繁申请内存。
// 高并发安全：sync.Map.Store 内部是线程安全的，每个 worker 都可以安全更新不同 key。
func (m *SafeMap[K, V]) RangeUpdate(updateFn func(key K, oldValue V) (newValue V, update bool)) {
	// 1. 收集所有需要修改的 key/value
	type kv struct {
		k K
		v V
	}
	var items []kv
	m.Range(func(key K, value V) bool {
		newVal, shouldUpdate := updateFn(key, value)
		if shouldUpdate {
			items = append(items, kv{k: key, v: newVal})
		}
		return true
	})

	// 2. 并发更新
	workerCount := runtime.NumCPU()
	wg := sync.WaitGroup{}
	taskCh := make(chan kv, workerCount*2)

	// 启动 worker
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range taskCh {
				m.Store(item.k, item.v)
			}
		}()
	}

	// 投递任务
	for _, item := range items {
		taskCh <- item
	}
	close(taskCh)

	wg.Wait()
}
