# Stage 1 Core Infra Implementation Plan

> 执行结果更新（2026-04-02）：本计划中的阶段一任务已完成并合并到 `main`。实际交付文件已包含 `go.mod`、`configs/config.example.yaml`、`pkg/config`、`pkg/logger`、`pkg/database`、`verification.md` 与 `docs/plans/2026-03-18-stage1-core-infra-tdd-spec.md`。本文档保留为实施记录与后续阶段参考。

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 完成 article-sentinel 阶段一核心基建，实现 `pkg/config`、`pkg/logger`、`pkg/database` 三个基础包，并为后续 `cmd/server/main.go` 装配提供稳定入口。

**Architecture:** 本阶段采用“显式配置加载 -> 显式 logger 初始化 -> 集中式 database bootstrap”的装配顺序，避免包级隐式状态扩散。`config` 只负责解析配置，`logger` 只负责输出与轮转，`database` 只负责 MySQL / Redis 初始化与关闭；真实 HTTP 启动与业务代码继续留在后续阶段。

**Tech Stack:** Go 1.22+, `http.ServeMux`, `github.com/ilyakaznacheev/cleanenv`, `log/slog`, `gopkg.in/natefinch/lumberjack.v2`, `gorm.io/gorm`, `gorm.io/driver/mysql`, `github.com/redis/go-redis/v9`

---

## Implementation Notes

- 仓库远程地址为 `git@github.com:dovetaill/article-sentinel.git`，本计划默认 module path 为 `github.com/dovetaill/article-sentinel`
- 默认运行配置文件路径是 `configs/config.yaml`
- 提交仓库的样例文件使用 `configs/config.example.yaml`
- 本阶段不创建 `cmd/server/main.go`
- 所有步骤保持小步提交，避免一次性堆大 diff

### Task 1: Bootstrap Go Module And Base Layout

**Files:**
- Create: `go.mod`
- Create: `go.sum`
- Create: `configs/config.example.yaml`
- Create: `pkg/config/.keep`
- Create: `pkg/logger/.keep`
- Create: `pkg/database/.keep`

**Step 1: Initialize module**

Run:

```bash
go mod init github.com/dovetaill/article-sentinel
```

Expected: 创建 `go.mod`

**Step 2: Add phase-one dependencies**

Run:

```bash
go get github.com/ilyakaznacheev/cleanenv \
  gopkg.in/natefinch/lumberjack.v2 \
  gorm.io/gorm \
  gorm.io/driver/mysql \
  github.com/redis/go-redis/v9
```

Expected: `go.mod` / `go.sum` 写入依赖

**Step 3: Create base directories and tracked example config**

Write:

```yaml
app:
  name: article-sentinel
  env: local
  host: 0.0.0.0
  port: 8080

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
  password: ""
  db: 0
  pool_size: 10
  min_idle_conns: 2

log:
  level: info
  format: json
  output: both
  dir: logs
  filename: app.log
  max_size_mb: 100
  max_backups: 14
  max_age_days: 30
  compress: false
  rotate_daily: true
```

Target file: `configs/config.example.yaml`

**Step 4: Tidy dependencies**

Run:

```bash
go mod tidy
```

Expected: `go.mod` / `go.sum` 干净可提交

**Step 5: Commit**

Run:

```bash
git add go.mod go.sum configs/config.example.yaml pkg/config/.keep pkg/logger/.keep pkg/database/.keep
git commit -m "chore: bootstrap phase1 infra module"
```

### Task 2: Implement `pkg/config`

**Files:**
- Remove: `pkg/config/.keep`
- Create: `pkg/config/config.go`
- Create: `pkg/config/load.go`
- Create: `pkg/config/config_test.go`

**Step 1: Write the failing tests**

Write tests covering:

```go
func TestLoadReadsYAML(t *testing.T) {}
func TestLoadEnvOverridesYAML(t *testing.T) {}
func TestLoadReturnsErrorForMissingRequiredFields(t *testing.T) {}
```

Test focus:

- `Load(path string)` 能读取 YAML
- 环境变量可覆盖 YAML 中的端口、日志级别等字段
- 缺失必要配置时返回错误

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./pkg/config -v
```

Expected: FAIL，报 `Load`、`Config` 或相关类型未定义

**Step 3: Write minimal implementation**

Implement:

```go
type Config struct {
    App   AppConfig
    MySQL MySQLConfig
    Redis RedisConfig
    Log   LogConfig
}

func Load(path string) (*Config, error)
```

Key rules:

- `config.yaml` 为主来源
- 使用 `cleanenv` 支持环境变量覆盖
- `App / MySQL / Redis / Log` 必须是分层强类型结构
- 用 tag 明确默认值、env 名称、必要字段

**Step 4: Run tests to verify they pass**

Run:

```bash
go test ./pkg/config -v
```

Expected: PASS

**Step 5: Commit**

Run:

```bash
git add pkg/config/config.go pkg/config/load.go pkg/config/config_test.go
git commit -m "feat: add config loading package"
```

### Task 3: Implement `pkg/logger`

**Files:**
- Remove: `pkg/logger/.keep`
- Create: `pkg/logger/logger.go`
- Create: `pkg/logger/rotate.go`
- Create: `pkg/logger/logger_test.go`

**Step 1: Write the failing tests**

Write tests covering:

```go
func TestNewReturnsJSONLogger(t *testing.T) {}
func TestNewSupportsStdoutOnly(t *testing.T) {}
func TestDailyRotatorCallsRotateAfterDayChange(t *testing.T) {}
```

Test focus:

- logger 输出为 JSON
- `output=stdout` 时不创建文件 writer
- 启用 `rotate_daily` 时，跨日逻辑会触发一次 `Rotate()`

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./pkg/logger -v
```

