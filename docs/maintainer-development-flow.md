# 维护开发手册

- 日期：2026-04-29
- 执行者：Codex
- 适用范围：`article-sentinel` 当前一期代码基线

## 1. 先读结论

### 1.1 当前 worker / scheduler 到底有没有在用

- `worker`：**仍然在承载真实业务主链路**
  - `POST /api/v1/article-inspect/tasks` 现在会先把 `task + task_keywords + outbox` 在一个事务里落到 MySQL。
  - 请求内会做一次 optimistic relay；如果 Redis / Asynq 暂时不可用，任务不会再被补偿删除，而是留下 pending outbox 供后续 retry。
  - 如果 server 启动时 dispatcher 初始化失败，`internal/api/register/router.go` 现在会直接打可观测日志，而不是静默退化。
  - `cmd/worker/main.go` 启动 Asynq worker。
  - `internal/queue/asynq/handlers.go` 已注册 `articleinspect:run-task` 的消费逻辑。
  - `internal/modules/articleinspect/worker/executor.go` 会真正分页扫描文稿、写入结果、更新任务状态。

- `scheduler`：**入口存在，默认关闭；启用后承载轻量控制面任务**
  - `configs/config.local.yaml` 默认是 `scheduler.enabled: false`。
  - `cmd/scheduler/main.go` 只有在 `scheduler.enabled=true` 时才会注册 job。
  - 当前注册的 job 至少包括：
    - `runtime:heartbeat`：验证 `cron -> queue -> worker` 链路
    - articleinspect task outbox relay：claim / relay 可投递消息，并对失败消息做 retry / dead-letter 收口
    - articleinspect task outbox cleanup：清理保留期到期的 `dispatched` / `dead_letter` 消息
  - 也就是说：**scheduler 仍然不是正文扫描这类重业务执行器，但现在已经是 outbox control-plane 的一部分。**

### 1.2 当前 outbox control-plane 的最小语义

- 状态机：`pending -> claimed -> dispatched/dead_letter`
- `claimed` 代表某个 relay 已拿到 lease；如果 `claim_until` 过期，后续 relay 可以重接管
- poison message（例如 payload 损坏、message type 不支持）不会无限重试，而会进入 `dead_letter`
- cleanup 只清理 `retained_until` 已到期的 `dispatched` / `dead_letter`；不应碰 `pending` / `claimed`

### 1.3 维护时应该怎么理解这四个入口

| 入口 | 角色 | 说明 |
| --- | --- | --- |
| `cmd/server` | API 服务 | 提供 HTTP 接口，负责同步校验、任务落库与 optimistic relay |
| `cmd/worker` | 异步消费者 | 消费 Asynq 任务，真正执行重活 |
| `cmd/scheduler` | 轻量控制面 | 负责 heartbeat、outbox relay / retry / cleanup 这类轻量后台任务，不做正文扫描 |
| `cmd/migrate` | 数据迁移入口 | 执行 `migrations/*.sql`，再同步已注册业务模型 |

### 1.4 哪些文件是“现状真相”

如果文档、设计稿、历史计划和代码说法不一致，请按下面优先级判断：

1. `internal/api/register/router.go`
2. `internal/api/register/router_test.go`
3. `internal/queue/asynq/handlers.go`
4. `internal/scheduler/`
5. `README.md`
6. 本手册

`docs/plans/` 下的内容更适合回看设计和实施过程，不建议直接拿来当运行手册。

## 2. 当前运行边界

### 2.1 应用层鉴权现状

当前仓库里已经有鉴权相关代码，例如：

- `internal/middleware/auth.go`
- `internal/middleware/authorize.go`
- `internal/identity/`

但要注意：

- 当前 `internal/api/register/router.go` 实际挂载的中间件只有：
  - `RequestID`
  - `Recover`
  - `Timeout`
  - `AccessLog`
- 也就是说，**当前应用层鉴权链路还没有真正接到 `NewRouter` 上**

维护含义：

- 如果线上访问控制依赖网关、Nginx、内网隔离或上层统一鉴权，这部分责任需要由外层系统承担
- 如果后续要启用应用层鉴权，应优先从 `internal/api/register/router.go` 接线，而不是在业务 handler 里零散判断

