# ArticleInspect Module Refactor Design

**Date:** 2026-05-08

## Goal

Refactor `internal/modules/articleinspect` from a flat single-directory implementation into a clearer business-subdomain structure while preserving all external contracts and runtime semantics.

This refactor must keep the following unchanged:

- HTTP paths, especially every `/api/v1/article-inspect/*` endpoint
- request/response JSON field names
- database table names, column names, GORM tags, migration behavior, and DDL
- task status semantics
- result disposition semantics
- action status semantics
- task outbox `pending / claimed / dispatched / dead_letter` semantics
- worker scan behavior, cursor-pagination behavior, retry behavior, and dead-letter behavior
- existing test behavior

## Scope

### In Scope

- reorganize `internal/modules/articleinspect` into clearer business groupings
- reduce navigation cost by grouping related files and introducing a phased target structure
- prepare the codebase for future package extraction by first removing implicit same-package coupling
- keep `module.go` and `routes.go` as stable composition and registration facades
- add design and implementation planning docs only in this phase

### Out of Scope

- changing runtime business behavior
- changing HTTP contracts or frontend call patterns
- changing DB schema or migration registration
- rewriting the worker/outbox/task execution model
- introducing a new architectural framework or heavy DDD abstraction layer
- code modifications in this round

## Current State

`internal/modules/articleinspect` currently contains 41 Go files in one package, `articleinspect`.

The module is already partially organized by file name, but still relies heavily on package-private cross-file reuse. That means many dependencies are currently hidden by the fact that everything lives in a single package. If the module is split into many real subpackages too early, those hidden edges become immediate import-cycle or over-export problems.

Two stable external module entrypoints already exist:

- `internal/modules/articleinspect/module.go:6` provides `NewRoutes`
- `internal/modules/articleinspect/routes.go:29` provides `RegisterRoutes`

Two more stable runtime entrypoints are used outside the module:

- `internal/modules/articleinspect/worker.go:23` provides `NewWorker`
- `internal/modules/articleinspect/task_outbox.go:44` provides `NewTaskOutboxRelay`

These are already consumed by:

- `internal/api/register/router.go:92`
- `internal/queue/asynq/handlers.go:21`
- `cmd/scheduler/main.go:36`

They should remain stable facades throughout the refactor.

## Current Responsibility Map

### Rules

Category management:

- `internal/modules/articleinspect/dto_categories.go`
- `internal/modules/articleinspect/repository_categories.go`
- `internal/modules/articleinspect/service_categories.go`
- `internal/modules/articleinspect/category_routes.go`

Keyword rule management:

- `internal/modules/articleinspect/dto_keywords.go`
- `internal/modules/articleinspect/validator_keywords.go`
- `internal/modules/articleinspect/repository_keywords.go`
- `internal/modules/articleinspect/service_keywords.go`
- `internal/modules/articleinspect/keyword_routes.go`

### Tasks

- `internal/modules/articleinspect/dto_tasks.go`
- `internal/modules/articleinspect/service_tasks.go`
- `internal/modules/articleinspect/task_routes.go`

### Outbox

- `internal/modules/articleinspect/task_outbox.go`

### Scan

- `internal/modules/articleinspect/scanner.go`

### Results

- `internal/modules/articleinspect/dto_results.go`
- `internal/modules/articleinspect/repository_results.go`
- `internal/modules/articleinspect/service_results.go`
- `internal/modules/articleinspect/result_routes.go`

### Actions

- `internal/modules/articleinspect/repository_actions.go`
- `internal/modules/articleinspect/action_service.go`
- `internal/modules/articleinspect/action_routes.go`

### Lifecycle

- `internal/modules/articleinspect/diff.go`
- `internal/modules/articleinspect/lifecycle_service.go`
- `internal/modules/articleinspect/lifecycle_routes.go`

### Articles

- `internal/modules/articleinspect/dto_articles.go`
- `internal/modules/articleinspect/repository_articles.go`
- `internal/modules/articleinspect/service_articles.go`
- `internal/modules/articleinspect/article_routes.go`

### Audit / Logs

- `internal/modules/articleinspect/audit_logs.go`
- `internal/modules/articleinspect/service_logs.go`
- `internal/modules/articleinspect/log_routes.go`

