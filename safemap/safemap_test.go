package safemap_test

import (
	"github.com/trancecho/ragnarok/safemap"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
)

func TestSafeMapBasicOperations(t *testing.T) {
	var m safemap.SafeMap[string, int]

	// Load on empty map
	if v, ok := m.Load("missing"); ok {
		t.Errorf("Load on empty map returned value %v, want not found", v)
	}

	// Store and Load
	m.Store("key1", 100)
	if v, ok := m.Load("key1"); !ok || v != 100 {
		t.Errorf("Load after Store got (%v, %v), want (100, true)", v, ok)
	}

	// LoadOrStore returns existing value if present
	actual, loaded := m.LoadOrStore("key1", 200)
	if !loaded || actual != 100 {
		t.Errorf("LoadOrStore existing key returned (%v, %v), want (100, true)", actual, loaded)
	}
	// LoadOrStore stores new value if absent
	actual, loaded = m.LoadOrStore("key2", 300)
	if loaded || actual != 300 {
		t.Errorf("LoadOrStore new key returned (%v, %v), want (300, false)", actual, loaded)
	}

	// LoadAndDelete returns old value and removes key
	val, loaded := m.LoadAndDelete("key1")
	if !loaded || val != 100 {
		t.Errorf("LoadAndDelete returned (%v, %v), want (100, true)", val, loaded)
	}
	// Subsequent Load should fail
	if _, ok := m.Load("key1"); ok {
		t.Errorf("Load after LoadAndDelete returned found key")
	}

	// Delete on existing and missing keys (no panic)
	m.Delete("key2")
	m.Delete("missing")

	// Swap returns previous value and updates
	m.Store("key3", 123)
	prev, loaded := m.Swap("key3", 456)
	if !loaded || prev != 123 {
		t.Errorf("Swap returned (%v, %v), want (123, true)", prev, loaded)
	}
	v, ok := m.Load("key3")
	if !ok || v != 456 {
		t.Errorf("Load after Swap returned (%v, %v), want (456, true)", v, ok)
	}
	// Swap on missing key returns zero and false
	prev, loaded = m.Swap("missing", 999)
	if loaded || prev != 0 {
		t.Errorf("Swap on missing returned (%v, %v), want (0, false)", prev, loaded)
	}

	// CompareAndSwap success and failure
	m.Store("key4", 10)
	swapped := m.CompareAndSwap("key4", 10, 20)
	if !swapped {
		t.Errorf("CompareAndSwap failed but should succeed")
	}
	swapped = m.CompareAndSwap("key4", 10, 30)
	if swapped {
		t.Errorf("CompareAndSwap succeeded but should fail")
	}

	// CompareAndDelete success and failure
	m.Store("key5", 50)
	deleted := m.CompareAndDelete("key5", 50)
	if !deleted {
		t.Errorf("CompareAndDelete failed but should succeed")
	}
	deleted = m.CompareAndDelete("key5", 50)
	if deleted {
		t.Errorf("CompareAndDelete succeeded but key already deleted")
	}

	// Clear removes all keys
	m.Store("key6", 60)
	m.Store("key7", 70)
	m.Clear()
	if _, ok := m.Load("key6"); ok {
		t.Errorf("Clear did not remove key6")
	}
	if _, ok := m.Load("key7"); ok {
		t.Errorf("Clear did not remove key7")
	}
}

func TestSafeMapRange(t *testing.T) {
	var m safemap.SafeMap[int, string]
	const itemCount = 1000

	// Prepare map with values
	for i := 0; i < itemCount; i++ {
		m.Store(i, "v"+strconv.Itoa(i))
	}

	count := 0
	keysSeen := make(map[int]struct{})
	m.Range(func(k int, v string) bool {
		if _, exists := keysSeen[k]; exists {
			t.Errorf("Range visited key %v twice", k)
		}
		keysSeen[k] = struct{}{}
		expected := "v" + strconv.Itoa(k)
		if v != expected {
			t.Errorf("Range key %v got value %v, want %v", k, v, expected)
		}
		count++
		return true
	})

	if count != itemCount {
		t.Errorf("Range visited %d items, want %d", count, itemCount)
	}
}

