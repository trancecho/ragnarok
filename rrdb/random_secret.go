package rrdb

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

type RandomSecretManager struct {
	rdb *redis.Client
}

func NewRandomSecretManager(rdb *redis.Client) *RandomSecretManager {
	return &RandomSecretManager{rdb: rdb}
}

func (this *RandomSecretManager) KeepSecrets(ctx context.Context, secrets []string) (err error) {
	rand.Seed(time.Now().UnixNano())
	ticker := time.NewTicker(10 * time.Minute)
	generate := func() {
		for i := range secrets {
			// 随机生成六位数
			num := rand.Intn(900000) + 100000 // 生成 100000 ~ 999999
			secret := fmt.Sprintf("%d", num)
			this.rdb.Set(ctx, "ragnarok:secret:"+secrets[i], secret, 11*time.Minute)
		}
	}
	go func() {
		generate()
		for {
			select {
			case <-ticker.C:
				{
					generate()
				}
			case <-ctx.Done():
				{
					//for _, key := range secrets {
					//	delCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					//	if err := this.rdb.Del(delCtx, "ragnarok:secret:"+key).Err(); err != nil {
					//		log.Printf("Failed to delete secret: %v", err)
					//	}
					//	cancel()
					//}
					ticker.Stop()
					return
				}
			}
		}
	}()
	return nil
}

func (this *RandomSecretManager) GetSecret(secretKey string) (string, error) {
	ctx := context.Background()
	secret, err := this.rdb.Get(ctx, "ragnarok:secret:"+secretKey).Result()
	if err != nil {
		return "", err
	}
	return secret, nil
}
