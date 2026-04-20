# Stage 1 Core Infra TDD Handoff

- 日期：2026-03-18
- 执行者：Codex
- 适用范围：article-sentinel 阶段一基建（`pkg/config`、`pkg/logger`、`pkg/database`）

## 1. 当前阶段交付边界

- 已完成：
  - `pkg/config`：强类型配置模型与 `Load(path string)` 入口
  - `pkg/logger`：JSON `slog` 初始化、`stdout/file/both` 输出、跨日主动轮转调度
  - `pkg/database`：MySQL / Redis 集中式 bootstrap、连接池参数设置、`Close()` 生命周期回收
- 明确未完成：
  - `cmd/server/main.go`
  - 原生 `http.ServeMux` 与 `huma/v2` 装配
  - middleware `Chain`
  - 统一 `code/msg/data` 响应封装
  - 业务 handler / service / repository

## 2. 核心 Struct 契约

### 2.1 `pkg/config`

#### `config.Config`

```go
type Config struct {
    App   AppConfig
    MySQL MySQLConfig
    Redis RedisConfig
    Log   LogConfig
}
```

- 责任：作为阶段一所有基础设施初始化的唯一配置输入对象
- 约束：只承载配置数据，不做任何资源初始化和缓存

#### `config.AppConfig`

- 关键字段：
  - `Name string`
  - `Env string`
  - `Host string`
  - `Port int`
- 关键约束：
  - `Name` 为必填
  - `Port` 默认 `8080`

#### `config.MySQLConfig`

- 关键字段：
  - `Host string`
  - `Port int`
  - `User string`
  - `Password string`
  - `DBName string`
  - `Charset string`
  - `ParseTime bool`
  - `Loc string`
  - `MaxOpenConns int`
  - `MaxIdleConns int`
  - `ConnMaxLifetimeMinutes int`
- 关键约束：
  - `Host` / `User` / `Password` / `DBName` 为必填
  - 连接池参数全部来自配置，不允许硬编码在 bootstrap 层

#### `config.RedisConfig`

- 关键字段：
  - `Addr string`
  - `Password string`
  - `DB int`
  - `PoolSize int`
  - `MinIdleConns int`
- 关键约束：
  - `Addr` 为必填

#### `config.LogConfig`

- 关键字段：
  - `Level string`
  - `Format string`
  - `Output string`
  - `Dir string`
  - `Filename string`
  - `MaxSizeMB int`
  - `MaxBackups int`
  - `MaxAgeDays int`
  - `Compress bool`
  - `RotateDaily bool`
- 关键约束：
  - 当前实现固定使用 JSON handler
  - `Output` 仅支持 `stdout` / `file` / `both`

### 2.2 `pkg/logger`

#### `logger.New`

```go
func New(cfg config.LogConfig) (*slog.Logger, func() error, error)
```

- 输入：
  - `config.LogConfig`
- 输出：
  - `*slog.Logger`
  - `cleanup func() error`
  - `error`
- 行为契约：
  - 输出格式固定为 JSON
  - `stdout`：仅写标准输出
  - `file`：仅写日志文件
  - `both`：同时写标准输出和日志文件
  - 文件输出由 `lumberjack.Logger` 承担
  - 当 `RotateDaily=true` 且启用文件输出时，启动后台跨日轮转 goroutine

### 2.3 `pkg/database`

#### `database.Resources`

```go
type Resources struct {
    MySQL *gorm.DB
    Redis *redis.Client
}
```

- 责任：聚合阶段一全部基础设施客户端，供未来 `main` 层注入

#### `database.Bootstrap`

```go
func Bootstrap(cfg *config.Config) (*Resources, error)
```

- 输入：
  - 完整 `*config.Config`
- 输出：
  - 聚合后的 `*Resources`
  - `error`
- 行为契约：
  - 固定顺序：`MySQL -> Redis`
  - 每个外部依赖在启动时即做 fail-fast 连通性检查
  - 任一步失败时要回收之前已初始化的资源

#### `(*Resources).Close`

```go
func (r *Resources) Close() error
```

- 行为契约：
  - `nil` receiver 安全
  - 空资源安全
  - 关闭 Redis 和 MySQL 时聚合错误返回

## 3. Interface 定义与抽象边界

### 3.1 对外公开接口

