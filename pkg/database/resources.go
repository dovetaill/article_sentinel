package database

import (
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
