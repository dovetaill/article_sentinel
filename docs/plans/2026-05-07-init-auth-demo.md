# Init Auth Demo Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Convert this branch into a reusable, runnable Go HTTP JWT authentication demo with protected API documentation and no frontend or business code.

**Architecture:** Keep a single `cmd/server` executable, a small runtime with config and logging, Huma-backed health/readiness/docs endpoints, and cookie-based JWT session authentication. Remove database, Redis, queues, schedulers, migrations, React/Vite admin code, and article-sentinel business modules. Protect documentation/schema endpoints by placing a docs guard after `SessionContext` in the middleware chain.

**Tech Stack:** Go, `net/http`, Huma v2, `github.com/golang-jwt/jwt/v5`, `cleanenv`, `slog`, bcrypt, lumberjack.

---

### Task 1: Reduce Configuration And Runtime Shape

**Files:**
- Modify: `pkg/config/config.go`
- Modify: `pkg/config/load.go`
- Modify: `pkg/config/config_test.go`
- Modify: `internal/app/bootstrap/runtime.go`
- Modify: `internal/app/bootstrap/server.go`
- Modify: `internal/app/bootstrap/bootstrap_test.go`
- Delete later: `internal/app/bootstrap/worker.go`, `internal/app/bootstrap/scheduler.go`, `internal/app/bootstrap/migrate.go`, `internal/app/bootstrap/schema.go`, and their obsolete tests

**Step 1: Write failing config tests**

Replace database-oriented expectations in `pkg/config/config_test.go` with reduced starter expectations:

```go
func TestDemoConfigTypeShape(t *testing.T) {
    tests := []struct {
        name        string
        typ         reflect.Type
        field       string
        wantPresent bool
    }{
        {name: "config drops database field", typ: reflect.TypeOf(config.Config{}), field: "Database", wantPresent: false},
        {name: "config drops redis field", typ: reflect.TypeOf(config.Config{}), field: "Redis", wantPresent: false},
        {name: "config drops queue field", typ: reflect.TypeOf(config.Config{}), field: "Queue", wantPresent: false},
        {name: "config drops scheduler field", typ: reflect.TypeOf(config.Config{}), field: "Scheduler", wantPresent: false},
        {name: "config keeps auth field", typ: reflect.TypeOf(config.Config{}), field: "Auth", wantPresent: true},
        {name: "config keeps docs field", typ: reflect.TypeOf(config.Config{}), field: "Docs", wantPresent: true},
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
```

Add a reduced load test:

```go
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
        t.Fatalf("App.Name = %q", cfg.App.Name)
    }
    if cfg.Auth.Session.Secret != "session-secret" {
        t.Fatalf("Auth.Session.Secret = %q", cfg.Auth.Session.Secret)
    }
    if cfg.HTTP.RequestTimeoutSeconds != 27 {
        t.Fatalf("HTTP.RequestTimeoutSeconds = %d", cfg.HTTP.RequestTimeoutSeconds)
    }
}
```

Update `internal/app/bootstrap/bootstrap_test.go` to expect config/logger only:

```go
func TestBuildServerRuntimeReturnsConfigAndLogger(t *testing.T) {
    origLoadConfig := loadConfigFn
    origNewLogger := newLoggerFn
    t.Cleanup(func() { loadConfigFn = origLoadConfig; newLoggerFn = origNewLogger })

    wantConfig := &config.Config{App: config.AppConfig{Name: "go-auth-demo"}}
    wantLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
    loadConfigFn = func(path string) (*config.Config, error) { return wantConfig, nil }
    newLoggerFn = func(cfg config.LogConfig) (*slog.Logger, func() error, error) {
        return wantLogger, func() error { return nil }, nil
    }

    rt, err := BuildServerRuntime("configs/config.yaml")
    if err != nil { t.Fatalf("BuildServerRuntime() error = %v", err) }
    if rt.Config != wantConfig { t.Fatal("runtime.Config mismatch") }
    if rt.Logger != wantLogger { t.Fatal("runtime.Logger mismatch") }
}
```

**Step 2: Run focused tests to verify failure**

Run:

```bash
go test ./pkg/config ./internal/app/bootstrap
```

Expected: compile/test failures because `Config` still has database/redis/queue/scheduler fields and runtime still bootstraps database resources.

