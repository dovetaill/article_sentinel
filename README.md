# article-sentinel

`article-sentinel` 是一个面向内容风控/文稿巡检场景的全栈项目：后端负责关键词规则管理、异步扫描任务、命中结果处理、整改与审计日志；前端提供运营后台，支持关键词配置、任务发起、命中审核、整改编辑与日志追踪。

项目基于 `PureMux` starter 初始化，目前已经落地一期文稿巡检主链路，并保留 starter 的基础设施能力与一个示例模块 `post` 作为参考。

## 项目解决什么问题

一期目标是把“文稿巡检”做成一条可追溯、可批量操作、可复核的业务链路：

1. 运营先配置关键词规则、匹配范围、风险等级、建议动作
2. 运营或系统发起巡检任务，筛选候选文稿并投递到异步队列
3. Worker 扫描文稿标题、摘要、关键词、正文等字段，写入命中结果与命中明细
4. 审核人员在后台查看结果，可批量下线、忽略、处理，或进入整改页修正文稿
5. 所有动作都会记录操作日志与字段变更日志，便于审计与问题回放

## 一期业务约束

### 文稿状态常量

| 数值 | 名称 |
| --- | --- |
| `0` | `del` |
| `1` | `audit` |
| `2` | `back` |
| `3` | `draft` |
| `5` | `step` |
| `7` | `offline_sync` |
| `8` | `offline` |
| `9` | `online` |

### 生命周期约定

- 新建巡检任务一期默认针对 `state=9` 在线文稿
- 下线、整改、重新提审统一走生命周期服务
- “保存后重新上线”一期不直接回 `9`
- 一期保守策略：整改后的默认目标状态按 `1 / audit` 处理

### 身份与审计约定

- 认证方案按 **OIDC-ready** 设计
- 本地/内网联调支持 `trusted_header` / `dev_header` fallback
- 关键操作会落 `operator_id`、`operator_name`、`request_id`、`source_ip`

## 系统怎么工作

### 后端主链路

后端入口位于 `cmd/`：

- `cmd/server`: API 服务
- `cmd/worker`: Asynq worker，消费巡检任务
- `cmd/scheduler`: 定时任务入口
- `cmd/migrate`: 启动时同步 GORM schema

HTTP 路由统一在 `internal/api/register/router.go` 装配：

- `post` 模块：starter 自带 demo
- `articleinspect` 模块：一期巡检业务模块

文稿巡检模块主要包括：

- 关键词管理
- 巡检任务创建与查询
- 扫描器与 diff 能力
- 命中结果查询
- 批量动作服务
- 生命周期服务
- 审计/字段变更日志服务

任务异步化通过 Asynq 完成，当前队列任务类型为：

- `articleinspect:run-task`

典型流程：

1. `POST /api/v1/article-inspect/tasks` 创建任务
2. Server 将任务 payload 投递到 Redis 队列
3. Worker 拉取任务，分页读取 `xt_article` 候选文稿
4. Worker 批量读取 `xt_article_info` 正文
5. 按规则扫描标题、摘要、关键词、正文等字段
6. 写入 `xt_article_inspect_results`、`xt_article_inspect_result_hits`
7. 后续批量下线/整改等动作写入 `xt_article_inspect_actions`、`xt_article_inspect_operation_logs`、`xt_article_inspect_field_change_logs`

### 前端主链路

前端位于 `web/admin`，是一个独立的 React + Vite 管理台。

当前页面：

- `/keywords`: 关键词管理
- `/tasks`: 巡检任务列表
- `/tasks/new`: 新建巡检任务
- `/results`: 命中结果列表
- `/articles/:articleId/rectify`: 整改编辑页
- `/logs`: 操作日志页

典型使用顺序：

1. 在 `/keywords` 配置或调整规则
2. 在 `/tasks/new` 发起巡检
3. 在 `/results` 查看命中结果并做批量处置
4. 在 `/articles/:articleId/rectify` 做整改
5. 在 `/logs` 回看操作链路

