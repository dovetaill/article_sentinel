# article-sentinel

`article-sentinel` 是一个面向内容风控/文稿巡检场景的全栈项目：后端负责关键词规则管理、异步扫描任务、命中结果处理、整改与审计日志；前端提供运营后台，支持关键词配置、任务发起、命中审核、整改编辑与日志追踪。

项目基于 `PureMux` starter 初始化，目前已经落地一期文稿巡检主链路，并保留 starter 的基础设施能力与一个示例模块 `post` 作为参考。

## 项目解决什么问题

一期目标是把“文稿巡检”做成一条可追溯、可批量操作、可复核的业务链路：

1. 运营先配置关键词规则、匹配范围、风险等级、建议动作
2. 运营或系统发起巡检任务，先把 task / task_keywords / outbox 落库，再进入异步投递与重试链路
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

补充说明：

- 当前仓库里已经有鉴权相关中间件与 identity 基础设施代码
- 但当前 `internal/api/register/router.go` 还没有把应用层鉴权中间件真正挂到 HTTP 链路上
- 因此如果线上访问控制依赖网关、Nginx、内网隔离或其他上层系统，需要在交接时明确责任边界

## 系统怎么工作

### 后端主链路

后端入口位于 `cmd/`：

- `cmd/server`: API 服务
- `cmd/worker`: Asynq worker，消费巡检任务
- `cmd/scheduler`: 定时任务入口
- `cmd/migrate`: 启动时同步 GORM schema

HTTP 路由总装配仍在 `internal/api/register/router.go`，但这个文件现在只负责顶层 wiring：

- `post` 模块：starter 自带 demo
- `articleinspect` 模块：一期巡检业务模块

`articleinspect` 的真实 route contract 与参数收口位于：

- `internal/modules/articleinspect/module.go`
- `internal/modules/articleinspect/routes.go`
- `internal/modules/articleinspect/routes_common.go`
- `internal/modules/articleinspect/*_routes.go`

如果 server 启动时 queue dispatcher 初始化失败，`internal/api/register/router.go` 现在会明确记录 `article inspect dispatcher unavailable` 日志；任务创建接口仍可先把 task 与 outbox 落库，等待后续 retry。

文稿巡检模块主要包括：

- 关键词管理
- 巡检任务创建与查询
- 扫描器与 diff 能力
- 命中结果查询
- 批量动作服务
- 生命周期服务
- 审计/字段变更日志服务

任务异步化仍通过 Asynq worker 完成，但任务创建链路已经升级成 `task + task_keywords + outbox` 同事务落库。当前控制面任务包括：

- `articleinspect:run-task`：真实巡检任务，worker 实际消费执行
- `runtime:heartbeat`：scheduler 基础连通性任务
- `articleinspect task outbox relay`：scheduler 启用时，周期性把 pending outbox message 重试投递到现有 Asynq 队列

典型流程：

1. `POST /api/v1/article-inspect/tasks` 创建任务
2. Server 在一个事务里写入 `xt_article_inspect_tasks`、`xt_article_inspect_task_keywords`、`xt_article_inspect_task_outbox`
3. 请求内会做一次 optimistic relay；如果 Redis / Asynq 不可用，则保留 pending outbox，不再补偿删除任务
4. scheduler 启用时会继续重试 pending outbox message
5. Worker 拉取 `articleinspect:run-task`，分页读取 `xt_article` 候选文稿
6. Worker 批量读取 `xt_article_info` 正文
7. 按规则扫描标题、摘要、关键词、正文等字段
8. 写入 `xt_article_inspect_results`、`xt_article_inspect_result_hits`
9. 后续批量下线/整改等动作写入 `xt_article_inspect_actions`、`xt_article_inspect_operation_logs`、`xt_article_inspect_field_change_logs`

### 前端主链路

前端位于 `web/admin`，是一个独立的 React + Vite 管理台。

当前页面：

- `/rules/categories`: 规则分类
- `/rules/keywords`: 规则管理
- `/tasks`: 巡检任务列表
- `/tasks/new`: 新建巡检任务（按规则执行）
- `/tasks/:taskId/results`: 单任务结果工作台
- `/results`: 全局风险结果列表（辅助入口）
- `/articles`: 文稿中心（独立浏览真实文稿）
- `/articles/:articleId`: 文稿详情
- `/articles/:articleId/rectify`: 整改编辑页
- `/logs`: 操作日志页

