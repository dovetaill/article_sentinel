package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

// Load 从 YAML 文件读取配置，并允许环境变量覆盖字段值。
func Load(path string) (*Config, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("config path is required")
	}

	cfg := &Config{
		Database: DatabaseConfig{
			MySQL: MySQLConfig{ParseTime: true},
		},
		Docs: DocsConfig{Enabled: true},
		Log:  LogConfig{RotateDaily: true},
	}
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}
	if err := validatePrimaryDatabaseConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validatePrimaryDatabaseConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("config is required")
	}

	switch strings.ToLower(strings.TrimSpace(cfg.Database.Driver)) {
	case "", "mysql":
		return validateRequiredFields("database.mysql", map[string]string{
			"host":     cfg.Database.MySQL.Host,
			"user":     cfg.Database.MySQL.User,
			"password": cfg.Database.MySQL.Password,
			"dbname":   cfg.Database.MySQL.DBName,
		})
	case "postgres":
		return validateRequiredFields("database.postgres", map[string]string{
			"host":     cfg.Database.Postgres.Host,
			"user":     cfg.Database.Postgres.User,
			"password": cfg.Database.Postgres.Password,
			"dbname":   cfg.Database.Postgres.DBName,
		})
	default:
		return fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}
}

func validateRequiredFields(prefix string, fields map[string]string) error {
	missing := make([]string, 0, len(fields))
	for name, value := range fields {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, prefix+"."+name)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("missing required database config: %s", strings.Join(missing, ", "))
}
