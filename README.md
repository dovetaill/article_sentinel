# article-sentinel

`article-sentinel` 基于 `PureMux` starter 初始化，当前仓库保留 `server` / `worker` / `scheduler` / `migrate` 四个入口、共享基础设施层，以及一个 starter 自带的示例模块 `post`，便于在接入文稿巡检能力前先完成基础环境验证。

如果你只想尽快跑起当前项目，请先按下面的本地开发流程执行；PureMux 更完整的多业务、多 surface 示例仍在 `showcase/multisurface` 分支。

## 本地启动

### 1. 启动依赖

```bash
make up
```

这会启动 starter 自带的本地依赖，并等待它们进入可用状态：

- MySQL
- Redis

对应配置已经放在 `configs/config.local.yaml`，默认指向 `docker-compose.yml` 里的服务（主机端口为 MySQL `3307`、Redis `6380`，数据库默认名为 `article_sentinel`，以减少与本机常驻服务冲突）。

### 2. 同步 starter schema

```bash
make migrate
```

`make migrate` 会通过 starter 的 AutoMigrate 同步当前业务模型；如果需要审阅一期文稿巡检的显式 SQL 结构，也可以直接查看 `migrations/20260420_01_article_inspection.sql`。

### 3. 启动 API

```bash
make dev
```

启动后默认访问：

- `http://127.0.0.1:8080/healthz`
- `http://127.0.0.1:8080/readyz`
- `http://127.0.0.1:8080/openapi.json`
- `http://127.0.0.1:8080/docs`

### 4. 启动前端控制台

```bash
cd web/admin
npm install
npm run dev
```

默认访问：`http://127.0.0.1:5173`

### 5. 导入一期 demo 数据（可选但推荐）

```bash
mysql -h127.0.0.1 -P3307 -uroot -proot article_sentinel < scripts/article_inspection_seed.sql
```

这会写入一组关键词、任务、命中结果、整改日志，便于直接验收 `/keywords`、`/tasks`、`/results`、`/logs` 流程。

### 6. 关闭依赖

```bash
make down
```

## 一键验证

开发过程中常用这些命令：

```bash
make test
make verify
make smoke
```

- `make test`: 运行 `go test ./...`
- `make verify`: 执行当前项目的标准校验脚本
- `make smoke`: 启动本地依赖、执行 migrate、拉起 server，并检查核心端点

文稿巡检附加文档：

- `docs/article-inspection-api.md`: API 路由、请求示例、生命周期说明
- `docs/article-inspection-pages.md`: 页面职责、验收路径、前端联调说明
- `scripts/article_inspection_seed.sql`: 本地演示数据

## 默认暴露什么

starter 默认只暴露这些入口：

- `GET /healthz`
- `GET /readyz`
- `GET /openapi.json`
- `GET /docs`
- `GET /api/v1/posts`
- `GET /api/v1/posts/{id}`
- `POST /api/v1/posts`
- `PATCH /api/v1/posts/{id}`
- `DELETE /api/v1/posts/{id}`

## 官方 demo 模块

`main` 分支只保留一个官方 demo 模块：`internal/modules/post`。

它用最小成本展示了 starter 推荐的模块分层：

- `model.go`
- `repository.go`
- `service.go`
- `handler.go`
- `post_test.go`

相关 wiring 入口位于：

- `internal/api/register/router.go`
- `internal/app/bootstrap/schema.go`

如果你要把 starter 改造成自己的项目，先运行 `bash scripts/new-module.sh article` 复制出自己的模块骨架，再按领域语义重命名实现。更具体的替换步骤见 `internal/modules/example/README.md`。

## 如何替换 demo 模块

推荐顺序：

1. 运行 `bash scripts/new-module.sh <module_name>`
2. 调整 `model.go` / `repository.go` / `service.go` / `handler.go`
3. 在 `internal/api/register/router.go` 注册新模块
4. 在 `internal/app/bootstrap/schema.go` 注册模型
5. 扩展你的模块测试与路由测试

## 运行时基础能力

`main` 分支保留的是 starter 级共享能力，而不是业务示例集合：

- Huma v2 + OpenAPI
- GORM 主数据库接入（MySQL / PostgreSQL）
- Redis bootstrap
- `slog` 结构化日志
- 统一 JSON envelope
- `/healthz` 与 `/readyz`
- 通用 middleware / identity / bootstrap 约定

## 完整 showcase 在哪里

如果你想参考更丰富的业务叙事、多角色 API surface 和更多模块组合，请切换到 `showcase/multisurface`：

```bash
git switch showcase/multisurface
```

简要说明见 `docs/showcase/multisurface.md`。