### 2.2 健康检查与 smoke 的真实边界

- `/healthz`：更偏“进程活着”
- `/readyz`：当前只检查 runtime 里的依赖资源是否已装配，不等于真实数据库/Redis 深度探活
- `make smoke`：当前主要验证基础 HTTP 端点和 starter demo，不等于完整覆盖巡检业务主链路

所以不要把下面两件事混为一谈：

- `server 能启动`
- `articleinspect 主业务链路已经可用`

如果你要验证巡检主链路，至少还需要补一遍：

1. `server` 已启动
2. `worker` 已启动
3. 新建一条巡检任务
4. 观察任务状态从 `pending` 进入 `running / success / partial_success / failed`

## 3. 本地开发的推荐路径

### 3.1 最短启动方式

推荐优先使用：

```bash
make up
make migrate
make dev
```

注意：

- `make dev` 当前会同时启动：
  - API server
  - worker
  - scheduler 进程
  - 前端 admin dev server
- 但 **scheduler 进程启动 != scheduler job 生效**，是否真的注册 heartbeat / outbox relay 仍然取决于 `scheduler.enabled`
- 如果你只想单独调试某个进程，可以用：

```bash
make dev-api
make dev-worker
make dev-scheduler
make dev-admin
```

### 3.2 日常验证最小闭环

1. 启动 `make up`
2. 执行 `make migrate`
3. 启动 `make dev` 或至少启动 `server + worker`
4. 打开 `http://127.0.0.1:8080/docs`
5. 新建一条巡检任务
6. 确认任务从 `pending` 进入 `running / success / partial_success / failed`

如果任务一直停留在 `pending`：

- 先确认 `worker` 是否真的启动
- 再确认 Redis 是否连通
- 再检查 `xt_article_inspect_task_outbox` 里是否有 pending message
- 如果有 pending outbox，确认 `scheduler.enabled` 是否已打开，以及 `internal/scheduler/jobs.go` 的 relay / cleanup job 是否在跑
- 最后检查 `internal/queue/asynq/handlers.go` 是否注册了对应 task type

### 3.3 Outbox 排障与人工恢复最短路径

1. 先确认 backlog 属于哪一类：
   - `pending`：等待 relay 或下一次 `next_attempt_at`
   - `claimed`：正在被某个 relay 处理，或 lease 已过期
   - `dead_letter`：需要人工判断是否 requeue
2. 优先查表：
   - pending backlog：`status = 'pending'`
   - 过期 lease：`status = 'claimed' AND claim_until < UTC_TIMESTAMP()`
   - 死信：`status = 'dead_letter'`
3. 如果需要人工恢复，不要临场手写 SQL，优先使用：
   - `scripts/articleinspect_outbox_requeue.sql`
4. 默认保留 `attempt_count`
   - 这能保留失败历史，避免人工恢复把问题证据抹掉
5. cleanup 的边界要记住：
   - 只应删除 `retained_until` 已过期的 `dispatched` / `dead_letter`
   - 不应删除仍待处理的 `pending` / `claimed`

## 4. 目录地图：改什么应该去哪里

| 场景 | 主要目录 / 文件 |
| --- | --- |
| 配置模型与加载 | `pkg/config/`、`configs/config.example.yaml`、`configs/config.local.yaml` |
| 运行时资源装配 | `internal/app/bootstrap/` |
| HTTP 路由总装配 | `internal/api/register/router.go` |
| 路由回归真相 | `internal/api/register/router_test.go` |
| 业务模块 | `internal/modules/<module>/` |
| 队列 payload / enqueue helper | `internal/queue/tasks/`、`internal/queue/asynq/client.go` |
| Worker 消费注册 | `internal/queue/asynq/handlers.go` |
| Scheduler 注册与 job | `internal/scheduler/` |
| 迁移与 schema 注册 | `migrations/`、`internal/app/bootstrap/schema.go` |
| 本地开发脚本 | `Makefile`、`scripts/dev.sh` |
| 前端页面与导航 | `web/admin/src/pages/`、`web/admin/src/routes.tsx` |
| 前端接口调用 | `web/admin/src/services/` |

