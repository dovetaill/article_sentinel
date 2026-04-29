# ArticleInspect Outbox Phase 3 Design

- 日期：2026-04-29
- 执行者：Codex
- 范围：`internal/modules/articleinspect`、`internal/scheduler`、`pkg/config`、`configs/`、运维文档与辅助脚本

## 背景

截至当前分支最新的 3 个提交：

- `40ff759` `docs(plans): capture articleinspect refactor and outbox upgrade design`
- `48b8c44` `refactor(articleinspect): modularize HTTP layer and transactional enqueue flow`
- `9616f15` `docs(articleinspect): refresh architecture and development guidance`

`articleinspect` 已经完成两类收口：

1. HTTP 结构重构  
   `handler.go` 已拆成 `module.go`、`routes.go`、`routes_common.go` 与 feature `*_routes.go`。
2. outbox skeleton  
   task create 已经从“补偿删除”升级为“task/task_keywords/outbox 同事务写入 + optimistic relay + scheduler fallback”。

当前 outbox 已经能运行，但仍然停留在“可用骨架”阶段，仍有以下问题：

- 只有 `pending / dispatched` 两个核心状态，无法表达“已被某个 relay 占有”
- optimistic relay 与 scheduler relay 并发时，仍然可能围绕同一条消息产生竞争
- 失败后缺少明确 backoff 语义，调度行为仍偏粗糙
- 缺少 poison message 的终态收口，坏消息会持续污染 backlog
- 缺少自动 cleanup，终态消息会持续累积
- 缺少结构化的人工恢复位与运维文档

## 目标

本轮 phase 3 的目标是把当前 outbox 从“可运行 skeleton”升级到“可运维 control plane”：

1. 引入 `claim / lease` 语义，避免多 relay 并发争抢同一条消息
2. 引入 retry/backoff 与 `dead_letter` 终态
3. 为 `dispatched` / `dead_letter` 消息增加 retention 与 cleanup
4. 补齐日志、最小指标位与人工恢复流程
5. 保持当前业务边界不变：worker 仍只消费 `articleinspect:run-task`

## 非目标

本轮明确不做：

- 不新增 HTTP API
- 不新增 admin 页面
- 不把 outbox 抽成全局跨模块基础设施
- 不改变 worker 的扫描业务语义
- 不引入 MySQL 8.0 `SKIP LOCKED` 依赖

## 方案对比

### 方案 A：最小增强版

只补 `claimed` 与最大重试次数，不做完整 dead-letter / cleanup / 运维位。

优点：

- 改动最小

缺点：

- phase 3 做完后仍然不够工程化
- backlog 与坏消息的治理能力不足

### 方案 B：单表完整运维版（推荐）

继续使用 `xt_article_inspect_task_outbox`，但把状态机、lease、backoff、dead-letter、cleanup、运维恢复全部补齐。

优点：

- 与当前 worktree 最连续
- 不需要引入新表或新 API
- 可以在不扩大业务面的前提下把 control plane 做完整

缺点：

- 单表字段会比 skeleton 更厚

### 方案 C：主 outbox 表 + dead-letter 专表

把失败终态消息迁移到单独的 dead-letter 表中。

优点：

- 概念最清晰

缺点：

- 迁移、恢复、代码路径和测试面都会明显膨胀
- 对当前分支目标来说偏重

## 推荐方案

采用方案 B：单表完整运维版。

## 设计总览

### 1. 状态机

当前 `pending -> dispatched` 升级为：

- `pending`
- `claimed`
- `dispatched`
- `dead_letter`

状态流转：

- `pending -> claimed -> dispatched`
- `pending -> claimed -> pending`：可恢复失败，等待后续 retry
- `pending -> claimed -> dead_letter`：不可恢复失败或达到最大尝试次数
- `claimed` 但 lease 过期：允许其他 relay 重新接管

### 2. ownership 划分

- `server` 请求内 optimistic relay  
  只负责“顺手尝试一次”，不拥有长期 retry 责任
- `scheduler`  
  负责 claim、retry、cleanup、dead-letter 收口，是 outbox control plane
- `worker`  
  继续只负责消费 `articleinspect:run-task` 并执行巡检主业务

### 3. MySQL 5.7 兼容性约束

本轮不依赖 `SKIP LOCKED` / `NOWAIT` 一类 MySQL 8.0 风格特性。

原因：

- 一期约束要求兼容 MySQL 5.7
- 显式 `claim / lease` 字段更容易在多实例下解释和排查

claim 应通过条件更新实现，而不是依赖锁读抢占：

- 先读候选消息
- 再用 `WHERE id = ? AND ...` 做原子 claim
- claim 成功者才真正 dispatch

## 数据模型变更

在现有 `xt_article_inspect_task_outbox` 上新增字段：

- `claimed_by`：谁抢到了 lease
- `claimed_at`：claim 时间
- `claim_until`：lease 到期时间
- `next_attempt_at`：下一次允许尝试的时间
- `last_error_code`：结构化错误码
- `dead_lettered_at`：进入死信时间
- `retained_until`：终态清理时间

保留当前已有字段：

- `attempt_count`
- `last_attempt_at`
- `last_error`
- `dispatched_at`

建议的错误码：

- `dispatch_error`
- `dispatcher_unavailable`
- `payload_decode_error`
- `unsupported_message_type`
- `db_update_error`

## relay 算法

