package initialize

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"golang-learn/internal/config"
)

// InitRedis 初始化 Redis 连接（带 Ping 探活，确保连接可用）
func InitRedis(cfg *config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// 探活重试（Docker 环境下 Redis 可能稍后就绪）
	var err error
	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err = rdb.Ping(ctx).Err()
		cancel()
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("连接 Redis 失败: %w", err)
	}
	return rdb, nil
}
