# 文稿智能巡检系统一期设计文档

> 维护提示：本文件主要用于保留一期设计背景，不保证逐段与当前代码完全同步。判断“现在到底怎么运行”时，请优先查看 `README.md`、`docs/maintainer-development-flow.md`、`internal/api/register/router.go` 与对应测试。

## 1. 需求拆解

### 1.1 一期业务目标

基于现有文稿业务表 `xt_article` 与 `xt_article_info`，构建一个面向运营后台 / 内容审核后台 / 风控后台的“文稿智能巡检系统”一期版本，形成“规则配置 -> 巡检执行 -> 命中结果 -> 批量处置 / 单篇整改 -> 审计追踪”的完整闭环。

### 1.2 一期范围

一期必须覆盖以下能力：

1. 关键词库管理
   - 按 `orgid` 隔离
   - 支持分类、风险级别、建议动作、命中范围、启停状态
   - 支持 `contains` / `exact` / `regex` 匹配方式
2. 巡检任务
   - 基于 `orgid + publish_at_time` 按批次扫描现网在线文稿
   - 支持关键词库选择、可选字段范围覆盖、可选标题/文章 ID 筛选
   - 通过 worker 异步执行，避免阻塞接口与一次性加载大量数据
3. 命中结果
   - 结果列表、详情、命中字段、命中片段、命中关键词、处置状态
4. 批量处置
   - 批量下线、批量忽略、批量标记已处理、批量导出
   - 保证幂等、可追溯、可看到部分成功 / 部分失败
5. 单篇处理
   - 单篇下线
   - 单篇整改编辑（标题、摘要、正文 HTML 等）
   - 修改后重新上线（不直接写死回到 `9`，统一走生命周期服务）
6. 审计与日志
   - 任务日志、命中日志、操作日志、字段变更日志全部结构化落库

### 1.3 强约束

1. 必须基于 PureMux main 分支既有结构开发，沿用其：
   - `server` / `worker` / `scheduler` / `migrate` 入口
   - `internal/modules/*` 模块分层
   - `internal/api/response` 统一响应 envelope
   - `internal/api/register/router.go` 路由注册方式
   - `internal/app/bootstrap/schema.go` schema 注册方式
2. 不能直接对 `xt_article_info.body` 做 MySQL 粗暴全表 `LIKE` 扫描。
3. 所有查询必须默认带 `orgid`。
4. 所有文稿状态变更必须统一走生命周期服务。
5. 一期不接入第三方 AI 审核服务，但要保留扩展点。
6. 数据库兼容 MySQL 5.7。

### 1.4 已知现网文稿状态

沿用现有业务常量：

- `STATE_DEL = 0`
- `STATE_AUDIT = 1`
- `STATE_BACK = 2`
- `STATE_DRAFT = 3`
- `STATE_STEP = 5`
- `STATE_OFFLINE_SYNC = 7`
- `STATE_OFFLINE = 8`
- `STATE_ONLINE = 9`

一期巡检候选默认只处理 `STATE_ONLINE = 9`。

## 2. 模块划分

## 2.1 后端模块

项目初始化后，在 PureMux 风格下新增：

- `internal/modules/articleinspect`

当前代码已经按能力拆成多个文件；如果只看今天的落地结构，建议把这里理解成下面这组更接近现状的文件职责：

- `model.go`：巡检领域模型与 GORM 模型
- `constants.go`：状态、风险级别、动作、字段范围等常量
- `dto_*.go`：服务层输入输出 DTO
- `validator_keywords.go`：关键词输入校验与规范化
- `repository_*.go`：巡检表读写、候选文稿读取、日志查询
- `service_*.go`：关键词、任务、结果、日志的应用服务
- `scanner.go`：扫描器接口与一期关键词扫描实现
- `lifecycle_service.go`：统一文稿生命周期服务
- `action_service.go`：批量下线 / 忽略 / 已处理 / 导出等编排
- `diff.go`：字段变更 diff helper
- `module.go` / `routes.go` / `routes_common.go` / `*_routes.go`：Huma 路由装配、参数收口与 OpenAPI 注释
- `task_outbox.go`：task enqueue 的 outbox relay / retry 协调层
- `worker.go`：巡检任务执行器
- `articleinspect_test.go`：基础模块测试

