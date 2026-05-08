# 部署手册

- 日期：2026-05-08
- 执行者：Codex

本文档描述 `article-sentinel` 在共享环境、Staging、UAT 和生产环境中的 artifact 部署流程。

## 部署原则

- 构建机或 CI 负责 `build once`
- 共享环境以上只部署版本化 artifact
- 目标机不再承担源码编译职责
- 应用回滚必须支持；数据库自动回滚不保证支持

## 构建侧步骤

在仓库根目录执行：

```bash
make print-version
make build
make release
```

产出：

- `release/article-sentinel_<version>_linux_amd64.tar.gz`
- `release/article-sentinel_<version>_linux_amd64.tar.gz.sha256`

如需在发包前手工校验：

```bash
cd release
sha256sum -c article-sentinel_<version>_linux_amd64.tar.gz.sha256
```

## 目标机目录

```text
/srv/article-sentinel/
  releases/
    <version>/
  current -> /srv/article-sentinel/releases/<version>
  previous -> /srv/article-sentinel/releases/<old-version>
```

```text
/etc/article-sentinel/
  config.yaml
  article-sentinel.env
```

## 标准部署顺序

1. 把 tarball 和 tarball checksum 传到目标机
2. 在目标机上再次校验 checksum
3. 执行 `install-release.sh` 解压到 `releases/<version>`
4. 执行 `activate-release.sh`：
   - 先跑 `article-sentinel-migrate@<version>.service`
   - 再切换 `current`
   - 再重启 `server -> worker -> scheduler`
   - 最后做 `/healthz`、`/readyz`、`/` smoke check
5. 只有 smoke 成功后，`previous` 才会更新到旧版

示例：

```bash
sha256sum -c article-sentinel_<version>_linux_amd64.tar.gz.sha256

sudo bash deploy/scripts/install-release.sh \
  --tarball /tmp/article-sentinel_<version>_linux_amd64.tar.gz \
  --version <version> \
  --checksum /tmp/article-sentinel_<version>_linux_amd64.tar.gz.sha256

sudo bash deploy/scripts/activate-release.sh --version <version>
```

## 迁移顺序

- 不要直接在 `current` 上手工执行“最新 migrate”
- 应始终让版本化 release 对应的 `article-sentinel-migrate@<version>.service` 执行迁移
- 当前迁移依赖 release 工作目录下的 `migrations/`，因此必须和候选版本绑定
- 所有 schema 变更都应按 backward-compatible / expand-contract 策略设计

## 回滚顺序

如果激活后需要回滚：

```bash
sudo bash deploy/scripts/rollback-release.sh
```

或显式指定版本：

```bash
sudo bash deploy/scripts/rollback-release.sh --version <old-version>
```

回滚行为：

1. 切换 `current`
2. 重启 `server -> worker -> scheduler`
3. 重新做 smoke check

注意：

- 回滚脚本不会自动回滚数据库
- 如果新版本迁移已经改动 schema，旧版本是否还能运行取决于迁移是否向后兼容

## 发布前后核对项

发布前：

- `make print-version`
- `make build`
- `make release`
- `sha256sum -c ...tar.gz.sha256`

发布后：

- `systemctl status article-sentinel-server.service`
- `systemctl status article-sentinel-worker.service`
- `systemctl status article-sentinel-scheduler.service`
- `curl -fsS http://127.0.0.1/healthz`
- `curl -fsS http://127.0.0.1/readyz`