### Shared Transport / Request Utilities

- `internal/modules/articleinspect/routes_common.go`
- `internal/modules/articleinspect/request_scope.go`
- `internal/modules/articleinspect/operator.go`
- `internal/modules/articleinspect/paging.go`

### Domain / Models

- `internal/modules/articleinspect/constants.go`
- `internal/modules/articleinspect/model.go`

### Worker

- `internal/modules/articleinspect/worker.go`

### Contract Tests / Regression Guardrails

- `internal/modules/articleinspect/articleinspect_test.go`

## Dependency Analysis

### Stable composition root

The module wiring is clean and should remain the long-term facade:

- `internal/modules/articleinspect/module.go:6` assembles repositories and services
- `internal/modules/articleinspect/routes.go:29` mounts feature routes under `/api/v1/article-inspect`

This is the correct place to preserve external compatibility while internals evolve.

### Task creation and outbox chain

Task creation is a workflow, not CRUD.

- `internal/modules/articleinspect/service_tasks.go:105` creates `InspectionTask` and `InspectionTaskOutboxMessage` inside one transaction
- `internal/modules/articleinspect/service_tasks.go:136` immediately calls concrete `*TaskOutboxRelay`
- `internal/modules/articleinspect/service_tasks.go:183` serializes keyword snapshots into `RuleSnapshot`
- `internal/modules/articleinspect/service_tasks.go:269` deletes downstream rows across multiple business areas

This creates several hidden dependency edges:

- tasks -> rules DTO conversion
- tasks -> outbox concrete implementation
- tasks -> results/actions/audit tables via deletion cascade

### Worker and scan chain

Worker execution is another workflow seam.

- `internal/modules/articleinspect/worker.go:34` orchestrates task execution
- `internal/modules/articleinspect/worker.go:201` supports both new rule snapshot structure and legacy `[]KeywordDTO`
- `internal/modules/articleinspect/repository_articles.go:32` provides candidate article loading with cursor pagination
- `internal/modules/articleinspect/scanner.go:63` performs keyword matching
- `internal/modules/articleinspect/worker.go:134` persists `InspectionResult` and `InspectionResultHit`

This means worker currently depends on:

- scan algorithm types
- article candidate store behavior
- task snapshot compatibility
- result persistence models

### Action and lifecycle chain

This is the highest-risk cycle zone.

- `internal/modules/articleinspect/action_service.go:51` implements batch offline as a workflow
- `internal/modules/articleinspect/action_service.go:94` creates `LifecycleService` inside the same transaction
- `internal/modules/articleinspect/lifecycle_service.go:59` creates an `ActionRepository`
- `internal/modules/articleinspect/lifecycle_service.go:336` writes operation logs through `ActionRepository`
- `internal/modules/articleinspect/audit_logs.go` provides shared audit snapshot and summary helpers

This means:

- actions depends on lifecycle
- lifecycle depends on action repository
- both depend on audit helpers

If split naively into `actions`, `lifecycle`, and `audit`, this becomes an import-cycle trap.

### Results and logs chain

This is the second largest read-side coupling zone.

- `internal/modules/articleinspect/service_results.go:46` returns result detail plus operation logs plus field change logs
- `internal/modules/articleinspect/repository_results.go` owns both result queries and audit-log queries
- `internal/modules/articleinspect/service_logs.go` is mostly a thin shell over `ResultRepository`

That means results and logs are not actually separated yet. They only look separated at the route/service filename level.

### Articles repository mixed responsibility

`internal/modules/articleinspect/repository_articles.go` currently mixes two concerns:

- worker candidate article scanning input at `internal/modules/articleinspect/repository_articles.go:32`
- article center list/detail read models at `internal/modules/articleinspect/repository_articles.go:100`

This file should be split by responsibility before any package extraction.

## Import Cycle Risk Map

### Highest risk: `actions` <-> `lifecycle` <-> `audit`

Why it is risky:

- `ActionService.BatchOffline` directly constructs `LifecycleService`
- `LifecycleService` directly writes logs through `ActionRepository`
- both paths rely on shared audit helpers

Consequence of naive split:

