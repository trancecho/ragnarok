package rlog

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level = zapcore.Level

const (
	DebugLevel = zapcore.DebugLevel
	InfoLevel  = zapcore.InfoLevel
	ErrorLevel = zapcore.ErrorLevel
	FatalLevel = zapcore.FatalLevel
)

type Config struct {
	Mode         string `json:"mode"` // "dev" or "prod"
	Level        Level  `json:"level"`
	LogFile      string `json:"log_file"`
	EnableFile   bool   `json:"enable_file"` // 本地日志开关
	SentryDSN    string `json:"sentry_dsn"`
	EnableSentry bool   `json:"enable_sentry"` // Sentry开关
}

var logger *zap.Logger

func Init(cfg Config) bool {
	// 创建日志目录
	if cfg.EnableFile && cfg.LogFile != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.LogFile), 0755); err != nil {
			log.Fatalln(err)
		}
	}

	// 初始化 Sentry
	if cfg.EnableSentry && cfg.SentryDSN != "" {
		err := sentry.Init(sentry.ClientOptions{Dsn: cfg.SentryDSN})
		if err != nil {
			log.Fatalln(err)
		}
	}

	var cores []zapcore.Core

	// 控制台输出 - dev模式全打印，prod模式不打印debug
	consoleLevel := cfg.Level
	if cfg.Mode == "prod" && consoleLevel == DebugLevel {
		consoleLevel = InfoLevel
	}

	consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
	cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), consoleLevel))

	// 文件输出 - 全打印
	if cfg.EnableFile && cfg.LogFile != "" {
		fileEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		})

		fileWriter := getFileWriter(cfg)
		cores = append(cores, zapcore.NewCore(fileEncoder, zapcore.AddSync(fileWriter), DebugLevel))
	}

	core := zapcore.NewTee(cores...)
	logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(ErrorLevel))
	return true
}

// 简洁的日志接口 - 删除Warn级别
func Debugln(msg string, fields ...zap.Field) { logger.Debug(msg, fields...) }
func Println(msg string, fields ...zap.Field) { logger.Info(msg, fields...) }
func Errln(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
	sendToSentry(ErrorLevel, msg, fields...)
}
func Fatalln(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
	sendToSentry(FatalLevel, msg, fields...)
}

// 便捷字段构造器
func String(key, val string) zap.Field                 { return zap.String(key, val) }
func Int(key string, val int) zap.Field                { return zap.Int(key, val) }
func Float64(key string, val float64) zap.Field        { return zap.Float64(key, val) }
func Bool(key string, val bool) zap.Field              { return zap.Bool(key, val) }
func Any(key string, val interface{}) zap.Field        { return zap.Any(key, val) }
func Err(err error) zap.Field                          { return zap.Error(err) }
func Duration(key string, val time.Duration) zap.Field { return zap.Duration(key, val) }

// 获取原始logger
func Logger() *zap.Logger { return logger }

// 时间格式化
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("15:04:05.000"))
}

// 发送到Sentry - 只处理Error和Fatal级别
func sendToSentry(level Level, msg string, fields ...zap.Field) {
	if level < ErrorLevel {
		return
	}

	extra := make(map[string]interface{})
	for _, field := range fields {
		switch field.Type {
		case zapcore.StringType:
			extra[field.Key] = field.String
		case zapcore.Int64Type:
			extra[field.Key] = field.Integer
		case zapcore.Float64Type:
			extra[field.Key] = field.Interface
		default:
			extra[field.Key] = field.Interface
		}
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetExtras(extra)
		switch level {
		case ErrorLevel:
			sentry.CaptureMessage(msg)
		case FatalLevel:
			sentry.CaptureException(fmt.Errorf(msg))
		}
	})
}

// 文件写入器
func getFileWriter(cfg Config) *os.File {
	file, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("open log file failed: %v", err))
	}
	return file
}
