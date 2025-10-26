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

func (this *RandomSecretManager) KeepSecrets(ctx context.Context, secrets []string) (err error) {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				{
					for i := range secrets {
						// 随机生成六位数
						rand.Seed(time.Now().UnixNano())  // 每次运行随机不同
						num := rand.Intn(900000) + 100000 // 生成 100000 ~ 999999
						secret := fmt.Sprintf("%d", num)
						this.rdb.Set(ctx, "ragnarok:secret:"+secrets[i], secret, 11*time.Minute)
					}
				}
			case <-ctx.Done():
				{
					for i := range secrets {
						this.rdb.Del(ctx, secrets[i])
					}
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
	secret, err := this.rdb.Get(ctx, secretKey).Result()
	if err != nil {
		return "", err
	}
	return secret, nil
}