典型使用顺序：

1. 在 `/rules/categories` 建立规则分类
2. 在 `/rules/keywords` 配置具体规则
3. 在 `/tasks/new` 选择规则并发起巡检
4. 在 `/tasks/:taskId/results` 查看单任务结果并做批量处置
5. 在 `/articles` 直接浏览真实文稿，必要时进入详情或整改
6. 在 `/logs` 回看操作链路

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

### 1. 准备本地配置

先复制示例配置：

```bash
cp .env.example .env
cp configs/config.example.yaml configs/config.local.yaml
```

说明：

- `.env` 控制 Docker Compose 的 MySQL / Redis 端口、密码、数据库名、业务账号
- `configs/config.local.yaml` 控制后端连接哪个 MySQL / Redis
- 如果你改了 `.env` 中的端口或密码，记得同步更新 `configs/config.local.yaml`

`.env.example` 默认值：

- MySQL: `127.0.0.1:3307`
- Redis: `127.0.0.1:6380`
- 数据库名: `article_sentinel`
- MySQL 业务账号: `article_sentinel`

### 2. 启动依赖

```bash
make up
```

会启动：

- MySQL 5.7
- Redis 7.2

默认由根目录 `.env` 驱动：

- `MYSQL_PORT`
- `MYSQL_ROOT_PASSWORD`
- `MYSQL_DATABASE`
- `MYSQL_APP_USER`
- `MYSQL_APP_PASSWORD`
- `REDIS_PORT`
- `REDIS_PASSWORD`

配置文件：

- `docker-compose.yml`
- `.env`
- `.env.example`
- `configs/config.local.yaml`

补充说明：

- MySQL `root` 会允许远程登录
- Compose 还会自动创建一个业务账号，并授予 `MYSQL_DATABASE` 的全部权限
- MySQL 官方镜像只会在数据目录首次初始化时应用这些账号/密码；如果你改了 `.env` 里的 MySQL 初始化参数，需要重新建卷，例如执行 `docker compose down -v`
- `make up` 只负责启动 MySQL / Redis 容器；它不会顺带启动后端 API、worker、scheduler 或前端开发服务器

### 3. 同步数据库结构

```bash
make migrate
```

说明：

- 会通过当前注册的 GORM model 做 schema 同步
- 当前 `cmd/migrate` 的实际语义是：**先执行 `migrations/*.sql`，再执行已注册业务模型的 `AutoMigrate`**
- 一期巡检的显式 MySQL DDL 在 `migrations/20260420_01_article_inspection.sql`
- 参考快照也放在 `ddl/xt_article_inspection.sql`
- `make migrate` 只负责巡检业务表，不会把 `xt_article` / `xt_article_info` 重建成线上真实结构
- 如果你要导入线上拉下来的真实文稿，必须额外执行 `ddl/xt_article.sql` 和 `ddl/xt_article_info.sql`

### 4. 启动开发联调栈（推荐）

```bash
make dev
```

当前 `make dev` 会同时启动：

- 后端 API
- 异步 worker
- scheduler 进程
- 前端 admin dev server

默认访问：

- 后端 API：`http://127.0.0.1:8080`
- 前端后台：`http://127.0.0.1:5173`

常用入口：

- `GET /healthz`
- `GET /readyz`
- `GET /openapi.json`
- `GET /docs`

补充说明：

- `scheduler` 进程是否真的注册任务，仍取决于 `scheduler.enabled`
- 默认 `configs/config.local.yaml` 中 `scheduler.enabled: false`；启用后除了 `runtime:heartbeat`，还会负责 articleinspect task outbox 的轻量 relay / retry
- 如果你只想单独启动某个进程，请使用：

```bash
make dev-api
make dev-worker
make dev-scheduler
make dev-admin
```

### 5. 单独启动 Worker

如果你没有使用 `make dev`，而是只想手动启动后端进程，那么实际消费异步巡检任务时还需要单独启动 worker：

```bash
go run ./cmd/worker -config configs/config.local.yaml
```

### 6. 单独启动 Scheduler

如果你没有使用 `make dev`，且需要单独验证定时触发与 outbox retry：

```bash
go run ./cmd/scheduler -config configs/config.local.yaml
```

默认 `configs/config.local.yaml` 中：

