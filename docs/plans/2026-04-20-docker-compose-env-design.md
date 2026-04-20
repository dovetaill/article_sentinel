# Docker Compose 环境变量化设计文档

## 1. 目标

为当前项目提供一套可直接落地的 Docker Compose 本地依赖配置：

- 使用 `mysql:5.7`
- 使用 `redis:7.2`
- 允许 MySQL 通过宿主机端口对外访问
- 支持 `root` 远程登录
- 额外创建一个独立业务账号，并对目标数据库拥有全部权限
- MySQL / Redis 的宿主机端口、账号密码通过项目根目录 `.env` 配置

## 2. 设计选择

### 2.1 推荐方案

采用“单个 `docker-compose.yml` + 单个 `.env.example`”方案：

- `docker-compose.yml` 只保留服务定义与变量引用
- 真实端口、密码、数据库名、业务账号名全部放在 `.env`
- 仓库提交 `.env.example`，不提交真实 `.env`

该方案足够覆盖当前需求，并且对本项目协作者最友好：复制 `.env.example` 为 `.env` 后即可直接启动。

### 2.2 不采用的方案

不新增自定义 Dockerfile 或额外初始化 SQL 目录。

原因：

- 当前需求只涉及标准镜像能力即可完成
- MySQL 官方镜像已经支持通过环境变量创建数据库、业务用户，并通过 `MYSQL_ROOT_HOST=%` 允许 `root` 远程连接
- 保持文件数量最少，降低后续维护成本

## 3. 服务设计

### 3.1 MySQL

MySQL 服务采用以下约束：

- 镜像版本固定为 `mysql:5.7`
- 通过 `${MYSQL_PORT}` 映射宿主机端口
- 使用 `${MYSQL_ROOT_PASSWORD}` 初始化 root 密码
- 使用 `${MYSQL_DATABASE}` 创建业务数据库
- 使用 `${MYSQL_APP_USER}` 与 `${MYSQL_APP_PASSWORD}` 创建业务账号
- 设置 `MYSQL_ROOT_HOST=%`，允许 root 从远端登录
- 业务账号默认对 `${MYSQL_DATABASE}` 拥有该库的全部权限（MySQL 官方镜像初始化逻辑会授予该用户对该数据库的全部权限）
- 保持 `utf8mb4` 与 `utf8mb4_unicode_ci`

### 3.2 Redis

Redis 服务采用以下约束：

- 镜像版本固定为 `redis:7.2`
- 通过 `${REDIS_PORT}` 映射宿主机端口
- 通过启动命令设置 `--requirepass ${REDIS_PASSWORD}`
- 健康检查使用带密码的 `redis-cli`

## 4. 配置文件设计

### 4.1 `.env.example`

新增根目录 `.env.example`，至少包含：

- `MYSQL_PORT`
- `MYSQL_ROOT_PASSWORD`
- `MYSQL_DATABASE`
- `MYSQL_APP_USER`
- `MYSQL_APP_PASSWORD`
- `REDIS_PORT`
- `REDIS_PASSWORD`

### 4.2 README 调整

README 的本地运行章节同步调整为：

- 明确先复制 `.env.example` 为 `.env`
- 说明 MySQL 版本已切换为 5.7
- 说明端口、密码、数据库名都以 `.env` 为准
- 提供 root / 业务账号连接示例

## 5. 安全边界

该设计会让 MySQL 与 Redis 通过宿主机端口暴露到外部网络。安全边界如下：

- 访问控制主要依赖 `.env` 中设置的强密码
- 生产或公网服务器仍应通过防火墙 / 安全组限制来源 IP
- `.env` 保持不入库，避免凭据泄漏

## 6. 验证方式

本次改动以配置文件为主，采用配置验证而非 TDD：

1. 运行 `docker compose config`
2. 确认 MySQL / Redis 服务均成功解析
3. 确认端口、环境变量引用、健康检查命令结构正确
