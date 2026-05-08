# Secrets 与配置约定

- 日期：2026-05-08
- 执行者：Codex

## 当前支持的方式

当前版本已经支持：

- `/etc/article-sentinel/config.yaml`：主配置文件
- `/etc/article-sentinel/article-sentinel.env`：环境变量覆盖
- `*_FILE` 风格 secrets：
  - `auth.session.legacy_secret_file`
  - `auth.session.secret_file`
  - `database.mysql.password_file`
  - `database.postgres.password_file`
  - `redis.password_file`

当前版本还不支持：

- systemd credentials
- 外部密钥管理器自动拉取

因此现阶段的原则是：示例配置进仓库，真实敏感值只放在目标机。

## 需要从仓库移出的敏感值

至少包括：

- `AUTH_SESSION_LEGACY_SECRET`
- `AUTH_SESSION_SECRET`
- `DB_MYSQL_PASSWORD`
- `DB_POSTGRES_PASSWORD`
- `REDIS_PASSWORD`

如果生产环境保留 docs 或额外调试入口，对应访问令牌也应放在目标机本地，而不是写进 release 包。

## 推荐放置方式

`configs/config.example.yaml` 只保留占位值。目标机上：

- 非敏感或结构化配置放到 `/etc/article-sentinel/config.yaml`
- 高敏感覆盖值优先放到宿主机私有 secret 文件，再通过 `*_file` 字段引用
- 如果暂时没有文件挂载能力，再退回 `/etc/article-sentinel/article-sentinel.env`

文件引用示例：

```yaml
database:
  mysql:
    password_file: /etc/article-sentinel/secrets/mysql-password
redis:
  password_file: /etc/article-sentinel/secrets/redis-password
auth:
  session:
    legacy_secret_file: /etc/article-sentinel/secrets/auth-legacy-secret
    secret_file: /etc/article-sentinel/secrets/auth-session-secret
```

说明：

- secret 文件内容会自动做首尾空白裁剪
- 空文件会在启动时直接报错，避免把空密码当成合法配置
- `*_file` 与内联值同时存在时，以文件内容为准

示例：

```bash
APP_ENV=prod
APP_HOST=127.0.0.1
AUTH_SESSION_SECURE_COOKIE=true
AUTH_SESSION_LEGACY_SECRET=replace-me
AUTH_SESSION_SECRET=replace-me
DB_MYSQL_PASSWORD=replace-me
REDIS_PASSWORD=replace-me
LOG_OUTPUT=stdout
DOCS_ENABLED=false
```

## 权限建议

- `/etc/article-sentinel/` 应只允许受控用户维护
- `config.yaml` 与 `article-sentinel.env` 不应加入 Git
- 建议使用最小可读权限，并把变更纳入宿主机审计流程

## 当前 production-safe 默认建议

至少建议在目标机运行时满足：

- `APP_HOST=127.0.0.1`
- `AUTH_SESSION_SECURE_COOKIE=true`
- `LOG_OUTPUT=stdout`
- `DOCS_ENABLED=false`

## 注意事项

- release 包只会携带 `configs/config.example.yaml`，不会携带真实 `/etc/article-sentinel/config.yaml`
- 不要把真实密钥写回仓库里的 `configs/`、`.env` 或 `docs/`
- 如果后续引入 systemd credentials 或外部密钥管理器，应把本文件一起升级
