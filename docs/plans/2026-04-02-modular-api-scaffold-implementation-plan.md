# Modular API Scaffold Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 把 article-sentinel 从“基础设施库集合”升级为可运行的模块化单体 API 脚手架，支持 `server / worker / scheduler / migrate` 多入口、可切换 `MySQL / PostgreSQL` 主数据库，以及 `Redis + Asynq + cron` 的后台任务骨架。

**Architecture:** 本阶段采用“薄入口 + 集中式 bootstrap + 分层 runtime 模块”的方式推进。`cmd/*` 只负责启动，`internal/app/bootstrap` 负责组装共享资源，`internal/api` 负责 HTTP skeleton，`internal/queue` 与 `internal/scheduler` 负责后台执行模型，数据库驱动通过配置切换而不是在业务层分叉。

**Tech Stack:** Go 1.25+, `http.ServeMux`, `github.com/danielgtaylor/huma/v2`, `gorm.io/gorm`, `gorm.io/driver/mysql`, `gorm.io/driver/postgres`, `github.com/redis/go-redis/v9`, `github.com/hibiken/asynq`, `github.com/robfig/cron/v3`, `github.com/golang-migrate/migrate/v4`

---

## Implementation Notes

- 当前仓库已完成阶段一：`pkg/config`、`pkg/logger`、`pkg/database`
- 阶段二要尽量复用阶段一成果，避免推翻已有契约
- 所有功能按 TDD 推进：先写失败测试，再写最小实现，再验证
- 每个任务保持小步提交，避免跨多个子系统的大型混合 diff
- 推荐在新 worktree 中执行本计划

### Task 1: Expand Config Model For Runtime Modes

**Files:**
- Modify: `pkg/config/config.go`
- Modify: `pkg/config/load.go`
- Modify: `pkg/config/config_test.go`
- Modify: `configs/config.example.yaml`

**Step 1: Write the failing tests**

Add tests covering:

```go
func TestLoadReadsDatabaseDriver(t *testing.T) {}
func TestLoadReadsPostgresConfig(t *testing.T) {}
func TestLoadReadsQueueAndSchedulerConfig(t *testing.T) {}
func TestLoadAppliesHTTPTimeoutDefaults(t *testing.T) {}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/config -v`
Expected: FAIL with undefined fields/types for `Database`, `Postgres`, `Queue`, `Scheduler`, or `HTTP`

**Step 3: Write minimal implementation**

Implement new config structs and fields for:

- `HTTPConfig`
- `DatabaseConfig` with `Driver`, `MySQL`, `Postgres`
- `PostgresConfig`
- `QueueConfig` with `AsynqConfig`
- `SchedulerConfig`
- `DocsConfig`

Update `config.example.yaml` to include the new sections.

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/config -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/config/config.go pkg/config/load.go pkg/config/config_test.go configs/config.example.yaml
git commit -m "feat: expand runtime config model"
```

### Task 2: Add PostgreSQL Support To `pkg/database`

**Files:**
- Modify: `pkg/database/resources.go`
- Modify: `pkg/database/database_test.go`
- Modify: `pkg/database/mysql.go`
- Create: `pkg/database/postgres.go`
- Create: `pkg/database/driver.go`

**Step 1: Write the failing tests**

Add tests covering:

```go
func TestBootstrapUsesMySQLDialectorWhenDriverIsMySQL(t *testing.T) {}
func TestBootstrapUsesPostgresDialectorWhenDriverIsPostgres(t *testing.T) {}
func TestBuildPostgresDSN(t *testing.T) {}
func TestBootstrapReturnsErrorForUnsupportedDriver(t *testing.T) {}
```

Use test seams/helpers instead of hitting a real database.

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/database -v`
Expected: FAIL with undefined postgres builder/driver selection code

**Step 3: Write minimal implementation**

Implement:

- `buildPostgresDSN(cfg config.PostgresConfig) string`
- driver switch based on `cfg.Database.Driver`
- unsupported driver guard
- keep `Resources.Close()` contract unchanged

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/database -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/database/resources.go pkg/database/database_test.go pkg/database/mysql.go pkg/database/postgres.go pkg/database/driver.go
git commit -m "feat: support postgres driver selection"
```

### Task 3: Build Shared Bootstrap And Lifecycle Layer

**Files:**
- Create: `internal/app/bootstrap/runtime.go`
- Create: `internal/app/bootstrap/server.go`
- Create: `internal/app/bootstrap/worker.go`
- Create: `internal/app/bootstrap/scheduler.go`
- Create: `internal/app/lifecycle/shutdown.go`
- Create: `internal/app/bootstrap/bootstrap_test.go`

**Step 1: Write the failing tests**

Add tests covering:

```go
func TestBuildServerRuntimeReturnsSharedResources(t *testing.T) {}
func TestBuildWorkerRuntimeReturnsSharedResources(t *testing.T) {}
func TestShutdownRunsClosersInReverseOrder(t *testing.T) {}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/app/... -v`
Expected: FAIL with missing bootstrap/lifecycle packages and symbols

**Step 3: Write minimal implementation**

Implement a shared runtime struct that carries:

- loaded config
- logger
- database resources
- cleanup closers

Implement shutdown helper that closes resources in reverse registration order.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/app/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/app/bootstrap internal/app/lifecycle
git commit -m "feat: add shared bootstrap lifecycle layer"
```

### Task 4: Add HTTP Server Skeleton With Huma And Middleware