可按职责再拆分为多个子文件，但必须保持 PureMux 风格，不将所有逻辑堆进单文件。

## 2.2 核心子域划分

### 2.2.1 Rule 子域

负责关键词库管理：

- 关键词主表
- 命中范围表
- 分类、风险级别、动作建议
- 规则快照生成

### 2.2.2 Task 子域

负责巡检任务：

- 任务创建
- 任务状态流转
- 任务统计更新
- 任务参数快照、规则快照、执行摘要

### 2.2.3 Scan 子域

负责扫描引擎：

- 候选文稿 DTO
- `Scanner` 接口
- `KeywordScanner` 实现
- 片段截取、命中字段、命中次数计算

### 2.2.4 Result 子域

负责文章级结果与字段级命中明细：

- 文章级结果
- 字段级 hit 明细
- 命中详情聚合查询
- 筛选与分页

### 2.2.5 Action / Lifecycle 子域

负责统一处置：

- 批量下线
- 单篇下线
- 忽略
- 标记已处理
- 整改保存
- 修改后重新上线

### 2.2.6 Audit 子域

负责审计日志：

- 任务日志
- 操作日志
- 字段变更日志
- 请求参数快照
- 操作人、IP、request_id

## 2.3 前端模块

新增后台项目：

- `web/admin`

技术栈：

- React
- TypeScript
- Ant Design
- ProComponents
- Vite

页面划分：

1. 关键词库管理页
2. 巡检任务页
3. 新建巡检任务页
4. 命中结果列表页
5. 命中详情页
6. 整改编辑页
7. 操作日志页

## 2.4 身份与权限模块

一期采用“正式项目基线 + 可兼容接入方式”设计：

- 生产基线：`Bearer JWT / OIDC-ready`
- 可选兼容：`trusted_header`
- 本地联调：`dev_header`

统一抽象为：

- `OperatorResolver`

并在上下文中统一产出：

- `operator_id`
- `operator_name`
- `operator_role`
- `request_id`
- `source_ip`

## 3. 数据流

## 3.1 关键词配置流

1. 运营人员进入关键词库页面
2. 创建 / 编辑关键词与命中范围
3. 后端写入：
   - `xt_article_inspect_keywords`
   - `xt_article_inspect_keyword_scopes`
4. 审计字段记录创建人 / 更新人 / 时间

## 3.2 巡检任务创建流

1. 前端提交：
   - `orgid`
   - `publish_at_time` 起止时间
   - 关键词集
   - 是否只扫描 `state = 9`
   - 是否包含 `body`
   - 可选文章 ID / 标题筛选
2. 后端校验参数
3. 加载关键词规则并生成快照
4. 在一个事务里创建：
   - 任务主记录
   - 任务-关键词关联
   - task outbox message
5. 请求内尝试做一次 optimistic relay 到现有 queue
6. 如果 queue 暂时不可用，保留 pending outbox，交给后续 retry
7. API 返回“任务已创建并进入异步投递流程”

## 3.3 巡检执行流

1. worker 拉取任务
2. 任务状态 `pending -> running`
3. repository 先按：
   - `orgid`
   - `state = 9`
   - `publish_at_time`
   - 可选文章 ID / 标题筛选
   分页读取 `xt_article`
4. 每批拿到候选 article ID 后，批量读取 `xt_article_info.body`
5. 应用层将 `xt_article` 与 `xt_article_info` 拼装为 `CandidateArticle`
6. `KeywordScanner` 逐篇扫描
7. 若命中：
   - 写 `xt_article_inspect_results`
   - 写 `xt_article_inspect_result_hits`
8. 每批更新任务统计
9. 完成后写入结束时间、耗时、最终状态

## 3.4 批量处置流

1. 前端从结果列表选中目标或按筛选条件执行批量动作
2. 后端创建 `xt_article_inspect_actions`
3. 遍历目标结果：
   - 如果是下线，调用 `ArticleLifecycleService.OfflineArticle(...)`
   - 如果是忽略 / 标记处理，更新结果处置状态
   - 每篇都写 `xt_article_inspect_operation_logs`
4. 汇总成功数 / 跳过数 / 失败数
5. 更新 action 主记录状态
6. 前端展示处理汇总

## 3.5 单篇整改流

1. 前端加载文章原始可编辑字段
2. 运营修改：
   - `title`
   - `short_title`
   - `rich_title`
   - `keyword`
   - `desc`
   - `body`
