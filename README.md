# Ragnarok 诸神黄昏 🚀  

**Go 泛型化高性能数据结构与工具库** | All-in-One | 简洁 | 高性能 | 广适配

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 📦 功能模块

### 数据结构

#### 基础数据结构
- **链表**（`list/`）：单/双向链表，支持泛型
- **无锁链表**（`lockfreelist/`）：并发安全的无锁实现
- **栈**（`stack/`）：泛型栈实现
- **堆**（`heap/`）：优先队列、堆排序
- **B+ 树**（`bptree/`）：支持范围查询的索引结构
- **跳表/ZSet**（`zset/`）：类 Redis ZSET 实现，支持排序和范围查询
- **布隆过滤器**（`bloom_filter/`）：高效的存在性判断
- **滑动窗口**（`rollingwindows/`）：时间窗口统计

### 工具模块（`util/`）

#### 配置初始化
- **`InitViper()`**：自动读取配置文件（dev/prod 模式切换）
- **`InitMysql(models...)`**：MySQL 连接 + 自动迁移
- **`InitRedis()`**：Redis 客户端初始化（带连接池和超时配置）
- **`InitNats()`**：NATS JetStream 客户端初始化
- **`InitClickHouse(models...)`**：ClickHouse 连接 + 表验证

#### 数据模型
- **`BaseModel`**：通用 GORM 模型基类（ID, CreatedAt, UpdatedAt, DeletedAt）

#### HTTP 工具
- **`RespSuccess()` / `RespError()`**：统一 JSON 响应格式
- **`GetUID()`**：从 Gin Context 获取用户 ID
- **`GetIP()`**：获取真实客户端 IP

#### 分页工具
- **`PageRequest` / `PageResponse`**：统一的分页请求/响应结构
- **`Paginate()`**：GORM 分页查询辅助函数

#### 其他工具
- **`SafeGet*()`**：安全地从 map 获取值，避免 panic
- **`GenerateSecret()`**：生成随机密钥

### 第三方服务集成

- **FastGPT**（`fastgpt/`）：FastGPT API 客户端
- **MinIO**（`rminio/`）：对象存储客户端封装
- **NATS**（`rnats/`）：NATS JetStream 客户端封装
- **日志**（`rlog/`）：结构化日志工具

### 数据库增强（`rrdb/`）

- **`RandomSecret`**：生成随机密钥
- **`Stream`**：数据库流式处理

## 🚀 快速开始

### 安装

```bash
go get github.com/trancecho/ragnarok
```

### 基础使用示例

#### 1. 配置初始化

```go
package main

import (
    "github.com/trancecho/ragnarok/util"
)

func main() {
    // 初始化配置（自动读取 config.dev.yaml 或 config.prod.yaml）
    util.InitViper()
    
    // 初始化 MySQL
    db := util.InitMysql(&User{}, &Post{})
    
    // 初始化 Redis
    rdb := util.InitRedis()
    
    // 初始化 NATS
    nats := util.InitNats()
    
    // 初始化 ClickHouse
    clickhouse := util.InitClickHouse(&Event{})
}
```



#### 2. 数据结构使用

**布隆过滤器：**

```go
import "github.com/trancecho/ragnarok/bloom_filter"

filter := bloom_filter.NewBloomFilter(1000, 0.01)
filter.Add("user_123")

if filter.Contains("user_123") {
    // 可能存在（有误判率）
}
```

**跳表/ZSet：**

```go
import "github.com/trancecho/ragnarok/zset"

zs := zset.New()
zs.Add("member1", 100.0)
zs.Add("member2", 200.0)

rank := zs.Rank("member1") // 获取排名
members := zs.RangeByRank(0, 10) // 获取前 10 名
```

**无锁链表（并发安全）：**

```go
import "github.com/trancecho/ragnarok/lockfreelist"

list := lockfreelist.New[int]()
list.PushBack(1)
list.PushBack(2)
val, ok := list.PopFront()
```

#### 3. HTTP 工具使用

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/trancecho/ragnarok/util"
)

func GetUserHandler(c *gin.Context) {
    // 获取当前用户 ID
    uid := util.GetUID(c)
    
    // 获取真实 IP
    ip := util.GetIP(c)
    
    // 返回成功响应
    util.RespSuccess(c, gin.H{
        "uid": uid,
        "ip": ip,
    })
}

