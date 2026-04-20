package bootstrap

import (
	"strings"
	"testing"

	"github.com/dovetaill/article-sentinel/pkg/config"
	"github.com/dovetaill/article-sentinel/pkg/database"
	"gorm.io/gorm"
)

func TestBuildMigrateConfigUsesPrimaryDatabaseConfig(t *testing.T) {
	tests := []struct {
		name       string
		cfg        *config.Config
		wantDriver string
		wantURL    string
	}{
		{
			name: "mysql",
			cfg: &config.Config{
				Database: config.DatabaseConfig{
					Driver: "mysql",
					MySQL: config.MySQLConfig{
						Host:      "127.0.0.1",
						Port:      3306,
						User:      "root",
						Password:  "root",
						DBName:    "article_sentinel",
						Charset:   "utf8mb4",
						ParseTime: true,
						Loc:       "Local",
					},
				},
			},
			wantDriver: "mysql",
			wantURL:    "mysql://root:root@tcp(127.0.0.1:3306)/article_sentinel?charset=utf8mb4&loc=Local&parseTime=true",
		},
		{
			name: "postgres",
			cfg: &config.Config{
				Database: config.DatabaseConfig{
					Driver: "postgres",
					Postgres: config.PostgresConfig{
						Host:     "127.0.0.1",
						Port:     5432,
						User:     "postgres",
						Password: "secret",
						DBName:   "article_sentinel",
						SSLMode:  "disable",
						TimeZone: "Asia/Shanghai",
					},
				},
			},
			wantDriver: "postgres",
			wantURL:    "postgres://postgres:secret@127.0.0.1:5432/article_sentinel?TimeZone=Asia%2FShanghai&sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildMigrateConfig(tt.cfg)
			if err != nil {
				t.Fatalf("BuildMigrateConfig() error = %v", err)
			}
			if got.Driver != tt.wantDriver {
				t.Fatalf("Driver = %q, want %q", got.Driver, tt.wantDriver)
			}
			if got.SourceURL != "file://migrations" {
				t.Fatalf("SourceURL = %q, want %q", got.SourceURL, "file://migrations")
			}
			if got.DatabaseURL != tt.wantURL {
				t.Fatalf("DatabaseURL = %q, want %q", got.DatabaseURL, tt.wantURL)
			}
		})
	}
}

func TestRunMigrateCommandRunsStarterSchemaSync(t *testing.T) {
	origLoadConfig := loadConfigFn
	origBootstrapDatabase := bootstrapDatabaseFn
	origAutoMigrate := autoMigrateBusinessTablesFn
	t.Cleanup(func() {
		loadConfigFn = origLoadConfig
		bootstrapDatabaseFn = origBootstrapDatabase
		autoMigrateBusinessTablesFn = origAutoMigrate
	})

	bootstrapCalls := 0
	autoMigrateCalls := 0
	loadConfigFn = func(path string) (*config.Config, error) {
		return &config.Config{
			Database: config.DatabaseConfig{
				Driver: "mysql",
				MySQL: config.MySQLConfig{
					Host:      "127.0.0.1",
					Port:      3306,
					User:      "root",
					Password:  "root",
					DBName:    "article_sentinel",
					Charset:   "utf8mb4",
					ParseTime: true,
					Loc:       "Local",
				},
			},
			Redis: config.RedisConfig{Addr: "127.0.0.1:6379"},
		}, nil
	}
	bootstrapDatabaseFn = func(cfg *config.Config) (*database.Resources, error) {
		bootstrapCalls++
		return &database.Resources{DB: &gorm.DB{}}, nil
	}
	autoMigrateBusinessTablesFn = func(migrator schemaMigrator) error {
		autoMigrateCalls++
		return nil
	}

	if err := RunMigrateCommand("configs/config.yaml"); err != nil {
		t.Fatalf("RunMigrateCommand() error = %v", err)
	}
	if bootstrapCalls != 1 {
		t.Fatalf("bootstrap database call count = %d, want %d", bootstrapCalls, 1)
	}
	if autoMigrateCalls != 1 {
		t.Fatalf("auto migrate call count = %d, want %d", autoMigrateCalls, 1)
	}
}

func TestMigrateCommandRejectsUnsupportedDriver(t *testing.T) {
	origLoadConfig := loadConfigFn
	t.Cleanup(func() {
		loadConfigFn = origLoadConfig
	})

	loadConfigFn = func(path string) (*config.Config, error) {
		return &config.Config{
			Database: config.DatabaseConfig{Driver: "sqlite"},
		}, nil
	}

	err := RunMigrateCommand("configs/config.yaml")
	if err == nil {
		t.Fatal("RunMigrateCommand() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "unsupported database driver") {
		t.Fatalf("error = %v, want contains %q", err, "unsupported database driver")
	}
}