3. 后端调用 `ArticleLifecycleService.UpdateArticleFields(...)`
4. 生成字段级 diff
5. 写字段变更日志与操作日志
6. 若选择整改后重新上线，则调用 `RepublishArticle(...)`

## 4. 数据库表设计

以下表均为一期新增巡检系统业务表，默认包含：

- `id BIGINT UNSIGNED PRIMARY KEY`
- `orgid BIGINT UNSIGNED NOT NULL`
- `create_at DATETIME NOT NULL`
- `update_at DATETIME NOT NULL`

具体字段类型可在 migration 中按 PureMux 现有风格细化，但整体语义与索引约束按本设计实现。

## 4.1 xt_article_inspect_keywords

关键词主表。

字段建议：

- `id`
- `orgid`
- `name`
- `category`
- `match_type`
- `risk_level`
- `suggest_action`
- `enabled`
- `remark`
- `creator_id`
- `creator_name`
- `updater_id`
- `updater_name`
- `create_at`
- `update_at`

索引建议：

- `idx_org_enabled_category (orgid, enabled, category)`
- `idx_org_name (orgid, name)`
- 允许同 org 下重名是否唯一，由业务再决定；一期默认不做强唯一，避免现网运营配置受限

## 4.2 xt_article_inspect_keyword_scopes

关键词命中范围表，一条关键词对应多条 scope。

字段建议：

- `id`
- `orgid`
- `keyword_id`
- `scope`
- `create_at`
- `update_at`

scope 取值：

- `title`
- `short_title`
- `rich_title`
- `keyword`
- `desc`
- `body`

索引建议：

- `idx_org_keyword (orgid, keyword_id)`
- `uk_org_keyword_scope (orgid, keyword_id, scope)`

## 4.3 xt_article_inspect_tasks

巡检任务主表。

字段建议：

- `id`
- `orgid`
- `task_no`
- `status`
- `article_state_filter`
- `publish_time_start`
- `publish_time_end`
- `article_id`
- `title_like`
- `include_body`
- `scope_override_mode`
- `scope_snapshot`
- `request_snapshot`
- `rule_snapshot`
- `total_scanned`
- `hit_articles`
- `hit_count`
- `skip_count`
- `fail_count`
- `batch_count`
- `started_at`
- `finished_at`
- `duration_ms`
- `error_message`
- `creator_id`
- `creator_name`
- `create_at`
- `update_at`

索引建议：

- `uk_task_no (task_no)`
- `idx_org_status_create (orgid, status, create_at)`
- `idx_org_creator_create (orgid, creator_id, create_at)`
- `idx_org_time_range (orgid, publish_time_start, publish_time_end)`

## 4.4 xt_article_inspect_task_keywords

任务与关键词关联表。

字段建议：

- `id`
- `orgid`
- `task_id`
- `keyword_id`
- `create_at`
- `update_at`

索引建议：

- `idx_org_task (orgid, task_id)`
- `idx_org_keyword (orgid, keyword_id)`
- `uk_org_task_keyword (orgid, task_id, keyword_id)`

## 4.5 xt_article_inspect_results

文章级命中结果表。

字段建议：

- `id`
- `orgid`
- `task_id`
- `article_id`
- `article_title`
- `article_state`
- `publish_at_time`
- `risk_level`
- `suggest_action`
- `hit_fields_count`
- `hit_keywords_count`
- `hit_count`
- `disposition_status`
- `latest_action_id`
- `latest_operator_id`
- `latest_operator_name`
- `latest_action_at`
- `create_at`
- `update_at`

索引建议：

- `uk_org_task_article (orgid, task_id, article_id)`
- `idx_org_task_risk (orgid, task_id, risk_level)`
- `idx_org_disposition_update (orgid, disposition_status, update_at)`
- `idx_org_article_update (orgid, article_id, update_at)`

## 4.6 xt_article_inspect_result_hits

字段级命中明细表。

字段建议：

- `id`
- `orgid`
- `task_id`
- `result_id`
- `article_id`
- `keyword_id`
- `keyword_text`
- `category`
- `field_name`
- `match_type`
- `risk_level`
- `suggest_action`
- `matched_text`
- `snippet`
- `position_start`
- `position_end`
- `rule_snapshot`
- `create_at`
- `update_at`

