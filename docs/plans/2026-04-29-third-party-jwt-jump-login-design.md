# Third-Party JWT Jump Login Design

- 日期：2026-04-29
- 执行者：Codex
- 范围：后端登录桥接、会话 cookie、中间件租户隔离、前端登录态初始化、`orgid` 收口

## 背景

当前仓库具备以下现状：

- 后端已有通用 JWT 基础能力，位于 `internal/identity/jwt.go`，但当前 claims 只覆盖 `subject/username/role/status`
- HTTP 路由由 `internal/api/register/router.go` 统一组装，目前没有第三方跳转登录入口
- `articleinspect` 全量接口目前仍依赖前端显式传 `orgid`，并由后端按请求参数执行查询与写入
- 前端管理台通过 `web/admin/src/context/org-context.tsx` 拉取 `/api/v1/article-inspect/orgs` 并允许用户手动切换机构
- 前端当前没有基于 cookie 的登录态检查，也没有统一的 401 跳转处理

业务新增需求如下：

1. 第三方系统通过同域 GET 接口跳转到管理台，并附带旧 JWT：`/auth/login?jwt=...`
2. 后端必须先使用旧密钥校验旧 JWT，只有校验成功后才能换签
3. 后端使用新密钥签发新的 JWT，并写入浏览器 cookie，有效期 1 天
4. 后端后续必须只从 cookie 中的新 JWT 读取 `orgid`，作为唯一数据隔离依据
5. 前端必须通过会话接口判断是否已登录；若未登录则跳转固定登录页 `https://appadmin.cq.qiludev.com/cq-admin/index.html`
6. 如果第三方再次带新 JWT 跳转，则必须覆盖旧 cookie，确保用户和机构实时更新
7. 如果 cookie 中的 JWT 过期或无效，则必须清理 cookie，并让前端回到登录页

## 目标

本轮设计目标：

1. 新增同域第三方跳转登录桥接接口
2. 建立“旧 JWT 验签 -> 新 JWT 换签 -> cookie 会话”完整链路
3. 把 `articleinspect` 的租户隔离从“前端传 `orgid`”收口为“后端从 cookie 会话取 `orgid`”
4. 建立前端启动时的登录态检查、401 清理与重定向机制
5. 保持现有页面主体结构不变，只替换认证与组织选择模型

## 非目标

本轮明确不做：

- 不引入 SSO/OIDC 标准协议协商
- 不做多机构切换能力；当前会话只绑定一个 `orgid`
- 不重构整个权限系统；仅为现有管理台补齐第三方跳转登录与租户收口
- 不在仓库代码中硬编码生产密钥；旧密钥、新密钥通过配置注入
- 不移除现有 `/api/v1/article-inspect/orgs` 后端接口，先只让前端停止依赖它

## 约束

- 登录接口与前端部署在同一域名下，因此浏览器 cookie 共享可行
- 第三方入口必须使用 GET，并通过 query string 传入 `jwt`
- 旧 JWT 必须使用旧密钥 `488f468853ddf19671ad060b31d0d0e8` 校验签名与过期时间
- 新 JWT 必须使用新密钥 `11151eb6d745d86e97007439d8c1687b` 签发，但该值应通过运行时配置提供，而不是写死在仓库
- 登录失败、JWT 失效、cookie 过期时，都必须显式清 cookie，避免脏会话残留
- 当前固定登录地址为 `https://appadmin.cq.qiludev.com/cq-admin/index.html`

## 方案对比

### 方案 A：后端桥接换签 + HttpOnly cookie + 会话接口（推荐）

链路：

1. 第三方访问 `GET /auth/login?jwt=...`
2. 后端校验旧 JWT
3. 后端重签新 JWT，写入 HttpOnly cookie
4. 后端 302 跳转到管理台首页
5. 前端启动时请求 `GET /api/v1/auth/session` 判断是否已登录

优点：

- 旧 JWT 不需要暴露给前端业务代码
- 可以把 session cookie 设为 `HttpOnly`，降低前端脚本读取风险
- 后端更容易把 `orgid` 收口为唯一可信来源
- 多次第三方跳转时，只要覆盖 cookie 即可完成用户切换

缺点：

- 前端不能直接读 cookie，需要额外的会话查询接口

### 方案 B：前端先接旧 JWT，再请求后端换签

链路：第三方先落到前端页面，前端从 URL 取 `jwt` 后再调用后端。

优点：

- 前端流程可见、可控

缺点：

- 旧 JWT 会暴露在浏览器地址栏、history 和记忆中
- 更容易出现刷新、重复交换、前端状态竞争问题
- 与“后端只认 cookie”目标不一致

### 方案 C：可读 cookie + 前端自己解析 JWT

