package util

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"
)

// GenerateUUIDv7 生成 UUIDv7 (时间排序的 UUID)
// UUIDv7 格式：
//   - 48 位 Unix 时间戳（毫秒）
//   - 12 位随机数（亚毫秒精度）
//   - 2 位版本号（0b0111）
//   - 62 位随机数
func GenerateUUIDv7() string {
	// 获取当前时间戳（毫秒）
	now := time.Now()
	timestamp := uint64(now.UnixMilli())

	var uuid [16]byte

	// 前 6 字节：时间戳（48位）
	binary.BigEndian.PutUint64(uuid[0:8], timestamp<<16)

	// 生成随机数填充剩余部分
	randBytes := make([]byte, 10)
	_, err := rand.Read(randBytes)
	if err != nil {
		// 如果随机数生成失败，使用时间戳的纳秒部分
		nano := uint64(now.Nanosecond())
		binary.BigEndian.PutUint64(randBytes[0:8], nano)
	}
	copy(uuid[6:], randBytes)

	// 设置版本号（version 7）
	uuid[6] = (uuid[6] & 0x0f) | 0x70

	// 设置变体（variant 2, RFC 4122）
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	// 格式化为标准 UUID 字符串
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(uuid[0:4]),
		binary.BigEndian.Uint16(uuid[4:6]),
		binary.BigEndian.Uint16(uuid[6:8]),
		binary.BigEndian.Uint16(uuid[8:10]),
		uuid[10:16],
	)
}

// MustGenerateUUIDv7 生成 UUIDv7，出错时 panic
func MustGenerateUUIDv7() string {
	uuid := GenerateUUIDv7()
	if uuid == "" {
		panic("failed to generate UUIDv7")
	}
	return uuid
}

// ParseUUIDv7Timestamp 从 UUIDv7 中解析时间戳
func ParseUUIDv7Timestamp(uuidStr string) (time.Time, error) {
	// 移除连字符
	cleaned := ""
	for _, ch := range uuidStr {
		if ch != '-' {
			cleaned += string(ch)
		}
	}

	if len(cleaned) != 32 {
		return time.Time{}, fmt.Errorf("invalid UUID format")
	}

	// 解析前 12 个十六进制字符（48位时间戳）
	var timestamp uint64
	_, err := fmt.Sscanf(cleaned[0:12], "%012x", &timestamp)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	// 转换为时间
	return time.UnixMilli(int64(timestamp)), nil
}