索引建议：

- `idx_org_result (orgid, result_id)`
- `idx_org_task_keyword (orgid, task_id, keyword_id)`
- `idx_org_article_field (orgid, article_id, field_name)`
- `idx_org_risk_create (orgid, risk_level, create_at)`

## 4.7 xt_article_inspect_actions

批量 / 单篇动作批次主表。

字段建议：

- `id`
- `orgid`
- `action_no`
- `action_type`
- `task_id`
- `batch_scope`
- `target_count`
- `success_count`
- `fail_count`
- `skip_count`
- `status`
- `reason`
- `request_snapshot`
- `operator_id`
- `operator_name`
- `request_id`
- `source_ip`
- `started_at`
- `finished_at`
- `create_at`
- `update_at`

索引建议：

- `uk_action_no (action_no)`
- `idx_org_action_type_create (orgid, action_type, create_at)`
- `idx_org_task_create (orgid, task_id, create_at)`
- `idx_org_operator_create (orgid, operator_id, create_at)`

## 4.8 xt_article_inspect_operation_logs

对象级操作日志表。

字段建议：

- `id`
- `orgid`
- `action_id`
- `task_id`
- `result_id`
- `article_id`
- `operation_type`
- `before_state`
- `after_state`
- `status`
- `reason`
- `summary`
- `request_snapshot`
- `operator_id`
- `operator_name`
- `request_id`
- `source_ip`
- `create_at`
- `update_at`

索引建议：

- `idx_org_article_create (orgid, article_id, create_at)`
- `idx_org_task_create (orgid, task_id, create_at)`
- `idx_org_operator_create (orgid, operator_id, create_at)`
- `idx_org_result (orgid, result_id)`

## 4.9 xt_article_inspect_field_change_logs

字段变更日志表。

字段建议：

- `id`
- `orgid`
- `action_id`
- `operation_log_id`
- `article_id`
- `field_name`
- `old_value`
- `new_value`
- `diff_summary`
- `operator_id`
- `operator_name`
- `create_at`
- `update_at`

索引建议：

- `idx_org_article_create (orgid, article_id, create_at)`
- `idx_org_action (orgid, action_id)`
- `idx_org_field_create (orgid, field_name, create_at)`

## 5. API 设计

所有 API 统一挂在：

- `/api/v1/article-inspect/*`

响应遵循 PureMux 既有 `response.Envelope` 结构。

## 5.1 关键词管理 API

- `GET /api/v1/article-inspect/keywords`
- `POST /api/v1/article-inspect/keywords`
- `GET /api/v1/article-inspect/keywords/{id}`
- `PUT /api/v1/article-inspect/keywords/{id}`
- `PATCH /api/v1/article-inspect/keywords/{id}/status`
- `DELETE /api/v1/article-inspect/keywords/{id}`

查询支持：

- `orgid`
- `category`
- `enabled`
- `keyword`
- 分页

## 5.2 巡检任务 API

- `GET /api/v1/article-inspect/tasks`
- `POST /api/v1/article-inspect/tasks`
- `GET /api/v1/article-inspect/tasks/{id}`
- `DELETE /api/v1/article-inspect/tasks/{id}`

查询支持：

- `orgid`
- `status`
- `task_no`
- 分页

## 5.3 命中结果 API

- `GET /api/v1/article-inspect/results`
- `GET /api/v1/article-inspect/results/{id}`

筛选支持：

- `orgid`
- `task_id`
- `risk_level`
- `disposition_status`
- `article_id`
- `title`
- 分页

## 5.4 批量处置 API

- `POST /api/v1/article-inspect/actions/batch-offline`
- `POST /api/v1/article-inspect/actions/batch-ignore`
- `POST /api/v1/article-inspect/actions/batch-process`

批量动作请求支持两种目标表达：

1. `result_ids`
2. `task_id + result_ids`

## 5.5 单篇整改与生命周期 API

- `POST /api/v1/article-inspect/articles/{article_id}/offline`
- `PUT /api/v1/article-inspect/articles/{article_id}/rectify`
- `POST /api/v1/article-inspect/articles/{article_id}/republish`
- `GET /api/v1/article-inspect/articles/{article_id}/change-logs`
- `GET /api/v1/article-inspect/articles/{article_id}/operation-logs`

## 5.6 审计日志 API

