package stack

import "errors"

type IStack[T any] interface {
	Push(v interface{})
	Pop() (*T, error)
	Peek() (*T, error)
	Size() int
	IsEmpty() bool
	Clear()
}

type Stack[T any] struct {
	items []T
}

// NewStack 初始化栈
func NewStack[T any]() *Stack[T] {
	return &Stack[T]{items: make([]T, 0)}
}

// Push 入栈
func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

// Pop 出栈
func (s *Stack[T]) Pop() (*T, error) {
	if len(s.items) == 0 {
		return nil, errors.New("empty stack")
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return &item, nil
}

// Peek 获取栈顶元素
func (s *Stack[T]) Peek() (*T, error) {
	if len(s.items) == 0 {
		return nil, errors.New("empty stack")
	}
	item := s.items[len(s.items)-1]
	return &item, nil
}

// Size 获取栈的大小
func (s *Stack[T]) Size() int {
	return len(s.items)
}

// Clear 清空栈
func (s *Stack[T]) Clear() {
	s.items = make([]T, 0)
}

// IsEmpty 检查栈是否为空
func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}
