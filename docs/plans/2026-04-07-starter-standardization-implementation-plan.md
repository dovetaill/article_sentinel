# article-sentinel Starter Standardization Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Turn `main` into a clean, standard, easy-to-start backend starter that keeps only shared runtime infrastructure plus one official demo module.

**Architecture:** Keep the four entrypoints (`server`, `worker`, `scheduler`, `migrate`) and the shared bootstrap / middleware / config layers, but remove showcase-only business modules from `main`. Standardize runtime behavior around one explicit schema-sync path, one primary database configuration model, one local development flow, and one official module template (`post`).

**Tech Stack:** Go 1.25+, Huma v2, `net/http`, GORM, `cleanenv`, `slog`, Docker Compose, Makefile, POSIX shell.

---

### Task 1: Lock The Starter Contract In Tests First

**Files:**
- Modify: `internal/api/register/router_test.go`
- Modify: `pkg/config/config_test.go`
- Modify: `internal/app/bootstrap/migrate_test.go`
- Modify: `internal/app/bootstrap/schema_test.go`

**Step 1: Tighten the router contract test**

Update `internal/api/register/router_test.go` so it asserts the default starter only exposes:

```go
/healthz
/readyz
/openapi.json
/docs
/api/v1/posts
/api/v1/posts/{id}
```

Add explicit absent-path assertions for all removed showcase routes:

```go
assertPathAbsent(t, doc.Paths, "/api/v1/auth/login")
assertPathAbsent(t, doc.Paths, "/api/v1/member/auth/login")
assertPathAbsent(t, doc.Paths, "/api/v1/admin/users")
assertPathAbsent(t, doc.Paths, "/api/v1/articles")
```

**Step 2: Add the new config-shape expectations**

Update `pkg/config/config_test.go` so the tests describe the target starter config model:

- one primary database config under `database`
- no legacy top-level `mysql` section
- no `seed_admin` defaults in starter config
- docs enable/disable and request-timeout semantics are explicit

Use table-driven tests where possible.

**Step 3: Add explicit migrate/server behavior tests**

Update `internal/app/bootstrap/migrate_test.go` and `internal/app/bootstrap/schema_test.go` so the target behavior is:

- `cmd/migrate` performs starter schema sync
- `BuildServerRuntime` no longer auto-migrates schemas during server startup
- no seed-admin bootstrap path exists in starter mode

Add a failing assertion like:

```go
if autoMigrateCalls != 0 {
    t.Fatalf("auto migrate call count = %d, want %d", autoMigrateCalls, 0)
}
```

**Step 4: Run the focused test set and confirm it fails for the right reasons**

Run:

```bash
go test ./internal/api/register ./pkg/config ./internal/app/bootstrap -v
```

Expected: FAIL because config structure, migrate semantics, and starter bootstrap behavior have not been refactored yet.

**Step 5: Commit the red test baseline**

```bash
git add internal/api/register/router_test.go pkg/config/config_test.go internal/app/bootstrap/migrate_test.go internal/app/bootstrap/schema_test.go
git commit -m "test: lock starter standardization contract"
```

### Task 2: Remove Showcase Business Modules And Stale Main-Branch Narratives

**Files:**
- Delete: `internal/modules/article/admin_handler.go`
- Delete: `internal/modules/article/article_test.go`
- Delete: `internal/modules/article/handler.go`
- Delete: `internal/modules/article/model.go`
- Delete: `internal/modules/article/public_handler.go`
- Delete: `internal/modules/article/repository.go`
- Delete: `internal/modules/article/service.go`
- Delete: `internal/modules/auth/auth_test.go`
- Delete: `internal/modules/auth/handler.go`
- Delete: `internal/modules/auth/jwt.go`
- Delete: `internal/modules/auth/model.go`
- Delete: `internal/modules/auth/password.go`
- Delete: `internal/modules/auth/repository.go`
- Delete: `internal/modules/auth/service.go`
- Delete: `internal/modules/category/admin_handler.go`
- Delete: `internal/modules/category/category_test.go`
- Delete: `internal/modules/category/handler.go`
- Delete: `internal/modules/category/model.go`
- Delete: `internal/modules/category/public_handler.go`
- Delete: `internal/modules/category/repository.go`
- Delete: `internal/modules/category/service.go`
- Delete: `internal/modules/engagement/engagement_test.go`
- Delete: `internal/modules/engagement/handler.go`
- Delete: `internal/modules/engagement/model.go`
- Delete: `internal/modules/engagement/repository.go`
- Delete: `internal/modules/engagement/service.go`
- Delete: `internal/modules/member/member_test.go`
- Delete: `internal/modules/member/model.go`
- Delete: `internal/modules/member/public_handler.go`
- Delete: `internal/modules/member/repository.go`
- Delete: `internal/modules/member/self_handler.go`
- Delete: `internal/modules/member/service.go`
- Delete: `internal/modules/user/handler.go`
- Delete: `internal/modules/user/model.go`
- Delete: `internal/modules/user/repository.go`
- Delete: `internal/modules/user/service.go`
- Delete: `internal/modules/user/user_test.go`
- Modify: `README.md`
- Modify: `docs/showcase/multisurface.md`
- Delete: `verification.md`

