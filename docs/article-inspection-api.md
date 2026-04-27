# Article Inspection API

本文档汇总 `article-sentinel` 一期文稿巡检 API 约定，便于联调前后端、补充 OpenAPI 注释、以及准备灰度验收脚本。

> 维护提示：当前接口现状请优先以 `internal/api/register/router.go` 和 `internal/api/register/router_test.go` 为准。本文件保留了部分设计阶段说明，如与代码冲突，请先修正文档再继续联调。

## 1. Base URL 与认证

- Base URL: `http://127.0.0.1:8080`
- API 前缀: `/api/v1/article-inspect`
- 文档入口: `/openapi.json`、`/docs`
- 认证策略:
  - 一期按 **OIDC-ready** 设计保留标准身份扩展位
  - 本地与内网联调可使用 `trusted_header` / `dev_header` fallback 注入操作人信息
  - 审计日志需要落 `operator_id`、`operator_name`、`request_id`、`source_ip`
  - 当前应用层鉴权中间件尚未真正挂到 `NewRouter`，如果线上依赖外层网关鉴权，请在交接时明确责任边界

## 2. 生命周期与通用常量

### 2.1 文稿状态常量

| 数值 | 名称 |
| --- | --- |
| `0` | `del` |
| `1` | `audit` |
| `2` | `back` |
| `3` | `draft` |
| `5` | `step` |
| `7` | `offline_sync` |
| `8` | `offline` |
| `9` | `online` |

### 2.2 一期处置约定

- 新建扫描任务默认筛选 `state=9` 在线文稿
- 下线动作通过统一生命周期服务处理
- 整改保存后的“重新上线”一期不直接回 `9`
- 一期保守策略：整改后默认目标状态按 `1 / audit` 处理

## 3. 路由总览

### 3.1 关键词管理

- `GET /api/v1/article-inspect/keywords`
- `POST /api/v1/article-inspect/keywords`
- `GET /api/v1/article-inspect/keywords/{id}`
- `PUT /api/v1/article-inspect/keywords/{id}`
- `PATCH /api/v1/article-inspect/keywords/{id}/status`
- `DELETE /api/v1/article-inspect/keywords/{id}`

### 3.2 巡检任务

- `GET /api/v1/article-inspect/tasks`
- `POST /api/v1/article-inspect/tasks`
- `GET /api/v1/article-inspect/tasks/{id}`
- `GET /api/v1/article-inspect/tasks/{id}/results`

### 3.3 命中结果与处置

- `GET /api/v1/article-inspect/results`
- `GET /api/v1/article-inspect/results/{id}`
- `POST /api/v1/article-inspect/results/export`
- `POST /api/v1/article-inspect/actions/batch-offline`
- `POST /api/v1/article-inspect/actions/batch-ignore`
- `POST /api/v1/article-inspect/actions/batch-process`
- `POST /api/v1/article-inspect/actions/batch-whitelist`

### 3.4 文章整改与生命周期

- `POST /api/v1/article-inspect/articles/{article_id}/offline`
- `PUT /api/v1/article-inspect/articles/{article_id}/rectify`
- `POST /api/v1/article-inspect/articles/{article_id}/republish`
- `GET /api/v1/article-inspect/articles/{article_id}/change-logs`
- `GET /api/v1/article-inspect/articles/{article_id}/operation-logs`

### 3.5 日志

- `GET /api/v1/article-inspect/logs/operations`
- `GET /api/v1/article-inspect/logs/field-changes`
- `GET /api/v1/article-inspect/logs/task-events`

## 4. 关键请求示例

### 4.1 创建关键词

```http
POST /api/v1/article-inspect/keywords
Content-Type: application/json

{
  "orgid": 100,
  "name": "spam",
  "category": "policy",
  "match_type": "contains",
  "risk_level": "high",
  "suggest_action": "offline",
  "enabled": true,
  "remark": "seed rule for acceptance",
  "scopes": ["title", "body"]
}
```

### 4.2 创建巡检任务

```http
POST /api/v1/article-inspect/tasks
Content-Type: application/json

{
  "orgid": 100,
  "keyword_ids": [10001],
  "include_body": true,
  "article_state": 9,
  "title_like": "spam"
}
```

示例响应:

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 20001,
    "task_no": "inspect-20260420-20001",
    "status": "pending"
  }
}
```

### 4.3 查询命中结果

```http
GET /api/v1/article-inspect/results?orgid=100&task_id=20001&risk_level=high&page=1&page_size=20
```

示例响应:

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "page": 1,
    "page_size": 20,
    "total": 1,
    "items": [
      {
        "id": 30001,
        "orgid": 100,
        "task_id": 20001,
        "article_id": 9001001,
        "article_title": "Spam headline needs audit",
        "article_state": 9,
        "risk_level": "high",
        "suggest_action": "offline",
        "hit_count": 3,
        "disposition_status": "pending"
      }
    ]
  }
}
```

### 4.4 批量下线

```http
POST /api/v1/article-inspect/actions/batch-offline
Content-Type: application/json

{
  "orgid": 100,
  "result_ids": [30001],
  "reason": "manual batch offline"
}
```

也支持通过 `filter_snapshot` 表达“对当前筛选结果全部处理”。

### 4.5 整改保存

```http
PUT /api/v1/article-inspect/articles/9001001/rectify
Content-Type: application/json

{
  "orgid": 100,
  "title": "Spam headline edited",
  "desc": "summary after manual cleanup",
  "body": "<p>clean html body</p>",
  "target_article_state": 1
}
```

说明:

- 仅保存整改草稿时可省略 `target_article_state`
- 保存后进入统一生命周期服务时，一期默认建议传 `1`

## 5. 查询参数建议

### 5.1 关键词列表

- `orgid`
- `category`
- `enabled`
- `keyword`
- `page`
- `page_size`

### 5.2 任务列表

- `orgid`
- `status`
- `task_no`
- `creator_id`
- `page`
- `page_size`

### 5.3 结果列表

- `orgid`
- `task_id`
- `keyword_id`
- `keyword_text`
- `risk_level`
- `disposition_status`
- `field_name`
- `article_id`
- `title`
- `page`
- `page_size`

### 5.4 操作日志

- `orgid`
- `article_id`
- `task_id`
- `operator_name`
- `page`
- `page_size`

## 6. 联调建议

1. 先执行 `make up && make migrate`
2. 通过 `mysql -h127.0.0.1 -P3307 -uroot -proot article_sentinel < scripts/article_inspection_seed.sql` 导入 demo 数据
3. 后端跑 `make dev`
4. 前端跑 `cd web/admin && npm install && npm run dev`
5. 打开 `/keywords` → `/tasks` → `/results` → `/logs` 走一遍验收路径