额外建议：

- 看“当前接口是否真的存在”，优先看 `internal/api/register/router_test.go`
- 看“当前前端页面是否真的存在”，优先看 `web/admin/src/routes.tsx`

## 5. 如何增加配置

配置链路是：

```text
configs/*.yaml / 环境变量
        -> pkg/config/config.go
        -> pkg/config/load.go
        -> bootstrap.Runtime.Config
        -> server / worker / scheduler / migrate 使用
```

### 标准步骤

1. 在 `pkg/config/config.go` 增加字段
   - 放到正确的配置分组里，例如 `AppConfig`、`QueueConfig`、`SchedulerConfig`
   - 同时补上 `yaml` 和 `env` 标签
2. 如果有默认值或必填校验
   - 在 `pkg/config/load.go` 里补默认初始化或校验逻辑
3. 在 `configs/config.example.yaml` 增加示例值和中文注释
4. 如果本地开发马上要用
   - 同步更新 `configs/config.local.yaml`
5. 在实际消费位置读取
   - 统一从 `rt.Config` 进入，不要在业务里自己重新读环境变量
6. 如果会影响启动或运维流程
   - 同步更新 `README.md` 和本手册

### 当前代码里的参考点

- 配置入口：`pkg/config/config.go`
- 加载入口：`pkg/config/load.go`
- server 读取配置：`cmd/server/main.go`
- worker 读取配置：`cmd/worker/main.go`
- scheduler 读取配置：`cmd/scheduler/main.go`

### 额外提醒：默认配置路径并不完全一样

- 裸跑 `go run ./cmd/server` 等入口时，默认读取的是 `configs/config.yaml`
- 但 `make dev` / `scripts/dev.sh` 默认使用的是 `configs/config.local.yaml`

所以：

- 如果你是命令行单独调试，记得显式带上 `-config configs/config.local.yaml`
- 如果你改了本地配置却感觉“不生效”，先确认自己到底走的是哪条启动路径

## 6. 如何增加路由

项目当前采用“**模块内定义 contract，router 只做 wiring**”的方式。

### 标准步骤

1. 先确认是扩展现有模块，还是新增模块
   - 现有模块：直接改 `internal/modules/articleinspect/` 下的 owner package；root 目录只保留装配和注册，不要把实现塞回 root
   - 新模块：可先参考 `internal/modules/post/`，或者用 `bash scripts/new-module.sh <module_name>` 起骨架
2. 在 owner package 的 `routes.go` / `transport.go` / `dto.go` 里增加请求结构、响应结构、`huma.Register(...)`
3. 把业务逻辑放进 service / repository，不要堆在 route handler 里
4. 在 `internal/api/register/router.go` 里完成顶层 wiring，再由模块自己的 `RegisterRoutes(...)` 接管具体路由注册
5. 如果新增了数据表，记得补：
   - `internal/app/bootstrap/schema.go`
   - `migrations/`
6. 为模块测试和 router 装配测试补回归

### 当前代码里的参考点

- router 总装配：`internal/api/register/router.go`
- 巡检模块注册入口：`internal/modules/articleinspect/module.go`、`internal/modules/articleinspect/routes.go`
- 巡检模块 feature owner：`internal/modules/articleinspect/{rules,tasks,results,actions,lifecycle,articles,audit}`
- 巡检模块 runtime owner：`internal/modules/articleinspect/{scan,worker,outbox}`
- starter 示例模块：`internal/modules/post/`
- 新模块模板说明：`internal/modules/example/README.md`

### 维护原则

- 路由 contract 以源码和 `router_test` 为准，不以旧设计文档为准
- 不要把前端页面路由和后端 API 路由混为一谈
- 如果文档与代码冲突，以代码为准，再回头修文档

## 7. 如何写新的业务逻辑

建议始终沿用下面这一层分工：

```text
*_routes.go -> service -> repository / model
```

### 分工约定

- `*_routes.go`
  - 负责 HTTP 入参、出参、状态码、OpenAPI
  - 对 malformed numeric path/query input 保持项目 envelope `400` 契约，不直接把 Huma `422` 暴露给前端
  - 不写大段业务判断
- `service.go`
  - 放业务规则、流程编排、事务边界