## 目录结构

```text
article-sentinel/
├── cmd/                         # server / worker / scheduler / migrate 入口
├── configs/                     # 本地与示例配置
├── ddl/                         # 上游文稿表 + 本项目 MySQL DDL 快照
├── docs/                        # API / 页面 / 设计 / 计划文档
├── internal/
│   ├── api/                     # Huma 路由注册、handler、统一响应
│   ├── app/bootstrap/           # runtime / schema / 资源初始化
│   ├── middleware/              # request id、recover、timeout、access log
│   ├── modules/articleinspect/  # 文稿巡检业务模块
│   ├── modules/post/            # starter 示例模块
│   └── queue/                   # Asynq 集成与任务定义
├── migrations/                  # inspection schema migration
├── scripts/                     # verify / smoke / seed / helper scripts
└── web/admin/                   # React 管理后台
```

## 本地运行

### 1. 启动依赖

```bash
make up
```

会启动：

- MySQL 8.4
- Redis 7.2

本地默认配置：

- MySQL: `127.0.0.1:3307`
- Redis: `127.0.0.1:6380`
- 数据库名: `article_sentinel`

配置文件：

- `docker-compose.yml`
- `configs/config.local.yaml`

### 2. 同步数据库结构

```bash
make migrate
```

说明：

- 会通过当前注册的 GORM model 做 schema 同步
- 一期巡检的显式 MySQL DDL 在 `migrations/20260420_01_article_inspection.sql`
- 参考快照也放在 `ddl/xt_article_inspection.sql`

### 3. 启动后端 API

```bash
make dev
```

默认监听：`http://127.0.0.1:8080`

常用入口：

- `GET /healthz`
- `GET /readyz`
- `GET /openapi.json`
- `GET /docs`

### 4. 启动 Worker

如果你要实际消费异步巡检任务，还需要单独启动 worker：

```bash
go run ./cmd/worker -config configs/config.local.yaml
```

### 5. 启动 Scheduler

如果你需要验证调度入口：

```bash
go run ./cmd/scheduler -config configs/config.local.yaml
```

一期默认 `configs/config.local.yaml` 中：

- `scheduler.enabled: false`

### 6. 启动前端后台

```bash
cd web/admin
npm install
npm run dev
```

默认访问：`http://127.0.0.1:5173`

### 7. 导入 demo 数据

```bash
mysql -h127.0.0.1 -P3307 -uroot -proot article_sentinel < scripts/article_inspection_seed.sql
```

这会导入：

- demo 关键词
- demo 巡检任务
- demo 命中结果
- demo 批量动作
- demo 操作日志
- demo 字段变更日志

### 8. 关闭依赖

```bash
make down
```

## 前后端如何联调

### 最短联调路径

1. `make up`
2. `make migrate`
3. 导入 `scripts/article_inspection_seed.sql`
4. 启动 `make dev`
5. 启动 `go run ./cmd/worker -config configs/config.local.yaml`
6. 启动 `cd web/admin && npm run dev`
7. 访问前端页面：
   - `/keywords`
   - `/tasks`
   - `/results`
   - `/logs`

### 推荐验收路径

1. 在 `/keywords` 确认 demo 关键词存在
2. 在 `/tasks` 确认任务状态与统计
3. 在 `/results` 查看高风险命中、打开详情抽屉
4. 在整改页修改标题/摘要/正文并提交
5. 在 `/logs` 按文章 ID、任务 ID、操作人过滤，检查日志链路

## 常用命令

### 仓库根目录

```bash
make up
make down
make dev
make migrate
make test
make verify
make smoke
```

说明：

- `make test`: 运行 `go test ./...`
- `make verify`: 执行 `scripts/verify.sh`，当前等价于后端测试集
- `make smoke`: 启动依赖、执行 migrate、拉起 server，并验证核心 HTTP 端点

### 前端目录 `web/admin`

```bash
npm install
npm run dev
npm run build
npm test -- --runInBand
```

## API 与页面文档

