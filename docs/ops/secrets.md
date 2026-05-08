# Secrets 与配置约定

- 日期：2026-05-08
- 执行者：Codex

## 当前支持的方式

当前版本已经支持：

- `/etc/article-sentinel/config.yaml`：主配置文件
- `/etc/article-sentinel/article-sentinel.env`：环境变量覆盖

当前版本还不支持：

- `*_FILE` 风格 secrets
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
- 高敏感覆盖值放到 `/etc/article-sentinel/article-sentinel.env`

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
- 未来如果引入 `*_FILE` 或 systemd credentials，应把本文件一起升级
