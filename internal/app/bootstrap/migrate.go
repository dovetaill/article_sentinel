package bootstrap

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/dovetaill/article-sentinel/pkg/config"
)

const migrationsSourceURL = "file://migrations"

// MigrateConfig 定义 migrate 命令运行所需的最小配置。
type MigrateConfig struct {
	Driver      string
	DatabaseURL string
	SourceURL   string
}

// BuildMigrateConfig 根据当前主库驱动生成 migrate 使用的连接配置。
func BuildMigrateConfig(cfg *config.Config) (*MigrateConfig, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}

	driver := normalizeDatabaseDriver(cfg.Database.Driver)
	switch driver {
	case "mysql":
		return &MigrateConfig{
			Driver:      driver,
			DatabaseURL: buildMySQLMigrateURL(cfg.Database.MySQL),
			SourceURL:   migrationsSourceURL,
		}, nil
	case "postgres":
		return &MigrateConfig{
			Driver:      driver,
			DatabaseURL: buildPostgresMigrateURL(cfg.Database.Postgres),
			SourceURL:   migrationsSourceURL,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}
}

// RunMigrateCommand 承载 starter schema sync 的执行流程。
func RunMigrateCommand(configPath string) error {
	cfg, err := loadConfigFn(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if _, err := BuildMigrateConfig(cfg); err != nil {
		return fmt.Errorf("build migrate config: %w", err)
	}

	resources, err := bootstrapDatabaseFn(cfg)
	if err != nil {
		return fmt.Errorf("bootstrap database resources: %w", err)
	}
	if resources != nil {
		defer func() {
			_ = resources.Close()
		}()
	}
	if resources == nil || resources.DB == nil {
		return errors.New("starter database is required")
	}
	if err := autoMigrateBusinessTablesFn(resources.DB); err != nil {
		return fmt.Errorf("auto migrate starter schema: %w", err)
	}

	return nil
}

func normalizeDatabaseDriver(driver string) string {
	driver = strings.TrimSpace(driver)
	if driver == "" {
		return "mysql"
	}
	return strings.ToLower(driver)
}

func buildMySQLMigrateURL(cfg config.MySQLConfig) string {
	query := url.Values{}
	query.Set("charset", cfg.Charset)
	query.Set("loc", cfg.Loc)
	query.Set("parseTime", strconv.FormatBool(cfg.ParseTime))

	return fmt.Sprintf(
		"mysql://%s@tcp(%s:%d)/%s?%s",
		url.UserPassword(cfg.User, cfg.Password).String(),
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		query.Encode(),
	)
}

func buildPostgresMigrateURL(cfg config.PostgresConfig) string {
	query := url.Values{}
	query.Set("sslmode", cfg.SSLMode)
	query.Set("TimeZone", cfg.TimeZone)

	return (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Path:     "/" + cfg.DBName,
		RawQuery: query.Encode(),
	}).String()
}
