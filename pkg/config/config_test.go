package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/dovetaill/go-auth-demo/internal/identity"
	"github.com/dovetaill/go-auth-demo/pkg/config"
)

func TestStarterConfigTypeShape(t *testing.T) {
	tests := []struct {
		name        string
		typ         reflect.Type
		field       string
		wantPresent bool
	}{
		{name: "config keeps app field", typ: reflect.TypeOf(config.Config{}), field: "App", wantPresent: true},
		{name: "config keeps http field", typ: reflect.TypeOf(config.Config{}), field: "HTTP", wantPresent: true},
		{name: "config keeps auth field", typ: reflect.TypeOf(config.Config{}), field: "Auth", wantPresent: true},
		{name: "config keeps docs field", typ: reflect.TypeOf(config.Config{}), field: "Docs", wantPresent: true},
		{name: "config keeps log field", typ: reflect.TypeOf(config.Config{}), field: "Log", wantPresent: true},
		{name: "config drops database field", typ: reflect.TypeOf(config.Config{}), field: "Database", wantPresent: false},
		{name: "config drops redis field", typ: reflect.TypeOf(config.Config{}), field: "Redis", wantPresent: false},
		{name: "config drops queue field", typ: reflect.TypeOf(config.Config{}), field: "Queue", wantPresent: false},
		{name: "config drops scheduler field", typ: reflect.TypeOf(config.Config{}), field: "Scheduler", wantPresent: false},
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

func TestLoadReadsAuthDemoConfigWithoutDatabase(t *testing.T) {
	clearEnv(t)

	path := writeConfigFile(t, `
app:
  name: go-auth-demo
http:
  request_timeout_seconds: 27
auth:
  session:
    legacy_secret: legacy-secret
    secret: session-secret
    secure_cookie: false
    login_url: https://example.com/login
    redirect_url: /docs
docs:
  enabled: true
  openapi_path: /openapi.json
  ui_path: /docs
log:
  level: info
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.App.Name != "go-auth-demo" {
		t.Fatalf("App.Name = %q, want %q", cfg.App.Name, "go-auth-demo")
	}
	if cfg.Auth.Session.Secret != "session-secret" {
		t.Fatalf("Auth.Session.Secret = %q, want %q", cfg.Auth.Session.Secret, "session-secret")
	}
	if cfg.HTTP.RequestTimeoutSeconds != 27 {
		t.Fatalf("HTTP.RequestTimeoutSeconds = %d, want %d", cfg.HTTP.RequestTimeoutSeconds, 27)
	}
}

func TestLoadReadsExplicitDocsAndRequestTimeoutConfig(t *testing.T) {
	clearEnv(t)

	path := writeConfigFile(t, `
app:
  name: go-auth-demo
http:
  request_timeout_seconds: 27
auth:
  session:
    legacy_secret: legacy-secret
    secret: session-secret
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
}

func TestLoadRejectsMissingAuthSessionSecrets(t *testing.T) {
	tests := []struct {
		name        string
		sessionYAML string
		wantFields  []string
	}{
		{
			name: "missing legacy secret",
			sessionYAML: `
    secret: session-secret`,
			wantFields: []string{"auth.session.legacy_secret"},
		},
		{
			name: "missing session secret",
			sessionYAML: `
    legacy_secret: legacy-secret`,
			wantFields: []string{"auth.session.secret"},
		},
		{
			name:        "missing both secrets",
			sessionYAML: "",
			wantFields:  []string{"auth.session.legacy_secret", "auth.session.secret"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)

			path := writeConfigFile(t, `
app:
  name: go-auth-demo
auth:
  session:`+tt.sessionYAML+`
docs:
  enabled: true
log:
  level: info
`)

			_, err := config.Load(path)
			if err == nil {
				t.Fatal("Load() error = nil, want missing auth session secret error")
			}
			for _, field := range tt.wantFields {
				if !strings.Contains(err.Error(), field) {
					t.Fatalf("Load() error = %v, want contains %q", err, field)
				}
			}
		})
	}
}

func TestLoadPreservesExplicitFalseSessionSecureCookie(t *testing.T) {
	clearEnv(t)

	path := writeConfigFile(t, `
app:
  name: go-auth-demo
auth:
  session:
    legacy_secret: legacy-secret
    secret: session-secret
    secure_cookie: false
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
}

func TestLoadIgnoresLegacySessionCookieNameOverride(t *testing.T) {
	clearEnv(t)

	path := writeConfigFile(t, `
app:
  name: go-auth-demo
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

func clearEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"APP_NAME",
		"APP_ENV",
		"APP_HOST",
		"APP_PORT",
		"HTTP_REQUEST_TIMEOUT_SECONDS",
		"HTTP_READ_TIMEOUT_SECONDS",
		"HTTP_WRITE_TIMEOUT_SECONDS",
		"HTTP_IDLE_TIMEOUT_SECONDS",
		"AUTH_SESSION_LEGACY_SECRET",
		"AUTH_SESSION_SECRET",
		"AUTH_SESSION_ISSUER",
		"AUTH_SESSION_TTL_HOURS",
		"AUTH_SESSION_SECURE_COOKIE",
		"AUTH_SESSION_LOGIN_URL",
		"AUTH_SESSION_REDIRECT_URL",
		"DOCS_ENABLED",
		"DOCS_OPENAPI_PATH",
		"DOCS_UI_PATH",
		"LOG_LEVEL",
		"LOG_FORMAT",
		"LOG_OUTPUT",
		"LOG_DIR",
		"LOG_FILENAME",
		"LOG_MAX_SIZE_MB",
		"LOG_MAX_BACKUPS",
		"LOG_MAX_AGE_DAYS",
		"LOG_COMPRESS",
		"LOG_ROTATE_DAILY",
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
