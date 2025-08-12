# RLog 日志库使用手册

简洁高效的 Go 日志库，基于 zap 和 Sentry 构建。

## 特性

- **模式控制**: dev/prod 模式自动调整日志级别
- **多输出**: 控制台 + 文件 + Sentry
- **开关控制**: 文件日志和 Sentry 独立开关
- **四级日志**: Debug、Info、Error、Fatal

## 快速开始

```go
import "your-project/server/pkg/ragnarok/rlog"

// 配置初始化
cfg := rlog.Config{
    Mode:         "dev",           // "dev" 或 "prod"
    LogFile:      "/var/log/app.log",
    EnableFile:   true,            // 开启文件日志
    SentryDSN:    "your-sentry-dsn",
    EnableSentry: true,            // 开启 Sentry
    EnableMysql:  true,            // 开启 MySQL 日志
}

rlog.Init(cfg)

// 使用日志
rlog.Debugln("调试信息", rlog.String("key", "value"))
rlog.Println("普通信息", rlog.Int("count", 100))
rlog.Errln("错误信息", rlog.Err(err))
rlog.Fatalln("致命错误", rlog.Any("data", obj))
```

## 模式差异

| 模式 | 控制台输出 | 文件输出 | Sentry |
|------|------------|----------|--------|
| dev  | 全部级别   | 全部级别 | Error/Fatal |
| prod | Info及以上 | 全部级别 | Error/Fatal |

## 字段构造器

```go
rlog.String("key", "value")      // 字符串
rlog.Int("count", 42)            // 整数
rlog.Float64("score", 3.14)      // 浮点数
rlog.Bool("success", true)       // 布尔值
rlog.Err(err)                    // 错误
rlog.Duration("elapsed", time.Second) // 时长
rlog.Any("data", obj)            // 任意对象
```

## 配置说明

```go
type Config struct {
    Mode         string // "dev" 或 "prod"
    LogFile      string // 日志文件路径
    EnableFile   bool   // 是否启用文件日志
    SentryDSN    string // Sentry DSN
    EnableSentry bool   // 是否启用 Sentry
    EnableMysql  bool   // 是否启用 MySQL 日志
}
```

## 高级用法

```go
// 获取原始 zap.Logger
logger := rlog.Logger()
logger.With(rlog.String("module", "auth")).Info("模块日志")
```

## 注意事项

- Error/Fatal 级别自动上报 Sentry
- 文件日志采用 JSON 格式
- 控制台日志带彩色输出
- Fatal 级别会终止程序 