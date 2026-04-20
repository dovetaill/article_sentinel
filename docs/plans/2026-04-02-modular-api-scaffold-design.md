# Modular API Scaffold Design

- 日期：2026-04-02
- 执行者：Codex
- 状态：已评审通过，待进入实施计划

## 背景

article-sentinel 当前已经完成阶段一核心基建，仓库中已有：

- `pkg/config`：YAML + env override 的强类型配置加载
- `pkg/logger`：JSON `slog`、`stdout/file/both` 输出、跨日主动轮转
- `pkg/database`：MySQL / Redis bootstrap、连接池参数、资源关闭

但项目仍停留在“基础设施库”层面，尚未成为一个可直接启动、可直接复用的 API 脚手架。下一阶段的目标不是补单点功能，而是把项目定型为一个可长期维护的综合型后端模板。

## 产品定位

article-sentinel 的目标定位为：`模块化单体 API 脚手架`。

它不是：

- 纯示例仓库
- 微服务平台模板
- 一开始就包含复杂插件系统的重量级框架

它应该是：

- 一个仓库、多个独立运行入口
- 一套统一的配置、日志、数据库、Redis、任务定义
- 支持 HTTP API、后台 worker、定时调度、数据库迁移
- 可按配置在 `MySQL` 和 `PostgreSQL` 之间切换主数据库
- 适合业务项目从 0 到 1 起步，再逐步演化

## 目标

下一阶段要完成的核心目标：

1. 让仓库具备可运行的 `server / worker / scheduler / migrate` 入口
2. 让主数据库可以按配置选择 `MySQL` 或 `PostgreSQL`
3. 让 Redis 同时服务于缓存/任务系统，并接入 `Asynq`
4. 让 `cron` 只负责触发，把真实任务统一投递给队列
5. 让 API 层具备中间件、健康检查、OpenAPI、统一响应等脚手架能力
6. 让资源初始化、生命周期管理、优雅停机具备统一装配方式

## 非目标

本阶段明确不做：

- 完整认证鉴权体系
- 多租户、权限系统、限流平台
- 复杂工作流引擎
- 大量示例业务模块
- 分布式微服务拆分
- 监控平台、追踪平台的完整接入

## 核心设计决策

### 1. 运行形态

采用 `一个仓库，多个入口/子命令` 的模式：

- `cmd/server`
- `cmd/worker`
- `cmd/scheduler`
- `cmd/migrate`

原因：

- HTTP API、队列消费、定时调度、迁移命令的生命周期不同
- 部署时通常需要独立扩缩容
- 单进程同时承载 API + worker + scheduler 会使职责边界模糊

### 2. 数据库策略

主数据库采用 `同一套代码，按配置选择 MySQL 或 PostgreSQL`。

做法：

- `config.Database.Driver` 显式声明 `mysql` 或 `postgres`
- `pkg/database` 根据配置选择对应的 GORM dialector
- 业务层不感知具体数据库，只依赖 `*gorm.DB`

原因：

- 对脚手架使用者最友好
- 模型与仓储代码可以保持一致
- 避免为“同时双写数据库”引入额外复杂度

### 3. Redis / Queue / Scheduler 策略

任务体系采用：`Redis + Asynq + cron`。

约束如下：

- Redis 是统一基础设施
- `Asynq` 负责异步任务、延迟任务、重试
- `cron` 只负责按时间触发 enqueue
- 真正业务任务统一由 worker 消费执行

推荐链路：

```text
cron -> enqueue asynq task -> worker consume -> service -> repository/db/cache
```

原因：

- 定时任务与异步任务统一到一套执行/重试模型里
- worker 能独立扩容
- 失败重试、优先级、延迟投递可以交给成熟库处理

### 4. HTTP 层策略

HTTP 层采用：`http.ServeMux + Huma v2`。

必须提供：

- `/healthz`
- `/readyz`
- `/openapi.json`
- `/docs`

必须具备：

- request id
- panic recover
- timeout
- access log
- 统一响应结构

原因：

- `ServeMux` 简单、原生、稳定
- `Huma` 能让 OpenAPI 描述与 handler 定义保持一致
- 对脚手架场景来说，生成文档与类型化输入输出比“自写路由层”更值钱