func TestSafeMapRangeEarlyStop(t *testing.T) {
	var m safemap.SafeMap[string, int]
	m.Store("a", 1)
	m.Store("b", 2)

	count := 0
	m.Range(func(k string, v int) bool {
		count++
		return false // stop after first
	})
	if count != 1 {
		t.Errorf("Range early stop visited %d items, want 1", count)
	}
}

func TestSafeMapConcurrentAccess(t *testing.T) {
	var m safemap.SafeMap[int, int]
	var wg sync.WaitGroup
	const goroutines = 50
	const keysPerGoroutine = 1000

	// Writer goroutines
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			for i := 0; i < keysPerGoroutine; i++ {
				m.Store(base+i, base+i)
			}
		}(g * keysPerGoroutine)
	}

	// Reader goroutines
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			for i := 0; i < keysPerGoroutine; i++ {
				m.Load(base + i)
			}
		}(g * keysPerGoroutine)
	}

	wg.Wait()

	// Verify keys exist
	total := goroutines * keysPerGoroutine
	for i := 0; i < total; i++ {
		v, ok := m.Load(i)
		if !ok || v != i {
			t.Errorf("Concurrent test key %d missing or wrong value %v", i, v)
		}
	}
}

func TestSafeMapRangeUpdate(t *testing.T) {
	var m safemap.SafeMap[int, int]
	for i := 0; i < 1000; i++ {
		m.Store(i, i)
	}

	// 更新所有偶数 key 的值，值加 1000
	m.RangeUpdate(func(k int, old int) (int, bool) {
		if k%2 == 0 {
			return old + 1000, true
		}
		return old, false
	})

	// 校验偶数键的值被更新
	m.Range(func(k int, v int) bool {
		if k%2 == 0 {
			want := k + 1000
			if v != want {
				t.Errorf("RangeUpdate: key %d got %d, want %d", k, v, want)
			}
		} else {
			if v != k {
				t.Errorf("RangeUpdate: odd key %d changed to %d", k, v)
			}
		}
		return true
	})
}

func TestSafeMapRangeUpdateConcurrency(t *testing.T) {
	var m safemap.SafeMap[int, int]
	const N = 10000
	for i := 0; i < N; i++ {
		m.Store(i, i)
	}

	var updates int64

	m.RangeUpdate(func(k int, old int) (int, bool) {
		atomic.AddInt64(&updates, 1)
		return old + 1, true
	})

	if updates != int64(N) {
		t.Errorf("RangeUpdate concurrency: updateFn called %d times, want %d", updates, N)
	}

	// 验证所有值都被加一
	m.Range(func(k int, v int) bool {
		if v != k+1 {
			t.Errorf("RangeUpdate concurrency: key %d got value %d, want %d", k, v, k+1)
		}
		return true
	})
}

func BenchmarkSafeMap_Store(b *testing.B) {
	var m safemap.SafeMap[int, int]
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Store(i, i)
			i++
		}
	})
}

func BenchmarkSafeMap_Load(b *testing.B) {
	var m safemap.SafeMap[int, int]
	const N = 100000
	for i := 0; i < N; i++ {
		m.Store(i, i)
	}
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Load(i % N)
			i++
		}
	})
}

func BenchmarkSafeMap_LoadOrStore(b *testing.B) {
	var m safemap.SafeMap[int, int]
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.LoadOrStore(i, i)
			i++
		}
	})
}

func BenchmarkSafeMap_RangeUpdate(b *testing.B) {
	var m safemap.SafeMap[int, int]
	const N = 100000
	for i := 0; i < N; i++ {
		m.Store(i, i)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.RangeUpdate(func(k int, old int) (int, bool) {
			return old + 1, true
		})
	}
}

func BenchmarkSafeMap_RangeUpdate_Parallel(b *testing.B) {
	var m safemap.SafeMap[int, int]
	const N = 100000
	for i := 0; i < N; i++ {
		m.Store(i, i)
	}
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.RangeUpdate(func(k int, old int) (int, bool) {
				return old + 1, true
			})
		}
	})
}

//运行go test -v测试
//运行go test -bench=. -cpuprofile=cpu.prof
//若生成文件为cpu，重命名为cpu.prof
//运行go tool pprof cpu.prof，用top等指令查看参数
