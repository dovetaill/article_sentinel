# Frontend API Response Snake Case Design

## Context

The current `web/admin` frontend consumes a bounded set of backend HTTP APIs under:

- `/api/v1/article-inspect/*`
- `/api/v1/posts*`

Those APIs already share a common envelope shape:

- `code`
- `message`
- `data`

Most list and detail payloads are already close to snake_case, but the current contract is still inconsistent in two ways:

1. Some action-oriented responses serialize exported Go struct fields without explicit JSON tags, producing PascalCase keys such as `ActionID`, `SuccessCount`, `ArticleID`, and `BeforeState`.
2. Some older payloads use historical field names like `orgid`, `cateid`, `create_at`, and `update_at`, which are not standard snake_case and force the frontend to carry compatibility logic.

The missing `batch-offline` route exposed one contract gap already. The next step is to normalize every response shape used by the current admin so the frontend can rely on one naming style everywhere.

## Product Goals

- Make every response field returned to the current admin UI use standard snake_case.
- Keep the outer response envelope unchanged: `code`, `message`, `data`.
- Avoid introducing new route changes or new business behavior while normalizing field names.
- Minimize long-term frontend compatibility code by converging backend and frontend on one contract.

## Scope

This change only covers backend HTTP APIs currently used by `web/admin`:

- article inspection org/category/keyword/task/result/article/log endpoints
- article inspection action and lifecycle endpoints
- posts endpoints currently registered under `/api/v1/posts`

Out of scope:

- request parameter renames for now
- internal repository/model field names
- APIs not used by the current frontend

## Confirmed Decisions

### 1. Response-only contract change

This normalization applies to response bodies only.

Examples:

- keep request/query `orgid` as-is for now
- normalize response fields from `orgid` to `org_id`
- keep route and request shapes stable to reduce breakage during rollout

This avoids turning one cleanup into a broader request-contract migration.

### 2. Explicit response DTOs over implicit struct serialization

Handlers should stop returning raw structs that rely on default Go JSON field naming when those structs are part of the public frontend contract.

Instead:

- add explicit DTOs with JSON tags
- map domain/service values into those DTOs before returning from handlers or service methods

This is required for:

- batch action summaries
- article lifecycle action results
- field change records returned by rectify
- any remaining frontend-visible payloads using historical names

### 3. Standard naming rules

The normalized naming rules are:

- `camelCase` and `PascalCase` response fields become `snake_case`
- abbreviated concatenations become expanded snake_case where the field represents multiple words
- timestamp fields use `created_at` / `updated_at`
- organization/category identifiers use `org_id` / `cate_id`

Representative examples:

- `ActionID` -> `action_id`
- `SuccessCount` -> `success_count`
- `ArticleID` -> `article_id`
- `BeforeState` -> `before_state`
- `orgid` -> `org_id`
- `cateid` -> `cate_id`
- `create_at` -> `created_at`
- `update_at` -> `updated_at`

### 4. No dual-field compatibility in backend responses

The backend should not emit both legacy and normalized keys in the same payload.

Why:

- duplicate keys prolong drift
- frontend migration is local to this repo and can be updated together
- one clean cut is easier to test and document than mixed compatibility responses

The frontend will be updated in the same change set.

## Affected API Families

### 1. Action endpoints

These currently need explicit normalization:

- `POST /api/v1/article-inspect/actions/batch-offline`
- `POST /api/v1/article-inspect/actions/batch-ignore`
- `POST /api/v1/article-inspect/actions/batch-process`

Current risk:

- return PascalCase summary fields because `BatchActionSummary` lacks JSON tags

Target response shape:

```json
{
  "code": 0,
  "message": "batch action applied",
  "data": {
    "action_id": 1,
    "target_count": 1,
    "success_count": 1,
    "fail_count": 0,
    "skip_count": 0,
    "status": "success",
    "action_type": "offline"
  }
}
```

### 2. Article lifecycle endpoints

These currently need explicit normalization:

