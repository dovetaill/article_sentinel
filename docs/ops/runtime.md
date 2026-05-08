# 运行手册

- 日期：2026-05-08
- 执行者：Codex

## 运行拓扑

生产运行时由以下组件组成：

- `nginx`：对外入口，托管管理台静态资源并反代 API
- `article-sentinel-server.service`：HTTP API
- `article-sentinel-worker.service`：异步任务消费
- `article-sentinel-scheduler.service`：定时任务与 outbox relay
- `article-sentinel-migrate@.service`：版本绑定迁移
- `article-sentinel.target`：整套后端栈的聚合目标

## 关键路径

- 运行目录：`/srv/article-sentinel/current`
- 版本目录：`/srv/article-sentinel/releases/<version>`
- 主配置：`/etc/article-sentinel/config.yaml`
- 环境覆盖：`/etc/article-sentinel/article-sentinel.env`

运行语义：

- `server` / `worker` / `scheduler` 使用 `WorkingDirectory=/srv/article-sentinel/current`
- `migrate@.service` 使用 `WorkingDirectory=/srv/article-sentinel/releases/%i`
- 应用日志统一走 `stdout -> journald`

## 常用 systemd 命令

首次加载或模板更新后：

```bash
sudo systemctl daemon-reload
```

整套服务启停：

```bash
sudo systemctl start article-sentinel.target
sudo systemctl stop article-sentinel.target
sudo systemctl restart article-sentinel.target
```

按单服务查看状态：

```bash
sudo systemctl status article-sentinel-server.service
sudo systemctl status article-sentinel-worker.service
sudo systemctl status article-sentinel-scheduler.service
```

按单服务重启：

```bash
sudo systemctl restart article-sentinel-server.service
sudo systemctl restart article-sentinel-worker.service
sudo systemctl restart article-sentinel-scheduler.service
```

执行版本绑定迁移：

```bash
sudo systemctl start article-sentinel-migrate@<version>.service
```

## 常用 journald 命令

查看最近日志：

```bash
sudo journalctl -u article-sentinel-server.service -n 200 --no-pager
sudo journalctl -u article-sentinel-worker.service -n 200 --no-pager
sudo journalctl -u article-sentinel-scheduler.service -n 200 --no-pager
```

实时跟随：

```bash
sudo journalctl -u article-sentinel-server.service -f
sudo journalctl -u article-sentinel-worker.service -f
sudo journalctl -u article-sentinel-scheduler.service -f
```

查看某次迁移日志：

```bash
sudo journalctl -u article-sentinel-migrate@<version>.service -n 200 --no-pager
```

## 运行期检查

同机基础探活：

```bash
curl -fsS http://127.0.0.1/healthz
curl -fsS http://127.0.0.1/readyz
curl -fsS http://127.0.0.1/
```

如果由 Nginx 对外提供 HTTPS，请用站点域名再做一次外部链路验证。

## 当前运行基线

- API 应只监听 `127.0.0.1`
- 生产应设置 `AUTH_SESSION_SECURE_COOKIE=true`
- 生产日志应设置 `LOG_OUTPUT=stdout`
- 共享环境与生产环境不应直接从源码目录启动进程
