package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/dovetaill/article-sentinel/internal/identity"
	"github.com/dovetaill/article-sentinel/pkg/config"
)

func TestStarterConfigTypeShape(t *testing.T) {
	tests := []struct {
		name        string
		typ         reflect.Type
		field       string
		wantPresent bool
	}{
		{name: "config keeps database field", typ: reflect.TypeOf(config.Config{}), field: "Database", wantPresent: true},
		{name: "config drops legacy top level mysql field", typ: reflect.TypeOf(config.Config{}), field: "MySQL", wantPresent: false},
		{name: "auth drops seed admin field", typ: reflect.TypeOf(config.AuthConfig{}), field: "SeedAdmin", wantPresent: false},
		{name: "auth drops unused jwt field", typ: reflect.TypeOf(config.AuthConfig{}), field: "JWT", wantPresent: false},
		{name: "session drops configurable cookie name field", typ: reflect.TypeOf(config.SessionConfig{}), field: "CookieName", wantPresent: false},
		{name: "http exposes request timeout field", typ: reflect.TypeOf(config.HTTPConfig{}), field: "RequestTimeoutSeconds", wantPresent: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := tt.typ.FieldByName(tt.field)
			if ok != tt.wantPresent {
				t.Fatalf("field %s present = %t, want %t", tt.field, ok, tt.wantPresent)
			}
		})
	}
}

func TestStarterHTTPAndDocsConfigTagsAreExplicit(t *testing.T) {
	tests := []struct {
		name     string
		typ      reflect.Type
		field    string
		wantYAML string
		wantEnv  string
	}{
		{
			name:     "request timeout field uses starter tags",
			typ:      reflect.TypeOf(config.HTTPConfig{}),
			field:    "RequestTimeoutSeconds",
			wantYAML: "request_timeout_seconds",
			wantEnv:  "HTTP_REQUEST_TIMEOUT_SECONDS",
		},
		{
			name:     "docs enabled field is explicit",
			typ:      reflect.TypeOf(config.DocsConfig{}),
			field:    "Enabled",
			wantYAML: "enabled",
			wantEnv:  "DOCS_ENABLED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, ok := tt.typ.FieldByName(tt.field)
			if !ok {
				t.Fatalf("missing field %s", tt.field)
			}
			if got := field.Tag.Get("yaml"); got != tt.wantYAML {
				t.Fatalf("yaml tag = %q, want %q", got, tt.wantYAML)
			}
			if got := field.Tag.Get("env"); got != tt.wantEnv {
				t.Fatalf("env tag = %q, want %q", got, tt.wantEnv)
			}
		})
	}
}

func TestAuthConfigExposesSessionFields(t *testing.T) {
	typ := reflect.TypeOf(config.AuthConfig{})
	field, ok := typ.FieldByName("Session")
	if !ok {
		t.Fatal("AuthConfig.Session missing")
	}
	if field.Type != reflect.TypeOf(config.SessionConfig{}) {
		t.Fatalf("AuthConfig.Session type = %v", field.Type)
	}
}