**Step 1: Remove unused showcase module directories from `main`**

Delete every file listed above from the six showcase-only modules.

Do not touch `internal/modules/post` or `internal/modules/example/README.md`.

**Step 2: Rewrite the main README so it only describes the starter**

Update `README.md` so it:

- no longer mentions old default role-driven flows
- only advertises one official demo module: `post`
- clearly says the richer business sample lives in `showcase/multisurface`
- points users to the starter quickstart first

**Step 3: Reduce showcase docs to a short branch note**

Trim `docs/showcase/multisurface.md` down to:

- what the branch is
- why it is separate
- how to switch to it

Remove any text that implies the `main` branch still carries those business modules.

**Step 4: Delete stale verification narrative**

Delete `verification.md` instead of trying to partially update it. The file reflects the old mixed starter/showcase state and will keep drifting.

**Step 5: Run the full test suite and fix compile fallout**

Run:

```bash
go test ./... 
```

Expected: FAIL at first due to stale references to removed modules in docs, tests, or package imports. Remove those references until the suite compiles again.

**Step 6: Commit the starter-boundary cleanup**

```bash
git add README.md docs/showcase/multisurface.md internal/modules verification.md
git commit -m "refactor: remove showcase business modules from main"
```

### Task 3: Collapse Config To One Primary Database Model And Remove Seed Admin Drift

**Files:**
- Modify: `pkg/config/config.go`
- Modify: `pkg/config/load.go`
- Modify: `pkg/config/config_test.go`
- Modify: `configs/config.example.yaml`
- Modify: `pkg/database/driver.go`
- Modify: `pkg/database/resources.go`
- Modify: `pkg/database/database_test.go`
- Modify: `internal/app/bootstrap/migrate.go`
- Modify: `internal/app/bootstrap/migrate_test.go`
- Modify: `internal/app/bootstrap/schema.go`
- Modify: `internal/app/bootstrap/schema_test.go`

**Step 1: Remove legacy top-level MySQL config and seed-admin config**

Refactor `pkg/config/config.go` so:

- `Config` no longer has top-level `MySQL`
- `AuthConfig` no longer has `SeedAdmin`
- `DatabaseConfig` becomes the single source of truth for the primary database

The target shape is:

```go
type Config struct {
    App       AppConfig
    HTTP      HTTPConfig
    Redis     RedisConfig
    Database  DatabaseConfig
    Auth      AuthConfig
    Queue     QueueConfig
    Scheduler SchedulerConfig
    Docs      DocsConfig
    Log       LogConfig
}
```

**Step 2: Rename the runtime primary database handle**

In `pkg/database/resources.go`, rename:

```go
MySQL *gorm.DB
```

to:

```go
DB *gorm.DB
```

Then update all starter-owned callers to use `Resources.DB`.

**Step 3: Remove seed-admin bootstrap helpers from starter bootstrap**

Delete the seed-admin-specific constants, interfaces, and helper functions from `internal/app/bootstrap/schema.go` and their tests from `internal/app/bootstrap/schema_test.go`.

Keep only the schema registration / auto-migrate helpers that the starter still needs.

**Step 4: Simplify migrate config resolution**

Update `internal/app/bootstrap/migrate.go` and `pkg/database/driver.go` so they read only from `cfg.Database`, with no legacy fallback to the removed top-level MySQL config.

**Step 5: Rewrite the example config**

Update `configs/config.example.yaml` so it contains:

- one `database` section
- no top-level `mysql`
- no `seed_admin`
- comments that match the starter story

**Step 6: Run the focused config/bootstrap test suite**

Run:

```bash
go test ./pkg/config ./pkg/database ./internal/app/bootstrap -v
```

Expected: PASS once the config and runtime model is fully consolidated.

**Step 7: Commit the config simplification**

```bash
git add pkg/config pkg/database internal/app/bootstrap configs/config.example.yaml
git commit -m "refactor: simplify starter config and bootstrap model"
```

### Task 4: Make Migrate Explicit And Fix HTTP Runtime Semantics

