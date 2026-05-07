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

	// 这里集中放启动默认值，避免业务代码散落兜底逻辑。
	cfg := &Config{
		Docs: DocsConfig{Enabled: true},
		Auth: AuthConfig{
			Session: SessionConfig{
				SecureCookie: true,
			},
		},
		Log: LogConfig{RotateDaily: true},
	}
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}
	if err := validateAuthSessionConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validateAuthSessionConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("config is required")
	}

	missing := make([]string, 0, 2)
	if strings.TrimSpace(cfg.Auth.Session.LegacySecret) == "" {
		missing = append(missing, "auth.session.legacy_secret")
	}
	if strings.TrimSpace(cfg.Auth.Session.Secret) == "" {
		missing = append(missing, "auth.session.secret")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required auth session config: %s", strings.Join(missing, ", "))
	}

	return nil
}
