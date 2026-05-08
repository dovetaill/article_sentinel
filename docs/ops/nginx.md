# Nginx 运行说明

- 日期：2026-05-08
- 执行者：Codex

## 模板位置

- 模板文件：`deploy/nginx/article-sentinel.conf`

该模板假设：

- 管理台静态资源根目录是 `/srv/article-sentinel/current/admin`
- 后端 API 监听 `127.0.0.1:8080`
- Nginx 作为同域入口，同时提供静态资源与 API 反代

## 关键行为

模板包含以下约定：

- `/api/`、`/auth/`、`/healthz`、`/readyz` 反代到 `127.0.0.1:8080`
- `/` 使用 `try_files $uri $uri/ /index.html` 支持 SPA 路由
- `/assets/` 使用长缓存
- `index.html` 设置 `Cache-Control: no-store`
- 覆盖 `X-Real-IP`、`X-Forwarded-For`、`X-Forwarded-Proto`、`X-Forwarded-Host`

原因：

- 后端现在只会在 `http.trusted_proxy_cidrs` 命中时信任 `X-Forwarded-For` / `X-Real-IP`
- 因此 Nginx 侧应写入受控值，而不是盲目透传客户端伪造链
- 同机部署时建议把 `/etc/article-sentinel/config.yaml` 中的 `http.trusted_proxy_cidrs` 设为：

```yaml
http:
  trusted_proxy_cidrs:
    - 127.0.0.1/32
    - ::1/128
```

## 登录桥接说明

当前主流程：

1. 上游系统直接跳转到 `/auth/login?jwt=...`
2. 后端校验 legacy JWT、写入 `as_admin_session`
3. 后端 302 跳转到管理台

模板仍然对 `/auth/login` 做 `access_log off`，避免在 access log 中记录 query 中的 bearer material。

如后续需要进一步降低 URL 中 token 暴露面，可扩展 `POST /api/v1/auth/exchange` -> `/auth/login?code=...` 的一次性 code 方案。

## 上线步骤

1. 把模板复制到目标机 Nginx 配置目录
2. 替换 `server_name`
3. 如使用 HTTPS，再补齐证书与 443 server block
4. 校验配置
5. Reload Nginx

常见命令：

```bash
sudo nginx -t
sudo systemctl reload nginx
```

## 缓存策略

- `admin/assets/*`：适合长期缓存，因为文件名由前端构建哈希驱动
- `index.html`：不做长期缓存，避免客户端继续引用旧资源图谱

## 联通性检查

站点配置生效后，至少检查：

```bash
curl -I http://127.0.0.1/
curl -I http://127.0.0.1/healthz
curl -I http://127.0.0.1/readyz
```

如果使用正式域名或 HTTPS，再补一轮面向外部入口的校验。
