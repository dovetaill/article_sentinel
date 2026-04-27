# API Response Snake Case Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Normalize every backend HTTP response field used by the current admin UI to standard snake_case while keeping existing routes and request shapes stable.

**Architecture:** Keep the outer response envelope unchanged and replace implicit frontend-facing struct serialization with explicit DTOs and JSON tags. Update `web/admin` services and tests in the same change so the repo converges on one response contract without backend compatibility shims.

**Tech Stack:** Go, Huma, GORM, React, TypeScript, Vitest

---

### Task 1: Lock the current batch-offline route fix into its own commit

**Files:**
- Modify: `internal/modules/articleinspect/handler.go`
- Modify: `internal/modules/articleinspect/action_service.go`
- Modify: `internal/modules/articleinspect/articleinspect_test.go`
- Modify: `internal/api/register/router_test.go`

**Step 1: Inspect the current diff for the batch-offline route fix**

Run: `git diff -- internal/modules/articleinspect/handler.go internal/modules/articleinspect/action_service.go internal/modules/articleinspect/articleinspect_test.go internal/api/register/router_test.go`
Expected: only the batch-offline route/service/test changes appear.

**Step 2: Run the focused backend regression tests**

Run: `go test ./internal/modules/articleinspect -run 'TestBatchAction|TestHandlerBatchActionsValidateTargets|TestRouteRegistrationRegistersArticleInspectPaths' && go test ./internal/api/register -run TestRouterRegistersArticleInspectRoutes`
Expected: PASS.

**Step 3: Commit the isolated route fix**

```bash
git add internal/modules/articleinspect/handler.go \
  internal/modules/articleinspect/action_service.go \
  internal/modules/articleinspect/articleinspect_test.go \
  internal/api/register/router_test.go
git commit -m "fix(api): add article inspect batch offline route"
```

### Task 2: Add failing backend contract tests for normalized response keys

**Files:**
- Modify: `internal/modules/articleinspect/articleinspect_test.go`
- Modify: `internal/modules/post/post_test.go` or the existing post handler/service test file if needed

**Step 1: Write failing tests for batch action and lifecycle response keys**

Add tests that decode `Envelope.data` from handler responses and assert keys such as:

- `action_id`
- `target_count`
- `success_count`
- `article_id`
- `before_state`
- `after_state`
- `field_name`
- `before_value`
- `after_value`

**Step 2: Run the focused backend tests to verify they fail for the expected reason**

Run: `go test ./internal/modules/articleinspect -run 'TestHandler.*SnakeCase|TestBatchAction.*SnakeCase'`
Expected: FAIL because the current handler payload still exposes PascalCase and legacy field names.

**Step 3: Add a small failing test for at least one historical field rename**

Use a frontend-visible endpoint such as org/category/task/log payloads and assert:

- `org_id` instead of `orgid`
- `cate_id` instead of `cateid`
- `created_at` instead of `create_at`

**Step 4: Re-run the focused backend tests**

Run: `go test ./internal/modules/articleinspect -run 'TestHandler.*SnakeCase|Test.*Normalized.*'`
Expected: FAIL with contract-key mismatches.

### Task 3: Implement explicit backend DTO normalization

**Files:**
- Modify: `internal/modules/articleinspect/action_service.go`
- Modify: `internal/modules/articleinspect/lifecycle_service.go`
- Modify: `internal/modules/articleinspect/diff.go`
- Modify: `internal/modules/articleinspect/dto_categories.go`
- Modify: `internal/modules/articleinspect/dto_keywords.go`
- Modify: `internal/modules/articleinspect/dto_tasks.go`
- Modify: `internal/modules/articleinspect/dto_results.go`
- Modify: `internal/modules/articleinspect/dto_articles.go`
- Modify: `internal/modules/articleinspect/service_results.go`
- Modify: `internal/modules/articleinspect/service_logs.go`
- Modify: `internal/modules/articleinspect/handler.go`
- Modify: `internal/modules/post/model.go`
- Modify: `internal/modules/post/service.go` if post list/detail DTOs need explicit response wrappers

**Step 1: Add explicit JSON tags or response DTOs for action summaries**

Implement snake_case fields for batch action responses:

