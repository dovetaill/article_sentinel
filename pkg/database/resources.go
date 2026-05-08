package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/dovetaill/article-sentinel/pkg/config"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	openMySQLFn    = openMySQL
	openPostgresFn = openPostgres
	openRedisFn    = openRedis
)

// Resources 聚合 starter 初始化出来的基础设施客户端。
type Resources struct {
	DB    *gorm.DB
	Redis *redis.Client
}

// PingDB 对主数据库执行一次轻量探活，供启动和 readiness 复用。
func PingDB(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return errors.New("database is not configured")
	}
	if db.Config == nil {
		return errors.New("database is not initialized")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	return nil
}

// PingRedis 对 Redis 执行一次轻量探活，供启动和 readiness 复用。
func PingRedis(ctx context.Context, client *redis.Client) error {
	if client == nil {
		return errors.New("redis is not configured")
	}

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}

	return nil
}

// Bootstrap 按固定顺序初始化主数据库与 Redis，并在失败时回收已创建资源。
func Bootstrap(cfg *config.Config) (*Resources, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}

	resources := &Resources{}

	db, err := openPrimaryDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("bootstrap database: %w", err)
	}
	resources.DB = db

	redisClient, err := openRedisFn(cfg.Redis)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("bootstrap redis: %w", err), resources.Close())
	}
	resources.Redis = redisClient

	return resources, nil
}

// Close 安全释放已经初始化的资源。
func (r *Resources) Close() error {
	if r == nil {
		return nil
	}

	var err error
	if r.Redis != nil {
		err = errors.Join(err, r.Redis.Close())
	}

	if r.DB != nil {
		if r.DB.Config == nil {
			return err
		}
		sqlDB, dbErr := r.DB.DB()
		if dbErr != nil {
			err = errors.Join(err, dbErr)
		} else {
			err = errors.Join(err, sqlDB.Close())
		}
	}

	return err
}