func ErrorHandler(c *gin.Context) {
    // 返回错误响应
    util.RespError(c, 400, "Invalid request")
}
```

#### 4. 分页查询

```go
import (
    "github.com/trancecho/ragnarok/util"
    "gorm.io/gorm"
)

func GetUsers(db *gorm.DB, page, pageSize int) util.PageResponse[User] {
    var users []User
    var total int64
    
    query := db.Model(&User{})
    
    // 使用分页工具
    result := util.Paginate(query, page, pageSize, &users, &total)
    
    return result
}
```

#### 5. NATS JetStream 使用

```go
import (
    "context"
    "github.com/trancecho/ragnarok/rnats"
)

func main() {
    client, _ := rnats.NewClient(rnats.Config{
        URL: "nats://localhost:4222",
    })
    defer client.Close()
    
    // 发布消息
    client.Publish(context.Background(), "events.user.login", map[string]any{
        "user_id": 123,
        "ip": "192.168.1.1",
    })
    
    // 订阅消息
    client.Subscribe(context.Background(), "events.>", "consumer-1", func(data []byte) error {
        // 处理消息
        return nil
    })
}
```

## 📚 详细文档

- [布隆过滤器使用指南](doc/bloom_filter.md)
- [B+ 树使用指南](doc/bptree.md)
- [堆/优先队列](heap/README.md)
- [栈实现](stack/README.md)
- [链表实现](list/readme.md)
- [无锁链表](lockfreelist/readme.md)
- [日志工具](rlog/README.md)

## 🎯 设计理念

### All-in-One
提供从基础数据结构到第三方服务集成的全套工具，减少项目中的重复代码。

### 简洁实用
- 统一的初始化函数（`Init*`）
- 统一的响应格式（`Resp*`）
- 统一的数据模型（`BaseModel`）
- 简洁的 API 设计

### 高性能
- 泛型支持，避免类型断言开销
- 针对 Go 并发模型优化
- 无锁数据结构实现
- 连接池和超时控制

### 广适配
- 支持主流数据库：MySQL、ClickHouse
- 支持主流缓存：Redis
- 支持消息队列：NATS JetStream
- 支持对象存储：MinIO
- 支持 AI 服务：FastGPT

## 🛠 工具函数速查

### 配置与初始化
```go
util.InitViper()                    // 配置文件初始化
util.InitMysql(models...)           // MySQL 初始化 + 迁移
util.InitRedis()                    // Redis 初始化
util.InitNats()                     // NATS 初始化
util.InitClickHouse(models...)      // ClickHouse 初始化
```

### HTTP 响应
```go
util.RespSuccess(c, data)           // 成功响应
util.RespError(c, code, msg)        // 错误响应
util.GetUID(c)                      // 获取用户 ID
util.GetIP(c)                       // 获取客户端 IP
```

### 分页
```go
util.Paginate(query, page, size, &items, &total)  // 分页查询
util.NewPageRequest(page, size)                   // 创建分页请求
```

### 安全访问
```go
util.SafeGetString(m, key, defaultVal)    // 安全获取 string
util.SafeGetInt(m, key, defaultVal)       // 安全获取 int
util.SafeGetBool(m, key, defaultVal)      // 安全获取 bool
```

## 📈 性能对比

### 布隆过滤器
- 内存占用：比标准 map 节省 **90%+**
- 查询速度：O(k)，k 为哈希函数个数

### 无锁链表
- 并发写入：比带锁链表快 **2-3 倍**
- 适用场景：高并发读写、生产者-消费者模式

### ZSet（跳表）
- 插入/查询：O(log n)
- 范围查询：O(log n + m)，m 为返回元素个数

## 🤝 贡献

欢迎提交 PR 或 Issue！

### 贡献指南
1. Fork 本仓库
2. 创建特性分支：`git checkout -b feature/amazing-feature`
3. 提交更改：`git commit -m 'Add amazing feature'`
4. 推送到分支：`git push origin feature/amazing-feature`
5. 提交 Pull Request

### 开发规范
- 代码需要通过 `go test ./...`
- 新增功能需要添加测试用例
- 遵循 Go 代码规范

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 🌟 Star History

如果这个项目对你有帮助，请给个 Star ⭐️！

---

**目标**：打造 Go 生态最实用的数据结构与工具库！  
**理念**：All-in-One | 简洁 | 高性能 | 广适配