- `scheduler.enabled: false`

启用后，scheduler 会注册：

- `runtime:heartbeat`
- articleinspect task outbox relay / retry job

### 7. 单独启动前端后台

```bash
cd web/admin
npm install
npm run dev
```

默认访问：

- 本机：`http://127.0.0.1:5173`
- 远程联调：`http://<server-ip>:5173`

补充说明：

- `web/admin/vite.config.ts` 已默认将 Vite dev server 绑定到 `0.0.0.0:5173`，方便远程联调
- 如果公网仍然访问不到，请检查服务器防火墙、云安全组是否放行 `5173/tcp`
- `5173` 仅建议用于开发联调；生产环境请执行 `npm run build`，并用 Nginx 或其他静态文件服务托管 `web/admin/dist`

### 8. 导入真实文稿数据（推荐）

如果你要让“文稿中心”直接展示线上拉下来的真实文稿，不要只导入 demo seed。推荐按下面顺序导入：

建议在 `make migrate` 之后、启动 API / worker / 前端之前完成这一步。

1. 先确保容器已启动：`make up`
2. 如果之前已经用简化版本地表结构跑过环境，先重建 `xt_article` / `xt_article_info`
3. 再导入真实数据文件：
   - `/home/wwwroot/article_sentinel/scripts/xt_article.sql`
   - `/home/wwwroot/article_sentinel/scripts/xt_article_info.sql`

在仓库根目录执行：

```bash
set -a
. ./.env
set +a

docker compose exec -T mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" -e "SET FOREIGN_KEY_CHECKS=0; DROP TABLE IF EXISTS xt_article_info; DROP TABLE IF EXISTS xt_article; SET FOREIGN_KEY_CHECKS=1;"
docker compose exec -T mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" < ddl/xt_article.sql
docker compose exec -T mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" < ddl/xt_article_info.sql
docker compose exec -T mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" < /home/wwwroot/article_sentinel/scripts/xt_article.sql
docker compose exec -T mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" < /home/wwwroot/article_sentinel/scripts/xt_article_info.sql
```

导入完成后可以快速确认行数：

```bash
set -a
. ./.env
set +a

docker compose exec -T mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" -e "SELECT COUNT(*) AS article_count FROM xt_article; SELECT COUNT(*) AS info_count FROM xt_article_info;"
```

说明：

- 这一步会直接替换本地 `xt_article` / `xt_article_info` 内容
- `ddl/xt_article.sql`、`ddl/xt_article_info.sql` 是必须的，因为本地默认简化表结构无法直接吃下线上 dump
- 当前代码已经按真实字段结构读取：`publish_at_time/create_at/update_at` 为时间戳，`xt_article_info.id` 直接对应文稿 ID
- 文稿中心默认只展示在线文稿（`state=9`），更符合实际浏览场景

### 9. 导入 demo 数据（可选）

如果你沿用了 `.env.example` 默认值，可以直接使用 root 导入：

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

如果你改过 `.env`，请把命令里的端口、密码、数据库名替换成你自己的值。

业务账号也可以直连，并且对业务库拥有全部权限，例如：

```bash
mysql -h127.0.0.1 -P3307 -uarticle_sentinel -particle_sentinel article_sentinel
```

Redis 连通性示例：

```bash
redis-cli -h 127.0.0.1 -p 6380 -a article_sentinel_redis ping
```

### 10. 关闭依赖

```bash
make down
```

## 开发环境启动顺序

推荐按下面的顺序启动，本地联调最省事：

### 0. 首次启动或改过密码时先确认配置

```bash
cp .env.example .env
cp configs/config.example.yaml configs/config.local.yaml
```

然后检查两件事：

- `.env` 里的 MySQL / Redis 账号、密码、端口
- `configs/config.local.yaml` 里的 `database.mysql.*`、`redis.*` 是否和 `.env` 保持一致

### 1. 如果你改过 `.env` 里的 MySQL 初始化账号或密码，先重建卷

这是最容易踩坑的地方：

- `.env` 只是 Compose 和 MySQL 初始化参数来源
- MySQL 官方镜像只会在数据目录首次初始化时读取 `MYSQL_ROOT_PASSWORD`、`MYSQL_USER`、`MYSQL_PASSWORD`
- 如果你已经执行过一次 `make up`，后来才改 `.env`，旧卷里的账号密码不会自动更新
- 这时直接 `make dev`，就可能出现 `Access denied` 之类的报错