**Files:**
- Modify: `internal/app/bootstrap/server.go`
- Modify: `internal/app/bootstrap/migrate.go`
- Modify: `cmd/migrate/main.go`
- Modify: `cmd/server/main.go`
- Modify: `internal/api/register/router.go`
- Modify: `internal/api/register/router_test.go`
- Modify: `pkg/config/config.go`
- Modify: `pkg/config/config_test.go`

**Step 1: Add a dedicated request-timeout field**

Extend `HTTPConfig` with a separate application timeout such as:

```go
RequestTimeoutSeconds int `yaml:"request_timeout_seconds" env:"HTTP_REQUEST_TIMEOUT_SECONDS" env-default:"15"`
```

This is distinct from server-level read/write/idle timeouts.

**Step 2: Stop auto-migrating on server startup**

Update `internal/app/bootstrap/server.go` so `BuildServerRuntime` only builds runtime resources.

Do not run schema migration inside server startup anymore.

**Step 3: Make `cmd/migrate` perform the starter schema sync**

Refactor `internal/app/bootstrap/migrate.go` so `RunMigrateCommand`:

- loads config
- builds runtime or database resources
- runs `AutoMigrateBusinessTables`
- returns a real result instead of only validating config

Keep the command GORM-based for starter simplicity; do not add a SQL migration framework in this pass.

**Step 4: Apply HTTP timeout settings consistently**

Update `cmd/server/main.go` to set:

```go
ReadTimeout
WriteTimeout
IdleTimeout
ReadHeaderTimeout
```

from config, and update `internal/api/register/router.go` to use only `RequestTimeoutSeconds` for `middleware.Timeout(...)`.

**Step 5: Make docs enable/disable explicit**

Update `internal/api/register/router.go` so `docs.enabled=false` disables both docs UI and OpenAPI endpoint registration.

Add or update router tests to cover:

- docs enabled: `/openapi.json` and `/docs` return success
- docs disabled: both endpoints return 404

**Step 6: Run the starter-facing test suite**

Run:

```bash
go test ./cmd/... ./internal/api/register ./internal/app/bootstrap ./pkg/config -v
```

Expected: PASS with explicit migrate behavior and consistent timeout/docs semantics.

**Step 7: Commit the runtime cleanup**

```bash
git add cmd/server/main.go cmd/migrate/main.go internal/app/bootstrap internal/api/register pkg/config
git commit -m "refactor: standardize starter runtime semantics"
```

### Task 5: Improve Request Tracing, Access Logs, And Readiness Semantics

**Files:**
- Modify: `internal/middleware/requestid.go`
- Modify: `internal/middleware/accesslog.go`
- Modify: `internal/api/handlers/ready.go`
- Modify: `internal/api/handlers/health_test.go`
- Create: `internal/middleware/requestid_test.go`
- Create: `internal/middleware/accesslog_test.go`

**Step 1: Write failing request-id and access-log tests**

Create `internal/middleware/requestid_test.go` and `internal/middleware/accesslog_test.go`.

Cover at least:

- generated request id is returned in `X-Request-ID`
- inbound request id is preserved
- request id is available via context
- access log records request id, method, path, status code, duration

Use a stub logger or in-memory handler to assert fields.

**Step 2: Implement request-id context propagation**

Refactor `internal/middleware/requestid.go` to:

- define a private context key
- store the request id in request context
- expose a helper like `RequestIDFromContext(ctx context.Context) (string, bool)`

**Step 3: Capture status codes in access logs**

Wrap `http.ResponseWriter` in `internal/middleware/accesslog.go` so the logger records the final status code instead of only method/path/duration.

Add `request_id` to the log fields if present in context.

**Step 4: Make readiness semantics explicit**

Refactor `internal/api/handlers/ready.go` so the response distinguishes between:

- dependency wiring state
- dependency health state

If true pings are too invasive for this pass, return structured fields such as:

```json
{
  "database": {"configured": true, "healthy": true},
  "redis": {"configured": true, "healthy": true}
}
```

**Step 5: Run the middleware and handler tests**

Run:

```bash
go test ./internal/middleware ./internal/api/handlers -v
```

Expected: PASS with request tracing, richer logs, and clearer readiness output.

**Step 6: Commit the observability pass**

```bash
git add internal/middleware internal/api/handlers
git commit -m "feat: improve starter observability defaults"
```

### Task 6: Add A Real Local Development Path And One-Command Verification

**Files:**
- Create: `Makefile`
- Create: `docker-compose.yml`
- Create: `configs/config.local.yaml`
- Create: `scripts/verify.sh`
- Create: `scripts/smoke.sh`
- Modify: `README.md`

