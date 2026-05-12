# Article Inspection Pages

本文档面向运营、测试、前端联调，描述当前管理台中“巡检任务 -> 任务详情 -> 命中结果 -> 文稿/整改”的实际工作流、页面职责与推荐验收路径。

> 维护提示：当前页面路由请优先以 `web/admin/config/routes.ts` 与 `web/admin/src/components/PageTabs/route-meta.ts` 为准。如本文件与代码冲突，以代码现状为准。

## 1. 页面与路由

| 路由 | 页面 | 目标 |
| --- | --- | --- |
| `/inspection/tasks` | 巡检任务列表 | 查看任务状态、统计与详情入口 |
| `/inspection/tasks/create` | 新建巡检任务 | 发起一次新的异步扫描 |
| `/inspection/tasks/:taskId` | 任务详情工作台 | 查看任务摘要、规则快照、请求快照、关联日志，并在命中结果工作台内执行处置 |
| `/content/articles/:articleId` | 文稿详情 | 查看真实文稿、最近巡检留痕，并按 `return_to` 返回来源工作区 |
| `/content/articles/:articleId/rectify` | 整改编辑页 | 对标题、摘要、正文做修订并提交下一状态 |
| `/audit/logs` | 操作日志页 | 按文章、任务、操作人回溯处理轨迹 |
| `/inspection/results` | 兼容入口 | 不再作为独立业务页，仅将旧链接导入任务详情结果视图 |
| `/tasks`、`/tasks/:taskId/results`、`/results`、`/logs` | 旧路由兼容入口 | 兼容早期链接，自动跳转到当前业务路由 |

## 2. 页面说明

### 2.1 规则中心 `/rules/categories`、`/rules/keywords`

- `规则分类`：维护分类名称、启停状态、排序，并可跳转查看分类下规则
- `关键词规则`：维护关键词、风险等级、建议动作、适用范围和启停状态
- 适合校验规则配置是否完整、启停是否即时生效、分类联动是否正确

### 2.2 任务列表 `/inspection/tasks`

- 展示任务编号、状态、命中统计、创建人、创建时间
- 任务编号可直接进入任务详情
- 操作列统一提供 `查看任务`
- 支持跳转 `/inspection/tasks/create`
- 待执行/失败任务允许删除，已执行任务仅保留查看入口

### 2.3 新建任务 `/inspection/tasks/create`

- 选择关键词集
- 配置发布时间范围
- 一期默认扫描 `state=9` 在线文稿

### 2.4 任务详情工作台 `/inspection/tasks/:taskId`

- 顶部展示任务摘要、执行统计、规则快照与请求快照入口
- `命中结果` 标签页已经升级为主工作台，而不是简单预览列表
- 结果工作台支持：
  - 当前页多选
  - 批量下线处置
  - 行级 `查看详情`
  - 行级 `下线处置`
  - 行级 `进入整改`
- 通过 `?tab=results&page=N` 保持结果视图与分页状态
- `规则快照`、`请求快照`、`关联日志` 继续保留在同一路由内

### 2.5 文稿详情 `/content/articles/:articleId`

- 展示真实文稿正文、最近巡检命中记录、操作记录与字段变更
- 通过 `return_to` 返回来源工作区
- 典型来源包括：
  - 任务详情结果工作台
  - 操作日志页
  - 文稿中心

### 2.6 整改编辑 `/content/articles/:articleId/rectify`

- 使用 `ProForm`
- 同屏对照旧值与新值
- `body` 一期按 HTML 源码编辑
- 支持按任务上下文携带 `task_id`、`result_id` 与 `return_to`
- 保存后可回到来源任务结果工作台

### 2.7 操作日志 `/audit/logs`

- 支持按文章 ID、任务 ID、操作人筛选
- 列出操作类型、状态流转、摘要、操作人、时间
- 支持查看请求参数快照
- 文章链接进入文稿详情
- 任务链接进入 `/inspection/tasks/:taskId?tab=results`

### 2.8 兼容结果入口 `/inspection/results`

- 保留是为了兼容旧链接与外部书签
- 该路由不再作为独立业务工作区
- 当存在 `task_id` 时，自动跳转到：
  - `/inspection/tasks/:taskId?tab=results`
- 当没有任务上下文时，自动跳转到：
  - `/inspection/tasks`

## 3. 推荐验收路径

### 3.1 最短 happy path

1. 打开 `/rules/keywords`，确认规则可正常展示与启停
2. 打开 `/inspection/tasks`，确认可从任务编号或 `查看任务` 进入详情
3. 进入 `/inspection/tasks/:taskId?tab=results`，检查命中结果工作台、批量下线与分页状态
4. 从结果工作台点击 `查看详情` 或 `进入整改`，确认 `return_to` 回跳到原任务结果视图
5. 打开 `/audit/logs`，点击任务链接，确认再次落回对应任务的结果视图

### 3.2 联调关注点

- 规则创建/编辑成功后列表是否刷新
- 任务列表是否始终可进入任务详情
- 任务详情结果工作台是否带上 `result_ids`、`task_id` 与正确分页状态
- 文稿详情/整改页是否能按 `return_to` 返回来源任务结果视图
- 日志页是否能回溯 `request_snapshot`、状态变化摘要、操作人，并正确跳回任务工作台

## 4. 本地运行参考

```bash
make up
make migrate
mysql -h127.0.0.1 -P3307 -uroot -proot article_sentinel < scripts/article_inspection_seed.sql
make dev
cd web/admin && npm install && npm run dev
```

- 后端地址: `http://127.0.0.1:8080`
- 前端地址: `http://127.0.0.1:5173`
- API 细节见 `docs/article-inspection-api.md`