- `GET /api/v1/article-inspect/logs/operations`
- `GET /api/v1/article-inspect/logs/field-changes`

## 6. 前端页面设计

## 6.1 前端技术栈

建议采用：

- Vite
- React
- TypeScript
- Ant Design
- ProComponents

原则：

- 符合企业后台风格
- 简洁、稳定、效率优先
- 不引入重型状态库
- 与 PureMux API 直接对接

## 6.2 页面清单

### 6.2.1 关键词库管理页

- 列表：`ProTable`
- 表单：`ModalForm`
- 支持：
  - 新增
  - 编辑
  - 删除
  - 启用 / 禁用
  - 分类与风险级别筛选

### 6.2.2 巡检任务页

- 列表：`ProTable`
- 展示任务编号、状态、统计、创建人、执行时间
- 支持详情 drawer
- 支持跳转新建任务页

### 6.2.3 新建巡检任务页

- 表单：`ProForm`
- 支持填写：
  - `orgid`
  - 时间范围
  - 关键词集
  - 是否仅扫描 `state=9`
  - 是否包含 `body`
  - article ID 精确筛选
  - 标题模糊筛选

### 6.2.4 命中结果列表页

- 列表：`ProTable`
- 多选、当前页全选、筛选结果批量操作
- 风险级别 `Tag`
- 命中片段关键字高亮
- 操作按钮：
  - 查看详情
  - 单篇下线
  - 去整改

### 6.2.5 命中详情页

- 采用 `Drawer + Tabs`
- 包含：
  - 文章基础信息
  - 命中明细
  - 正文片段
  - 操作历史
  - 字段变更

### 6.2.6 整改编辑页

- 使用 `ProForm`
- 每个字段展示旧值与新值
- `body` 一期采用 HTML 源码 textarea 编辑
- 支持保存整改与“保存后重新上线”

### 6.2.7 操作日志页

- 列表：`ProTable`
- 按 `orgid`、文章 ID、任务 ID、操作人、时间筛选
- 支持查看请求参数快照、状态变化摘要

## 7. 批量任务设计

## 7.1 队列模型

基于 PureMux 现有 Asynq 集成扩展：

- 在 `internal/queue/tasks` 增加：
  - `articleinspect:run-task`

payload 建议字段：

- `task_id`
- `orgid`
- `trigger_source`
- `operator_id`
- `operator_name`

## 7.2 执行器流程

1. worker 加载任务
2. 任务状态从 `pending` 原子更新为 `running`
3. 使用游标分页读取候选文章
4. 按批次读取正文
5. 调用 `Scanner.ScanArticle(...)`
6. 命中则写结果与 hit 明细
7. 每批更新统计
8. 最终更新任务状态为：
   - `success`
   - `partial_success`
   - `failed`

## 7.3 分页 / 分批策略

- 优先使用基于 `id` 的游标分页，避免大 offset
- 每批大小支持配置，默认建议：
  - 候选文章每批 `100 ~ 300`
- 批次内只保留当前批文章在内存中

## 7.4 失败容忍策略

- 单篇失败：记录 `fail_count`，继续执行
- 单批失败：记录错误，继续剩余批次
- 全局初始化失败或全部失败：标记为 `failed`
- 有成功也有失败：标记为 `partial_success`

## 7.5 幂等要求

- 重复下线已下线文章：跳过，不作为致命失败
- 重复忽略已忽略结果：安全返回
- 重复标记已处理：安全返回
- 所有批量任务都要汇总：
  - 成功数
  - 失败数
  - 跳过数

## 8. 日志 / 审计设计

## 8.1 任务日志

至少记录：

- 谁创建了任务
- 任务参数
- 任务状态变化
- 执行结果汇总

实现方式：

- 任务主表保存完整快照与状态、统计信息
- 如有必要，可通过 operation log 补充任务级事件摘要

## 8.2 命中日志

至少记录：

- 哪篇文章
- 命中了什么关键词
- 命中了哪个字段
- 命中片段是什么
- 当时生效规则快照是什么

实现方式：

- `results` + `result_hits`

## 8.3 操作日志

覆盖：

- 批量下线
- 单篇下线
- 单篇修改
- 修改后重新上线
- 忽略
- 标记处理完成

实现方式：

- `actions` 保存一次批次动作汇总
- `operation_logs` 保存单对象级操作明细

