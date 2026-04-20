# Article Inspection Pages

本文档面向运营、测试、前端联调，描述一期控制台页面职责、入口路由与推荐验收路径。

## 1. 页面与路由

| 路由 | 页面 | 目标 |
| --- | --- | --- |
| `/keywords` | 关键词库 | 管理关键词、作用范围、风险等级、建议动作 |
| `/tasks` | 巡检任务列表 | 查看任务执行状态、统计与详情抽屉 |
| `/tasks/new` | 新建巡检任务 | 发起一次异步扫描 |
| `/results` | 命中结果列表 | 审核命中结果、批量下线、进入整改 |
| `/articles/:articleId/rectify` | 整改编辑页 | 对标题、摘要、正文做修订并提交下一状态 |
| `/logs` | 操作日志页 | 按文章、任务、操作人回溯处理轨迹 |

## 2. 页面说明

### 2.1 关键词库 `/keywords`

- 使用 `ProTable` 展示关键词、分类、风险、作用域、状态
- 支持打开新建弹窗与编辑弹窗
- 表单使用 `ModalForm`
- 适合校验关键词配置、风险等级与建议动作是否正确

### 2.2 任务列表 `/tasks`

- 展示任务编号、状态、命中统计、创建人、创建时间
- 支持打开详情抽屉查看规则快照
- 支持跳转 `/tasks/new`

### 2.3 新建任务 `/tasks/new`

- 填写 `orgid`
- 选择关键词集
- 配置时间范围、标题模糊匹配、文章 ID 精确匹配
- 支持 `include_body`
- 一期默认扫描 `state=9` 在线文稿

### 2.4 命中结果 `/results`

- 使用 `ProTable`
- 支持当前页多选、批量下线确认
- 风险级别使用 `Tag`
- 列表片段对命中词做高亮
- 行动作支持：
  - 查看详情
  - 单篇下线
  - 去整改

### 2.5 命中详情抽屉

- 采用 `Drawer + Tabs`
- 默认展示文章基础信息
- Tab 包含：
  - 命中明细
  - 正文片段
  - 操作历史
  - 字段变更

### 2.6 整改编辑 `/articles/:articleId/rectify`

- 使用 `ProForm`
- 同屏对照旧值与新值
- `body` 一期按 HTML 源码编辑
- 按钮区分：
  - `Save Rectification`
  - `Save & Send To Audit`
- “保存后重新上线”一期走统一生命周期服务，默认目标状态为 `1 / audit`

### 2.7 操作日志 `/logs`

- 支持按文章 ID、任务 ID、操作人筛选
- 列出操作类型、状态流转、摘要、操作人、时间
- 支持查看请求参数快照

## 3. 推荐验收路径

### 3.1 最短 happy path

1. 打开 `/keywords`，确认 seed 规则已存在
2. 打开 `/tasks`，查看 demo 任务状态
3. 打开 `/results`，检查高风险命中与详情抽屉
4. 点击 `Rectify` 进入整改页，保存一次整改
5. 打开 `/logs`，用文章 ID 与任务 ID 过滤，确认日志链路完整

### 3.2 联调关注点

- 关键词创建/编辑成功后列表是否刷新
- 任务详情是否正确展示规则快照
- 结果页批量动作是否带上 `result_ids` 或 `filter_snapshot`
- 整改提交是否落字段变更日志
- 日志页是否能回溯 `request_snapshot`、状态变化摘要、操作人

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