- `POST /api/v1/article-inspect/articles/{article_id}/offline`
- `POST /api/v1/article-inspect/articles/{article_id}/republish`
- `PUT /api/v1/article-inspect/articles/{article_id}/rectify`

Current risk:

- lifecycle result and field change values serialize with PascalCase when returned directly

Target examples:

```json
{
  "article_id": 9001,
  "status": "success",
  "before_state": 9,
  "after_state": 8
}
```

```json
[
  {
    "field_name": "body",
    "before_value": "old",
    "after_value": "new",
    "diff_summary": "body: old -> new"
  }
]
```

### 3. List/detail payloads with legacy field names

These need DTO cleanup so every frontend-visible field becomes canonical snake_case:

- org list
- category list/detail/mutation responses
- keyword list/detail/mutation responses
- task list/detail responses
- result list/detail responses
- article list/detail responses
- operation log / field change log responses
- posts list/detail responses already align and mainly need verification

Specific historical names to normalize:

- `orgid` -> `org_id`
- `cateid` -> `cate_id`
- `create_at` -> `created_at`
- `update_at` -> `updated_at`

## Backend Design

### Response DTO strategy

Introduce or expand dedicated DTOs for every frontend-facing response whose current field names are not canonical snake_case.

Guidelines:

- service-layer domain structs may stay unchanged internally
- handler/service boundaries should return DTOs with explicit JSON tags
- embedded GORM models should not leak legacy field names into frontend contracts unless already canonical

### Where to normalize

- `internal/modules/articleinspect/action_service.go`
  - batch action summary DTOs
- `internal/modules/articleinspect/lifecycle_service.go`
  - lifecycle action result DTOs
- `internal/modules/articleinspect/diff.go`
  - field change DTO tags or response mapping
- `internal/modules/articleinspect/dto_*.go`
  - org/category/keyword/task/result/article DTO field tags
- `internal/modules/articleinspect/service_logs.go`
  - paged log result item field names
- `internal/modules/post/*.go`
  - verify no response renames needed beyond consistency checks

### Handler behavior

Routes stay the same.

Success messages stay the same unless tests show a mismatch. The main contract change is field naming inside `data`.

### Documentation and OpenAPI

OpenAPI-backed handler registration will automatically pick up explicit JSON tags once DTOs are used. Tests should lock the contract by asserting key payloads and registered paths.

## Frontend Design

### Service contract update

Update `web/admin/src/services/*` to consume only normalized snake_case response fields.

Examples:

- `OrgRecord.cateid` -> `OrgRecord.cate_id`
- `CategoryRecord.orgid` -> `CategoryRecord.org_id`
- `CategoryRecord.create_at` -> `CategoryRecord.created_at`
- `TaskRecord.created_at` becomes the only supported create timestamp field

### Remove compatibility glue

Where frontend code currently normalizes mixed backend shapes, remove compatibility fallback once the backend contract is updated.

Examples:

- remove `create_at ?? created_at` style merges where no longer needed
- remove assumptions that batch action endpoints return `action_no`
- align lifecycle service return types with the actual backend DTOs

### Test updates

Update affected frontend tests to use only normalized keys in mocked payloads.

## Testing Strategy

### Backend

Add or extend tests for:

- batch action responses use snake_case fields
- lifecycle action responses use snake_case fields
- rectify field change responses use snake_case fields
- org/category/task/log/article/result payloads expose only normalized keys where relevant
- router coverage still includes the current frontend route set

### Frontend

Update service and page tests to assert the normalized response contract.

### Full verification

Run:

- targeted Go package tests while iterating
- targeted Vitest suites for changed frontend services/pages
- full `go test ./...` before completion

## Rollout Strategy

1. Commit the already-fixed `batch-offline` route as its own isolated backend fix.
2. Land the snake_case response normalization as a separate change.
3. Restart the local dev stack and verify the admin against the normalized contract.

## Commit Strategy

Recommended commit split:

1. `fix(api): add article inspect batch offline route`
2. `refactor(api): normalize frontend response fields to snake_case`
