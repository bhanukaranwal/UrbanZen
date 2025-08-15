package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/bhanukaranwal/urbanzen/internal/config"
)

type RedisDB struct {
	*redis.Client
}

func NewRedis(cfg *config.Config) (*RedisDB, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Database.Redis.Host, cfg.Database.Redis.Port),
		Password: cfg.Database.Redis.Password,
		DB:       cfg.Database.Redis.DB,
		PoolSize: 20,
		MinIdleConns: 5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisDB{rdb}, nil
}

func (r *RedisDB) SetEX(key, value string, expiration time.Duration) error {
	ctx := context.Background()
	return r.Client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisDB) Get(key string) (string, error) {
	ctx := context.Background()
	return r.Client.Get(ctx, key).Result()
}