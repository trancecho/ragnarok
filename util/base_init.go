package util

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"github.com/trancecho/ragnarok/rnats"
	"gorm.io/driver/mysql"
)

func InitViper() {
	var mode string
	flag.StringVar(&mode, "mode", "dev", "运行模式")
	flag.Parse()
	if mode == "dev" {
		// 开发模式
		viper.SetConfigName("config.dev")
	} else if mode == "prod" {
		// 生产模式
		viper.SetConfigName("config.prod")
	} else {
		panic("未知的运行模式: " + mode)
	}
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic("读取配置文件失败: " + err.Error())
	}
}

func InitMysql(models ...any) *gorm.DB {
	var check []string
	check = []string{"mysql.user", "mysql.password", "mysql.host", "mysql.port", "mysql.db"}
	for _, s := range check {
		if !viper.IsSet(s) {
			log.Fatalf("配置项 %s 未设置，请检查配置文件", s)
		}
	}

	// MySQL 连接配置
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetInt("mysql.port"),
		viper.GetString("mysql.db"),
	)
	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// 自动迁移 (创建或更新表结构)
	err = db.AutoMigrate(models...)
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	return db
}

func InitRedis() *redis.Client {
	var check []string
	check = []string{"redis.addr", "redis.password", "redis.db"}
	for _, s := range check {
		if !viper.IsSet(s) {
			log.Fatalf("配置项 %s 未设置，请检查配置文件", s)
		}
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: viper.GetString("redis.password"), // 密码，如果没有设置则为空
		DB:       viper.GetInt("redis.db"),          // 数据库索引，默认为0
		PoolSize: 100,                               // 连接池大小

		// 连接超时设置
		DialTimeout:  5 * time.Second, // 连接建立超时时间
		ReadTimeout:  3 * time.Second, // 读超时
		WriteTimeout: 3 * time.Second, // 写超时
		PoolTimeout:  4 * time.Second, // 获取连接池连接超时时间

		ConnMaxIdleTime: 5 * time.Minute, // 连接最大空闲时间
	})

	// 测试连接是否正常
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
		return nil
	}
	return redisClient
}

func InitNats() *rnats.Client {
	var check []string
	check = []string{"nats.url"}
	for _, s := range check {
		if !viper.IsSet(s) {
			log.Fatalf("配置项 %s 未设置，请检查配置文件", s)
		}
	}

	client, err := rnats.NewClient(rnats.Config{
		URL:      viper.GetString("nats.url"),
		Username: viper.GetString("nats.username"),
		Password: viper.GetString("nats.password"),
		Name:     viper.GetString("nats.name"),
	})
	if err != nil {
		log.Fatalf("failed to connect to NATS: %v", err)
	}

	// 测试连接是否正常
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		log.Fatalf("failed to ping NATS: %v", err)
	}

	log.Println("NATS client initialized successfully")
	return client
}
