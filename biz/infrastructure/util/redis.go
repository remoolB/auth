package util

import (
	"auth/biz/infrastructure/config"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	redisOnce   sync.Once
	redisError  error
)

// GetRedisClient 获取Redis客户端连接
func GetRedisClient() (*redis.Client, error) {
	redisOnce.Do(func() {
		// 获取配置
		conf := config.GetConfig().Redis

		// 创建Redis客户端
		redisClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", conf.Host, conf.Port),
			Password: conf.Password,
			DB:       conf.DB,
		})

		// 检查连接
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := redisClient.Ping(ctx).Result()
		if err != nil {
			redisError = fmt.Errorf("redis连接失败: %w", err)
			fmt.Println("Redis连接失败:", err)
		}
	})

	return redisClient, redisError
}

// SetWithExpire 设置键值对，带过期时间
func SetWithExpire(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	client, err := GetRedisClient()
	if err != nil {
		return err
	}
	return client.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func Get(ctx context.Context, key string) (string, error) {
	client, err := GetRedisClient()
	if err != nil {
		return "", err
	}
	return client.Get(ctx, key).Result()
}

// Del 删除键
func Del(ctx context.Context, key string) error {
	client, err := GetRedisClient()
	if err != nil {
		return err
	}
	return client.Del(ctx, key).Err()
}

// Exists 检查键是否存在
func Exists(ctx context.Context, key string) (bool, error) {
	client, err := GetRedisClient()
	if err != nil {
		return false, err
	}
	result, err := client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// GetTTL 获取键的剩余过期时间
func GetTTL(ctx context.Context, key string) (time.Duration, error) {
	client, err := GetRedisClient()
	if err != nil {
		return 0, err
	}
	return client.TTL(ctx, key).Result()
}

// Incr 自增值，如果键不存在则创建
func Incr(ctx context.Context, key string) (int64, error) {
	client, err := GetRedisClient()
	if err != nil {
		return 0, err
	}
	return client.Incr(ctx, key).Result()
}

// Expire 设置过期时间
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	client, err := GetRedisClient()
	if err != nil {
		return err
	}
	return client.Expire(ctx, key, expiration).Err()
}

// Ttl 获取键的剩余过期时间（为了统一命名风格）
func Ttl(ctx context.Context, key string) (time.Duration, error) {
	return GetTTL(ctx, key)
}

// IsRedisNil 检查是否是Redis的Nil错误(键不存在)
func IsRedisNil(err error) bool {
	return errors.Is(err, redis.Nil)
}