**Step 3: Implement reduced config/runtime**

In `pkg/config/config.go`, reduce `Config` to:

```go
type Config struct {
    App  AppConfig  `yaml:"app"`
    HTTP HTTPConfig `yaml:"http"`
    Auth AuthConfig `yaml:"auth"`
    Docs DocsConfig `yaml:"docs"`
    Log  LogConfig  `yaml:"log"`
}
```

Delete `RedisConfig`, `DatabaseConfig`, `MySQLConfig`, `PostgresConfig`, `QueueConfig`, `AsynqConfig`, `OutboxConfig`, and `SchedulerConfig`.

In `pkg/config/load.go`, remove database defaults and validation:

```go
cfg := &Config{
    Docs: DocsConfig{Enabled: true},
    Auth: AuthConfig{Session: SessionConfig{SecureCookie: true}},
    Log:  LogConfig{RotateDaily: true},
}
if err := cleanenv.ReadConfig(path, cfg); err != nil {
    return nil, fmt.Errorf("read config %q: %w", path, err)
}
return cfg, nil
```

In `internal/app/bootstrap/runtime.go`, remove database imports, `Resources`, `bootstrapDatabaseFn`, and database shutdown registration. Keep `Config`, `Logger`, and registered closers.

**Step 4: Run focused tests to verify pass**

Run:

```bash
go test ./pkg/config ./internal/app/bootstrap
```

Expected: PASS after obsolete bootstrap tests are removed or rewritten.

**Step 5: Commit**

```bash
git add pkg/config internal/app/bootstrap
git commit -m "refactor: reduce runtime to auth demo core"
```

---

### Task 2: Protect Documentation Endpoints And Simplify Router

**Files:**
- Create: `internal/middleware/docs.go`
- Create: `internal/middleware/docs_test.go`
- Modify: `internal/api/register/router.go`
- Modify: `internal/api/register/router_test.go`
- Modify: `internal/api/handlers/ready.go`
- Modify: `internal/api/handlers/health_test.go`

**Step 1: Write failing docs protection tests**

Create `internal/middleware/docs_test.go`:

```go
package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/dovetaill/article-sentinel/internal/identity"
)

func TestProtectDocumentationRequiresActorOnDocsPaths(t *testing.T) {
    protected := ProtectDocumentation(true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusNoContent)
    }))

    for _, path := range []string{"/docs", "/docs/", "/openapi.json", "/openapi.yaml", "/openapi-3.0.json", "/openapi-3.0.yaml", "/schemas/ErrorModel.json"} {
        t.Run(path, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodGet, path, nil)
            rec := httptest.NewRecorder()
            protected.ServeHTTP(rec, req)
            if rec.Code != http.StatusUnauthorized {
                t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
            }
        })
    }
}

func TestProtectDocumentationAllowsActorOnDocsPaths(t *testing.T) {
    protected := ProtectDocumentation(true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusNoContent)
    }))
    req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
    req = req.WithContext(identity.ContextWithActor(req.Context(), identity.NewActor(1, "admin", "admin", "active")))
    rec := httptest.NewRecorder()

    protected.ServeHTTP(rec, req)

    if rec.Code != http.StatusNoContent {
        t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
    }
}
```

Rewrite router tests in `internal/api/register/router_test.go` to expect only generic paths in OpenAPI:

```go
wantPaths := []string{"/api/v1/demo/me", "/healthz", "/readyz"}
```

Add anonymous docs tests:

```go
func TestRouterProtectsDocsEndpoints(t *testing.T) {
    handler := register.NewRouter(newRouterTestRuntime(true))
    for _, path := range []string{"/docs", "/openapi.json", "/openapi.yaml", "/openapi-3.0.json", "/openapi-3.0.yaml", "/schemas/ErrorModel.json"} {
        req := httptest.NewRequest(http.MethodGet, path, nil)
        rec := httptest.NewRecorder()
        handler.ServeHTTP(rec, req)
        if rec.Code != http.StatusUnauthorized {
            t.Fatalf("%s status = %d, want %d", path, rec.Code, http.StatusUnauthorized)
        }
    }
}
```

Add authenticated OpenAPI test using a signed session cookie from the manager in `newRouterTestRuntime`.

**Step 2: Run focused tests to verify failure**

Run:

```bash
go test ./internal/middleware ./internal/api/register ./internal/api/handlers
```

