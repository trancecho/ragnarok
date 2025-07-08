package stack

import (
	"log"
	"testing"
)

func TestStack(t *testing.T) {
	t.Run("stack test", func(t *testing.T) {
		s := NewStack[int]()
		if !s.IsEmpty() {
			t.Error("创建栈应为空")
		}
		if s.Size() != 0 {
			t.Error("栈大小应为0，但实际为", s.Size())
		}

		//测试Pop空栈
		_, err := s.Pop()
		if err == nil {
			t.Error("Pop空栈应返回错误")
		}
		// 测试Peek空栈
		_, err = s.Peek()
		if err == nil {
			t.Error("Peek空栈应返回错误")
		}
		// 测试Push
		s.Push(1)
		if s.IsEmpty() {
			t.Error("Push后栈不应为空")
		}
		s.Push(2)
		s.Push(3)
		//测试PeeK
		item, err := s.Peek()
		if err != nil {
			t.Error("Peek不应返回错误，但实际返回:", err)
		}
		if *item != 3 {
			t.Error("Peek应返回栈顶元素3，但实际返回:", *item)
		}
		// 测试Pop
		item, err = s.Pop()
		if err != nil {
			t.Error("Pop不应返回错误，但实际返回:", err)
		}
		if *item != 3 {
			t.Error("Pop应返回栈顶元素3，但实际返回:", *item)
		}
		log.Println(s.Peek())
		//测试Clear
		s.Clear()
		if !s.IsEmpty() {
			t.Error("Clear后栈应为空")
		}
	})
}