- `docs/article-inspection-api.md`: API 路由、请求/响应示例、生命周期说明
- `docs/article-inspection-pages.md`: 页面职责、联调路径、验收建议
- `docs/article-inspection-design.md`: 业务设计文档
- `docs/plans/2026-04-20-article-sentinel-implementation.md`: 实施计划

## DDL 与数据库资料

`ddl/` 目录统一收纳参考 SQL 快照：

- `ddl/xt_article.sql`: 上游原始文稿主表
- `ddl/xt_article_info.sql`: 上游原始文稿详情表
- `ddl/xt_article_inspection.sql`: 一期巡检 MySQL DDL 快照

执行用 migration：

- `migrations/20260420_01_article_inspection.sql`

说明：

- `ddl/` 里的 SQL 主要用于参考、对照与留档
- 真正参与项目迁移/初始化的仍是 `migrations/` 与 `cmd/migrate`

## 技术栈

### 后端

- Go `1.25`
- Huma v2
- GORM
- MySQL / PostgreSQL driver
- Redis
- Asynq
- robfig/cron
- `slog`

### 前端

- React `18`
- TypeScript `6`
- Vite `8`
- React Router `7`
- Ant Design `5`
- Ant Design Pro Components `2`

### 测试与开发工具

- Go testing
- Vitest
- Testing Library
- jsdom
- Docker Compose

## 使用了哪些开源库

### 后端核心库

- `github.com/danielgtaylor/huma/v2`
  - HTTP API / OpenAPI 框架
- `gorm.io/gorm`
  - ORM 与 model 映射
- `gorm.io/driver/mysql`
  - MySQL 驱动
- `gorm.io/driver/postgres`
  - PostgreSQL 驱动
- `github.com/hibiken/asynq`
  - 异步任务队列
- `github.com/redis/go-redis/v9`
  - Redis 客户端
- `github.com/robfig/cron/v3`
  - 定时任务
- `github.com/ilyakaznacheev/cleanenv`
  - 配置加载
- `github.com/golang-jwt/jwt/v5`
  - JWT 能力
- `gopkg.in/natefinch/lumberjack.v2`
  - 日志滚动

### 前端核心库

- `react`
- `react-dom`
- `react-router-dom`
- `antd`
- `@ant-design/pro-components`
- `@ant-design/icons`

### 前端测试库

- `vitest`
- `@testing-library/react`
- `@testing-library/user-event`
- `@testing-library/jest-dom`
- `jsdom`

## 当前已实现的后端能力

- 关键词 CRUD 与启停用
- 关键词 scope / 风险 / 动作校验
- 巡检任务创建、详情、结果查询
- 扫描器与字段 diff
- 异步任务投递与 worker 执行
- 结果查询、命中明细、日志查询
- 批量下线/处理/忽略动作
- 整改提交与字段变更日志
- 生命周期状态流转

## 当前已实现的前端能力

- 关键词管理页
- 巡检任务列表页
- 新建任务页
- 命中结果列表页
- 命中详情抽屉
- 整改编辑页
- 操作日志页
- 路由懒加载与 vendor chunk 拆分

## 默认暴露的接口

### 基础接口

- `GET /healthz`
- `GET /readyz`
- `GET /openapi.json`
- `GET /docs`

### starter demo 接口

- `GET /api/v1/posts`
- `GET /api/v1/posts/{id}`
- `POST /api/v1/posts`
- `PATCH /api/v1/posts/{id}`
- `DELETE /api/v1/posts/{id}`

### 文稿巡检接口

详见 `docs/article-inspection-api.md`，主要包括：

- keywords
- tasks
- results
- actions
- article rectify / republish / offline
- operation logs / field change logs / task events

## 补充说明

- `internal/modules/post` 仍保留为 starter 示例模块，可作为新模块分层参考
- 如果你要继续扩展业务模块，可先运行 `bash scripts/new-module.sh <module_name>`
- 更完整的 starter showcase 仍在 `showcase/multisurface` 分支