Expected: FAIL because `ProtectDocumentation` and `/api/v1/demo/me` do not exist and business routes are still registered.

**Step 3: Implement docs guard**

Create `internal/middleware/docs.go`:

```go
package middleware

import (
    "net/http"
    "strings"

    "github.com/dovetaill/article-sentinel/internal/identity"
)

func ProtectDocumentation(enabled bool) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if enabled && isDocumentationPath(r.URL.Path) {
                if _, ok := identity.ActorFromContext(r.Context()); !ok {
                    writeAuthError(w, http.StatusUnauthorized, "unauthorized")
                    return
                }
            }
            next.ServeHTTP(w, r)
        })
    }
}

func isDocumentationPath(path string) bool {
    switch path {
    case "/docs", "/openapi.json", "/openapi.yaml", "/openapi-3.0.json", "/openapi-3.0.yaml":
        return true
    }
    return strings.HasPrefix(path, "/docs/") || strings.HasPrefix(path, "/schemas/")
}
```

**Step 4: Simplify router**

In `internal/api/register/router.go`:

- Remove imports for `articleinspectmodule`, `postmodule`, `queueasynq`, and database-related route builders.
- Keep Huma setup and auth routes.
- Register `handlers.RegisterHealth`, `handlers.RegisterReady`, and `handlers.RegisterDemoRoutes`.
- Insert `middleware.ProtectDocumentation(docsEnabled(rt))` after `middleware.SessionContext(adminSessionManager)`.

The chain should be:

```go
return middleware.Chain(
    apiMux,
    middleware.RequestID(),
    middleware.SessionContext(adminSessionManager),
    middleware.ProtectDocumentation(docsEnabled(rt)),
    middleware.Recover(),
    middleware.Timeout(timeout),
    middleware.AccessLog(nilLogger(rt)),
)
```

In `internal/api/handlers/ready.go`, return a database-free payload:

```go
Body: response.OK("ready", map[string]any{"status": "ready"})
```

**Step 5: Run focused tests to verify pass**

Run:

```bash
go test ./internal/middleware ./internal/api/register ./internal/api/handlers
```

Expected: PASS.

**Step 6: Commit**

```bash
git add internal/middleware internal/api/register internal/api/handlers
git commit -m "feat: protect api documentation with session auth"
```

---

### Task 3: Add Minimal Protected Demo Endpoint

**Files:**
- Create: `internal/api/handlers/demo.go`
- Create: `internal/api/handlers/demo_test.go`
- Modify: `internal/api/register/router.go`

**Step 1: Write failing handler test**

Create `internal/api/handlers/demo_test.go`:

```go
package handlers_test

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/danielgtaylor/huma/v2"
    "github.com/danielgtaylor/huma/v2/adapters/humago"
    "github.com/dovetaill/article-sentinel/internal/api/handlers"
    "github.com/dovetaill/article-sentinel/internal/api/response"
    "github.com/dovetaill/article-sentinel/internal/identity"
)

func TestDemoMeRequiresAuthenticatedActor(t *testing.T) {
    mux := http.NewServeMux()
    api := humago.New(mux, huma.DefaultConfig("go-auth-demo", "0.1.0"))
    handlers.RegisterDemoRoutes(huma.NewGroup(api))

    req := httptest.NewRequest(http.MethodGet, "/api/v1/demo/me", nil)
    rec := httptest.NewRecorder()
    mux.ServeHTTP(rec, req)

    if rec.Code != http.StatusUnauthorized {
        t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
    }
}

func TestDemoMeReturnsActorFromContext(t *testing.T) {
    mux := http.NewServeMux()
    api := humago.New(mux, huma.DefaultConfig("go-auth-demo", "0.1.0"))
    handlers.RegisterDemoRoutes(huma.NewGroup(api))

    req := httptest.NewRequest(http.MethodGet, "/api/v1/demo/me", nil)
    req = req.WithContext(identity.ContextWithActor(req.Context(), identity.NewActor(7, "demo", "admin", "active")))
    rec := httptest.NewRecorder()
    mux.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
    }
    var got response.Envelope
    if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil { t.Fatalf("decode: %v", err) }
    data := got.Data.(map[string]any)
    if data["username"] != "demo" { t.Fatalf("data = %#v", data) }
}
```

