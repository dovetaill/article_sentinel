package database

import (
	"errors"
	"testing"

	"github.com/dovetaill/article-sentinel/pkg/config"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func TestBuildMySQLDSN(t *testing.T) {
	cfg := config.MySQLConfig{
		Host:                   "127.0.0.1",
		Port:                   3306,
		User:                   "root",
		Password:               "root",
		DBName:                 "article_sentinel",
		Charset:                "utf8mb4",
		ParseTime:              true,
		Loc:                    "Local",
		MaxOpenConns:           20,
		MaxIdleConns:           10,
		ConnMaxLifetimeMinutes: 60,
	}

	got := buildMySQLDSN(cfg)
	want := "root:root@tcp(127.0.0.1:3306)/article_sentinel?charset=utf8mb4&parseTime=true&loc=Local"
	if got != want {
		t.Fatalf("buildMySQLDSN() = %q, want %q", got, want)
	}
}

func TestBuildRedisOptions(t *testing.T) {
	cfg := config.RedisConfig{
		Addr:         "127.0.0.1:6379",
		Password:     "secret",
		DB:           2,
		PoolSize:     32,
		MinIdleConns: 8,
	}

	opts := buildRedisOptions(cfg)
	if opts.Addr != cfg.Addr {
		t.Fatalf("Addr = %q, want %q", opts.Addr, cfg.Addr)
	}
	if opts.Password != cfg.Password {
		t.Fatalf("Password = %q, want %q", opts.Password, cfg.Password)
	}
	if opts.DB != cfg.DB {
		t.Fatalf("DB = %d, want %d", opts.DB, cfg.DB)
	}
	if opts.PoolSize != cfg.PoolSize {
		t.Fatalf("PoolSize = %d, want %d", opts.PoolSize, cfg.PoolSize)
	}
	if opts.MinIdleConns != cfg.MinIdleConns {
		t.Fatalf("MinIdleConns = %d, want %d", opts.MinIdleConns, cfg.MinIdleConns)
	}
}

func TestResourcesCloseIsSafe(t *testing.T) {
	var nilResources *Resources
	if err := nilResources.Close(); err != nil {
		t.Fatalf("nil Close() error = %v", err)
	}

	emptyResources := &Resources{}
	if err := emptyResources.Close(); err != nil {
		t.Fatalf("empty Close() error = %v", err)
	}
}

func TestBootstrapUsesMySQLDialectorWhenDriverIsMySQL(t *testing.T) {
	origOpenMySQL := openMySQLFn
	origOpenPostgres := openPostgresFn
	origOpenRedis := openRedisFn
	t.Cleanup(func() {
		openMySQLFn = origOpenMySQL
		openPostgresFn = origOpenPostgres
		openRedisFn = origOpenRedis
	})

	mysqlCalled := false
	postgresCalled := false
	openMySQLFn = func(cfg config.MySQLConfig) (*gorm.DB, error) {
		mysqlCalled = true
		return &gorm.DB{}, nil
	}
	openPostgresFn = func(cfg config.PostgresConfig) (*gorm.DB, error) {
		postgresCalled = true
		return &gorm.DB{}, nil
	}
	openRedisFn = func(cfg config.RedisConfig) (*redis.Client, error) {
		return nil, nil
	}

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Driver: "mysql",
			MySQL: config.MySQLConfig{
				Host:   "127.0.0.1",
				User:   "root",
				DBName: "article_sentinel",
			},
		},
		Redis: config.RedisConfig{Addr: "127.0.0.1:6379"},
	}

	if _, err := Bootstrap(cfg); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	if !mysqlCalled {
		t.Fatal("openMySQLFn was not called")
	}
	if postgresCalled {
		t.Fatal("openPostgresFn was called, want not called")
	}
}

func TestBootstrapUsesPostgresDialectorWhenDriverIsPostgres(t *testing.T) {
	origOpenMySQL := openMySQLFn
	origOpenPostgres := openPostgresFn
	origOpenRedis := openRedisFn
	t.Cleanup(func() {
		openMySQLFn = origOpenMySQL
		openPostgresFn = origOpenPostgres
		openRedisFn = origOpenRedis
	})

	mysqlCalled := false
	postgresCalled := false
	openMySQLFn = func(cfg config.MySQLConfig) (*gorm.DB, error) {
		mysqlCalled = true
		return &gorm.DB{}, nil
	}
	openPostgresFn = func(cfg config.PostgresConfig) (*gorm.DB, error) {
		postgresCalled = true
		return &gorm.DB{}, nil
	}
	openRedisFn = func(cfg config.RedisConfig) (*redis.Client, error) {
		return nil, nil
	}

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Driver: "postgres",
			Postgres: config.PostgresConfig{
				Host:   "127.0.0.1",
				User:   "postgres",
				DBName: "article_sentinel",
			},
		},
		Redis: config.RedisConfig{Addr: "127.0.0.1:6379"},
	}

	if _, err := Bootstrap(cfg); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	if !postgresCalled {
		t.Fatal("openPostgresFn was not called")
	}
	if mysqlCalled {
		t.Fatal("openMySQLFn was called, want not called")
	}
}

func TestBuildPostgresDSN(t *testing.T) {
	cfg := config.PostgresConfig{
		Host:     "127.0.0.1",
		Port:     5432,
		User:     "postgres",
		Password: "secret",
		DBName:   "article_sentinel",
		SSLMode:  "disable",
		TimeZone: "Asia/Shanghai",
	}

	got := buildPostgresDSN(cfg)
	want := "host=127.0.0.1 port=5432 user=postgres password=secret dbname=article_sentinel sslmode=disable TimeZone=Asia/Shanghai"
	if got != want {
		t.Fatalf("buildPostgresDSN() = %q, want %q", got, want)
	}
}

func TestBootstrapReturnsErrorForUnsupportedDriver(t *testing.T) {
	origOpenRedis := openRedisFn
	t.Cleanup(func() {
		openRedisFn = origOpenRedis
	})
	openRedisFn = func(cfg config.RedisConfig) (*redis.Client, error) {
		return nil, errors.New("redis should not be called for unsupported driver")
	}

	cfg := &config.Config{
		Database: config.DatabaseConfig{Driver: "sqlite"},
		Redis:    config.RedisConfig{Addr: "127.0.0.1:6379"},
	}

	_, err := Bootstrap(cfg)
	if err == nil {
		t.Fatal("Bootstrap() error = nil, want non-nil")
	}
}
