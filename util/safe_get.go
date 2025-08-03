package util

import (
	"log"
	"reflect"
)

func SafeGet[T any](m map[string]interface{}, key string, defaultValue T) T {
	if val, exists := m[key]; exists {
		if typedVal, ok := val.(T); ok {
			return typedVal
		}
	}
	return defaultValue
}

func SafeAssert[T any](val any, defaultValue T) T {
	if typedVal, ok := val.(T); ok {
		return typedVal
	} else {
		log.Println("断言失败", val, reflect.TypeOf(val))
	}
	return defaultValue
}
