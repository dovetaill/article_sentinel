# Deployment / Build / Release / Ops Design

- 日期：2026-05-08
- 执行者：Codex
- 范围：`article-sentinel` 的生产部署、构建制品、Nginx、systemd、Jenkins 发布流程、运维基线

## 背景

当前仓库已经具备以下基础：

- 后端拆分为 4 个入口：`cmd/server`、`cmd/worker`、`cmd/scheduler`、`cmd/migrate`
- 前端为独立的 `React + Vite` 管理台，位于 `web/admin`
- 前端当前通过同源相对路径访问 `/api/...` 和 `/auth/...`
- 本地开发已有 `Makefile` 和 `scripts/dev.sh`，但尚未形成生产制品、版本化发布目录、systemd 单元、Nginx 模板与 Jenkins 发布规范

同时，当前代码里存在几个会直接影响生产部署设计的事实：

- `internal/app/bootstrap/migrate.go` 通过 `file://migrations` 读取迁移目录，依赖工作目录
- `internal/api/handlers/ready.go` 当前 `/readyz` 只检查资源是否已装配，不做实时 DB/Redis 深探活
- `internal/middleware/auth.go` 会直接读取 `X-Forwarded-For` / `X-Real-IP`
- `internal/api/handlers/auth.go` 当前仍接受 `/auth/login?jwt=...` 的 query-string 登录桥接
- `pkg/logger/logger.go` 支持 `stdout` / `file` / `both`，但生产环境更适合统一交给 journald

## 目标

本轮设计目标：

1. 建立适合 Debian / CentOS / RHEL 单机的标准生产部署形态
2. 同时满足：
   - Jenkins 自动构建、发布、回滚
   - 本地或个人沙箱可从源码目录构建相同制品
3. 为前端静态文件、后端多进程、配置、日志、迁移、回滚提供统一规范
4. 把“手工经验”收敛为可模板化的 `Makefile + deploy/ + docs/ops + Jenkins` 流程

## 非目标

本轮明确不做：

- 不引入 Docker / Kubernetes / Nomad 作为首选生产方案
- 不把前端静态资源嵌入 Go 二进制
- 不把 Nginx、systemd、宿主机用户管理全部自动化到 Ansible / Terraform
- 不承诺数据库迁移天然可逆；数据库回滚能力需由迁移策略单独保证
- 不在本轮强行改造现有登录桥接协议为完整 OIDC；只先把部署和生产治理框架搭起来

## 方案对比

### 方案 A：服务器源码目录直接构建并运行

优点：

- 上手简单
- 个人机器操作成本低

缺点：

- 共享环境和生产环境无法做到 `build once, deploy many`
- Jenkins 难以保证每个环境运行的是同一份制品
- 回滚、审计、版本追踪、依赖稳定性都较差
- 目标机成为构建机，Node / Go / 依赖版本漂移风险高

### 方案 B：版本化发布包 + 版本目录 + `current` 软链接（推荐）

优点：

- 适合 Jenkins 产出 immutable artifact
- 适合单机 Nginx + systemd 的成熟运维模式
- 支持同包多环境推广、版本追踪、快速回滚
- 可以兼容“源码目录先 build 再 package”的个人开发路径

缺点：

- 需要一次性补齐 `Makefile`、`deploy/`、文档与运维模板

### 方案 C：容器优先部署

优点：

- 在容器平台上迁移性强

缺点：

- 当前用户的明确运行环境是单台 Linux + Nginx + systemd
- 仓库当前也没有成熟的生产容器和编排资产
- 会显著增加第一版交付复杂度

## 推荐方案

采用方案 B：

- 非正式环境允许从源码目录构建同样的制品
- 共享环境和生产环境只接受版本化 artifact
- 宿主机运行结构使用 `/srv/article-sentinel/releases/<version>` + `/srv/article-sentinel/current`
- 宿主机配置使用 `/etc/article-sentinel`
- 对外流量统一由 Nginx 承接
- 进程管理统一交给 systemd

## 最终架构

### 1. 构建与发布治理

- Jenkins 负责 `build once`
- 产出一个版本化 tarball，例如 `article-sentinel_<version>_linux_amd64.tar.gz`
- 共享测试、Staging、UAT、生产都只部署这个 tarball
- 生产环境禁止：
  - `git pull` 后直接运行
  - 目标机现场重新 build
  - 通过源码工作区直接切换线上版本

### 2. 宿主机目录结构

```text
/srv/article-sentinel/
  releases/
    <version>/
      bin/
      admin/
      migrations/
      deploy/
      manifest.json
      manifest.sha256
  current -> /srv/article-sentinel/releases/<version>
  previous -> /srv/article-sentinel/releases/<old-version>
```

```text
/etc/article-sentinel/
  config.yaml
  article-sentinel.env
```

```text
/var/lib/article-sentinel/
  ... mutable state if later needed ...
```

```text
/run/article-sentinel/
  ... runtime-only files if later needed ...
```

### 3. Nginx 结构

