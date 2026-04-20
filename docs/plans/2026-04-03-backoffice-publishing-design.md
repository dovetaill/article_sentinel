# Backoffice Publishing System Design

## Goal

在 article-sentinel 当前的模块化单体骨架上，落地一套可运行的“后台多用户文稿发布系统”第一版，提供：

- 登录与 JWT 鉴权
- 基于角色与资源归属的权限控制
- 用户管理
- 文稿分类管理
- 文稿增删改查、列表、发布/取消发布

## Scope

本阶段只实现后台系统所需的基础同步业务闭环，不在本阶段扩展：

- refresh token
- 审计日志
- 文稿版本历史
- 审核流
- 真实 migration SQL 执行
- 与异步任务/定时任务的深度业务耦合

## Architecture

系统继续沿用当前仓库的分层约定：

- `cmd/server` 提供 HTTP 入口
- `internal/api/register` 负责统一装配模块依赖与注册路由
- 业务按模块拆分到 `internal/modules/*`
- 每个业务模块遵循 `handler -> service -> repository` 的依赖方向
- 鉴权与权限控制通过 `internal/middleware` 和 service 层协同完成

共享运行时仍由 `internal/app/bootstrap` 提供：

- 配置
- logger
- 数据库与 Redis 资源

## Business Roles

系统使用两类角色：

- `admin`
  - 可管理所有用户
  - 可管理所有分类
  - 可管理所有文稿
- `user`
  - 不可管理用户
  - 不可管理分类
  - 只能管理自己创建的文稿

## Authentication Model

采用账号密码登录 + Bearer JWT 鉴权：

- 登录接口接收 `username + password`
- 登录成功返回 JWT
- 每次请求从 `Authorization: Bearer <token>` 解析当前用户信息
- 当前用户信息写入 request context
- JWT 至少包含：
  - `sub`（用户 ID）
  - `username`
  - `role`
  - `exp`

如果用户状态为 `disabled`，即使 token 本身有效，也禁止访问受保护接口。

## Domain Model

### User

字段：

- `id`
- `username`
- `password_hash`
- `role`：`admin | user`
- `status`：`active | disabled`
- `created_at`
- `updated_at`

### Category

字段：

- `id`
- `name`
- `slug`
- `description`
- `created_at`
- `updated_at`

### Article

字段：

- `id`
- `title`
- `summary`
- `content`
- `status`：`draft | published`
- `author_id`
- `category_id`
- `published_at`
- `created_at`
- `updated_at`

## API Shape

统一挂在 `/api/v1` 下。

### Auth

- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me`

### Admin Users

- `POST /api/v1/admin/users`
- `GET /api/v1/admin/users`
- `GET /api/v1/admin/users/{id}`
- `PATCH /api/v1/admin/users/{id}`
- `DELETE /api/v1/admin/users/{id}`

### Admin Categories

- `POST /api/v1/admin/categories`
- `GET /api/v1/admin/categories`
- `GET /api/v1/admin/categories/{id}`
- `PATCH /api/v1/admin/categories/{id}`
- `DELETE /api/v1/admin/categories/{id}`

### Articles

- `POST /api/v1/articles`
- `GET /api/v1/articles`
- `GET /api/v1/articles/{id}`
- `PATCH /api/v1/articles/{id}`
- `DELETE /api/v1/articles/{id}`
- `POST /api/v1/articles/{id}/publish`
- `POST /api/v1/articles/{id}/unpublish`

## Permission Rules

### Role-Level

- `/api/v1/admin/users/*` 仅 `admin`
- `/api/v1/admin/categories/*` 仅 `admin`

### Ownership-Level

对文稿资源：

- `admin` 可查看与修改所有文稿
- `user` 只能查看与修改自己创建的文稿
- `user` 不能越权发布、删除、修改其他人的文稿

## Project Structure

建议新增：

```text
internal/
  middleware/
    auth.go
    authorize.go
  modules/
    auth/
      handler.go
      jwt.go
      model.go
      password.go
      repository.go
      service.go
    user/
      handler.go
      model.go
      repository.go
      service.go
    category/
      handler.go
      model.go
      repository.go
      service.go
    article/
      handler.go
      model.go
      repository.go
      service.go
```

## Runtime Wiring

- `internal/api/register/router.go`
  - 继续作为统一依赖装配点
  - 组装 repository / service / handler
  - 注册 auth、admin、article 路由
- `internal/middleware/auth.go`
  - 负责 Bearer JWT 解析与当前用户注入
- `internal/middleware/authorize.go`
  - 负责角色级别校验 helper
- 资源归属校验
  - 放在 service 层，避免把权限逻辑散落到 handler

## Persistence Strategy

本阶段直接使用 GORM + 当前主数据库：

- 通过 `database.driver` 切换 `mysql` / `postgres`
- repository 基于 `*gorm.DB` 工作
- 为降低首版复杂度，暂不引入复杂查询抽象层

由于当前仓库的 `cmd/migrate` 仍是 skeleton，本阶段会补：

- GORM `AutoMigrate` 初始化路径
- 默认 admin seed 能力

这样可以让系统在没有真实 migration SQL 的前提下先跑通业务闭环。

## Seed Strategy

为了让系统开箱可测，增加一个最小 seed 机制：

- 若系统中不存在 admin 用户，则初始化一个默认 admin
- 默认 admin 账号密码来自配置或固定开发默认值
- 仅用于当前开发阶段的可运行性，后续可替换为正式 migration / seed 流程

## Response Shape

继续沿用当前统一响应结构：

- 成功响应带 `code / message / data`
- 列表响应在 `data` 中包含：
  - `page`
  - `page_size`
  - `total`
  - `items`

## Error Handling

统一错误语义：

- 未认证：`401`
- 无权限：`403`
- 资源不存在：`404`
- 参数错误：`400`
- 冲突（如用户名或分类 slug 重复）：`409`

## Testing Strategy

测试重点放在以下层次：

- `auth`：登录、JWT 解析、`me` 接口、中间件
- `user`：admin CRUD、列表、禁用用户
- `category`：admin CRUD、列表
- `article`：CRUD、列表、发布/取消发布、归属权限
- middleware：缺 token、坏 token、disabled 用户、非 admin 越权

优先保证：

- handler 行为测试
- service 权限测试
- 全仓 `go test ./... -v` 可通过

## Delivery Phases

建议按以下顺序实现：

1. 配置与 bootstrap 扩展（JWT、seed、AutoMigrate）
2. `auth` 模块
3. `user` 模块
4. `category` 模块
5. `article` 模块
6. 路由装配与整体验证
7. README / verification 更新

## Deferred Items

本阶段明确延后：

- refresh token
- 审计日志
- 软删除
- 文稿版本历史
- 审核流
- 真实 SQL migration 文件
- 文章业务与 Asynq / scheduler 的深度联动
