package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	client *redis.Client
}

// NewRedisStorage creates a new Redis storage instance
func NewRedisStorage(addr, password string, db int) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStorage{client: client}, nil
}

func (r *RedisStorage) GetRequestCount(ctx context.Context, key string) (int64, error) {
	count, err := r.client.Get(ctx, fmt.Sprintf("count:%s", key)).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

func (r *RedisStorage) IncrementRequestCount(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	countKey := fmt.Sprintf("count:%s", key)
	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, countKey)
	pipe.Expire(ctx, countKey, expiration)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	
	return incr.Val(), nil
}

func (r *RedisStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
	exists, err := r.client.Exists(ctx, fmt.Sprintf("blocked:%s", key)).Result()
	return exists == 1, err
}

func (r *RedisStorage) Block(ctx context.Context, key string, duration time.Duration) error {
	return r.client.Set(ctx, fmt.Sprintf("blocked:%s", key), 1, duration).Err()
}

func (r *RedisStorage) Close() error {
	return r.client.Close()
}