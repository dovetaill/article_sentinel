# 文档索引

- 日期：2026-04-27
- 执行者：Codex

## 1. 先看哪里

为了避免交接时被历史设计稿误导，建议按下面顺序阅读：

1. `README.md`
   - 看项目定位、启动方式、部署命令
2. `docs/maintainer-development-flow.md`
   - 看日常维护、扩展配置、加路由、写 worker、写 scheduler 的标准流程
3. `internal/api/register/router.go`
   - 看当前后端如何做顶层 wiring、依赖装配、dispatcher 初始化与日志收口
4. `internal/modules/articleinspect/routes.go` / `internal/modules/articleinspect/*_routes.go`
   - 看 articleinspect 的真实 route contract、参数解析与 envelope 契约
5. `internal/api/register/router_test.go`
   - 看当前路由面回归测试，也是“接口是否真实存在”的快速佐证
6. `internal/modules/articleinspect/task_outbox.go` / `internal/queue/asynq/handlers.go` / `internal/scheduler/`
   - 看当前异步投递、outbox retry 与 worker 消费的真实接线方式

## 2. 文档分层规则

### 2.1 权威实现说明

这些内容应优先视为“当前真实现状”：

- `README.md`
- `docs/maintainer-development-flow.md`
- `internal/api/register/router.go`
- `internal/api/register/router_test.go`
- `Makefile`
- `scripts/dev.sh`

### 2.2 业务说明文档

这些文档适合帮助理解业务背景和使用方式，但如果和代码冲突，以代码为准：

- `docs/article-inspection-api.md`
- `docs/article-inspection-pages.md`
- `docs/article-inspection-design.md`

### 2.3 历史计划文档

`docs/plans/` 下的文件主要用于保留设计和实施过程，不应直接当作当前运行手册。

如果你要判断“现在到底是怎么跑的”，请优先看：

- `README.md`
- `docs/maintainer-development-flow.md`
- 对应源码

## 3. 当前最重要的事实

- `worker` 仍是一期真实业务主链路，会消费 `articleinspect:run-task`
- task create 已从“直接 enqueue”升级成“task/task_keywords/outbox 同事务落库 + optimistic relay + scheduler retry”
- `scheduler` 默认仍关闭，但启用后会承担轻量 outbox relay / retry，而不做正文扫描这类重业务
- `make dev` 当前会同时启动 `api + worker + scheduler + admin`
- 应用层鉴权中间件代码已存在，但当前 `NewRouter` 还没有真正挂载鉴权链路

## 4. 维护时的判断原则

- 判断接口是否存在：先看 `internal/api/register/router.go` 的 wiring，再看 `internal/modules/articleinspect/*_routes.go` 和 `internal/api/register/router_test.go`
- 判断 malformed numeric path/query 是否走项目 envelope：看 `internal/modules/articleinspect/articleinspect_test.go`
- 判断任务是否真的会被消费：看 `internal/modules/articleinspect/task_outbox.go`、`internal/queue/asynq/handlers.go`
- 判断定时任务 / outbox retry 是否真的会执行：看 `scheduler.enabled`、`internal/scheduler/scheduler.go`、`internal/scheduler/jobs.go`
- 判断迁移会做什么：看 `internal/app/bootstrap/migrate.go`、`internal/app/bootstrap/schema.go`