## 8.4 字段变更日志

针对整改编辑，记录：

- `field_name`
- `old_value`
- `new_value`
- `diff_summary`
- `operator_id`
- `operator_name`
- 时间

实现方式：

- `field_change_logs`
- 对 `body` 等长文本，页面展示摘要，表中保留必要值与 diff 摘要；如现网长文本太大，可在实现中截断摘要并保留请求快照

## 8.5 审计统一字段

所有任务 / 动作 / 操作日志都应尽量落以下信息：

- `orgid`
- `operator_id`
- `operator_name`
- `request_id`
- `source_ip`
- `request_snapshot`
- 时间戳

## 9. 状态机设计

## 9.1 任务状态机

任务状态：

- `pending`
- `running`
- `success`
- `failed`
- `partial_success`

状态流转：

- `pending -> running`
- `running -> success`
- `running -> partial_success`
- `running -> failed`

## 9.2 命中结果处置状态机

建议一期统一处置状态：

- `pending`
- `ignored`
- `processed`
- `offlined`
- `republished`
- `failed`

说明：

- 与现网 `xt_article.state` 分离，避免把“巡检结果处置状态”和“文章发布状态”混淆

## 9.3 文稿生命周期状态机

沿用现网状态常量：

- 在线：`9`
- 下线：`8`
- 同步下线：`7`
- 草稿：`3`
- 待审：`1`
- 待审退回：`2`
- 删除：`0`
- 转审：`5`

一期处理约束：

1. 巡检候选默认只扫描 `9`
2. 下线由 `ArticleLifecycleService.OfflineArticle(...)` 统一处理，默认目标 `9 -> 8`
3. 已为 `8/7` 时，下线动作幂等跳过
4. 整改保存不直接改回在线状态
5. 重新上线统一走 `RepublishArticle(...)`
6. 默认保守策略：
   - 当前为 `8` 的文章，整改后重新上线默认提交回 `1`（待审）
   - 目标状态通过配置驱动，后续可按 org / 业务线扩展

## 10. 后续扩展点

## 10.1 AI 审核器扩展

预留扫描器接口：

```go
type Scanner interface {
    ScanArticle(ctx context.Context, article CandidateArticle, rules []KeywordRule) ([]Hit, error)
}
```

一期实现：

- `KeywordScanner`

后续可接：

- `AIScanner`
- `CompositeScanner`

## 10.2 白名单 / 忽略词

预留：

- `WhitelistProvider`
- 忽略规则短路能力

一期先不做完整白名单功能，但批量加入白名单接口位置预留。

## 10.3 定时巡检

当前阶段 scheduler 仍不直接做正文扫描，但已经不只是占位：

- 会注册 `runtime:heartbeat`
- 在 `queue.outbox.enabled=true` 时，会注册 articleinspect task outbox relay / retry

后续可新增：

- 巡检计划表
- cron 配置
- 定时任务管理页面

## 10.4 导出报表

一期可先做 CSV 导出。

后续可扩展：

- 异步导出任务
- 对象存储
- 报表中心

## 10.5 通知告警

后续可新增：

- 高风险命中通知
- 批量下线结果通知
- 任务失败告警

## 10.6 权限细化

一期先预留统一 `OperatorResolver` 与操作人上下文。

后续可扩展：

- 菜单级权限
- 按 org 授权
- 动作级 RBAC
- 审批流

## 11. 实施建议与落地顺序

按用户要求顺序执行：

1. 初始化 PureMux starter 到当前项目目录
2. 将项目命名调整为 `article-sentinel`
3. 新增设计文档
4. 新增数据库 migration
5. 实现后端模块
6. 注册 schema、router、worker handler
7. 新增前端后台
8. 补测试
9. 补 API / 页面文档与示例数据

## 12. 一期保守处理说明

以下点因现网细节尚未完全提供，一期采用保守策略：

1. 重新上线不直接回 `9`，默认通过配置决定目标状态，缺省走回待审 `1`
2. 现网若已有统一上下线服务，一期先以内聚式 `ArticleLifecycleService` 封装，后续再接 adapter
3. 鉴权默认设计为 `OIDC-ready`，同时兼容 `trusted_header` 与 `dev_header` 以适配不同部署环境
4. 白名单与 AI 审核器先预留接口与扩展点，不在一期硬上复杂实现