**Step 2: Run test to verify failure**

Run:

```bash
go test ./internal/api/handlers -run DemoMe
```

Expected: FAIL because `RegisterDemoRoutes` does not exist.

**Step 3: Implement demo handler**

Create `internal/api/handlers/demo.go`:

```go
package handlers

import (
    "context"
    "net/http"

    "github.com/danielgtaylor/huma/v2"
    "github.com/dovetaill/article-sentinel/internal/api/response"
    "github.com/dovetaill/article-sentinel/internal/identity"
)

type demoMeOutput struct {
    Body response.Envelope
}

func RegisterDemoRoutes(api huma.API) {
    huma.Register(api, huma.Operation{
        OperationID: "demo-me",
        Method:      http.MethodGet,
        Path:        "/api/v1/demo/me",
        Summary:     "current authenticated actor",
    }, func(ctx context.Context, input *struct{}) (*demoMeOutput, error) {
        actor, ok := identity.ActorFromContext(ctx)
        if !ok {
            return nil, huma.Error401Unauthorized("unauthorized")
        }
        return &demoMeOutput{Body: response.OK("me", actor)}, nil
    })
}
```

Register it in `internal/api/register/router.go`.

**Step 4: Run tests to verify pass**

Run:

```bash
go test ./internal/api/handlers ./internal/api/register
```

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/api/handlers internal/api/register
git commit -m "feat: add protected demo actor endpoint"
```

---

### Task 4: Remove Frontend, Business, Queue, Scheduler, Database, And Migration Code

**Files:**
- Delete: `web/`
- Delete: `internal/modules/`
- Delete: `internal/queue/`
- Delete: `internal/scheduler/`
- Delete: `cmd/worker/`
- Delete: `cmd/scheduler/`
- Delete: `cmd/migrate/`
- Delete: `pkg/database/`
- Delete: `migrations/`
- Delete: `ddl/`
- Delete or replace: `scripts/`
- Modify: `Makefile`
- Modify: `go.mod`
- Modify: `go.sum`

**Step 1: Confirm tree before deletion**

Run:

```bash
git status --short
rg --files web internal/modules internal/queue internal/scheduler cmd/worker cmd/scheduler cmd/migrate pkg/database migrations ddl scripts | sed -n '1,120p'
```

Expected: only planned files are listed; no unrelated dirty files.

**Step 2: Delete obsolete code**

Run:

```bash
rm -rf web internal/modules internal/queue internal/scheduler cmd/worker cmd/scheduler cmd/migrate pkg/database migrations ddl scripts
```

**Step 3: Simplify Makefile**

Replace `Makefile` with minimal targets:

```makefile
CONFIG ?= configs/config.example.yaml

.PHONY: run test verify build

run:
	go run ./cmd/server -config $(CONFIG)

test:
	go test ./...

verify:
	go test ./...

build:
	go build ./cmd/server
```

**Step 4: Tidy modules**

Run:

```bash
go mod tidy
```

Expected: unused dependencies such as asynq, gorm drivers, Redis, cron, and sqlite are removed if no retained code imports them.

**Step 5: Run tests**

Run:

```bash
go test ./...
```

Expected: compile failures only for imports still referencing deleted packages. Fix those references, then rerun until PASS.

**Step 6: Commit**

```bash
git add -A
git commit -m "refactor: remove frontend and business services"
```

---

### Task 5: Rename Demo Surface And Update Config Files

**Files:**
- Modify: `go.mod`
- Modify all Go imports under: `cmd/`, `internal/`, `pkg/`
- Modify: `configs/config.example.yaml`
- Modify: `configs/config.local.yaml`
- Modify: `.env.example`

**Step 1: Rename module imports**

Choose generic module path `github.com/dovetaill/go-auth-demo` and update `go.mod`:

```go
module github.com/dovetaill/go-auth-demo
```

Replace imports:

```bash
rg -l 'github.com/dovetaill/article-sentinel' --glob '*.go' | xargs sed -i 's#github.com/dovetaill/article-sentinel#github.com/dovetaill/go-auth-demo#g'
```

**Step 2: Update app defaults in config**

Set `configs/config.example.yaml` to the minimal runnable config:

```yaml
app:
  name: go-auth-demo
  env: local
  host: 0.0.0.0
  port: 8080