- either direct cycle
- or forcing many internal helpers and writer functions to become exported
- or pushing too much unrelated code into a bad `shared` dumping-ground package

### High risk: `tasks` <-> `outbox`

Why it is risky:

- task creation writes outbox rows directly
- task service calls concrete outbox relay directly
- outbox relies on task state constants, errors, message models, and dispatcher contract

Consequence of naive split:

- direct tasks/outbox cycle
- or a flood of compatibility exports

### High risk: `tasks` <-> `rules` <-> `worker`

Why it is risky:

- task snapshot generation reuses keyword DTO building
- worker must still parse legacy `[]KeywordDTO` snapshot data

Consequence of naive split:

- worker ends up importing rules package for historical compatibility
- tasks ends up importing rules internals for snapshot construction

### Medium-high risk: `results` <-> `audit/logs`

Why it is risky:

- result detail endpoint aggregates audit logs
- log service currently reuses result repository query types and implementations

Consequence of naive split:

- cycle between result detail assembly and audit data access
- or duplicate log query logic

### Medium risk: `articles` <-> `scan`

Why it is risky:

- candidate article types currently align more with worker/scan needs
- article repository also serves article center pages

Consequence of naive split:

- article repository may have to import scan-owned types
- or scan package may have to depend on article-center read models

## Design Decision

Adopt a phased refactor with three principles:

1. Keep the root package facade stable
2. Split responsibilities inside the same package before splitting into real subpackages
3. Only extract true leaf packages early; defer workflow packages until dependency seams are explicit

## Recommended Target Structure

This is the recommended end-state shape, not the first-step structure.

```text
internal/modules/articleinspect/
├── module.go
├── routes.go
├── ports.go
├── doc.go
├── domain/
├── shared/
├── scan/
├── rules/
├── articles/
├── results/
├── audit/
├── tasks/
├── outbox/
├── worker/
├── actions/
└── lifecycle/
```

### Important constraint

Not every directory above should become a real Go package in the first implementation phase.

The safe order is:

- first: root package file regrouping
- second: `domain`, `shared`, `scan`
- third: `rules`, `articles`, `results`, `audit`
- fourth: `tasks`, `outbox`
- last: `worker`, `actions`, `lifecycle`

## Recommended Phases

### Phase 1: Same-package structural cleanup

Stay in package `articleinspect`.

Goals:

- split oversized files into clearer same-package files
- surface hidden dependency edges without changing import paths
- make future package boundaries visible

Recommended Phase 1 work:

- split `articleinspect_test.go` into multiple test files
- split `model.go` into inspection-owned models vs upstream source article models
- split `routes_common.go` into transport params / envelopes / error mapping
- split `task_outbox.go` into config / relay / state transition / codec files
- split `worker.go` into orchestration / rule decoding / persistence helpers
- split `repository_articles.go` into article-center reads vs candidate scan reads
- split `repository_results.go` into result queries vs audit-log queries

### Phase 2: Extract leaf packages

Create true subpackages only where dependency direction is simple.

Safe first candidates:

- `domain`
- `shared`
- `scan`

Rules for these packages:

- `domain` owns models and constants only
- `shared` owns only tiny reusable request/transport helpers
- `scan` owns keyword matching algorithm and related value types only

### Phase 3: Extract rules and read-side packages

After hidden same-package reuse is removed:

- extract `rules`
- extract `articles`
- extract `results`
- extract `audit`

Required preparation before this phase:

- stop tasks from reusing private keyword DTO builders indirectly
- stop logs from piggybacking on `ResultRepository` without a clear audit read boundary
- split mixed repositories before moving them

### Phase 4: Extract tasks and outbox

Do this only after the task creation seam is made explicit.

Required preparation:

- remove `TaskService` direct dependency on concrete `*TaskOutboxRelay`
- move immediate post-create relay orchestration to route/module composition or to a tiny relay interface in `ports.go`
- keep delete-graph behavior low-level and DB-centric, not service-fanout-centric

### Phase 5: Extract actions, lifecycle, and worker

Do this last.

Required preparation:

- split audit write capability out of `ActionRepository`
- make lifecycle depend on an audit writer seam instead of `ActionRepository`
- make action service depend on a lifecycle seam instead of directly calling `NewLifecycleService(tx)`
- preserve worker rule-snapshot backward compatibility and cursor pagination behavior

