package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dovetaill/article-sentinel/pkg/config"
)

func TestLoadOverridesSecretsFromFiles(t *testing.T) {
	clearLegacyDatabaseEnv(t)

	dir := t.TempDir()
	legacySecretFile := writeSecretFile(t, dir, "legacy-secret.txt", "file-legacy-secret\n")
	sessionSecretFile := writeSecretFile(t, dir, "session-secret.txt", "file-session-secret\n")
	mysqlPasswordFile := writeSecretFile(t, dir, "mysql-password.txt", "file-mysql-password\n")
	redisPasswordFile := writeSecretFile(t, dir, "redis-password.txt", "file-redis-password\n")

	path := writeConfigFile(t, `
app:
  name: article-sentinel
database:
  driver: mysql
  mysql:
    host: 127.0.0.1
    user: root
    password: inline-mysql-password
    password_file: ` + mysqlPasswordFile + `
    dbname: article_sentinel
redis:
  addr: 127.0.0.1:6379
  password: inline-redis-password
  password_file: ` + redisPasswordFile + `
auth:
  session:
    legacy_secret: inline-legacy-secret
    legacy_secret_file: ` + legacySecretFile + `
    secret: inline-session-secret
    secret_file: ` + sessionSecretFile + `
docs:
  enabled: false
log:
  level: info
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got := cfg.Auth.Session.LegacySecret; got != "file-legacy-secret" {
		t.Fatalf("Auth.Session.LegacySecret = %q, want %q", got, "file-legacy-secret")
	}
	if got := cfg.Auth.Session.Secret; got != "file-session-secret" {
		t.Fatalf("Auth.Session.Secret = %q, want %q", got, "file-session-secret")
	}
	if got := cfg.Database.MySQL.Password; got != "file-mysql-password" {
		t.Fatalf("Database.MySQL.Password = %q, want %q", got, "file-mysql-password")
	}
	if got := cfg.Redis.Password; got != "file-redis-password" {
		t.Fatalf("Redis.Password = %q, want %q", got, "file-redis-password")
	}
}

func TestLoadReturnsErrorForEmptySecretFile(t *testing.T) {
	clearLegacyDatabaseEnv(t)

	dir := t.TempDir()
	emptySecretFile := writeSecretFile(t, dir, "empty-secret.txt", "\n")

	path := writeConfigFile(t, `
app:
  name: article-sentinel
database:
  driver: mysql
  mysql:
    host: 127.0.0.1
    user: root
    password: inline-mysql-password
    password_file: ` + emptySecretFile + `
    dbname: article_sentinel
redis:
  addr: 127.0.0.1:6379
docs:
  enabled: false
log:
  level: info
`)

	if _, err := config.Load(path); err == nil {
		t.Fatal("Load() error = nil, want non-nil")
	}
}

func writeSecretFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
	return path
}