Expected: FAIL，报 `New`、轮转调度或测试依赖类型未定义

**Step 3: Write minimal implementation**

Implement:

```go
func New(cfg config.LogConfig) (*slog.Logger, func() error, error)
```

Implementation rules:

- 输出格式固定为 JSON
- 支持 `stdout`、`file`、`both`
- 文件侧接入 `lumberjack.Logger`
- 将“跨日主动轮转”封装为独立调度逻辑，返回 cleanup 函数停止 goroutine / ticker
- 若未启用文件输出，则 cleanup 可以是空实现

**Step 4: Run tests to verify they pass**

Run:

```bash
go test ./pkg/logger -v
```

Expected: PASS

**Step 5: Commit**

Run:

```bash
git add pkg/logger/logger.go pkg/logger/rotate.go pkg/logger/logger_test.go
git commit -m "feat: add slog logger bootstrap"
```

### Task 4: Implement `pkg/database`

**Files:**
- Remove: `pkg/database/.keep`
- Create: `pkg/database/resources.go`
- Create: `pkg/database/mysql.go`
- Create: `pkg/database/redis.go`
- Create: `pkg/database/database_test.go`

**Step 1: Write the failing tests**

Write tests covering:

```go
func TestBuildMySQLDSN(t *testing.T) {}
func TestBuildRedisOptions(t *testing.T) {}
func TestResourcesCloseIsSafe(t *testing.T) {}
```

Test focus:

- MySQL DSN 构造符合配置预期
- Redis options 正确映射连接池参数
- `Close()` 在 nil / 空资源场景下可安全调用

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./pkg/database -v
```

Expected: FAIL，报 `Bootstrap`、`Resources`、辅助构造函数未定义

**Step 3: Write minimal implementation**

Implement:

```go
type Resources struct {
    MySQL *gorm.DB
    Redis *redis.Client
}

func Bootstrap(cfg *config.Config) (*Resources, error)
func (r *Resources) Close() error
```

Implementation rules:

- 启动阶段连接 MySQL 与 Redis，并执行 fail-fast 检查
- 连接池参数全部来自配置
- `Bootstrap` 内部顺序固定：MySQL -> Redis -> health check -> 聚合返回
- 任一步失败都要释放已初始化资源
- 不引入 Repository、事务封装、读写分离、多数据源

**Step 4: Run tests to verify they pass**

Run:

```bash
go test ./pkg/database -v
```

Expected: PASS

**Step 5: Optional local smoke check**

Prepare local file:

```bash
cp configs/config.example.yaml configs/config.yaml
```

Then run:

```bash
go test ./pkg/database -run TestBuild -v
```

Expected: PASS；如果后续补充 integration test，再单独使用本地 MySQL / Redis 验证 `Bootstrap`

**Step 6: Commit**

Run:

```bash
git add pkg/database/resources.go pkg/database/mysql.go pkg/database/redis.go pkg/database/database_test.go
git commit -m "feat: add database bootstrap package"
```

### Task 5: Verify Phase-One Deliverables

**Files:**
- Modify: `README.md`
- Create: `verification.md`

**Step 1: Update README quick-start section**

Add concise notes for:

- module 已初始化
- 默认读取 `configs/config.yaml`
- 仓库提交的是 `configs/config.example.yaml`
- 阶段一只完成基础包，不包含 server 启动

**Step 2: Run package tests**

Run:

```bash
go test ./pkg/... -v
```

Expected: PASS

**Step 3: Run full verification**

Run:

```bash
go test ./... -v
```

Expected: PASS

**Step 4: Write verification summary**

Write `verification.md` with:

- 执行命令
- 结果摘要
- 未覆盖项：真实 MySQL / Redis 联通性仍待后续启动阶段补充

**Step 5: Commit**

Run:

```bash
git add README.md verification.md
git commit -m "docs: record phase1 verification"
```

## Handoff Checklist

- `go.mod` 已使用 `github.com/dovetaill/article-sentinel`
- `configs/config.example.yaml` 已提交，`configs/config.yaml` 只作为本地运行文件
- `pkg/config` 只负责配置解析
- `pkg/logger` 返回 `*slog.Logger` 和 cleanup
- `pkg/database` 返回聚合 `Resources` 并支持 `Close()`
- 阶段一验证命令全部可执行
- `cmd/server/main.go` 保持未实现，留待后续阶段

## Deferred Work

- 原生 `http.ServeMux` + Huma 路由装配
- middleware `Chain`
- 统一 `code, msg, data` 响应包装
- 示例业务流转
