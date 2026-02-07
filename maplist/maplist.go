package maplist

// 线程不安全，使用的时候需要注意

type MapList[T any] struct {
	List []*T
	Map  map[int64]*T
}

func NewMapList[T any]() MapList[T] {
	return MapList[T]{
		List: make([]*T, 0),
		Map:  make(map[int64]*T),
	}
}
func (ml *MapList[T]) Add(key int64, value *T) {
	ml.List = append(ml.List, value)
	ml.Map[key] = value
}
func (ml *MapList[T]) Get(key int64) (*T, bool) {
	val, ok := ml.Map[key]
	return val, ok
}

func (ml *MapList[T]) GetList() []*T {
	return ml.List
}
func (ml *MapList[T]) GetMap() map[int64]*T {
	return ml.Map
}
func (ml *MapList[T]) Size() int {
	return len(ml.List)
}
func (ml *MapList[T]) Has(key int64) bool {
	_, ok := ml.Map[key]
	return ok
}
func (ml *MapList[T]) Remove(key int64) {
	delete(ml.Map, key)
	for i, v := range ml.List {
		// 这里假设 T 有一个字段 Key 用于唯一标识
		// 需要根据实际情况修改
		if vKey, ok := any(v).(interface{ GetKey() int64 }); ok {
			if vKey.GetKey() == key {
				ml.List = append(ml.List[:i], ml.List[i+1:]...)
				break
			}
		}
	}
}
func (ml *MapList[T]) Clear() {
	ml.List = make([]*T, 0)
	ml.Map = make(map[int64]*T)
}
func (ml *MapList[T]) IsEmpty() bool {
	return len(ml.List) == 0
}
func (ml *MapList[T]) Keys() []int64 {
	keys := make([]int64, 0, len(ml.Map))
	for k := range ml.Map {
		keys = append(keys, k)
	}
	return keys
}
func (ml *MapList[T]) Values() []*T {
	return ml.List
}
