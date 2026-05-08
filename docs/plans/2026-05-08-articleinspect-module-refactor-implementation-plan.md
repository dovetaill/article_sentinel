# ArticleInspect Module Refactor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Refactor `internal/modules/articleinspect` toward a clearer business-subdomain structure without changing API, DB schema, worker/outbox semantics, or test-visible behavior.

**Architecture:** Use a phased migration. First split responsibilities inside the existing `articleinspect` package, then extract leaf subpackages (`domain`, `shared`, `scan`), then move read-side and rule packages, and only last split workflow-heavy packages (`tasks`, `outbox`, `worker`, `actions`, `lifecycle`). Preserve `NewRoutes`, `RegisterRoutes`, `NewWorker`, and `NewTaskOutboxRelay` as stable facades throughout.

**Tech Stack:** Go, Huma v2, GORM, Asynq, SQLite-backed Go tests, OpenAPI contract tests

## Review-Approved Execution Update (2026-05-08)

Critical review before execution found several plan issues that must be honored during implementation:

- do not execute `Task 1-3` as a single first batch; the approved first code batch is `Task 2` only
- `Task 3` must treat `ChuangqiOrg` as an upstream source model, not an articleinspect-owned model
- `Task 4` is removed from batch 1 and should execute later as its own contract-sensitive same-package split
- `Task 8` must be rewritten to avoid a root-level `ports.go` seam and to preserve current task-create plus pending-outbox semantics
- `Task 5` and `Task 6` must use stronger regression commands than the original draft

Additional verification rule for every task that may affect route contracts or frontend clients:

```bash
go test ./internal/api/register -run TestRouterRegistersArticleInspectRoutes -v
cd web/admin && npm test
```

---

### Task 1: Establish refactor guardrails and worktree baseline

