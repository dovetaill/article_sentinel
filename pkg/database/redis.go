package database

import (
	"context"
	"fmt"
	"time"

	"github.com/dovetaill/article-sentinel/pkg/config"
	"github.com/redis/go-redis/v9"
)

func openRedis(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(buildRedisOptions(cfg))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}

func buildRedisOptions(cfg config.RedisConfig) *redis.Options {
	return &redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	}
}