```go
type BatchActionSummary struct {
    ActionID     uint64 `json:"action_id"`
    TargetCount  int64  `json:"target_count"`
    SuccessCount int64  `json:"success_count"`
    FailCount    int64  `json:"fail_count"`
    SkipCount    int64  `json:"skip_count"`
    Status       string `json:"status"`
    ActionType   string `json:"action_type"`
}
```

**Step 2: Add explicit JSON tags or response DTOs for lifecycle results and field changes**

Implement snake_case fields for:

- `LifecycleActionResult`
- `FieldChange`

**Step 3: Normalize legacy frontend-visible DTO fields**

Update JSON tags to canonical snake_case for frontend-facing DTOs, including representative changes:

- `OrgDTO.CateID` -> `json:"cate_id"`
- `CategoryDTO.OrgID` -> `json:"org_id"`
- `CategoryDTO.CreateAt` -> `json:"created_at"`
- `CategoryDTO.UpdateAt` -> `json:"updated_at"`
- same pattern for keyword/task/result/article/log DTOs where needed

**Step 4: Keep routes and request parsing unchanged**

Do not rename:

- route paths
- query names like `orgid`
- body field names used by the current frontend requests

**Step 5: Run the focused backend tests to verify they pass**

Run: `go test ./internal/modules/articleinspect -run 'TestHandler.*SnakeCase|Test.*Normalized.*|TestBatchAction|TestHandlerBatchActionsValidateTargets|TestRouteRegistrationRegistersArticleInspectPaths' && go test ./internal/api/register -run TestRouterRegistersArticleInspectRoutes`
Expected: PASS.

### Task 4: Update frontend services and tests to the normalized contract

**Files:**
- Modify: `web/admin/src/services/orgs.ts`
- Modify: `web/admin/src/services/categories.ts`
- Modify: `web/admin/src/services/keywords.ts`
- Modify: `web/admin/src/services/tasks.ts`
- Modify: `web/admin/src/services/results.ts`
- Modify: `web/admin/src/services/articles.ts`
- Modify: `web/admin/src/services/logs.ts`
- Modify: affected `web/admin/src/pages/**/*.test.tsx`
- Modify: affected `web/admin/src/services/*.test.ts`

**Step 1: Write failing frontend service/page tests for the normalized keys**

Update mocks to return only snake_case fields such as:

- `org_id`
- `cate_id`
- `created_at`
- `updated_at`
- `action_id`
- `success_count`

**Step 2: Run the focused frontend tests to verify they fail**

Run: `cd web/admin && npm test -- --runInBand src/services src/pages`
Expected: FAIL where the frontend still expects legacy response keys.

**Step 3: Update frontend types and adapters**

Examples:

- `OrgRecord.cateid` -> `cate_id`
- `CategoryRecord.orgid` -> `org_id`
- remove `create_at ?? created_at` compatibility logic
- align batch action/lifecycle return types with the real backend DTOs

**Step 4: Run the focused frontend tests to verify they pass**

Run: `cd web/admin && npm test -- --runInBand src/services src/pages`
Expected: PASS.

### Task 5: Verify the integrated contract and commit the normalization

**Files:**
- Modify: `docs/plans/2026-04-27-api-response-snake-case-design.md` only if implementation details materially changed
- Modify: any touched backend/frontend files from Tasks 3 and 4

**Step 1: Run full backend verification**

Run: `go test ./...`
Expected: PASS.

**Step 2: Run targeted frontend verification for touched suites**

Run: `cd web/admin && npm test -- --runInBand src/services src/pages`
Expected: PASS.

**Step 3: Smoke-check the dev API contract**

Run:

```bash
curl -sS -X POST 'http://127.0.0.1:8080/api/v1/article-inspect/actions/batch-offline' \
  -H 'Content-Type: application/json' \
  --data '{"orgid":29,"result_ids":[35],"reason":"task-offline-action"}'
```

Expected: response envelope data contains snake_case keys like `action_id`, `target_count`, `success_count`.

**Step 4: Commit the normalization change**

```bash
git add internal/modules/articleinspect \
  internal/modules/post \
  web/admin/src/services \
  web/admin/src/pages \
  web/admin/src/**/*.test.ts \
  web/admin/src/**/*.test.tsx \
  docs/plans/2026-04-27-api-response-snake-case-design.md \
  docs/plans/2026-04-27-api-response-snake-case.md
git commit -m "refactor(api): normalize frontend response fields to snake_case"
```