## Package Responsibility Rules

### Root package

The root directory should eventually keep only:

- `module.go` for dependency assembly
- `routes.go` for top-level Huma registration
- `ports.go` for cross-package seams like `TaskDispatcher`
- optional `doc.go`
- temporary compatibility aliases/facades during migration

### `domain`

Allowed:

- GORM models
- `TableName()` methods
- state constants and core value types

Forbidden:

- repositories
- HTTP DTO binding
- Huma route registration
- DB query logic

### `shared`

Allowed:

- tiny cross-subdomain helpers used by multiple packages
- request scope helpers
- paging helpers
- transport param helpers if truly shared

Forbidden:

- dumping business logic into `shared`
- broad util/helper/common grab-bag patterns

### `scan`

Allowed:

- scanner interface and implementation
- match result value types
- snippet / regex safety / matching logic

Forbidden:

- DB logic
- Huma DTOs
- task execution orchestration

### `outbox`

Allowed:

- reliable message relay semantics
- claim/retry/dead-letter logic
- payload encode/decode for outbox messages

Forbidden:

- scan business logic
- HTTP DTOs
- broad task orchestration

### `worker`

Allowed:

- async task execution orchestration only
- scan invocation
- candidate-article paging and result persistence coordination

Forbidden:

- HTTP DTO leakage
- route-specific transport logic

## Temporary Non-Goals Due to Cycle Risk

These boundaries should not be forced early:

- `actions` and `lifecycle`
- `tasks` and `outbox`
- `results` and `audit`
- `articles` candidate scan reads and article-center reads until repository responsibilities are split cleanly

## Compatibility Strategy

### HTTP compatibility

Keep unchanged:

- all route paths under `/api/v1/article-inspect`
- route methods
- `operationId` values
- current envelope shape
- current JSON field names

Guardrails already exist in:

- `internal/modules/articleinspect/articleinspect_test.go:2441`
- `internal/api/register/router_test.go:26`

### Database compatibility

Keep unchanged:

- all `TableName()` return values
- all GORM tags
- all column names
- all migration registrations and DDL behavior

Special caution:

- `Article` and `ArticleInfo` are upstream source tables, not articleinspect-owned tables
- future package extraction must not accidentally widen auto-migration scope in `internal/app/bootstrap/schema.go`

### Workflow compatibility

Keep unchanged:

- task creation transactional behavior
- outbox claim/retry/reclaim/dispatched/dead-letter semantics
- worker cursor pagination and break conditions
- rule snapshot backward compatibility in worker
- lifecycle state transition semantics
- batch action summary behavior

### Frontend compatibility

Current admin client paths are hard-coded in:

- `web/admin/src/services/categories.ts`
- `web/admin/src/services/keywords.ts`
- `web/admin/src/services/tasks.ts`
- `web/admin/src/services/results.ts`
- `web/admin/src/services/articles.ts`
- `web/admin/src/services/logs.ts`

Therefore route paths and response field names must remain unchanged.

## Verification Baseline

This design phase did not modify runtime code.

Baseline verification completed on 2026-05-08:

- `go test ./...` -> PASS
- `go vet ./...` -> PASS

These provide the current green baseline before any refactor work starts.

## Recommended First Implementation Slice

The first code-change slice should be intentionally boring:

1. create a dedicated worktree under `.worktrees`
2. split tests into multiple files in the same package
3. split `model.go` into multiple same-package files
4. split `routes_common.go` into multiple same-package files
5. split `task_outbox.go` into multiple same-package files
6. split `worker.go` into multiple same-package files
7. run `gofmt`, `go test ./...`, and `go vet ./...`

Only after that baseline lands should package extraction begin.

## References

- Go official module/package organization guidance: https://go.dev/doc/modules/layout
- Go official code organization guidance: https://go.dev/blog/organizing-go-code
- Huma route groups: https://github.com/danielgtaylor/huma/blob/main/docs/docs/features/groups.md
- Huma modular route registration/testing example: https://github.com/danielgtaylor/huma/blob/main/docs/docs/tutorial/writing-tests.md
