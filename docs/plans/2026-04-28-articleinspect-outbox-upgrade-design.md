# ArticleInspect Outbox Upgrade Design

- 日期：2026-04-28
- 执行者：Codex
- 范围：`internal/modules/articleinspect`、`internal/api/register`、`internal/scheduler`、项目运行文档

## 背景

上一轮已经完成两件事：

1. `articleinspect` 路由从单个 `handler.go` 拆成按能力分组的 route files。
2. malformed numeric query 已统一回到项目 envelope，而不是直接暴露 Huma `422`。
3. 任务创建在 enqueue 失败时，已经从“留下脏 pending task”收口成“补偿删除”。

但“补偿删除”仍然不是最终形态，因为它仍然属于典型的双写流程：

- 先写 MySQL 任务记录
- 再投递 Redis / Asynq

只要第二步失败，就必须做补偿；即使补偿成功，也意味着创建接口与异步投递仍然强耦合。

## 目标

本轮把任务创建从“补偿删除”升级到“事务内写 task + outbox message，事务外 relay 到 Asynq”的方向，同时保持当前 worker 主链路不变。

另外补上：

- router 侧 dispatcher 初始化失败的可观测性
- README 与维护开发文档对当前架构的更新

## 方案对比

### 方案 A：模块内垂直切片 outbox（推荐）

在 `articleinspect` 模块内新增一张 task outbox 表与 relay 服务：

- `TaskService` 在一个事务里同时创建：
  - `InspectionTask`
  - `InspectionTaskKeyword`
  - `InspectionTaskOutboxMessage`
- 事务提交后，尝试做一次 optimistic relay 到 Asynq
- relay 成功则标记 outbox 为 `dispatched`
- relay 失败则保留 outbox 为 `pending`，等待后续后台 retry
- scheduler 新增一个轻量 outbox relay job，周期性把 pending message 投递到现有 Asynq

优点：

- 不要求这轮就把全局 queue 基础设施重写掉
- articleinspect 可以先独立验证 outbox 方向
- 现有 worker / task payload / Asynq handler 全部可以保留

缺点：

- 这轮的 outbox 仍是模块内实现，不是全局抽象
- 后续如果别的模块也要 outbox，需要再抽公共层

### 方案 B：一次抽成全局 queue/outbox 基础设施

优点：

- 架构更整洁
- 后续多模块复用更好

缺点：

- 改动面过大
- 需要先统一 message envelope、publisher interface、polling ownership
- 容易打断当前 worktree 的收口节奏

### 方案 C：只写设计稿，不改运行链路

优点：

- 风险最低

缺点：

- 用户已经明确要求进入下一阶段
- 只能停留在文档层，不解决当前代码演进方向

## 推荐方案

采用方案 A。

## 设计细节

### 1. 数据模型

新增 `InspectionTaskOutboxMessage`：

- `id`
- `orgid`
- `task_id`
- `message_type`
- `payload`
- `status`：`pending` / `dispatched`
- `attempt_count`
- `last_error`
- `last_attempt_at`
- `dispatched_at`
- 标准 `create_at` / `update_at`

当前只承载一种消息类型：`articleinspect.task.run`。

### 2. 事务边界

`TaskService` 负责真正的业务事务边界：

- 校验输入
- 构建任务快照
- 创建 task 和 task-keywords
- 同事务写入 outbox message

HTTP handler 不再把“任务创建 + 队列投递”的细节展开在 route 里。

### 3. relay 策略

新增模块内 relay：

- 支持按 outbox ID 做一次投递
- 支持批量扫描 pending records 做周期性 retry
- 使用现有 `TaskDispatcher`，不直接把 Asynq 细节散落进 service / scheduler

策略：

- 创建任务后做一次 optimistic relay
- 失败时不删除 task；保留 outbox pending 供后续 retry
- 只要事务提交成功，HTTP 创建接口就返回成功

### 4. 后台 retry ownership

本轮把 retry ownership 放在 `scheduler`：

- scheduler 本身就是轻量后台触发器
- outbox relay job 只做“读 DB -> 调 dispatcher -> 回写 outbox 状态”，不做扫描主业务
- article inspect 真正重活仍然只在现有 Asynq worker 中执行

### 5. router 可观测性

`internal/api/register/router.go` 中如果 `queueasynq.NewClient(rt)` 失败：

- 不再静默返回 `nil`
- 记录明确日志，包含模块语义与错误原因

## 行为变化

### API 创建任务

从“enqueue 失败就返回 `500` 并补偿删除 task”调整为：

- 只要 task + outbox 事务提交成功，就返回 `201`
- 若 optimistic relay 失败，任务仍存在，outbox 留待 retry

这意味着接口语义从“同步确认已投递队列”转向“同步确认已持久化并进入异步投递流程”。

## 文档更新点

- `README.md`
  - task create 链路改成 `task + outbox + optimistic relay + scheduler retry`
  - 增加 dispatcher 初始化失败日志与 outbox retry 说明
- `docs/maintainer-development-flow.md`
  - 路由新增说明从 `handler.go` 改为 route files / module wiring
  - worker / scheduler / outbox ownership 重新描述
- `docs/README.md`
  - 更新“看哪里最权威”的入口描述

## 暂不做

- 不做全局通用 outbox 抽象
- 不做 outbox 死信、指数退避、人工重放后台页面
- 不做跨模块 event bus
