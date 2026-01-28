package rlog

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestSave2Mysql(t *testing.T) {
	var dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		"root", "123456", "localhost", "13306", "trace")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	msg := "unit test error"
	save2Mysql(
		db,
		ErrorLevel,
		msg,
		zap.String("user", "alice"),
		zap.Int("code", 500),
		zap.Bool("retry", true),
	)
}
