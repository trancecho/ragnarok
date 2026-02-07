package maplist

// 线程不安全，使用的时候需要注意

type MapStringList[T any] struct {
	List []*T
	Map  map[string]*T
}

func NewMapStringList[T any]() MapStringList[T] {
	return MapStringList[T]{
		List: make([]*T, 0),
		Map:  make(map[string]*T),
	}
}

func (ml *MapStringList[T]) Add(key string, value *T) {
	ml.List = append(ml.List, value)
	ml.Map[key] = value
}

func (ml *MapStringList[T]) Get(key string) (*T, bool) {
	val, ok := ml.Map[key]
	return val, ok
}

func (ml *MapStringList[T]) GetList() []*T {
	return ml.List
}

func (ml *MapStringList[T]) GetMap() map[string]*T {
	return ml.Map
}

func (ml *MapStringList[T]) Size() int {
	return len(ml.List)
}

func (ml *MapStringList[T]) Has(key string) bool {
	_, ok := ml.Map[key]
	return ok
}

func (ml *MapStringList[T]) Remove(key string) {
	delete(ml.Map, key)
	for i, v := range ml.List {
		// 这里假设 T 有一个字段 Key 用于唯一标识
		// 需要根据实际情况修改
		if vKey, ok := any(v).(interface{ GetKey() string }); ok {
			if vKey.GetKey() == key {
				ml.List = append(ml.List[:i], ml.List[i+1:]...)
				break
			}
		}
	}
}

func (ml *MapStringList[T]) Clear() {
	ml.List = make([]*T, 0)
	ml.Map = make(map[string]*T)
}

func (ml *MapStringList[T]) IsEmpty() bool {
	return len(ml.List) == 0
}

func (ml *MapStringList[T]) Keys() []string {
	keys := make([]string, 0, len(ml.Map))
	for k := range ml.Map {
		keys = append(keys, k)
	}
	return keys
}
