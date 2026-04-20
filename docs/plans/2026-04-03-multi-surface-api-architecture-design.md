# article-sentinel Multi-Surface API Architecture Design

## Goal

为 article-sentinel 定义一套长期可维护的 API 架构：同时支持后台管理接口、前台公开内容接口、前台会员身份接口，以及前台会员交互接口，并保持职责边界清晰、权限模型稳定、模块扩展成本可控。

## Current Assessment

当前仓库已经有一套可工作的模块化分层：

- `internal/modules/*` 采用 `handler -> service -> repository`
- `internal/api/register/router.go` 负责 Huma API 创建、依赖装配、middleware 链接入
- `article` 模块的 ownership 规则仍然位于 `service.go`

这说明当前架构对小到中等规模项目是合理的，但它还没有收敛到最省心的长期形态。

### Current Strengths

- 路由定义靠近模块 handler，HTTP contract 没有散落到多个业务文件
- 业务规则主要留在 service，repository 仍然只做数据访问
- OpenAPI 和 Huma 注册方式已经形成稳定习惯

### Current Risks

1. `router.go` 中重复声明业务路径权限，形成“两份路由知识”
2. 后台和前台未来会共享同一条 API 面，容易让 contract 污染
3. 前台会员能力如果继续塞进 `article` 或 `user`，模块会快速膨胀
4. 当前 `auth` 语义偏后台，未来难以同时承载 admin 与 member 两套身份

## Design Principles

### 1. Keep route ownership close to the module

路由字符串应该继续由模块 handler 持有，而不是集中塞进一个超大的 router 文件。模块最了解自己的 HTTP contract。

### 2. Move access policy next to route groups

访问策略不应该通过 `net/http` 路径前缀硬编码二次声明，而应该通过 Huma groups 或 group middleware 挂载到对应 surface。

### 3. Separate API surfaces by actor

article-sentinel 需要按访问主体拆 API 面，而不是只按资源名拆：

- admin
- public
- member public auth
- member protected

### 4. Share infrastructure, not business identity

后台管理员与前台会员可以复用 token、password hash、context 注入等底层能力，但不应该共用模糊的“同一类用户”业务语义。

### 5. Optimize for predictable growth

后续新增评论、历史记录、推荐、标签、专题、通知时，不应该打破现有边界。

## Target API Surfaces

### Admin API

用途：后台内容运营与管理。

特点：

- 需要管理员身份
- 暴露草稿、发布、审核、管理列表等能力
- 可以返回后台字段与内部状态

建议前缀：`/api/v1/admin`

### Public API

用途：前台公开内容访问。

特点：

- 不需要登录
- 只暴露已发布内容
- 返回前台需要的公开字段

建议前缀：`/api/v1`

### Member Auth API

用途：前台会员注册与登录。

特点：

- 不需要已登录状态
- 面向前台会员身份
- 不与后台管理员登录复用同一路由面

建议前缀：`/api/v1/member/auth`

### Member Self API

用途：前台会员登录后的“我的”和互动行为。

特点：

- 需要会员身份
- 只服务当前登录会员
- 涵盖资料、收藏、点赞、历史等自助能力

建议前缀：`/api/v1/me` 以及需要会员鉴权的内容互动接口

## Target Identity Model

建议把 principal 明确成三类：

- `Anonymous`
- `AdminPrincipal`
- `MemberPrincipal`

这三类 principal 通过统一的 identity 基础设施进入 context，但不会在业务模块中混用。

### Identity Infrastructure

建议新增公共身份基础设施目录：

```text
internal/identity/
  claims.go
  context.go
  password.go
  token.go
```

职责：

- token 签发与解析
- principal claims 定义
- context 注入与读取
- 密码 hash 能力

## Target Module Layout

推荐终态目录如下：

```text
internal/
  api/
    register/
      router.go
    response/
      response.go

  identity/
    claims.go
    context.go
    password.go
    token.go

  middleware/
    authn.go
    require_admin.go
    require_member.go
    request_id.go
    recover.go
    timeout.go
    access_log.go

  modules/
    adminuser/
      model.go
      repository.go
      service.go
      handler.go

    member/
      model.go
      repository.go
      service.go
      public_handler.go
      self_handler.go

    article/
      model.go
      repository.go
      service.go
      public_handler.go
      admin_handler.go

    category/
      model.go
      repository.go
      service.go
      public_handler.go
      admin_handler.go

    engagement/
      model.go
      repository.go
      service.go
      handler.go
```