- `repository.go`
  - 放查询条件、持久化细节
- `model.go`
  - 放 GORM model 与表结构映射
- `worker.go`
  - 只在需要异步执行时存在

### 标准步骤

1. 先判断逻辑是同步请求内完成，还是要改成异步任务
2. 如果涉及表结构改动：
   - 先写 migration
   - 再补 model
   - 再在 `internal/app/bootstrap/schema.go` 注册
3. 先补 / 改测试，再补实现
4. 改完后至少跑：

```bash
go test ./...
```

## 8. 如何写新的 worker

这里的“写 worker”不是新增一个独立二进制，而是**新增一个 Asynq task type 及其消费逻辑**。

### 标准步骤

1. 在 `internal/queue/tasks/` 定义新的 task type 与 payload
   - 例如参考 `internal/queue/tasks/articleinspect.go`
2. 如果入口需要先持久化，再异步投递
   - 优先考虑像 articleinspect 一样，先写业务行 + outbox，再由控制面做 relay
3. 在 `internal/queue/asynq/client.go` 增加 enqueue helper
   - 统一封装队列名与 payload 编码
4. 在 server / scheduler 侧找到触发入口并 dispatch 或 retry
   - 推荐保留一层 dispatcher seam，避免 route 或 cron 直接散落 Asynq 细节
5. 在 `internal/queue/asynq/handlers.go` 注册新的消费 handler
5. 在具体业务模块里实现执行器
   - 推荐放在 `internal/modules/<module>/worker.go`
6. 补测试
   - payload 编解码测试
   - enqueue helper 测试
   - handler 注册 / 调度测试
   - 业务 worker 执行测试

### 一条正确的心智模型

```text
HTTP / 业务入口
   -> 业务行 + outbox 同事务落库
   -> optimistic relay 或 scheduler retry
   -> worker 消费
   -> 业务执行器落库 / 更新状态
```

### 当前项目里的参考实现

- task 定义：`internal/queue/tasks/articleinspect.go`
- enqueue helper：`internal/queue/asynq/client.go`
- 消费注册：`internal/queue/asynq/handlers.go`
- 真实执行器：`internal/modules/articleinspect/worker/executor.go`

## 9. 如何写新的 scheduler

当前 scheduler 的定位必须记住一句话：

> scheduler 只负责轻量控制面工作（例如 heartbeat、outbox retry），不要把正文扫描这类重业务逻辑直接写进 cron 回调。

### 标准步骤

1. 先判断 cron 要做的是哪种轻量控制面动作
   - heartbeat / enqueue / outbox retry 都可以
   - 正文扫描、分页扫库等重业务不要直接放进 scheduler
2. 在 `internal/scheduler/jobs.go` 增加新的 job 函数
   - 优先保持“轻逻辑 + 明确日志 + 可重试”
3. 在 `internal/scheduler/scheduler.go` 的 `RegisterJobs(...)` 里注册
4. 如有必要，在 `pkg/config/config.go` 里增加更细的开关 / spec
5. 更新 `configs/config.example.yaml`
6. 补 scheduler 单测和 queue 单测
7. 本地验证时显式打开：

```bash
SCHEDULER_ENABLED=true go run ./cmd/scheduler -config configs/config.local.yaml
```

### 当前项目里的参考实现

- scheduler 入口：`cmd/scheduler/main.go`
- job 注册：`internal/scheduler/scheduler.go`
- 现有 job：`internal/scheduler/jobs.go`
- 单测：`internal/scheduler/scheduler_test.go`

### 当前阶段的特别提醒

- 当前 scheduler 至少会承载两类轻量任务：
  - `runtime:heartbeat`
  - articleinspect task outbox relay / retry
- 它们都不是正文扫描执行器；真正的巡检扫描仍然交给 worker
- 如果后续要做“每日自动巡检”“定时重扫”“超时任务补偿”，应继续保持这种轻控制面边界

## 10. 数据迁移与结构维护

当前 `cmd/migrate` 的实际语义是：

```text
先执行 migrations/*.sql
再执行已注册业务模型的 AutoMigrate
```

所以维护时要同时关注两件事：

