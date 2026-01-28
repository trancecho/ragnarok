package util

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateUUIDv7(t *testing.T) {
	uuid := GenerateUUIDv7()

	// 验证格式：8-4-4-4-12
	parts := strings.Split(uuid, "-")
	if len(parts) != 5 {
		t.Errorf("UUID should have 5 parts, got %d", len(parts))
	}

	if len(parts[0]) != 8 {
		t.Errorf("First part should be 8 chars, got %d", len(parts[0]))
	}
	if len(parts[1]) != 4 {
		t.Errorf("Second part should be 4 chars, got %d", len(parts[1]))
	}
	if len(parts[2]) != 4 {
		t.Errorf("Third part should be 4 chars, got %d", len(parts[2]))
	}
	if len(parts[3]) != 4 {
		t.Errorf("Fourth part should be 4 chars, got %d", len(parts[3]))
	}
	if len(parts[4]) != 12 {
		t.Errorf("Fifth part should be 12 chars, got %d", len(parts[4]))
	}

	t.Logf("Generated UUIDv7: %s", uuid)
}

func TestGenerateUUIDv7_Uniqueness(t *testing.T) {
	// 生成多个 UUID，确保不重复
	uuids := make(map[string]bool)
	count := 1000

	for i := 0; i < count; i++ {
		uuid := GenerateUUIDv7()
		if uuids[uuid] {
			t.Errorf("Duplicate UUID found: %s", uuid)
		}
		uuids[uuid] = true
	}

	if len(uuids) != count {
		t.Errorf("Expected %d unique UUIDs, got %d", count, len(uuids))
	}
}

func TestGenerateUUIDv7_Ordering(t *testing.T) {
	// UUIDv7 应该是时间排序的
	uuid1 := GenerateUUIDv7()
	time.Sleep(2 * time.Millisecond)
	uuid2 := GenerateUUIDv7()

	// 字典序比较，uuid2 应该大于 uuid1
	if uuid2 <= uuid1 {
		t.Errorf("UUIDv7 should be time-ordered, but %s is not greater than %s", uuid2, uuid1)
	}

	t.Logf("UUID1: %s", uuid1)
	t.Logf("UUID2: %s", uuid2)
}

func TestParseUUIDv7Timestamp(t *testing.T) {
	now := time.Now()
	uuid := GenerateUUIDv7()

	parsed, err := ParseUUIDv7Timestamp(uuid)
	if err != nil {
		t.Fatalf("Failed to parse timestamp: %v", err)
	}

	// 验证时间戳在合理范围内（误差 1 秒）
	diff := parsed.Sub(now).Abs()
	if diff > time.Second {
		t.Errorf("Timestamp difference too large: %v", diff)
	}

	t.Logf("Original time: %s", now.Format(time.RFC3339Nano))
	t.Logf("Parsed time:   %s", parsed.Format(time.RFC3339Nano))
	t.Logf("Difference:    %v", diff)
}

func TestMustGenerateUUIDv7(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustGenerateUUIDv7 should not panic with valid input")
		}
	}()

	uuid := MustGenerateUUIDv7()
	if uuid == "" {
		t.Errorf("MustGenerateUUIDv7 returned empty string")
	}
}

func BenchmarkGenerateUUIDv7(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateUUIDv7()
	}
}
