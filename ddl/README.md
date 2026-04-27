# DDL 说明

- 日期：2026-04-27
- 执行者：Codex

## 1. 这个目录放什么

本目录保存的是“参考 SQL 快照”，不是项目自动执行的迁移入口。

- `xt_article.sql`：上游真实文稿主表结构快照
- `xt_article_info.sql`：上游真实文稿详情表结构快照
- `xt_article_inspection.sql`：本项目巡检表结构快照

## 2. 和 migrations/ 的区别

- `ddl/`
  - 用于留档、对照、导入真实数据前的结构准备
  - 面向“参考”
- `migrations/`
  - 由 `cmd/migrate` 真正执行
  - 面向“迁移”

如果你要跑项目迁移，请看 `migrations/README.md` 和 `cmd/migrate`，不要把 `ddl/*.sql` 直接当成迁移链路。

## 3. 常见使用场景

### 3.1 导入真实文稿数据前

如果你要导入 `/scripts/xt_article.sql`、`/scripts/xt_article_info.sql` 这类真实数据 dump，通常需要先执行本目录中的结构 SQL，让本地表结构与上游保持兼容。

### 3.2 对照字段结构

当维护者排查：

- 为什么某个字段是时间戳
- 为什么 `xt_article_info.id` 对应文稿 ID
- 为什么正文不在 `xt_article` 表里

都应该先回到本目录核对结构。

## 4. 维护注意点

- 本目录内容不自动参与 `make migrate`
- 更新上游文稿结构时，应同步更新这里的快照
- `xt_article_inspection.sql` 是巡检表的参考副本；真正执行入口仍然是 `migrations/20260420_01_article_inspection.sql`