开发环境推荐直接重建：

```bash
make down
docker compose down -v
make up
```

如果你不想删库，就需要用当前仍然有效的旧密码登录 MySQL，手动执行 `ALTER USER` 来更新账号密码。下面这个示例已经脱敏，按你的真实旧密码和新密码替换占位符即可：

```bash
docker compose exec -T mysql mysql -uroot -p'<current-root-password>' <<'SQL'
ALTER USER 'root'@'%' IDENTIFIED BY '<new-root-password>';
ALTER USER 'root'@'localhost' IDENTIFIED BY '<new-root-password>';
ALTER USER 'article_sentinel'@'%' IDENTIFIED BY '<new-app-password>';
FLUSH PRIVILEGES;
SQL
```

### 2. 启动依赖容器

```bash
make up
```

说明：

- 只启动 `mysql` 和 `redis`
- 命令会等待这两个容器健康检查通过后返回

### 3. 同步数据库结构

```bash
make migrate
```

### 4. 推荐：导入真实文稿数据

```bash
set -a
. ./.env
set +a

docker compose exec -T mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" -e "SET FOREIGN_KEY_CHECKS=0; DROP TABLE IF EXISTS xt_article_info; DROP TABLE IF EXISTS xt_article; SET FOREIGN_KEY_CHECKS=1;"
docker compose exec -T mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" < ddl/xt_article.sql
docker compose exec -T mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" < ddl/xt_article_info.sql
docker compose exec -T mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" < /home/wwwroot/article_sentinel/scripts/xt_article.sql
docker compose exec -T mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" < /home/wwwroot/article_sentinel/scripts/xt_article_info.sql
```

### 5. 可选：导入 demo 数据

```bash
mysql -h127.0.0.1 -P3307 -uroot -p<your-root-password> article_sentinel < scripts/article_inspection_seed.sql
```

### 6. 启动开发联调栈（推荐）

在终端 A 执行：

```bash
make dev
```

说明：

- 这是前台常驻进程，不会自动返回 shell
- 正常现象不是“卡住”，而是服务在持续运行
- 会同时拉起 `server + worker + scheduler + admin`
- 默认访问：
  - 后端 API：`http://127.0.0.1:8080`
  - 前端后台：`http://127.0.0.1:5173`
- 默认 `scheduler.enabled: false`，所以本地不显式打开时 scheduler 不会执行 outbox relay / retry
- 如果只想单独调试某个进程，请使用：
  - `make dev-api`
  - `make dev-worker`
  - `make dev-scheduler`
  - `make dev-admin`

### 7. 如果不使用 `make dev`，可单独启动异步 worker

在终端 B 执行：

```bash
go run ./cmd/worker -config configs/config.local.yaml
```

说明：

- 不启动 worker 也能访问后端接口
- 但真正的异步巡检任务不会被消费

### 8. 如果不使用 `make dev`，可单独启动前端管理台

在终端 C 执行：

```bash
cd web/admin
npm install
npm run dev
```

默认访问：

- 本机：`http://127.0.0.1:5173`
- 远程联调：`http://<server-ip>:5173`

### 9. 如果不使用 `make dev`，可单独启动 scheduler

如果你要验证调度入口，再开一个终端：

```bash
go run ./cmd/scheduler -config configs/config.local.yaml
```

一期默认 `configs/config.local.yaml` 中 `scheduler.enabled: false`。

### 10. 开发环境最常用的检查点

- 后端健康检查：`http://127.0.0.1:8080/healthz`
- 后端就绪检查：`http://127.0.0.1:8080/readyz`
- OpenAPI：`http://127.0.0.1:8080/openapi.json`
- 文档页：`http://127.0.0.1:8080/docs`
- 前端开发页：`http://127.0.0.1:5173`

## 前后端如何联调

### 最短联调路径

1. `cp .env.example .env`
2. `cp configs/config.example.yaml configs/config.local.yaml`
3. `make up`
4. `make migrate`
5. 导入 `/home/wwwroot/article_sentinel/scripts/xt_article.sql` 和 `/home/wwwroot/article_sentinel/scripts/xt_article_info.sql`
6. 启动 `make dev`
7. 访问前端页面：
   - `/rules/categories`
   - `/rules/keywords`
   - `/tasks`
   - `/articles`
   - `/logs`

