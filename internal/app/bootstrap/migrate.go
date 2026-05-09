package bootstrap

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dovetaill/article-sentinel/pkg/config"
	"gorm.io/gorm"
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
	migrateCfg, err := BuildMigrateConfig(cfg)
	if err != nil {
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
	if err := applySQLMigrationsFn(resources.DB, migrateCfg.SourceURL); err != nil {
		return fmt.Errorf("apply sql migrations: %w", err)
	}
	if err := autoMigrateBusinessTablesFn(resources.DB); err != nil {
		return fmt.Errorf("auto migrate starter schema: %w", err)
	}

	return nil
}

func ApplySQLMigrations(db *gorm.DB, sourceURL string) error {
	if db == nil {
		return errors.New("database is required")
	}

	dir, err := migrationDirectoryFromSource(sourceURL)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migration directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration file %s: %w", entry.Name(), err)
		}

		statements := splitSQLStatements(string(content))
		for _, statement := range statements {
			if err := db.Exec(statement).Error; err != nil {
				return fmt.Errorf("execute migration statement from %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

func migrationDirectoryFromSource(sourceURL string) (string, error) {
	if !strings.HasPrefix(sourceURL, "file://") {
		return "", fmt.Errorf("unsupported migration source: %s", sourceURL)
	}

	dir := strings.TrimSpace(strings.TrimPrefix(sourceURL, "file://"))
	if dir == "" {
		return "", errors.New("migration directory is required")
	}
	return dir, nil
}

func splitSQLStatements(content string) []string {
	statements := make([]string, 0)
	var current strings.Builder

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "#") {
			continue
		}

		if current.Len() > 0 {
			current.WriteByte('\n')
		}
		current.WriteString(line)

		if strings.HasSuffix(line, ";") {
			statement := strings.TrimSpace(strings.TrimSuffix(current.String(), ";"))
			if statement != "" {
				statements = append(statements, statement)
			}
			current.Reset()
		}
	}

	if trailing := strings.TrimSpace(current.String()); trailing != "" {
		statements = append(statements, trailing)
	}

	return statements
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