**Step 1: Create the local config and dependency stack**

Create `docker-compose.yml` with starter-owned services only:

- MySQL
- Redis

Create `configs/config.local.yaml` that points to those defaults.

Do not add extra infrastructure that the starter does not use.

**Step 2: Add standard developer commands**

Create `Makefile` targets:

```make
up
down
dev
test
verify
smoke
migrate
```

Keep commands thin wrappers around existing Go entrypoints and scripts.

**Step 3: Add verification and smoke scripts**

Create:

- `scripts/verify.sh` to run format checks if added later, then `go test ./...`
- `scripts/smoke.sh` to:
  - ensure deps are up
  - run `cmd/migrate`
  - start `cmd/server`
  - `curl` `/healthz`, `/readyz`, `/openapi.json`, `/api/v1/posts`

Make both scripts strict:

```bash
set -euo pipefail
```

**Step 4: Validate the scripts**

Run:

```bash
bash -n scripts/verify.sh
bash -n scripts/smoke.sh
```

Expected: PASS with no shell syntax errors.

Then run:

```bash
make verify
```

Expected: PASS.

**Step 5: Commit the local DX layer**

```bash
git add Makefile docker-compose.yml configs/config.local.yaml scripts/verify.sh scripts/smoke.sh README.md
git commit -m "feat: add starter local development workflow"
```

### Task 7: Turn The Post Module Into The Official Copyable Template

**Files:**
- Modify: `internal/modules/example/README.md`
- Modify: `README.md`
- Create: `scripts/new-module.sh`
- Modify: `internal/modules/post/post_test.go`
- Modify: `internal/modules/post/handler.go`
- Modify: `internal/modules/post/service.go`
- Modify: `internal/api/register/router.go`

**Step 1: Tighten the post-module behavior tests**

Update `internal/modules/post/post_test.go` so it clearly demonstrates the starter contract:

- create
- list with pagination envelope
- detail
- patch
- delete
- validation failures

The test file should read like starter documentation in executable form.

**Step 2: Keep the router entrypoint thin**

Refactor `internal/api/register/router.go` only if needed so module wiring remains a small block:

```go
if postService := newPostService(rt); postService != nil {
    postmodule.RegisterRoutes(publicRoutes, postService)
}
```

Do not move post business logic into the router.

**Step 3: Add a module bootstrap script**

Create `scripts/new-module.sh` that:

- copies `internal/modules/post`
- renames the directory
- replaces obvious `post` naming tokens
- prints manual follow-up steps for model and route naming

Keep the script intentionally small and transparent.

**Step 4: Rewrite the module guide**

Update `internal/modules/example/README.md` to document the exact starter flow:

1. run `scripts/new-module.sh`
2. rename model / repository / service / handler details
3. register module in `internal/api/register/router.go`
4. register model in `internal/app/bootstrap/schema.go` if schema sync still relies on model registration
5. extend tests

**Step 5: Validate the module path**

Run:

```bash
bash -n scripts/new-module.sh
go test ./internal/modules/post ./internal/api/register -v
```

Expected: PASS.

**Step 6: Commit the module-template polish**

```bash
git add internal/modules/example/README.md internal/modules/post internal/api/register/router.go scripts/new-module.sh README.md
git commit -m "feat: standardize the official starter module template"
```

### Task 8: Final Verification And Release Polish

**Files:**
- Modify: `README.md`
- Modify: `docs/showcase/multisurface.md`
- Modify: `configs/config.example.yaml`
- Modify: `Makefile`
- Modify: `scripts/verify.sh`

**Step 1: Run the complete verification pass**

Run:

```bash
go test ./...
make verify
```

Expected: PASS.

If Docker is available locally, also run:

```bash
make up
make migrate
make smoke
make down
```

Expected: PASS with the starter booting successfully against local MySQL + Redis.

**Step 2: Read the README like a new user**

Manually verify `README.md` answers these questions in order:

- what is this repo?
- how do I start it locally?
- what does it expose by default?
- how do I replace the demo module?
- where is the old showcase?

If any answer requires reading a second document first, tighten the README again.

**Step 3: Commit the final starter polish**

```bash
git add README.md docs/showcase/multisurface.md configs/config.example.yaml Makefile scripts/verify.sh
git commit -m "docs: finalize pure starter onboarding"
```

**Step 4: Sanity-check the git diff**

Run:

```bash
git status --short
git diff --stat main...HEAD
```

Expected:

- no uncommitted files
- diff is concentrated in starter-owned runtime/config/docs/module paths
- no showcase business modules remain in `main`

