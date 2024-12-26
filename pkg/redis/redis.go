package redis

import (
	"campus2/pkg/config"
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient redis.UniversalClient
	redisOnce   sync.Once
)

// GetRedis 获取Redis客户端连接
func GetRedis(cfg config.Redis) redis.UniversalClient {
	redisOnce.Do(func() {
		if cfg.UseCluster {
			// 集群模式
			redisClient = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:    cfg.ClusterAddrs,
				Password: cfg.Password,
				PoolSize: cfg.PoolSize,
			})
		} else {
			// 单机模式
			redisClient = redis.NewClient(&redis.Options{
				Addr:     cfg.Addr,
				Password: cfg.Password,
				DB:       cfg.DB,
				PoolSize: cfg.PoolSize,
			})
		}

		// 测试连接
		ctx := context.Background()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			panic(fmt.Sprintf("Redis连接失败: %v", err))
		}
	})
	return redisClient
}