http:
  request_timeout_seconds: 15
  read_timeout_seconds: 15
  write_timeout_seconds: 15
  idle_timeout_seconds: 60

auth:
  session:
    legacy_secret: change-me-legacy-secret
    secret: change-me-session-secret
    issuer: go-auth-demo
    ttl_hours: 24
    secure_cookie: false
    login_url: /auth/login
    redirect_url: /docs

docs:
  enabled: true
  openapi_path: /openapi.json
  ui_path: /docs

log:
  level: info
  format: json
  output: stdout
  dir: logs
  filename: app.log
  max_size_mb: 100
  max_backups: 14
  max_age_days: 30
  compress: false
  rotate_daily: true
```

Make `configs/config.local.yaml` either identical with local comments or remove it if redundant.

Set `.env.example` to optional overrides only:

```dotenv
APP_NAME=go-auth-demo
APP_PORT=8080
AUTH_SESSION_LEGACY_SECRET=change-me-legacy-secret
AUTH_SESSION_SECRET=change-me-session-secret
AUTH_SESSION_SECURE_COOKIE=false
AUTH_SESSION_REDIRECT_URL=/docs
```

**Step 3: Format and test**

Run:

```bash
gofmt -w $(rg --files -g '*.go')
go test ./...
```

Expected: PASS.

**Step 4: Commit**

```bash
git add -A
git commit -m "chore: rename project to go auth demo"
```

---

### Task 6: Rewrite README And Remove Business Documentation

**Files:**
- Modify: `README.md`
- Keep: `docs/plans/2026-05-07-init-auth-demo-design.md`
- Keep: `docs/plans/2026-05-07-init-auth-demo.md`
- Delete: business-specific files under `docs/` except retained plan history if desired

**Step 1: Replace README with generic starter guide**

Write `README.md` with these sections:

```markdown
# go-auth-demo

A small Go HTTP backend starter that demonstrates JWT-based admin session authentication.

## Features

- Third-party HS256 legacy JWT exchange
- HTTP-only `as_admin_session` cookie
- Session middleware and current-session API
- Protected OpenAPI/docs/schema endpoints
- Health and readiness endpoints
- Minimal protected demo API

## Run Locally

```bash
go run ./cmd/server -config configs/config.example.yaml
```

## API

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| GET | `/healthz` | No | process health |
| GET | `/readyz` | No | app readiness |
| GET | `/auth/login?jwt=...` | No | exchange legacy JWT for session cookie |
| GET | `/api/v1/auth/session` | Cookie | current session |
| POST | `/api/v1/auth/logout` | No | clear cookie |
| GET | `/api/v1/demo/me` | Cookie | current actor demo |
| GET | `/docs` | Cookie | API docs |
| GET | `/openapi.json` | Cookie | OpenAPI 3.1 JSON |
| GET | `/openapi.yaml` | Cookie | OpenAPI 3.1 YAML |
| GET | `/openapi-3.0.json` | Cookie | OpenAPI 3.0 JSON |
| GET | `/openapi-3.0.yaml` | Cookie | OpenAPI 3.0 YAML |
| GET | `/schemas/{schema}` | Cookie | JSON schema browser |
```

Include a small Go or shell snippet that signs a legacy JWT for local testing.

**Step 2: Delete business docs**

Remove docs that describe article inspection, admin UI restyles, seed data, outbox, and category workflows. Keep the new init design/implementation docs.

**Step 3: Commit**

```bash
git add -A
git commit -m "docs: rewrite readme for auth demo"
```

---

### Task 7: Final Verification

**Files:**
- Modify only if final verification reveals issues.

**Step 1: Run full verification**

Run:

```bash
go test ./...
go build ./cmd/server
```

Expected: both commands PASS.

**Step 2: Smoke start the server briefly**

Run:

```bash
timeout 5s go run ./cmd/server -config configs/config.example.yaml
```

Expected: command starts the server and exits due to timeout with status 124; there should be no config, database, Redis, or frontend errors before timeout.

**Step 3: Inspect status**

Run:

```bash
git status --short --branch
git log --oneline -8
```

Expected: branch is `init`; working tree is clean; commits show design, plan, and implementation commits.

**Step 4: Fix any issues and commit**

If verification found issues, fix them and commit:

```bash
git add -A
git commit -m "fix: finalize auth demo verification"
```