## Module Responsibilities

### `adminuser`

后台管理员账户域。

负责：

- 后台管理员登录后的身份管理
- 后台管理员账户维护
- 后台侧权限判断所需的管理员角色信息

不负责：

- 前台会员注册
- 点赞收藏
- 前台个人中心

### `member`

前台会员身份域。

负责：

- 注册
- 登录
- token 刷新
- 会员资料
- 后续第三方登录能力

不负责：

- 后台管理员管理
- 内容 CRUD
- 点赞收藏业务本身

### `article`

内容主域。

负责：

- 文稿实体
- 标题、摘要、正文、slug、作者、分类、状态、发布时间
- 后台内容管理
- 前台已发布内容查询

不负责：

- 点赞收藏记录
- 前台会员身份

### `category`

分类域。

负责：

- 分类创建、修改、删除
- 前台分类公开读取

### `engagement`

会员互动域。

负责：

- 点赞
- 收藏
- 取消点赞
- 取消收藏
- 我的收藏
- 后续浏览历史、关注、阅读进度等互动能力

## Route Design

推荐路由如下：

```text
# Public
GET    /api/v1/articles
GET    /api/v1/articles/{slug}
GET    /api/v1/categories

# Member Auth
POST   /api/v1/member/auth/register
POST   /api/v1/member/auth/login
POST   /api/v1/member/auth/refresh

# Member Self / Member Protected
GET    /api/v1/me
GET    /api/v1/me/favorites
POST   /api/v1/articles/{id}/likes
DELETE /api/v1/articles/{id}/likes
POST   /api/v1/articles/{id}/favorites
DELETE /api/v1/articles/{id}/favorites

# Admin
POST   /api/v1/admin/auth/login
GET    /api/v1/admin/auth/me
GET    /api/v1/admin/users
POST   /api/v1/admin/users
GET    /api/v1/admin/categories
POST   /api/v1/admin/categories
PATCH  /api/v1/admin/categories/{id}
DELETE /api/v1/admin/categories/{id}
GET    /api/v1/admin/articles
POST   /api/v1/admin/articles
GET    /api/v1/admin/articles/{id}
PATCH  /api/v1/admin/articles/{id}
DELETE /api/v1/admin/articles/{id}
POST   /api/v1/admin/articles/{id}/publish
POST   /api/v1/admin/articles/{id}/unpublish
```

### Route Naming Notes

- 后台使用 `{id}` 更适合内部管理与精确查询
- 前台详情使用 `{slug}` 更适合 SEO、分享与可读 URL
- 点赞与收藏建议使用资源式复数名词接口，避免 `/like` 与 `/unlike` 这类动词式路径

## Router Composition Strategy

`internal/api/register/router.go` 的目标职责应该收敛为：

1. 创建 Huma API
2. 挂全局 middleware
3. 建立 route groups
4. 装配 repository 与 service
5. 把模块注册到对应 group

### Recommended Composition Sketch

```text
api := newHumaAPI(...)
publicV1 := huma.NewGroup(api, "/api/v1")
memberAuth := huma.NewGroup(api, "/api/v1/member/auth")
memberSelf := huma.NewGroup(api, "/api/v1/me")
adminV1 := huma.NewGroup(api, "/api/v1/admin")

memberSelf.UseMiddleware(requireMember)
adminV1.UseMiddleware(requireAdmin)

article.RegisterPublicRoutes(publicV1, articleQueryService)
category.RegisterPublicRoutes(publicV1, categoryService)
member.RegisterPublicRoutes(memberAuth, memberService)
member.RegisterSelfRoutes(memberSelf, memberService)
engagement.RegisterRoutes(publicV1, engagementService)
article.RegisterAdminRoutes(adminV1, articleAdminService)
category.RegisterAdminRoutes(adminV1, categoryService)
adminuser.RegisterRoutes(adminV1, adminUserService)
```

### Router Anti-Goals

`router.go` 不应该：

- 再手写 `/api/v1/articles` 需要什么权限的字符串规则
- 再承载业务级路径判断
- 直接内联业务规则

