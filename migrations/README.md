# Migrations 说明

- 日期：2026-04-27
- 执行者：Codex

## 1. 这个目录的职责

`migrations/` 存放项目真正会执行的数据库迁移脚本。

当前 `cmd/migrate` 的执行语义不是只跑 GORM `AutoMigrate`，而是：

```text
先执行 migrations/*.sql
再执行 internal/app/bootstrap/schema.go 中注册模型的 AutoMigrate
```

对应代码：

- `internal/app/bootstrap/migrate.go`
- `internal/app/bootstrap/schema.go`

## 2. 和 ddl/ 的区别

- `migrations/`
  - 真正参与项目启动迁移
  - 面向“执行”
- `ddl/`
  - 主要用于留档、对照、参考
  - 面向“说明”

不要把 `ddl/*.sql` 当成项目自动迁移入口。

## 3. 新增迁移时的要求

1. 文件名保持可排序，建议继续沿用 `YYYYMMDD_NN_<topic>.sql`
2. 脚本尽量写成可重复执行或至少对重复执行友好
3. 如果新增了业务表 / 字段：
   - 同步更新 GORM model
   - 同步更新 `internal/app/bootstrap/schema.go`
4. 如果只是改 model、没补 migration，交接后很容易出现环境不一致

## 4. 当前维护注意点

- 仓库当前没有单独暴露 migration history 管理界面
- 因此维护时要格外注意 SQL 脚本的幂等性与回滚预案
- 生产执行前建议先备份，再执行 `./bin/article-sentinel-migrate -config configs/config.prod.yaml`
