## 关于Ragnarok的Stack数据结构

**Stack**（栈）是一个后进先出（LIFO）的数据结构，支持基本的入栈、出栈和查看栈顶元素操作。
**下面列出当前Stack满足的方法**
- `Push(value T)`: 将元素压入栈顶。
- `Pop() (T, error)`: 弹出栈顶元素，返回值和错误。
- `Peek() (T, error)`: 查看栈顶元素，但不弹出。
- `Size() int`: 获取栈的当前大小。
- `IsEmpty() bool`: 检查栈是否为空。
- `Clear()`: 清空栈。

**亮点**:支持泛型类型，使用简单