- 当前阶段没有额外交付公开 `interface`
- 当前对外契约主要由以下导出对象构成：
  - `config.Config`
  - `config.Load`
  - `logger.New`
  - `database.Resources`
  - `database.Bootstrap`
  - `(*database.Resources).Close`

### 3.2 包内测试/调度抽象

#### `logger.rotationTicker`

```go
type rotationTicker interface {
    C() <-chan time.Time
    Stop()
}
```

- 作用：隔离真实 `time.Ticker`，便于测试跨日轮转调度

#### `logger.fileRotator`

```go
type fileRotator interface {
    Rotate() error
}
```

- 作用：隔离 `lumberjack.Logger.Rotate()`，便于验证跨日轮转逻辑

## 4. 当前装配链路与后续接入点

### 4.1 当前阶段推荐装配顺序

```go
cfg, err := config.Load("configs/config.yaml")
log, cleanupLog, err := logger.New(cfg.Log)
resources, err := database.Bootstrap(cfg)
defer cleanupLog()
defer resources.Close()
```

- 先加载配置，再初始化 logger，再初始化 database
- logger 不依赖数据库
- database 依赖完整配置，但不依赖 logger 实例

### 4.2 后续 `cmd/server/main.go` 推荐接入位置

- `main` 层负责：
  - 调用 `config.Load`
  - 调用 `logger.New`
  - 调用 `database.Bootstrap`
  - 组装 `http.ServeMux`
  - 组装 `huma.API`
  - 注入 logger / database resources / config 给 handler 或 service

### 4.3 Huma / middleware 流转链路说明

- 当前阶段尚未实现 Huma middleware
- 下一阶段建议测试与实现时采用如下链路：

```text
HTTP Request
-> net/http middleware（请求 ID、超时、恢复、访问日志）
-> huma API adapter
-> handler / use case
-> 访问 database.Resources
-> 写回 huma response
```

- 当前文档中的 Huma 链路属于“下一阶段装配契约”，不是本阶段现有代码

## 5. 下一阶段优先补测清单

### 5.1 `pkg/config`

- `Load("")` 返回明确错误
- YAML 语法错误时返回包装错误
- 默认值是否在未提供字段时生效
- 环境变量覆盖是否覆盖 `bool` / `int` / `string` 三类字段

### 5.2 `pkg/logger`

- `Output` 为非法值时返回错误
- `Dir` 不可创建时返回错误
- `cleanup()` 是否允许重复调用
- `RotateDaily=false` 时不启动后台调度
- `both` 模式下标准输出与文件写入是否同时成立

### 5.3 `pkg/database`

- `Bootstrap(nil)` 返回错误
- MySQL 打开失败时直接返回且不泄露资源
- Redis 失败时已创建的 MySQL 连接是否被关闭
- `Close()` 在 MySQL `DB()` 失败时是否正确聚合错误

## 6. 关键边缘情况（Edge Cases）

### 6.1 配置层

- `configs/config.yaml` 缺失
- YAML 字段类型错误，例如字符串写进整数字段
- 环境变量覆盖值非法，例如 `APP_PORT=abc`

### 6.2 日志层

- 日志目录无写权限
- `cleanup()` 被重复调用时，当前 `startDailyRotation` 的 `done` channel 可能出现重复关闭风险
- 跨日时钟跳变、系统时间回拨或时区变化导致轮转判断偏差

### 6.3 数据库层

- MySQL / Redis 启动不可达，fail-fast 直接返回错误
- MySQL 连接池耗尽导致请求阻塞或超时
- Redis 连接池过小导致吞吐抖动
- `PingContext` / `Ping` 依赖固定 5 秒超时，真实生产环境可能需要由上层统一 `context.Context`

### 6.4 并发与生命周期

- 当前阶段未引入共享 map，因此不存在当前实现内的并发 map 读写问题
- 但 logger 后台轮转 goroutine 与应用退出时序需要在下一阶段验证
- `resources.Close()` 与未来 HTTP server shutdown 的调用顺序需要统一管理，避免先关数据库再处理在途请求

## 7. 下一阶段 TDD 建议

- 先补 `cmd/server/main.go` 的装配测试，再写最小实现
- 先定义 middleware 行为测试，再接入 `http.ServeMux` / `huma`
- 对外部依赖的集成测试保持“可选本地 smoke test + 默认单元测试”策略
- 统一以“先失败测试 -> 最小实现 -> 全量验证”的节奏推进