### 候选消息

一轮 relay 中，可处理消息应满足：

1. `status = pending` 且 `next_attempt_at IS NULL OR next_attempt_at <= now`
2. 或 `status = claimed` 且 `claim_until < now`

### claim

claim 成功后写入：

- `status = claimed`
- `claimed_by`
- `claimed_at = now`
- `claim_until = now + lease_duration`

claim 失败说明该消息已被其他实例抢走，本轮跳过即可。

### 成功路径

dispatch 成功后写入：

- `status = dispatched`
- `dispatched_at = now`
- `attempt_count += 1`
- `last_attempt_at = now`
- `last_error = ''`
- `last_error_code = ''`
- `claimed_by = ''`
- `claimed_at = NULL`
- `claim_until = NULL`
- `retained_until = now + dispatched_retention`

### 可恢复失败

可恢复失败后写入：

- `status = pending`
- `attempt_count += 1`
- `last_attempt_at = now`
- `last_error_code`
- `last_error`
- `next_attempt_at = now + backoff(attempt_count)`
- `claimed_by = ''`
- `claimed_at = NULL`
- `claim_until = NULL`

### 不可恢复失败 / 超阈值

进入 `dead_letter` 时写入：

- `status = dead_letter`
- `dead_lettered_at = now`
- `attempt_count += 1`
- `last_attempt_at = now`
- `last_error_code`
- `last_error`
- `claimed_by = ''`
- `claimed_at = NULL`
- `claim_until = NULL`
- `retained_until = now + dead_letter_retention`

### poison message 判断

以下错误应直接判为不可恢复：

- payload 解码失败
- 不支持的 `message_type`
- 关键 payload 字段缺失且无法修复

原因：继续重试只会污染 backlog，不会自愈。

## backoff 策略

本轮不引入复杂指数退避，采用易读的分段策略：

- 第 1~3 次失败：`15s`
- 第 4~10 次失败：`1m`
- 第 11 次以后：`5m`

优点：

- 易读、易测、易排查
- 对当前业务规模足够

## scheduler 设计

phase 3 后，scheduler 应承担三类 control-plane 责任：

### 1. relay job

职责：

- 扫描候选消息
- claim
- dispatch
- 成功 / retry / dead-letter 收口

### 2. cleanup job

职责：

- 删除 `dispatched` 且 `retained_until < now` 的消息
- 删除 `dead_letter` 且 `retained_until < now` 的消息

cleanup job 应单独配置 cron，不与 relay job 共用 spec。

### 3. recover / repair 人工处理位

本轮不做 HTTP API，而是先落：

- SQL 模板或脚本
- 运维文档
- 测试覆盖 dead-letter -> pending 恢复语义

建议最小形态：

- `scripts/articleinspect_outbox_requeue.sql`

其职责是把指定 `id` 或 `task_id` 的 `dead_letter` 消息重置回：

- `status = pending`
- `next_attempt_at = NOW()`
- 清空 claim 字段

## 配置项

在现有 `queue.outbox` 下补充：

- `enabled`
- `relay_spec`
- `batch_size`
- `lease_duration_seconds`
- `max_attempts`
- `cleanup_spec`
- `dispatched_retention_hours`
- `dead_letter_retention_hours`

说明：

- `cleanup_spec` 独立于 `relay_spec`
- `lease_duration_seconds` 保守默认 `30` 或 `60`
- `max_attempts` 默认值应足够保守，例如 `10`

## 日志与可观测性

本轮先做最小可观测性，不强行接入完整指标系统。

单条消息日志建议包含：

- `outbox_id`
- `task_id`
- `status`
- `attempt_count`
- `claimed_by`
- `last_error_code`

批次级汇总日志建议包含：

- `scanned`
- `claimed`
- `dispatched`
- `retried`
- `dead_lettered`
- `cleaned`

## 人工处理流程

运维文档至少覆盖以下场景：

1. pending backlog 持续增长
2. claimed 卡死不释放
3. dead-letter 出现
4. dispatched / dead-letter 累积过多

每个场景都应明确：

- 查哪些字段
- 看哪些日志
- 如何安全恢复
- 哪些 SQL 可以执行

## 测试策略

phase 3 至少覆盖：

- claim 成功
- claim 冲突失败
- lease 过期后可重新 claim
- retryable failure 会回到 `pending` 并写 `next_attempt_at`
- poison message 进入 `dead_letter`
- 超过 `max_attempts` 进入 `dead_letter`
- cleanup job 删除过期终态消息
- dead-letter 人工恢复语义

## 提交计划

推荐拆成 4 个提交：

1. `docs(plans): add articleinspect outbox phase 3 design and plan`
2. `feat(articleinspect): add outbox lease and dead-letter state machine`
3. `feat(scheduler): add outbox relay and cleanup control-plane jobs`
4. `docs(articleinspect): document outbox operations and recovery flow`

## 风险与取舍

### 风险

- claim 逻辑如果实现粗糙，会引入新的并发边界 bug
- cleanup 如果筛选条件写错，可能误删未观察充分的终态消息
- overly aggressive backoff 会拖慢真正需要快速恢复的场景

### 取舍

- 本轮优先“可解释、可测试、可运维”，而不是“最抽象”
- 使用单表状态机，而不是一步到位拆 dead-letter 专表
- 优先文档化恢复能力，而不是立刻暴露 API/UI