## Middleware Strategy

### Global Middleware

保留在最外层：

- Request ID
- Recover
- Timeout
- Access log

### Group Middleware

按 API 面挂载：

- `RequireAdmin`
- `RequireMember`
- 可选的 `OptionalAuth`

### Authentication Flow

建议统一使用一个 `Authenticate` 中间件：

- 负责从 header/cookie 解析 token
- 成功时写入 principal 到 context
- 无 token 时允许公开接口继续执行
- 不直接承担 admin/member 授权决策

## Error Handling Strategy

建议采用“domain error + transport mapping”两层模型。

### Domain Errors

由 service 定义并返回，例如：

- `ErrNotFound`
- `ErrForbidden`
- `ErrConflict`
- `ErrInvalidInput`

### HTTP Mapping

由 handler 或公共 mapper 做统一转换：

- invalid input -> `400` or `422`
- unauthenticated -> `401`
- forbidden -> `403`
- not found -> `404`
- conflict -> `409`
- unexpected -> `500`

## Response Strategy

继续保留统一 envelope，但不要强行复用后台 DTO 与前台 DTO。

### Keep Consistent

- 统一响应 envelope
- 统一分页结构
- 统一错误消息模型

### Keep Separate

- 后台 article DTO
- 前台 article DTO
- member self DTO
- admin user DTO

## Testing Strategy

建议按四层测试：

### 1. Service Tests

重点验证：

- ownership
- 发布状态流转
- 重复收藏冲突
- 前台只可见已发布内容

### 2. Handler Tests

重点验证：

- 路由注册
- 请求绑定
- 状态码
- response envelope

### 3. Router Composition Tests

重点验证：

- public 路由无需登录
- member 路由未登录返回 `401`
- admin 路由 member 访问返回 `403`
- admin 路由 admin 访问成功

### 4. Repository Tests

重点验证：

- 分页
- 唯一约束
- 按状态筛选
- 按作者筛选
- slug 查询

## Recommended Migration Path

建议分五步做最小迁移：

### Phase 1: Refactor router composition

先把 `router.go` 调整为 Huma groups 方式，去掉路径字符串二次声明。

### Phase 2: Split article handlers by surface

把 `article/handler.go` 拆成：

- `public_handler.go`
- `admin_handler.go`

### Phase 3: Split category handlers by surface

分类跟文章一样需要双面 contract。

### Phase 4: Add member identity domain

新增 `member` 模块，只做前台身份：

- register
- login
- refresh
- me

### Phase 5: Add engagement domain

新增 `engagement` 模块，只做：

- likes
- favorites
- my favorites

## Decisions Explicitly Not Recommended Now

以下事项此阶段不建议优先做：

- 一开始就拆成多个独立服务
- 引入过重的 DI 框架作为首要任务
- 把前台会员强行塞入后台 `user` 模块
- 把点赞收藏继续塞到 `article/service.go`

## Why This Matches Industry Practice

- Shopify 明确拆分 `Admin API`、`Storefront API`、`Customer Account API`
- Medusa 明确拆分 `Admin API` 与 `Store API`，并在模块层拆分 `User Module` 与 `Customer Module`
- Strapi 默认就把 admin 面和 content API 面拆开
- Huma 官方支持用 groups 挂公共前缀与 middleware，适合做 article-sentinel 的 surface 编排

## References

- Shopify APIs overview: `https://shopify.dev/docs/api`
- Shopify Storefront API: `https://shopify.dev/docs/storefronts/headless/building-with-the-storefront-api`
- Shopify Customer Account API: `https://shopify.dev/docs/api/customer/latest`
- Medusa Admin API: `https://docs.medusajs.com/api/admin`
- Medusa Store API: `https://docs.medusajs.com/api/store`
- Medusa User Module: `https://docs.medusajs.com/resources/commerce-modules/user`
- Medusa Customer Module: `https://docs.medusajs.com/resources/commerce-modules/customer`
- Strapi admin panel configuration: `https://docs.strapi.io/cms/configurations/admin-panel`
- Huma groups: `https://github.com/danielgtaylor/huma/blob/main/docs/docs/features/groups.md`
- Huma testing utilities: `https://github.com/danielgtaylor/huma/blob/main/docs/docs/tutorial/writing-tests.md`