优点：

- 前端登录判断实现直观

缺点：

- 需要把 cookie 设成非 HttpOnly，显著降低安全性
- 会让前端重新承担租户与登录判定逻辑，不符合后端收口目标

## 推荐方案

采用方案 A：后端桥接换签 + HttpOnly cookie + 会话接口。

## 设计总览

### 1. 会话模型

新增一套专用于管理台第三方登录的会话结构，例如 `AdminSession`，包含：

- `id`：用户 ID
- `orgid`：机构 ID，后端唯一租户依据
- `orgname`：机构名称
- `platform`
- `priv`
- `roleid`
- `nickname`
- `avatar`
- `departmentid`
- `is_open_edu`
- 标准时间字段：`iat`、`exp`

旧 JWT 中的 `id/orgid/orgname/platform/priv/roleid/nickname/avatar` 必须映射到新 JWT；`departmentid`、`is_open_edu` 等保留字段继续透传。

### 2. JWT 策略

#### 2.1 旧 JWT

- 只用于第三方跳转入口
- 使用旧密钥校验签名
- 校验失败、过期、缺少关键字段时，视为登录失败

#### 2.2 新 JWT

- 由后端使用新密钥签发
- 使用独立 claims，不复用当前 `internal/identity/jwt.go` 的简化 claims
- 有效期固定 24 小时
- 存入 cookie，不暴露给前端读取

### 3. Cookie 策略

cookie 建议：

- 名称：`as_admin_session`
- `Path=/`
- `HttpOnly=true`
- `SameSite=Lax`
- `Secure=true`（HTTPS 正式环境）
- `Max-Age=86400`

清 cookie 时必须使用与写 cookie 一致的 `Path`、`SameSite`、`Secure` 组合，避免浏览器残留旧值。

### 4. 后端接口设计

#### 4.1 `GET /auth/login?jwt=...`

职责：第三方跳转登录桥接。

流程：

1. 读取 query 中的 `jwt`
2. 用旧密钥校验旧 JWT
3. 把旧 claims 映射为内部 `AdminSession`
4. 用新密钥签发 1 天有效的新 JWT
5. 写入 `as_admin_session` cookie
6. 302 跳转到 `/`

失败路径：

- 缺少 `jwt`
- 签名错误
- 旧 token 已过期
- claims 不完整

以上任一情况都执行：

1. 清理 `as_admin_session`
2. 302 跳到固定登录页 `https://appadmin.cq.qiludev.com/cq-admin/index.html`

#### 4.2 `GET /api/v1/auth/session`

职责：前端启动时判断登录态。

成功时返回当前会话信息：

- `id`
- `nickname`
- `avatar`
- `orgid`
- `orgname`
- `platform`
- `priv`
- `roleid`
- 保留字段

失败时：

- 返回 `401`
- 同时清理 `as_admin_session` cookie

#### 4.3 `POST /api/v1/auth/logout`

职责：主动退出。

行为：

- 无论当前 cookie 是否有效，都清理 `as_admin_session`
- 返回成功 envelope，前端收到后跳登录页

### 5. 后端认证中间件

新增专用 session 中间件，而不是挤进现有 `internal/middleware/auth.go` 的 bearer/header 逻辑：

1. 从 `as_admin_session` 读取新 JWT
2. 尝试解析成 `AdminSession`
3. 成功时把会话写入 request context
4. 同时派生出通用 `identity.Actor` / `identity.Principal`，供现有操作日志逻辑继续使用
5. 如果 cookie 存在但 JWT 无效或过期，则立即下发清 cookie 响应头，但不自动崩溃非受保护页面

这样可以保持：

- 登录页桥接接口可匿名访问
- 健康检查、文档接口不受影响
- 业务接口可统一从 context 中读取当前会话

### 6. `articleinspect` 的 `orgid` 收口

当前 `articleinspect` 路由广泛使用 query/body/path 中的 `orgid`。改造原则如下：

1. 后端真实执行时只使用 context 中的 `session.orgid`
2. 第一版允许接口暂时保留旧的 `orgid` 参数形状，以减少一次性接口签名改动
3. 但这些旧参数在业务执行时全部忽略或被覆盖
4. 创建、更新、删除类接口中的 `OrgID` 字段统一由会话注入
5. 列表、详情、日志查询类接口统一用会话 `orgid` 调 service/repository

这样即使前端或恶意请求传入别的 `orgid`，后端也只能读取当前 cookie 会话对应的数据。

### 7. 前端登录态模型

前端不再通过解析 JWT 或读取 cookie 字符串判断登录，而改为：

1. 应用启动时调用 `GET /api/v1/auth/session`
2. 成功则把返回的用户与机构信息放入 React context
3. 失败则直接跳到 `https://appadmin.cq.qiludev.com/cq-admin/index.html`