- 对外域名：`https://article-sentinel-admin.cq.qiludev.com`
- `root` 指向 `/srv/article-sentinel/current/admin`
- `/api/` 反向代理到 `127.0.0.1:8080`
- `/auth/` 反向代理到 `127.0.0.1:8080`
- `/healthz`、`/readyz` 反向代理到 `127.0.0.1:8080`
- SPA 路由通过 `try_files $uri $uri/ /index.html` 回退
- `assets/` 使用长缓存；`index.html` 不做长缓存
- 当前 `/auth/login?jwt=...` 至少要在 Nginx access log 层面做降噪或专门处理，避免 query token 被日志记录

### 4. systemd 结构

最终 systemd 单元：

- `article-sentinel.target`
- `article-sentinel-server.service`
- `article-sentinel-worker.service`
- `article-sentinel-scheduler.service`
- `article-sentinel-migrate@.service`

其中：

- `server` / `worker` / `scheduler`：长运行 unit
- `migrate@.service`：显式触发的 `oneshot` 模板 unit
- `target`：方便整套后端栈统一管理

### 5. 服务工作目录与启动方式

- `server` / `worker` / `scheduler`
  - `WorkingDirectory=/srv/article-sentinel/current`
- `migrate@.service`
  - `WorkingDirectory=/srv/article-sentinel/releases/%i`

原因：

- 当前迁移依赖相对 `migrations/` 目录
- 需要让迁移与候选 release 版本显式绑定

### 6. 配置与日志

- 主配置文件：`/etc/article-sentinel/config.yaml`
- 小量环境覆盖：`/etc/article-sentinel/article-sentinel.env`
- 生产日志策略：Go 服务统一 `stdout -> journald`
- Nginx 日志按宿主机标准保留文件日志
- 生产环境不建议继续使用应用 `both` 双写日志

### 7. 生产治理矩阵

| 环境 | 允许源码部署 | 允许 artifact 部署 | 是否可晋升到生产 |
| --- | --- | --- | --- |
| 本地开发 | 是 | 是 | 否 |
| 个人沙箱 | 是 | 是 | 否 |
| 共享测试 / 集成 | 否（原则上） | 是 | 是 |
| Staging / UAT | 否 | 是 | 是 |
| 生产 | 否 | 是 | 是 |

说明：

- “源码部署”只作为开发和个人环境的辅助能力保留
- 真正共享环境以上一律采用 artifact promotion

## 运行与回滚模型

### 标准部署流程

1. Jenkins 产出 tarball + checksum + manifest
2. 目标机下载或接收 tarball
3. 校验 checksum
4. 解压到 `/srv/article-sentinel/releases/<version>`
5. 执行 `systemctl start article-sentinel-migrate@<version>.service`
6. 更新 `previous`
7. 切换 `current`
8. 重启：`server -> worker -> scheduler`
9. 进行 smoke check：`/healthz`、`/readyz`、`/`
10. 失败时回滚 `current` 到上一版

### 回滚边界

- 应用回滚：必须支持
- 数据库回滚：不承诺自动支持
- 因此 DB migration 需采用兼容式或显式可回退策略

## 最小安全基线

### 第一版必须做到

- Go API 只监听 `127.0.0.1`
- 生产 `secure_cookie=true`
- systemd 单元使用非 root 运行用户
- systemd 至少启用：
  - `NoNewPrivileges=yes`
  - `PrivateTmp=yes`
  - `PrivateDevices=yes`
  - `ProtectHome=yes`
  - `ProtectSystem=full` 或 `strict`
  - `UMask=0077`
- Nginx 覆盖并规范化 `X-Forwarded-*` 头
- 共享环境和生产环境禁止 mutable source-tree deployment

### 后续增强项

- `/auth/login?jwt=...` 改为更安全的 code exchange
- `/readyz` 改为实时 DB/Redis 探活
- `*_FILE` / systemd credential 风格 secrets
- DB/Redis TLS 支持
- 更严格的 trusted proxy 收口

## 分阶段实施

### 阶段 1：发布骨架落地

- `Makefile` 增加 `build/package/release`
- 增加 `deploy/nginx/` 模板
- 增加 `deploy/systemd/` 模板
- 增加 `deploy/scripts/` 发布与回滚脚本
- 更新 README 与 ops 文档

### 阶段 2：运维可用性增强

- 加入 checksum / manifest
- journald-only app logging
- `article-sentinel.target`
- `migrate@.service`
- smoke check 脚本

### 阶段 3：安全与健康检查增强

- 强化 `/readyz`
- 强化 secrets 管理
- 强化 trusted proxy
- 替换 query-string JWT 登录桥接
- DB/Redis TLS 配置能力

## 结论

本轮最终结论：

- 宿主机运行形态采用 `/srv + /etc + Nginx + systemd`
- 共享环境与生产治理采用 `build once, artifact promote many`
- 源码工作区构建能力只保留给本地和个人沙箱，不作为正式发布路径
- 第一期先把发布、回滚、日志、配置、unit 模板和 Jenkins 流水线收口；更深的安全和健康检查增强放到第二阶段推进