**Files:**
- Create: `cmd/server/main.go`
- Create: `internal/api/register/router.go`
- Create: `internal/api/handlers/health.go`
- Create: `internal/api/handlers/ready.go`
- Create: `internal/api/response/response.go`
- Create: `internal/middleware/requestid.go`
- Create: `internal/middleware/recover.go`
- Create: `internal/middleware/timeout.go`
- Create: `internal/middleware/accesslog.go`
- Create: `internal/api/handlers/health_test.go`

**Step 1: Write the failing tests**

Add tests covering:

```go
func TestHealthzReturnsAlive(t *testing.T) {}
func TestReadyzReturnsDependencyStatus(t *testing.T) {}
func TestResponseHelpersReturnStandardShape(t *testing.T) {}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/api/... ./internal/middleware/... -v`
Expected: FAIL with missing router/handler/response code

**Step 3: Write minimal implementation**

Implement:

- Huma-backed router registration
- `/healthz`
- `/readyz`
- `/openapi.json`
- `/docs`
- standard response struct
- request id / recover / timeout / access log middleware chain
- `cmd/server/main.go` using bootstrap + graceful shutdown

**Step 4: Run test to verify it passes**

Run: `go test ./internal/api/... ./internal/middleware/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/server/main.go internal/api internal/middleware
git commit -m "feat: add huma server skeleton"
```

### Task 5: Add Redis Queue Skeleton With Asynq

**Files:**
- Create: `internal/queue/tasks/tasks.go`
- Create: `internal/queue/tasks/payload.go`
- Create: `internal/queue/asynq/client.go`
- Create: `internal/queue/asynq/server.go`
- Create: `internal/queue/asynq/handlers.go`
- Create: `internal/queue/asynq/asynq_test.go`
- Create: `cmd/worker/main.go`

**Step 1: Write the failing tests**

Add tests covering:

```go
func TestNewTaskBuildsStableTaskNameAndPayload(t *testing.T) {}
func TestEnqueueHelperBuildsAsynqTask(t *testing.T) {}
func TestRegisterHandlersReturnsMuxWithKnownTaskTypes(t *testing.T) {}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/queue/... -v`
Expected: FAIL with missing task/asynq symbols

**Step 3: Write minimal implementation**

Implement:

- canonical task names
- payload structs
- enqueue helper
- Asynq server builder
- handler registration mux
- `cmd/worker/main.go` using shared bootstrap

**Step 4: Run test to verify it passes**

Run: `go test ./internal/queue/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/worker/main.go internal/queue
git commit -m "feat: add asynq worker skeleton"
```

### Task 6: Add Scheduler Skeleton With Cron-Triggered Enqueue

**Files:**
- Create: `internal/scheduler/scheduler.go`
- Create: `internal/scheduler/jobs.go`
- Create: `internal/scheduler/scheduler_test.go`
- Create: `cmd/scheduler/main.go`

**Step 1: Write the failing tests**

Add tests covering:

```go
func TestRegisterJobsAddsCronEntries(t *testing.T) {}
func TestScheduledJobOnlyEnqueuesTask(t *testing.T) {}
```

Use an interface seam for enqueue behavior instead of running a real queue.

**Step 2: Run test to verify it fails**

Run: `go test ./internal/scheduler -v`
Expected: FAIL with missing scheduler/job code

**Step 3: Write minimal implementation**

Implement:

- cron scheduler builder
- job registration
- job function that only enqueues Asynq tasks
- `cmd/scheduler/main.go` using shared bootstrap

**Step 4: Run test to verify it passes**

Run: `go test ./internal/scheduler -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/scheduler/main.go internal/scheduler
git commit -m "feat: add scheduler skeleton"
```

### Task 7: Add Migration Command And Directory Baseline

**Files:**
- Create: `cmd/migrate/main.go`
- Create: `internal/app/bootstrap/migrate.go`
- Create: `migrations/.keep`
- Create: `internal/app/bootstrap/migrate_test.go`

**Step 1: Write the failing tests**

Add tests covering:

```go
func TestBuildMigrateConfigUsesSelectedDriver(t *testing.T) {}
func TestMigrateCommandRejectsUnsupportedDriver(t *testing.T) {}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/app/bootstrap -run Migrate -v`
Expected: FAIL with missing migrate bootstrap code

**Step 3: Write minimal implementation**

Implement:

- migration DSN selection from current database driver
- `cmd/migrate/main.go` command skeleton
- `migrations/` baseline directory

Do not add real migration SQL yet; just create the runtime skeleton and directory contract.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/app/bootstrap -run Migrate -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/migrate/main.go internal/app/bootstrap/migrate.go internal/app/bootstrap/migrate_test.go migrations/.keep
git commit -m "feat: add migrate command skeleton"
```

### Task 8: Wire README And End-To-End Verification

**Files:**
- Modify: `README.md`
- Modify: `verification.md`
- Create: `internal/modules/example/README.md`

**Step 1: Update README**

Document:

- available commands: `server`, `worker`, `scheduler`, `migrate`
- supported primary DBs: `mysql`, `postgres`
- Redis/Asynq/cron runtime roles
- basic startup flow

**Step 2: Add minimal example module doc**

Write a short stub doc under `internal/modules/example/README.md` describing the intended handler/service/repository split.

**Step 3: Run full verification**

Run:

```bash
go test ./... -v
go mod tidy
go fmt ./...
go vet ./...
go build ./...
```

Expected: PASS

**Step 4: Update verification.md**

Append:

- stage two verification commands
- result summary
- remaining deferred items

**Step 5: Commit**

```bash
git add README.md verification.md internal/modules/example/README.md
git commit -m "docs: record modular scaffold progress"
```