**Files:**
- Create worktree: `.worktrees/articleinspect-module-refactor`
- Verify in place first: `internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Create the dedicated worktree**

Run:

```bash
git worktree add .worktrees/articleinspect-module-refactor -b articleinspect-module-refactor
```

Expected: new worktree created on branch `articleinspect-module-refactor`

**Step 2: Verify baseline in the worktree**

Run:

```bash
go test ./...
go vet ./...
```

Expected: PASS for both commands

**Step 3: Record current route and frontend contract guardrails**

Check these files before edits:

- `internal/modules/articleinspect/articleinspect_test.go`
- `internal/api/register/router_test.go`
- `web/admin/src/services/categories.ts`
- `web/admin/src/services/keywords.ts`
- `web/admin/src/services/tasks.ts`
- `web/admin/src/services/results.ts`
- `web/admin/src/services/articles.ts`
- `web/admin/src/services/logs.ts`

**Step 4: Commit the worktree baseline note if needed**

If a small doc note or TODO file was added in the worktree, commit it. Otherwise skip.

```bash
git status --short
```

Expected: either clean or only intentional setup files

### Task 2: Split the giant test file into focused same-package test files

**Files:**
- Modify: `internal/modules/articleinspect/articleinspect_test.go`
- Create: `internal/modules/articleinspect/model_test.go`
- Create: `internal/modules/articleinspect/scanner_test.go`
- Create: `internal/modules/articleinspect/task_service_test.go`
- Create: `internal/modules/articleinspect/task_outbox_test.go`
- Create: `internal/modules/articleinspect/worker_test.go`
- Create: `internal/modules/articleinspect/lifecycle_test.go`
- Create: `internal/modules/articleinspect/http_routes_test.go`
- Create: `internal/modules/articleinspect/openapi_test.go`
- Create: `internal/modules/articleinspect/fixtures_test.go`

**Step 1: Move test helpers and fixtures first**

Move package-private helpers such as request senders, seed helpers, and local test seams into `fixtures_test.go` without changing their code.

**Step 2: Move focused test groups into new files**

Suggested grouping:

- model and table-name assertions -> `model_test.go`
- scanner and diff tests -> `scanner_test.go`
- task create/delete tests -> `task_service_test.go`
- outbox relay tests -> `task_outbox_test.go`
- worker execution tests -> `worker_test.go`
- lifecycle and action flow tests -> `lifecycle_test.go`
- route HTTP behavior tests -> `http_routes_test.go`
- OpenAPI/path/operationId/schema tests -> `openapi_test.go`

**Step 3: Run focused package tests**

Run:

```bash
go test ./internal/modules/articleinspect -v
```

Expected: PASS with unchanged behavior

**Step 4: Run full regression**

Run:

```bash
go test ./...
go test ./internal/api/register -run TestRouterRegistersArticleInspectRoutes -v
cd web/admin && npm test
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/modules/articleinspect/*_test.go
git commit -m "refactor(articleinspect): split module tests by concern"
```

### Task 3: Split `model.go` into inspection-owned vs upstream source models

**Files:**
- Modify: `internal/modules/articleinspect/model.go`
- Create: `internal/modules/articleinspect/model_inspection.go`
- Create: `internal/modules/articleinspect/model_article_source.go`
- Check: `internal/app/bootstrap/schema.go`

**Step 1: Move inspection-owned models into `model_inspection.go`**

Move these unchanged:

- `InspectionTimestamps`
- `InspectionCategory`
- `InspectionKeyword`
- `InspectionKeywordScope`
- `InspectionTask`
- `InspectionTaskKeyword`
- `InspectionTaskOutboxMessage`
- `InspectionResult`
- `InspectionResultHit`
- `InspectionAction`
- `InspectionOperationLog`
- `InspectionFieldChangeLog`

**Step 2: Move upstream source tables into `model_article_source.go`**

Move unchanged:

- `ChuangqiOrg`
- `Article`
- `ArticleInfo`

Keep every `gorm` tag, `json` tag, and `TableName()` method byte-for-byte identical.

**Step 3: Run focused tests**

Run:

```bash
go test ./internal/modules/articleinspect -run 'TestInspectionModelMetadata|TestMigrationFileContainsInspectionTables' -v
```

Expected: PASS

**Step 4: Run full verification**

Run:

```bash
gofmt -w internal/modules/articleinspect/model*.go
go test ./...
go vet ./...
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/modules/articleinspect/model*.go
git commit -m "refactor(articleinspect): split inspection and source models"
```

### Task 4: Split shared transport helpers without changing package boundary

Do not include this task in batch 1. Execute it only after the test split and model split have already landed cleanly.

**Files:**
- Modify: `internal/modules/articleinspect/routes_common.go`
- Create: `internal/modules/articleinspect/transport_params.go`
- Create: `internal/modules/articleinspect/transport_envelope.go`
- Create: `internal/modules/articleinspect/transport_errors.go`

**Step 1: Move custom param types into `transport_params.go`**

Move unchanged:

- `uint64Param`
- `optionalUint64Param`
- `optionalIntParam`
- `optionalBoolParam`
- `optionalInt8Param`
- `optionalTimeParam`

**Step 2: Move response envelope types/helpers into `transport_envelope.go`**

Move unchanged:

- `okEnvelopeOutput`
- `createdEnvelopeOutput`
- `successOKEnvelope`
- `successCreatedEnvelope`
- failure envelope helpers

**Step 3: Move error/status mapping into `transport_errors.go`**

Move unchanged:

- `articleInspectStatusFromError`

**Step 4: Run HTTP/OpenAPI guardrails**

Run:

```bash
go test ./internal/modules/articleinspect -run 'TestRouteRegistrationRegistersArticleInspectPaths|TestHandlerKeywordTaskAndResultsRoutes|TestHandlerOrgCategoryAndArticleCenterContracts' -v
go test ./internal/api/register -run TestRouterRegistersArticleInspectRoutes -v
cd web/admin && npm test
```

Expected: PASS

**Step 5: Run full verification and commit**

```bash
gofmt -w internal/modules/articleinspect/transport_*.go internal/modules/articleinspect/routes_common.go
go test ./...
go vet ./...
git add internal/modules/articleinspect/routes_common.go internal/modules/articleinspect/transport_*.go
git commit -m "refactor(articleinspect): split shared transport helpers"
```

### Task 5: Split `task_outbox.go` by concern inside the same package

**Files:**
- Modify: `internal/modules/articleinspect/task_outbox.go`
- Create: `internal/modules/articleinspect/task_outbox_settings.go`
- Create: `internal/modules/articleinspect/task_outbox_codec.go`
- Create: `internal/modules/articleinspect/task_outbox_relay.go`
- Create: `internal/modules/articleinspect/task_outbox_cleanup.go`

**Step 1: Move settings/config helpers**

Move unchanged:

- `TaskOutboxSettings`
- `NewTaskOutboxSettings`
- `defaultTaskOutboxSettings`
- `defaultTaskOutboxClaimOwner`

**Step 2: Move payload codec and dispatchability helpers**

Move unchanged:

- `decodeTaskOutboxPayload`
- `isDispatchableMessage`

**Step 3: Move relay orchestration and state transitions**

Move unchanged:

- `TaskOutboxRelay`
- `NewTaskOutboxRelay`
- `WithSettings`
- `CanDispatch`
- `TryDispatchMessage`
- `DispatchMessage`
- `DispatchPending`
- claim/retry/dead-letter/dispatched helpers

**Step 4: Move cleanup path**

Move unchanged:

- `CleanupArticleInspectTaskOutbox`

**Step 5: Run focused outbox tests**

Run:

```bash
go test ./internal/modules/articleinspect -run 'TestTaskCreationWithOutbox|TestTaskOutboxRelayDispatchesPendingMessage|TestTaskOutboxRelayReclaimsExpiredLease|TestTaskOutboxRelayRetryableFailureSchedulesNextAttempt|TestTaskOutboxRelayMovesPoisonMessageToDeadLetter|TestTaskOutboxRelayImplementsCleanerContract|TestTaskOutboxRelayDeadLettersPoisonRowWithoutBlockingLaterMessages|TestTaskCreateEnqueueFailureLeavesPendingOutbox|TestTaskCreateWithoutDispatcherStillCreatesPendingOutbox' -v
```

Expected: PASS

**Step 6: Run full verification and commit**

```bash
gofmt -w internal/modules/articleinspect/task_outbox*.go
go test ./...
go vet ./...
git add internal/modules/articleinspect/task_outbox*.go
git commit -m "refactor(articleinspect): split outbox implementation by concern"
```

### Task 6: Split `worker.go` by orchestration, snapshot decode, and persistence

**Files:**
- Modify: `internal/modules/articleinspect/worker.go`
- Create: `internal/modules/articleinspect/worker_rules.go`
- Create: `internal/modules/articleinspect/worker_persist.go`

**Step 1: Keep `Worker` and `ExecuteTask` in `worker.go`**

Retain:

- `Worker`
- `NewWorker`
- `ExecuteTask`
- `startTask`
- `finishTask`

**Step 2: Move rule snapshot helpers into `worker_rules.go`**

Move unchanged:

- `decodeTaskRules`
- `parseArticleStateFilter`
- `resolveTaskStatus`

Do not change legacy `KeywordDTO` snapshot compatibility behavior.

**Step 3: Move persistence/count helpers into `worker_persist.go`**

Move unchanged:

- `persistArticleResult`
- `uniqueFieldCount`
- `uniqueKeywordCount`

**Step 4: Run focused worker tests**

Run:

```bash
go test ./internal/modules/articleinspect -run 'TestArticleInspectWorker|TestDecodeTaskRulesFromKeywordDTOJSON' -v
```

Expected: PASS

**Step 5: Run full verification and commit**

```bash
gofmt -w internal/modules/articleinspect/worker*.go
go test ./...
go vet ./...
git add internal/modules/articleinspect/worker*.go
git commit -m "refactor(articleinspect): split worker orchestration and helpers"
```

### Task 7: Split mixed repositories before package extraction

**Files:**
- Modify: `internal/modules/articleinspect/repository_articles.go`
- Create: `internal/modules/articleinspect/repository_article_candidates.go`
- Create: `internal/modules/articleinspect/repository_article_center.go`
- Modify: `internal/modules/articleinspect/repository_results.go`
- Create: `internal/modules/articleinspect/repository_result_queries.go`
- Create: `internal/modules/articleinspect/repository_audit_queries.go`

**Step 1: Split article repository by responsibility**

Candidate-scan side:

- `ListCandidateArticles`
- candidate support helpers

Article-center side:

- `ListArticles`
- `GetArticleDetail`
- article-center support helpers

Keep the same `ArticleRepository` type for now.

**Step 2: Split result repository by responsibility**

Result side:

- `ResultListInput`
- `ListResults`
- `GetResult`
- `ListHits`
- result-hit helpers

Audit-log side:

- `OperationLogListInput`
- `FieldChangeLogListInput`
- `ListOperationLogs`
- `ListFieldChangeLogs`

Keep the same `ResultRepository` type for now.

**Step 3: Run focused tests**

Run:

```bash
go test ./internal/modules/articleinspect -run 'TestCandidateArticleLoading|TestHandlerOrgCategoryAndArticleCenterContracts|TestHandlerKeywordTaskAndResultsRoutes' -v
```

Expected: PASS

**Step 4: Run full verification and commit**

```bash
gofmt -w internal/modules/articleinspect/repository_*.go
go test ./...
go vet ./...
git add internal/modules/articleinspect/repository_*.go
git commit -m "refactor(articleinspect): split mixed article and result repositories"
```

### Task 8: Introduce a consumer-local immediate relay seam and remove direct concrete relay coupling from task service

**Files:**
- Modify: `internal/modules/articleinspect/service_tasks.go`
- Modify: `internal/modules/articleinspect/task_routes.go`
- Modify: `internal/modules/articleinspect/routes.go`
- Modify: `internal/modules/articleinspect/module.go`

This task must preserve current task-create runtime semantics:

- task row and outbox row are still written in one DB transaction
- immediate dispatch remains best-effort and out-of-transaction
- when immediate dispatch fails, HTTP create still succeeds and the outbox row remains `pending`
- `dispatcher == nil` and dispatcher failure paths must keep their current `attempt_count` behavior

Do not introduce a root-level `ports.go` seam for this task.

**Step 1: Add a minimal consumer-local relay seam**

Add the smallest interface needed by task creation near the consuming code, for example inside `service_tasks.go`:

```go
type TaskOutboxImmediateRelay interface {
	TryDispatchMessage(ctx context.Context, outboxID uint64) bool
}
```

**Step 2: Change `TaskService.CreateAndEnqueue` to depend on the seam, not `*TaskOutboxRelay`**

Update the signature only as much as needed.

**Step 3: Keep orchestration in route/module layer**

Let route/module composition continue constructing the concrete relay and pass it through the seam without changing the runtime behavior above.

**Step 4: Run focused task tests**

Run:

```bash
go test ./internal/modules/articleinspect -run 'TestTaskCreation|TestTaskCreationWithOutbox|TestTaskCreateEnqueueFailureLeavesPendingOutbox|TestTaskCreateWithoutDispatcherStillCreatesPendingOutbox|TestHandlerKeywordTaskAndResultsRoutes' -v
go test ./internal/api/register -run TestRouterRegistersArticleInspectRoutes -v
```

Expected: PASS

**Step 5: Run full verification and commit**

```bash
gofmt -w internal/modules/articleinspect/service_tasks.go internal/modules/articleinspect/task_routes.go internal/modules/articleinspect/routes.go internal/modules/articleinspect/module.go
go test ./...
go vet ./...
git add internal/modules/articleinspect/service_tasks.go internal/modules/articleinspect/task_routes.go internal/modules/articleinspect/routes.go internal/modules/articleinspect/module.go
git commit -m "refactor(articleinspect): decouple task service from concrete outbox relay"
```

### Task 9: Extract `domain`, `shared`, and `scan` as the first real subpackages

**Files:**
- Create dir/package: `internal/modules/articleinspect/domain`
- Create dir/package: `internal/modules/articleinspect/shared`
- Create dir/package: `internal/modules/articleinspect/scan`
- Modify imports across articleinspect module files
- Check: `internal/app/bootstrap/schema.go`

**Step 1: Move domain files**

Move to `domain`:

- inspection models
- source article models
- constants

If needed, add temporary root-level type aliases to keep external call sites stable during the transition.

**Step 2: Move shared helpers**

Move only truly shared items:

- paging helpers
- request scope helpers
- operator helpers
- transport helper types used by multiple route groups

Do not create a generic utility bucket.

**Step 3: Move scan implementation**

Move scanner types and implementation to `scan`.

**Step 4: Update root facade compatibility**

Keep these root exports working:

- `NewRoutes`
- `RegisterRoutes`
- `NewWorker`
- `NewTaskOutboxRelay`
- `NewTaskOutboxSettings`

**Step 5: Run focused regression**

Run:

```bash
go test ./internal/modules/articleinspect -v
go test ./internal/api/register -v
go test ./internal/queue/asynq -v
go test ./internal/scheduler -v
```

Expected: PASS

**Step 6: Run full verification and commit**

```bash
gofmt -w internal/modules/articleinspect
go test ./...
go vet ./...
git add internal/modules/articleinspect internal/app/bootstrap/schema.go
git commit -m "refactor(articleinspect): extract domain shared and scan packages"
```

### Task 10: Extract `rules` after removing hidden same-package DTO coupling

**Files:**
- Create dir/package: `internal/modules/articleinspect/rules`
- Modify: current category and keyword files
- Modify: `internal/modules/articleinspect/service_tasks.go`
- Add helper in rules package for snapshot construction if needed

**Step 1: Stop task service from reusing private keyword DTO builder**

Add a stable rule snapshot builder owned by rules logic so `TaskService` does not depend on package-private keyword DTO helpers.

**Step 2: Move category and keyword implementations**

Move:

- routes
- DTOs
- validator
- repositories
- services

**Step 3: Keep root route/module facade wiring stable**

`module.go` and `routes.go` should continue to assemble and mount through the root package.

**Step 4: Run focused tests**

Run:

```bash
go test ./internal/modules/articleinspect -run 'TestKeywordService|TestKeywordValidation|TestHandlerKeywordTaskAndResultsRoutes|TestHandlerOrgCategoryAndArticleCenterContracts' -v
```

Expected: PASS

**Step 5: Run full verification and commit**

```bash
gofmt -w internal/modules/articleinspect
go test ./...
go vet ./...
git add internal/modules/articleinspect
git commit -m "refactor(articleinspect): extract rules package"
```

### Task 11: Extract `articles`, `results`, and audit read-side concerns

**Files:**
- Create dir/package: `internal/modules/articleinspect/articles`
- Create dir/package: `internal/modules/articleinspect/results`
- Create dir/package: `internal/modules/articleinspect/audit`
- Modify root wiring and imports

**Step 1: Move article-center concerns into `articles`**

Move article DTOs, service, routes, and article-center repository code.

**Step 2: Move result concerns into `results`**

Move result DTOs, service, routes, and result query repository code.

**Step 3: Move audit/log read concerns into `audit`**

Move:

- log routes
- log service
- audit query repository code

Keep audit write helpers in the root package for now if they are still used by action/lifecycle write flows. Move those only after the audit-writer seam in Task 13 exists.

**Step 4: Preserve result detail assembly semantics**

`ResultDetail` must still include:

- result
- hits
- operation logs
- field change logs

Keep HTTP response shape identical.

**Step 5: Run focused regression**

Run:

```bash
go test ./internal/modules/articleinspect -run 'TestHandlerKeywordTaskAndResultsRoutes|TestHandlerOrgCategoryAndArticleCenterContracts|TestRouteRegistrationRegistersArticleInspectPaths' -v
```

Expected: PASS

**Step 6: Run full verification and commit**

```bash
gofmt -w internal/modules/articleinspect
go test ./...
go vet ./...
git add internal/modules/articleinspect
git commit -m "refactor(articleinspect): extract read-side packages"
```

### Task 12: Extract `tasks` and `outbox`

**Files:**
- Create dir/package: `internal/modules/articleinspect/tasks`
- Create dir/package: `internal/modules/articleinspect/outbox`
- Modify root wiring and imports

**Step 1: Move task DTOs, task service, and task routes**

Keep delete-graph logic behavior unchanged.

**Step 2: Move outbox relay implementation**

Keep unchanged:

- pending/claimed/dispatched/dead_letter semantics
- reclaim behavior
- retry scheduling behavior
- poison-message dead-letter behavior
- cleanup behavior

**Step 3: Keep scheduler and worker-facing facades stable**

External files must keep working with minimal changes:

- `cmd/scheduler/main.go`
- `internal/api/register/router.go`

**Step 4: Run focused regression**

Run:

```bash
go test ./internal/modules/articleinspect -run 'TestTaskCreation|TestTaskCreationWithOutbox|TestTaskOutboxRelayDispatchesPendingMessage|TestTaskOutboxRelayReclaimsExpiredLease|TestTaskOutboxRelayRetryableFailureSchedulesNextAttempt|TestTaskOutboxRelayMovesPoisonMessageToDeadLetter|TestTaskCreateEnqueueFailureLeavesPendingOutbox|TestTaskCreateWithoutDispatcherStillCreatesPendingOutbox|TestTaskOutboxRelayDeadLettersPoisonRowWithoutBlockingLaterMessages' -v
```

Expected: PASS

**Step 5: Run full verification and commit**

```bash
gofmt -w internal/modules/articleinspect
go test ./...
go vet ./...
git add internal/modules/articleinspect cmd/scheduler/main.go internal/api/register/router.go
git commit -m "refactor(articleinspect): extract task and outbox packages"
```

### Task 13: Extract audit write seam, then extract `actions`, `lifecycle`, and `worker`

**Files:**
- Modify: `internal/modules/articleinspect/repository_actions.go`
- Modify: `internal/modules/articleinspect/lifecycle_service.go`
- Modify: `internal/modules/articleinspect/action_service.go`
- Create dir/package: `internal/modules/articleinspect/actions`
- Create dir/package: `internal/modules/articleinspect/lifecycle`
- Create dir/package: `internal/modules/articleinspect/worker`
- Modify: `internal/queue/asynq/handlers.go`

**Step 1: Introduce an audit-writer seam**

Move operation-log and field-change-log writes behind a small interface or dedicated audit writer so lifecycle no longer depends directly on `ActionRepository`.

The seam must be transaction-aware. Acceptable shapes:

- inject a lifecycle factory that binds to the current `*gorm.DB` transaction
- or define seam methods that explicitly accept the current `*gorm.DB`

Do not break the existing transaction boundary by moving lifecycle work outside the action transaction.

**Step 2: Remove direct `NewLifecycleService(tx)` coupling from action service**

Make `BatchOffline` depend on a lifecycle seam instead of constructing lifecycle directly.

**Step 3: Extract action package**

Move routes, service, and action repository code.

**Step 4: Extract lifecycle package**

Move lifecycle routes, service, and diff logic.

**Step 5: Extract worker package last**

Move worker orchestration after scan, task, article-candidate, and result-persist seams are already stable.

Keep unchanged:

- cursor pagination loop behavior
- break conditions
- legacy rule snapshot compatibility
- task start/finish semantics

**Step 6: Run focused regression**

Run:

```bash
go test ./internal/modules/articleinspect -run 'TestArticleInspectWorker|TestHandlerKeywordTaskAndResultsRoutes|TestHandlerOrgCategoryAndArticleCenterContracts' -v
go test ./internal/queue/asynq -v
go test ./internal/scheduler -v
```

Expected: PASS

**Step 7: Run final full verification and commit**

```bash
gofmt -w internal/modules/articleinspect internal/queue/asynq
go test ./...
go vet ./...
git add internal/modules/articleinspect internal/queue/asynq/handlers.go
git commit -m "refactor(articleinspect): extract workflow packages"
```

### Task 14: Final compatibility audit

**Files:**
- Check only; modify if drift is found
- `internal/modules/articleinspect/*`
- `internal/api/register/router_test.go`
- `web/admin/src/services/*.ts`

**Step 1: Verify paths and operation IDs**

Run:

```bash
go test ./internal/modules/articleinspect -run 'TestRouteRegistrationRegistersArticleInspectPaths' -v
go test ./internal/api/register -v
```

Expected: PASS with unchanged paths and stable operation IDs

**Step 2: Verify frontend compatibility manually**

Check that these paths still exist and return the same field names:

- `/api/v1/article-inspect/categories`
- `/api/v1/article-inspect/keywords`
- `/api/v1/article-inspect/tasks`
- `/api/v1/article-inspect/results`
- `/api/v1/article-inspect/articles`
- `/api/v1/article-inspect/logs/operations`
- `/api/v1/article-inspect/logs/field-changes`

**Step 3: Verify full backend regression**

Run:

```bash
go test ./...
go vet ./...
```

Expected: PASS

**Step 4: Commit final cleanup**

```bash
git add .
git commit -m "refactor(articleinspect): complete modular subdomain reorganization"
```