说明：

- 当前 `make dev` 已经会同时启动 `server + worker + scheduler + admin`
- 如果你不想一键启动，再改用 `make dev-api` / `make dev-worker` / `make dev-admin`

### 推荐验收路径

1. 在 `/rules/categories` 新建或确认规则分类
2. 在 `/rules/keywords` 新建规则，并确认规则已归属到正确分类
3. 在 `/tasks/new` 选择规则发起任务，然后进入 `/tasks/:taskId/results`
4. 从结果页或 `/articles` 打开真实文稿详情，确认富文本标题/正文可正常渲染
5. 在整改页修改标题/摘要/正文并“保存并提交复核”，再到 `/logs` 检查链路

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
- `make smoke`: 启动依赖、执行 migrate、拉起 server，并验证基础 HTTP 端点；它当前还不覆盖文稿巡检主链路

### 前端目录 `web/admin`

```bash
npm install
npm run dev
npm run build
npm test -- --runInBand
```

## 生产部署命令清单

当前仓库只内置了开发用 `docker-compose.yml`，用于启动 MySQL / Redis；仓库里没有现成的生产 `Dockerfile`、`systemd`、`nginx` 或发布脚本。所以正式部署通常按“后端二进制 + 前端静态资源 + 外部 MySQL/Redis”来做。

### 1. 准备生产配置

```bash
cp configs/config.example.yaml configs/config.prod.yaml
```

然后至少修改这些内容：

- `app.env`
- `database.mysql.host` / `port` / `user` / `password` / `dbname`
- `redis.addr` / `password`
- `auth.jwt.secret`
- `log.*`

### 2. 构建后端二进制

```bash
mkdir -p bin
go build -o bin/article-sentinel-server ./cmd/server
go build -o bin/article-sentinel-worker ./cmd/worker
go build -o bin/article-sentinel-migrate ./cmd/migrate
go build -o bin/article-sentinel-scheduler ./cmd/scheduler
```

### 3. 构建前端静态资源

```bash
cd web/admin
npm ci
npm run build
```

构建产物默认在 `web/admin/dist`。

### 4. 发布前校验

```bash
make test
make smoke
```

如果是纯生产环境、没有本地 compose 依赖，也至少建议执行：

```bash
go test ./...
cd web/admin && npm test -- --runInBand
```

### 5. 执行数据库迁移

```bash
./bin/article-sentinel-migrate -config configs/config.prod.yaml
```

### 6. 启动生产服务

启动 API：

```bash
./bin/article-sentinel-server -config configs/config.prod.yaml
```

启动 worker：

```bash
./bin/article-sentinel-worker -config configs/config.prod.yaml
```

如果启用了调度任务，再启动 scheduler：

```bash
./bin/article-sentinel-scheduler -config configs/config.prod.yaml
```

### 7. 接入 Web 服务

- 用 Nginx 或同类反向代理把 API 转发到后端服务端口
- 用 Nginx、CDN 或静态文件服务托管 `web/admin/dist`
- 如果你需要进程守护，建议额外配置 `systemd`、Supervisor、PM2 或容器编排系统

## API 与页面文档

- `docs/README.md`: 文档索引，说明哪些文档是当前现状、哪些是历史计划
- `docs/article-inspection-api.md`: API 路由、请求/响应示例、生命周期说明
- `docs/article-inspection-pages.md`: 页面职责、联调路径、验收建议
- `docs/article-inspection-design.md`: 业务设计文档
- `docs/maintainer-development-flow.md`: 交接维护开发手册，涵盖配置、路由、业务、worker、scheduler 扩展方式
- `docs/plans/2026-04-20-article-sentinel-implementation.md`: 实施计划

## DDL 与数据库资料

`ddl/` 目录统一收纳参考 SQL 快照：

- `ddl/xt_article.sql`: 上游原始文稿主表
- `ddl/xt_article_info.sql`: 上游原始文稿详情表
- `ddl/xt_article_inspection.sql`: 一期巡检 MySQL DDL 快照
- `ddl/README.md`: DDL 快照与真实数据导入说明

执行用 migration：

- `migrations/20260420_01_article_inspection.sql`
- `migrations/README.md`: 迁移执行语义、命名约定与维护注意点

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