1. `migrations/` 是否补了显式 SQL
2. `internal/app/bootstrap/schema.go` 是否注册了对应 model

如果你只改了 model 没补 migration，或者只补了 migration 没补 model，后续交接很容易出现环境不一致。

### `ddl/` 与 `migrations/` 的边界

- `ddl/`：快照、对照、导入真实文稿前的结构准备
- `migrations/`：项目真正执行的迁移脚本

## 11. 任务删除、重跑、幂等注意点

- 任务删除当前只允许：
  - `pending`
  - `failed`
- 删除时会级联删除任务关联的：
  - task keywords
  - results
  - result hits
  - operation logs
  - field change logs
  - actions
- worker 重跑同一篇文稿结果时，当前策略是“先删旧结果，再写新结果”

这意味着：

- 删除任务不是软删，而是连关联数据一起清理
- 重跑结果是覆盖，不是追加
- 如果队列里残留了旧任务，而对应任务已经不再是 `pending`，`startTask` 会拒绝重复消费

## 12. org 29 与真实数据导入

当前前端组织上下文会优先选择 `orgid=29`，如果不存在再回退到第一个组织。

维护时请注意：

- 这更像当前 seed / 本地约定，不应默认为永久业务规则
- 如果你替换了真实组织数据，记得同步检查前端默认组织选择行为
- 导入真实文稿数据前，应先用 `ddl/xt_article.sql` 和 `ddl/xt_article_info.sql` 准备兼容结构

## 13. 提交前检查清单

每次改动后，至少过一遍这张表：

- 配置是否同时更新了：
  - `pkg/config/config.go`
  - `pkg/config/load.go`
  - `configs/config.example.yaml`
- 新表是否同时更新了：
  - `migrations/`
  - `model.go`
  - `internal/app/bootstrap/schema.go`
- 新路由是否同时更新了：
  - 模块 `module.go` / `routes.go` / `routes_common.go` / `*_routes.go`
  - `internal/api/register/router.go`
- 新异步任务是否同时更新了：
  - `internal/queue/tasks/`
  - `internal/queue/asynq/client.go`
  - `internal/queue/asynq/handlers.go`
  - 对应业务 `worker.go`
- 文档是否更新了：
  - `README.md`
  - 本手册

## 14. 改完后怎么验证

最常用命令如下：

```bash
make test
make verify
make smoke
make dev
make migrate
```

推荐按下面方式理解：

- `make test`
  - 主要跑后端 `go test ./...`
- `make verify`
  - 当前仍主要是后端验证集合
- `make smoke`
  - 验证基础 HTTP 端点，不代表覆盖巡检主链路
- `make dev`
  - 一键拉起联调栈
- `make migrate`
  - 同步数据库结构

### 测试定位表

- 路由注册：`internal/api/register/router_test.go`
- scheduler：`internal/scheduler/scheduler_test.go`
- queue glue：`internal/queue/asynq/asynq_test.go`
- 巡检模块级回归：`internal/modules/articleinspect/http_routes_test.go`、`internal/modules/articleinspect/openapi_test.go`
- 配置加载：`pkg/config/config_test.go`

## 15. 当前最容易踩的坑

1. **看到 scheduler 进程启动，就误以为 heartbeat / outbox relay / cleanup 已经生效**
   - 实际还要看 `scheduler.enabled`
2. **只启动 server，不启动 worker**
   - 结果是任务能创建，但不会被消费
3. **在 router 里堆业务逻辑**
   - 当前项目约定 router 只做装配
4. **在 cron job 里直接跑正文扫描、分页扫全量文稿**
   - 正确方式是 scheduler 只做轻量控制面动作，真正处理交给 worker
5. **改了配置模型，却忘了同步 example / README**
   - 交接后最容易给维护者留下隐性坑
6. **把旧设计文档当成现状说明**
   - 先看源码、router test、本手册，再回看历史设计稿
7. **以为 `/readyz` 和 `make smoke` 已经代表业务可用**
   - 它们目前还不能证明巡检主链路已跑通
8. **人工恢复 outbox 时顺手把 `attempt_count` 清零**
   - 默认不要这么做，先保留失败历史，再判断是否需要更激进修复
