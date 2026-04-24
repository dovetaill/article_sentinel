package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dovetaill/article-sentinel/pkg/config"
	"github.com/dovetaill/article-sentinel/pkg/database"
	"gorm.io/driver/sqlite"
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
	origApplySQLMigrations := applySQLMigrationsFn
	origAutoMigrate := autoMigrateBusinessTablesFn
	t.Cleanup(func() {
		loadConfigFn = origLoadConfig
		bootstrapDatabaseFn = origBootstrapDatabase
		applySQLMigrationsFn = origApplySQLMigrations
		autoMigrateBusinessTablesFn = origAutoMigrate
	})

	bootstrapCalls := 0
	callOrder := make([]string, 0, 2)
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
	applySQLMigrationsFn = func(db *gorm.DB, sourceURL string) error {
		callOrder = append(callOrder, "sql:"+sourceURL)
		return nil
	}
	autoMigrateBusinessTablesFn = func(migrator schemaMigrator) error {
		callOrder = append(callOrder, "auto")
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
	if got, want := strings.Join(callOrder, ","), "sql:file://migrations,auto"; got != want {
		t.Fatalf("call order = %q, want %q", got, want)
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

func TestApplySQLMigrationsExecutesStatementsFromFiles(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "migrate.db")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "001_create_widgets.sql"), []byte(`
-- create table
CREATE TABLE widgets (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL
);
`), 0o644); err != nil {
		t.Fatalf("write create migration: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "002_seed_widgets.sql"), []byte(`
-- seed row
INSERT INTO widgets (id, name)
VALUES (1, 'alpha');
`), 0o644); err != nil {
		t.Fatalf("write seed migration: %v", err)
	}

	if err := ApplySQLMigrations(db, "file://"+dir); err != nil {
		t.Fatalf("ApplySQLMigrations() error = %v", err)
	}

	var got struct {
		ID   int
		Name string
	}
	if err := db.Table("widgets").First(&got, "id = ?", 1).Error; err != nil {
		t.Fatalf("load seeded widget: %v", err)
	}
	if got.Name != "alpha" {
		t.Fatalf("seeded widget name = %q, want %q", got.Name, "alpha")
	}
}