### 5. Migration 策略

默认迁移工具选 `golang-migrate`。

原因：

- 足够主流，适合模板仓库
- 对 `MySQL` 与 `PostgreSQL` 都成熟
- 迁移文件结构清晰，便于项目接入方理解

Atlas 保留为后续增强方向，但不是这一阶段的默认方案。

## 目录设计

建议目录结构：

```text
cmd/
  server/
  worker/
  scheduler/
  migrate/

configs/
  config.example.yaml

pkg/
  config/
  logger/
  database/

internal/
  app/
    bootstrap/
    lifecycle/
  api/
    register/
    handlers/
    response/
  middleware/
  queue/
    asynq/
    tasks/
  scheduler/
  modules/
    example/
      handler/
      service/
      repository/
      dto/

migrations/
docs/plans/
```

设计原则：

- `cmd/*` 只保留薄入口
- 统一装配逻辑集中在 `internal/app/bootstrap`
- API、中间件、队列、调度彼此解耦
- 后续业务模块放到 `internal/modules/*`

## 配置模型

建议在现有配置上扩展为：

```yaml
app:
http:
database:
  driver:
  mysql:
  postgres:
redis:
log:
queue:
  asynq:
scheduler:
docs:
```

建议说明：

- `database.driver`：主数据库驱动开关
- `database.mysql` / `database.postgres`：分离两套连接配置
- `redis`：通用 Redis 配置，供缓存与 Asynq 共用
- `queue.asynq`：worker 并发、队列优先级、strict priority 等参数
- `scheduler`：cron 开关、时区、任务启用项
- `docs`：Swagger/OpenAPI 开关
- `http`：read/write/idle timeout

## 装配与生命周期

所有入口都采用统一装配顺序：

```text
load config -> init logger -> init database/redis -> init runtime component -> run -> graceful shutdown
```

- `server`：初始化 HTTP server、router、middleware、docs、health/readiness
- `worker`：初始化 Asynq server 与 task handler
- `scheduler`：初始化 cron，并在 job 内只做 enqueue
- `migrate`：执行 migration 命令后退出

优雅停机要求：

- 捕获系统信号
- 先停止接收新流量/新任务
- 再等待在途请求或任务处理完毕
- 最后关闭数据库、Redis、日志资源

## API 约定

响应结构推荐统一为：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

错误响应至少包含：

- `code`
- `msg`
- `request_id`
- 可选 `details`

健康检查定义：

- `/healthz`：进程活着即可返回成功
- `/readyz`：依赖资源 ready 才成功，例如数据库、Redis、任务系统

## 风险点

1. 阶段二如果一次性做太多业务示例，容易偏离“脚手架内核”目标
2. `pkg/database` 当前已经同时持有 SQL 与 Redis 资源，阶段二要谨慎演进，避免接口抖动过大
3. Asynq 与 cron 同时引入后，必须统一任务命名与 payload 结构，否则后续维护混乱
4. 主数据库驱动切换会扩展测试矩阵，必须保持配置层与 bootstrap 层足够清晰

## 推荐实施顺序

1. 先做 `cmd/server`、`cmd/worker`、`cmd/scheduler`、`cmd/migrate` 薄入口
2. 再做 `internal/app/bootstrap` 与生命周期管理
3. 再扩展 `pkg/database` 支持 `postgres`
4. 再落 HTTP skeleton、middleware、health/readiness、OpenAPI/docs
5. 再落 `Asynq` 与 `cron` 骨架
6. 最后补一个最小示例模块来验证整体链路

## 参考资料

- Huma middleware: `https://huma.rocks/features/middleware/`
- Huma graceful shutdown: `https://huma.rocks/how-to/graceful-shutdown`
- Go `http.Server`: `https://pkg.go.dev/net/http#Server`
- Go signal handling: `https://pkg.go.dev/os/signal#Notify`
- GORM connections: `https://gorm.io/docs/connecting_to_the_database.html`
- GORM DBResolver: `https://gorm.io/docs/dbresolver.html`
- GORM migration: `https://gorm.io/docs/migration.html`
- Asynq: `https://github.com/hibiken/asynq`
- robfig/cron v3: `https://pkg.go.dev/github.com/robfig/cron/v3`