原因：

- cookie 为 `HttpOnly` 时，前端无法直接读取，安全性更高
- 登录态、租户隔离都由后端统一裁决

### 8. 前端界面变化

#### 8.1 机构区

`OrgSwitcher` 不再是真正的切换器，而改成只读显示当前机构：

- 显示 `orgname`
- 不再允许用户手动切换 `orgid`
- `OrgProvider` 不再调用 `/api/v1/article-inspect/orgs` 获取组织列表作为工作态来源

#### 8.2 用户区

`UserMenu` 改为展示：

- `nickname`
- `avatar`

退出登录时：

1. 调用 `POST /api/v1/auth/logout`
2. 本地清空会话上下文
3. 跳到固定登录页

### 9. 前端请求层

`web/admin/src/lib/request.ts` 需要统一补两类行为：

1. `fetch` 必须带上同域 cookie（`credentials: 'same-origin'`）
2. 遇到 `401` 时统一触发登录页跳转

这样即使用户在页面停留期间 cookie 过期，任意业务请求拿到 `401` 后也能回到登录页。

### 10. 配置策略

#### 10.1 需要新增运行时配置

建议在 `pkg/config/config.go` 的 `AuthConfig` 下新增 session 配置，至少包括：

- `old_secret`
- `new_secret`
- `cookie_name`
- `issuer`
- `ttl_hours`
- `secure_cookie`

#### 10.2 固定登录页地址

用户已明确“现在写死为 `https://appadmin.cq.qiludev.com/cq-admin/index.html`”。

因此第一版设计：

- 后端桥接失败时使用固定常量跳转
- 前端 session 失败或 401 时使用固定常量跳转

后续若需要多环境可配，再把该值提升为前后端配置项。

## 异常处理

### 登录桥接失败

统一行为：

- 清 cookie
- 302 到固定登录页

### 会话查询失败

统一行为：

- 返回 `401`
- 清 cookie

### 业务接口会话失效

统一行为：

- 后端返回 `401`
- 中间件或 handler 清 cookie
- 前端请求层收到 `401` 后跳登录页

### 再次第三方跳转

统一行为：

- 用新的会话内容重新签发 JWT
- 覆盖已有 cookie
- 保证浏览器中的 `id/orgid/orgname/nickname/avatar` 与第三方最新值一致

## 安全注意事项

1. 旧密钥与新密钥不应提交到仓库代码中；通过配置文件或环境变量注入
2. cookie 使用 `HttpOnly`，避免前端脚本读取 JWT
3. cookie 使用 `Secure`，避免明文链路传输
4. 后端不再信任前端传入的 `orgid`
5. 登录桥接接口必须校验旧 JWT 的签名与过期时间，不能只做 base64 解码

## 测试策略

### 后端

- 会话 token 单测：旧 JWT 验签、新 JWT 签发、新 JWT 解析、过期与签名错误
- 登录 handler 单测：成功换签写 cookie、失败清 cookie 并 302
- 会话 handler 单测：有效 cookie 返回 session，无效 cookie 返回 401 并清 cookie
- 中间件单测：有效 cookie 写入 context，无效 cookie 触发清理
- `articleinspect` 回归测试：即使请求参数带错 `orgid`，最终仍只按 session `orgid` 返回数据

### 前端

- `apiRequest` 在 401 时跳登录页
- App 启动时 session 成功可进入系统，失败跳登录页
- `OrgSwitcher` 只读显示当前机构，不再允许切换
- `UserMenu` 展示 `nickname/avatar` 并支持退出
- 各 services 不再要求显式传 `orgid`

## 分阶段落地

### 阶段 1：会话链路落地

- 配置扩展
- 新增旧 JWT 校验与新 JWT 签发能力
- 新增 `/auth/login`、`/api/v1/auth/session`、`/api/v1/auth/logout`
- 新增 cookie session 中间件

### 阶段 2：后端 `orgid` 收口

- `articleinspect` 路由从 context 取 `orgid`
- 忽略前端传入的租户值
- 保留旧参数外形，避免接口一次性大破坏

### 阶段 3：前端登录态与 UI 调整

- App 启动改为先取 session
- 去掉机构切换，改成机构展示
- 用户区改为真实会话信息
- 统一 401 跳登录
- 前端 services 逐步删除 `orgid` 传参

## 结论

本方案以“后端桥接换签 + HttpOnly cookie + 会话接口”为中心，既满足第三方 GET 跳转登录需求，又把 `orgid` 租户隔离的唯一可信来源收回后端。这样可以在不重做整个权限系统的前提下，快速把现有管理台提升到可用、可控、可扩展的会话模型。