func TestLoadReadsStarterPrimaryDatabaseConfig(t *testing.T) {
	clearLegacyDatabaseEnv(t)

	path := writeConfigFile(t, `
app:
  name: article-sentinel
database:
  driver: mysql
  mysql:
    host: 127.0.0.1
    port: 3306
    user: root
    password: root
    dbname: article_sentinel
    charset: utf8mb4
    parse_time: true
    loc: Local
    max_open_conns: 20
    max_idle_conns: 10
    conn_max_lifetime_minutes: 60
redis:
  addr: 127.0.0.1:6379
docs:
  enabled: false
log:
  level: info
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Database.Driver != "mysql" {
		t.Fatalf("Database.Driver = %q, want %q", cfg.Database.Driver, "mysql")
	}
	if cfg.Database.MySQL.DBName != "article_sentinel" {
		t.Fatalf("Database.MySQL.DBName = %q, want %q", cfg.Database.MySQL.DBName, "article_sentinel")
	}
	if cfg.Docs.Enabled {
		t.Fatal("Docs.Enabled = true, want false")
	}
}

func TestLoadReadsExplicitDocsAndRequestTimeoutConfig(t *testing.T) {
	clearLegacyDatabaseEnv(t)

	path := writeConfigFile(t, `
app:
  name: article-sentinel
http:
  request_timeout_seconds: 27
  trusted_proxy_cidrs:
    - 127.0.0.1/32
    - ::1/128
database:
  driver: mysql
  mysql:
    host: 127.0.0.1
    user: root
    password: root
    dbname: article_sentinel
redis:
  addr: 127.0.0.1:6379
docs:
  enabled: false
  openapi_path: /schema.json
  ui_path: /reference
log:
  level: info
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTP.RequestTimeoutSeconds != 27 {
		t.Fatalf("HTTP.RequestTimeoutSeconds = %d, want %d", cfg.HTTP.RequestTimeoutSeconds, 27)
	}
	if cfg.Docs.Enabled {
		t.Fatal("Docs.Enabled = true, want false")
	}
	if cfg.Docs.OpenAPIPath != "/schema.json" {
		t.Fatalf("Docs.OpenAPIPath = %q, want %q", cfg.Docs.OpenAPIPath, "/schema.json")
	}
	if cfg.Docs.UIPath != "/reference" {
		t.Fatalf("Docs.UIPath = %q, want %q", cfg.Docs.UIPath, "/reference")
	}
	if !reflect.DeepEqual(cfg.HTTP.TrustedProxyCIDRs, []string{"127.0.0.1/32", "::1/128"}) {
		t.Fatalf("HTTP.TrustedProxyCIDRs = %v, want %v", cfg.HTTP.TrustedProxyCIDRs, []string{"127.0.0.1/32", "::1/128"})
	}
}

func TestLoadPreservesExplicitFalseSessionSecureCookie(t *testing.T) {
	clearLegacyDatabaseEnv(t)

	path := writeConfigFile(t, `
app:
  name: article-sentinel
database:
  driver: mysql
  mysql:
    host: 127.0.0.1
    user: root
    password: root
    dbname: article_sentinel
redis:
  addr: 127.0.0.1:6379
auth:
  session:
    legacy_secret: legacy-secret
    secret: session-secret
    secure_cookie: false
    allow_legacy_query_jwt: true
    login_url: https://example.com/login
    redirect_url: http://127.0.0.1:5173/
docs:
  enabled: false
log:
  level: info
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Auth.Session.SecureCookie {
		t.Fatal("Auth.Session.SecureCookie = true, want false")
	}
	if cfg.Auth.Session.LoginURL != "https://example.com/login" {
		t.Fatalf("Auth.Session.LoginURL = %q", cfg.Auth.Session.LoginURL)
	}
	if cfg.Auth.Session.RedirectURL != "http://127.0.0.1:5173/" {
		t.Fatalf("Auth.Session.RedirectURL = %q", cfg.Auth.Session.RedirectURL)
	}
	if !cfg.Auth.Session.AllowLegacyQueryJWT {
		t.Fatal("Auth.Session.AllowLegacyQueryJWT = false, want true")
	}
}

func TestLoadIgnoresLegacySessionCookieNameOverride(t *testing.T) {
	clearLegacyDatabaseEnv(t)

	path := writeConfigFile(t, `
app:
  name: article-sentinel
database:
  driver: mysql
  mysql:
    host: 127.0.0.1
    user: root
    password: root
    dbname: article_sentinel
redis:
  addr: 127.0.0.1:6379
auth:
  session:
    legacy_secret: legacy-secret
    secret: session-secret
    cookie_name: legacy_cookie
docs:
  enabled: false
log:
  level: info
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	manager := identity.NewAdminSessionManager(cfg.Auth.Session)
	if got := manager.CookieName(); got != "as_admin_session" {
		t.Fatalf("CookieName() = %q, want %q", got, "as_admin_session")
	}
}

func TestLoadRejectsInvalidTrustedProxyCIDR(t *testing.T) {
	clearLegacyDatabaseEnv(t)

	path := writeConfigFile(t, `
app:
  name: article-sentinel
http:
  trusted_proxy_cidrs:
    - definitely-not-a-cidr
database:
  driver: mysql
  mysql:
    host: 127.0.0.1
    user: root
    password: root
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

func clearLegacyDatabaseEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"MYSQL_HOST",
		"MYSQL_PORT",
		"MYSQL_USER",
		"MYSQL_PASSWORD",
		"MYSQL_DBNAME",
		"MYSQL_CHARSET",
		"MYSQL_PARSE_TIME",
		"MYSQL_LOC",
		"MYSQL_MAX_OPEN_CONNS",
		"MYSQL_MAX_IDLE_CONNS",
		"MYSQL_CONN_MAX_LIFETIME_MINUTES",
		"DB_MYSQL_HOST",
		"DB_MYSQL_PORT",
		"DB_MYSQL_USER",
		"DB_MYSQL_PASSWORD",
		"DB_MYSQL_DBNAME",
		"DB_MYSQL_CHARSET",
		"DB_MYSQL_PARSE_TIME",
		"DB_MYSQL_LOC",
		"DB_MYSQL_MAX_OPEN_CONNS",
		"DB_MYSQL_MAX_IDLE_CONNS",
		"DB_MYSQL_CONN_MAX_LIFETIME_MINUTES",
		"HTTP_REQUEST_TIMEOUT_SECONDS",
		"HTTP_TRUSTED_PROXY_CIDRS",
		"AUTH_SESSION_ALLOW_LEGACY_QUERY_JWT",
		"DOCS_ENABLED",
		"DOCS_OPENAPI_PATH",
		"DOCS_UI_PATH",
	} {
		value, ok := os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("Unsetenv(%q) error = %v", key, err)
		}
		t.Cleanup(func() {
			if !ok {
				if err := os.Unsetenv(key); err != nil {
					t.Fatalf("cleanup Unsetenv(%q) error = %v", key, err)
				}
				return
			}
			if err := os.Setenv(key, value); err != nil {
				t.Fatalf("cleanup Setenv(%q) error = %v", key, err)
			}
		})
	}
}

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	return path
}
