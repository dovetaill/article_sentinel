# Article Sentinel Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Initialize a new `article-sentinel` project from the PureMux starter and deliver the first production-ready version of the article inspection system with keyword rules, async scan tasks, result handling, rectification, audit logs, admin UI, and tests.

**Architecture:** Start from the PureMux starter in the current project root, rename starter-facing project metadata to `article-sentinel`, then add a single new backend business module `internal/modules/articleinspect`, Asynq worker integration, and a standalone `web/admin` React app. Keep document scanning application-side by paginating `xt_article`, batch-loading `xt_article_info`, scanning in Go, and routing all article state changes through a unified lifecycle service.

**Tech Stack:** Go 1.25, GORM, Huma v2, Asynq, MySQL 5.7, Redis, React, TypeScript, Vite, Ant Design, ProComponents.

---

### Task 1: Initialize the starter in the current project root

**Files:**
- Create: `/home/wwwroot/article_sentinel/*` copied from `/home/wwwroot/PureMux/*`
- Modify: `/home/wwwroot/article_sentinel/go.mod`
- Modify: `/home/wwwroot/article_sentinel/README.md`
- Modify: `/home/wwwroot/article_sentinel/configs/config.example.yaml`
- Modify: `/home/wwwroot/article_sentinel/configs/config.local.yaml`

**Step 1: Copy the PureMux starter into the current directory**

Run:

```bash
rsync -a /home/wwwroot/PureMux/ /home/wwwroot/article_sentinel/ --exclude .git --exclude .worktrees --exclude .worktree
```

Expected: starter files exist in `/home/wwwroot/article_sentinel` and there is no inherited `.git` directory.

**Step 2: Rename starter-facing metadata to `article-sentinel`**

Update:

- `go.mod` module path to the new project module name
- app name / docs title in config and bootstrap defaults
- `README.md` title and local startup instructions

Expected: project no longer presents itself as `PureMux` except where explicitly documenting the starter origin.

**Step 3: Run module sync and baseline formatting**

Run:

```bash
go mod tidy
```

Expected: module metadata is consistent for the renamed project.

**Step 4: Baseline smoke check**

Run:

```bash
go test ./...
```

Expected: starter baseline passes before feature work begins, or any existing failure is documented before proceeding.

**Step 5: Commit**

```bash
git add .
git commit -m "chore: initialize article-sentinel from starter"
```

### Task 2: Add article inspection constants, models, and schema registration

**Files:**
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/constants.go`
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/model.go`
- Modify: `/home/wwwroot/article_sentinel/internal/app/bootstrap/schema.go`
- Test: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Write the failing schema/model test**

Add tests asserting:

- inspection model table names match the expected `xt_article_inspect_*` names
- article lifecycle constants include the provided existing states
- result / task status constants expose only approved values

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/modules/articleinspect -run TestInspectionModelMetadata -v
```

Expected: FAIL because the module does not exist yet.

**Step 3: Write the minimal model implementation**

Implement:

- inspection table models
- existing article read models for `xt_article` and `xt_article_info`
- table name methods
- shared status / scope / risk constants

**Step 4: Register the business models**

Add inspection models to `internal/app/bootstrap/schema.go` so the starter migrate command can create the new tables.

**Step 5: Run test to verify it passes**

Run:

```bash
go test ./internal/modules/articleinspect -run TestInspectionModelMetadata -v
```

Expected: PASS.

**Step 6: Commit**

```bash
git add internal/modules/articleinspect internal/app/bootstrap/schema.go
git commit -m "feat: add article inspection domain models"
```

### Task 3: Add database migrations for inspection tables

**Files:**
- Create: `/home/wwwroot/article_sentinel/migrations/20260420_01_article_inspection.sql`
- Modify: `/home/wwwroot/article_sentinel/README.md`
- Test: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Write the failing migration test**

Add a test that verifies the migration file exists and includes all required table names:

- `xt_article_inspect_keywords`
- `xt_article_inspect_keyword_scopes`
- `xt_article_inspect_tasks`
- `xt_article_inspect_task_keywords`
- `xt_article_inspect_results`
- `xt_article_inspect_result_hits`
- `xt_article_inspect_actions`
- `xt_article_inspect_field_change_logs`
- `xt_article_inspect_operation_logs`

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/modules/articleinspect -run TestMigrationFileContainsInspectionTables -v
```

Expected: FAIL because the migration file is missing.

**Step 3: Write the migration**

Create MySQL 5.7-compatible SQL with:

- all required tables
- required indexes and uniqueness constraints
- `orgid`, `create_at`, `update_at` on every table

**Step 4: Re-run the test**

Run:

```bash
go test ./internal/modules/articleinspect -run TestMigrationFileContainsInspectionTables -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add migrations/20260420_01_article_inspection.sql internal/modules/articleinspect/articleinspect_test.go README.md
git commit -m "feat: add article inspection schema migration"
```

### Task 4: Implement keyword repository and service

**Files:**
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/repository_keywords.go`
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/service_keywords.go`
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/dto_keywords.go`
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/validator_keywords.go`
- Test: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Write the failing keyword service tests**

Cover:

- create keyword with multi-scope values
- reject missing `orgid`
- reject unsupported scope / risk / action values
- enable / disable keyword
- list keywords by `orgid`

**Step 2: Run the keyword tests to verify failure**

Run:

```bash
go test ./internal/modules/articleinspect -run 'TestKeyword(Service|Validation)' -v
```

Expected: FAIL because service and repository methods do not exist.

**Step 3: Implement repository methods**

Implement minimal repository methods for:

- create keyword + scopes
- update keyword + scopes
- delete keyword
- get keyword detail
- list keywords with paging and filtering
- patch keyword enabled status

**Step 4: Implement service validation and orchestration**

Implement:

- enum validation
- scope deduplication
- audit field propagation from operator context
- list result pagination wrapper

**Step 5: Re-run the keyword tests**

Run:

```bash
go test ./internal/modules/articleinspect -run 'TestKeyword(Service|Validation)' -v
```

Expected: PASS.

**Step 6: Commit**

```bash
git add internal/modules/articleinspect
git commit -m "feat: add article inspection keyword management"
```

### Task 5: Implement scanner and diff helpers

**Files:**
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/scanner.go`
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/diff.go`
- Test: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Write the failing scanner and diff tests**

Cover:

- `contains` match against title
- `exact` match against keyword field
- safe regex match
- body snippet extraction with context
- multiple hits from multiple scopes
- field diff generation for title/body changes
- no diff output when values are unchanged

**Step 2: Run tests to verify failure**

Run:

```bash
go test ./internal/modules/articleinspect -run 'Test(KeywordScanner|FieldDiff)' -v
```

Expected: FAIL because helpers are missing.

**Step 3: Implement minimal scanner**

Implement:

- `Scanner` interface
- `KeywordScanner`
- scope dispatch for `title`, `short_title`, `rich_title`, `keyword`, `desc`, `body`
- snippet generation and hit metadata
- regex safety guardrails (length cap / compile timeout avoidance / limited pattern acceptance)

**Step 4: Implement minimal diff helper**

Implement:

- editable field struct
- field-by-field compare
- long text summary truncation for body diffs

**Step 5: Re-run tests**

Run:

```bash
go test ./internal/modules/articleinspect -run 'Test(KeywordScanner|FieldDiff)' -v
```

Expected: PASS.

**Step 6: Commit**

```bash
git add internal/modules/articleinspect
git commit -m "feat: add article inspection scanner and diff helpers"
```

### Task 6: Implement candidate article loading and task creation

**Files:**
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/repository_articles.go`
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/service_tasks.go`
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/dto_tasks.go`
- Test: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Write the failing task service tests**

Cover:

- create task with valid `orgid` and keyword IDs
- reject missing `orgid`
- persist task snapshots
- paginate candidate article reads by `orgid`, `state=9`, `publish_at_time`
- support exact article ID and fuzzy title filters

**Step 2: Run tests to verify failure**

Run:

```bash
go test ./internal/modules/articleinspect -run 'Test(TaskCreation|CandidateArticleLoading)' -v
```

Expected: FAIL.

**Step 3: Implement candidate article repository methods**

Implement:

- cursor-based `xt_article` loading
- batch `xt_article_info` loading by article IDs
- candidate article DTO assembly

**Step 4: Implement task creation service**

Implement:

- parameter validation
- keyword snapshot capture
- task / task-keyword persistence
- initial task stats setup

**Step 5: Re-run tests**

Run:

```bash
go test ./internal/modules/articleinspect -run 'Test(TaskCreation|CandidateArticleLoading)' -v
```

Expected: PASS.

**Step 6: Commit**

```bash
git add internal/modules/articleinspect
git commit -m "feat: add inspection task creation and candidate loading"
```

### Task 7: Implement Asynq task payloads and worker execution

**Files:**
- Modify: `/home/wwwroot/article_sentinel/internal/queue/tasks/tasks.go`
- Create: `/home/wwwroot/article_sentinel/internal/queue/tasks/articleinspect.go`
- Modify: `/home/wwwroot/article_sentinel/internal/queue/asynq/handlers.go`
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/worker.go`
- Test: `/home/wwwroot/article_sentinel/internal/queue/asynq/asynq_test.go`
- Test: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Write the failing worker tests**

Cover:

- enqueue builds `articleinspect:run-task` payload
- worker handler dispatches to article inspect executor
- successful batch updates task counters and status
- mixed batch failures end in `partial_success`

**Step 2: Run tests to verify failure**

Run:

```bash
go test ./internal/queue/... ./internal/modules/articleinspect -run 'Test(ArticleInspectQueue|ArticleInspectWorker)' -v
```

Expected: FAIL.

**Step 3: Implement task payload and enqueue helpers**

Add a dedicated payload type and constructor for article inspection tasks.

**Step 4: Implement worker executor**

Implement:

- task loading
- state transition `pending -> running`
- batch scan loop
- result persistence
- task stat updates
- final status resolution

**Step 5: Re-run tests**

Run:

```bash
go test ./internal/queue/... ./internal/modules/articleinspect -run 'Test(ArticleInspectQueue|ArticleInspectWorker)' -v
```

Expected: PASS.

**Step 6: Commit**

```bash
git add internal/queue internal/modules/articleinspect
git commit -m "feat: add article inspection worker execution"
```

### Task 8: Implement lifecycle and batch action services

**Files:**
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/lifecycle_service.go`
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/action_service.go`
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/repository_actions.go`
- Test: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Write the failing lifecycle/action tests**

Cover:

- offline transitions `9 -> 8`
- already offline records as skipped
- rectify updates article fields and writes change logs
- republish defaults from `8 -> 1` unless configured otherwise
- batch ignore and batch processed are idempotent

**Step 2: Run tests to verify failure**

Run:

```bash
go test ./internal/modules/articleinspect -run 'Test(Lifecycle|BatchAction)' -v
```

Expected: FAIL.

**Step 3: Implement lifecycle service**

Implement:

- `OfflineArticle`
- `UpdateArticleFields`
- `RepublishArticle`
- central state transition rules and TODO-backed config hooks

**Step 4: Implement action service**

Implement:

- batch action envelope
- action summary counts
- operation log persistence
- result disposition updates

**Step 5: Re-run tests**

Run:

```bash
go test ./internal/modules/articleinspect -run 'Test(Lifecycle|BatchAction)' -v
```

Expected: PASS.

**Step 6: Commit**

```bash
git add internal/modules/articleinspect
git commit -m "feat: add article lifecycle and batch action services"
```

### Task 9: Implement list/detail/log query services

**Files:**
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/repository_results.go`
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/service_results.go`
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/service_logs.go`
- Test: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Write the failing query tests**

Cover:

- results list filters by `orgid`, task, risk, status, article title / ID
- result detail includes hit list and recent logs
- operation log query filters by article / task / operator / time
- field change log query filters by article / field / time

**Step 2: Run tests to verify failure**

Run:

```bash
go test ./internal/modules/articleinspect -run 'Test(ResultQuery|OperationLogQuery|FieldChangeLogQuery)' -v
```

Expected: FAIL.

**Step 3: Implement repository query methods**

Implement the paged list and detail queries using explicit `orgid` filters in every query.

**Step 4: Implement service aggregation**

Implement:

- paged response shaping
- detail aggregation for hits + operations + changes
- safe empty-state handling

**Step 5: Re-run tests**

Run:

```bash
go test ./internal/modules/articleinspect -run 'Test(ResultQuery|OperationLogQuery|FieldChangeLogQuery)' -v
```

Expected: PASS.

**Step 6: Commit**

```bash
git add internal/modules/articleinspect
git commit -m "feat: add inspection result and log queries"
```

### Task 10: Implement authentication context and operator resolver

**Files:**
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/operator.go`
- Modify: `/home/wwwroot/article_sentinel/internal/middleware/auth.go`
- Modify: `/home/wwwroot/article_sentinel/internal/identity/*.go`
- Test: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/articleinspect_test.go`
- Test: `/home/wwwroot/article_sentinel/internal/middleware/auth_test.go`

**Step 1: Write the failing operator resolution tests**

Cover:

- JWT / bearer context maps to operator fields
- trusted header mode only accepts configured header source
- request ID and IP extraction are preserved for audit logs

**Step 2: Run tests to verify failure**

Run:

```bash
go test ./internal/middleware ./internal/modules/articleinspect -run 'Test(OperatorResolver|TrustedHeaderAuth)' -v
```

Expected: FAIL.

**Step 3: Implement operator resolver**

Implement:

- operator DTO from identity context
- fallback extraction for trusted header / dev header
- helper methods used by handlers and services

**Step 4: Re-run tests**

Run:

```bash
go test ./internal/middleware ./internal/modules/articleinspect -run 'Test(OperatorResolver|TrustedHeaderAuth)' -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/middleware internal/identity internal/modules/articleinspect
git commit -m "feat: add operator resolver for inspection audit"
```

### Task 11: Implement Huma handlers and route wiring

**Files:**
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/handler.go`
- Modify: `/home/wwwroot/article_sentinel/internal/api/register/router.go`
- Test: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Write the failing handler tests**

Cover:

- keyword CRUD routes return unified envelope
- task creation route enqueues async work
- result list and detail routes return paged / detail payloads
- batch action routes validate target input
- rectify and republish routes reject missing `orgid`

**Step 2: Run tests to verify failure**

Run:

```bash
go test ./internal/modules/articleinspect -run 'Test(Handler|RouteRegistration)' -v
```

Expected: FAIL.

**Step 3: Implement handlers**

Implement Huma operations with:

- OpenAPI metadata
- DTO parsing
- error-to-envelope mapping
- service wiring

**Step 4: Register routes**

Wire the new service into `internal/api/register/router.go` following the existing `post` module pattern.

**Step 5: Re-run tests**

Run:

```bash
go test ./internal/modules/articleinspect -run 'Test(Handler|RouteRegistration)' -v
```

Expected: PASS.

**Step 6: Commit**

```bash
git add internal/modules/articleinspect internal/api/register/router.go
git commit -m "feat: expose article inspection APIs"
```

### Task 12: Build the admin frontend shell and shared API client

**Files:**
- Create: `/home/wwwroot/article_sentinel/web/admin/package.json`
- Create: `/home/wwwroot/article_sentinel/web/admin/tsconfig.json`
- Create: `/home/wwwroot/article_sentinel/web/admin/vite.config.ts`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/main.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/App.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/lib/request.ts`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/routes.tsx`
- Test: `/home/wwwroot/article_sentinel/web/admin/src/**/*.test.tsx`

**Step 1: Write the failing frontend shell tests**

Cover:

- app bootstraps
- router renders navigation shell
- API client unwraps PureMux envelope and throws readable errors

**Step 2: Run tests to verify failure**

Run:

```bash
cd /home/wwwroot/article_sentinel/web/admin && npm test -- --runInBand
```

Expected: FAIL because the frontend project does not exist.

**Step 3: Create minimal frontend shell**

Implement:

- Vite React app
- Ant Design / ProComponents layout shell
- lightweight API client
- base route structure

**Step 4: Re-run tests**

Run:

```bash
cd /home/wwwroot/article_sentinel/web/admin && npm test -- --runInBand
```

Expected: PASS.

**Step 5: Commit**

```bash
git add web/admin
git commit -m "feat: scaffold inspection admin frontend"
```

### Task 13: Build keyword and task management pages

**Files:**
- Create: `/home/wwwroot/article_sentinel/web/admin/src/pages/keywords/index.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/index.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/new.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/services/keywords.ts`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/services/tasks.ts`
- Test: `/home/wwwroot/article_sentinel/web/admin/src/pages/**/*.test.tsx`

**Step 1: Write the failing page tests**

Cover:

- keyword list renders data and opens create/edit modal
- task list renders status tags and detail action
- task creation page submits valid form data

**Step 2: Run tests to verify failure**

Run:

```bash
cd /home/wwwroot/article_sentinel/web/admin && npm test -- --runInBand keywords tasks
```

Expected: FAIL.

**Step 3: Implement pages**

Use:

- `ProTable` for lists
- `ProForm` / `ModalForm` for forms
- clear validation and submit feedback

**Step 4: Re-run tests**

Run:

```bash
cd /home/wwwroot/article_sentinel/web/admin && npm test -- --runInBand keywords tasks
```

Expected: PASS.

**Step 5: Commit**

```bash
git add web/admin
git commit -m "feat: add keyword and task admin pages"
```

### Task 14: Build results, detail, rectify, and logs pages

**Files:**
- Create: `/home/wwwroot/article_sentinel/web/admin/src/pages/results/index.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/pages/results/detail.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/rectify.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/pages/logs/index.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/services/results.ts`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/services/logs.ts`
- Test: `/home/wwwroot/article_sentinel/web/admin/src/pages/**/*.test.tsx`

**Step 1: Write the failing page tests**

Cover:

- results page supports row selection and batch action confirmation
- detail drawer shows hits and operation history
- rectify page displays old/new values and submits update
- logs page filters by article / operator / task

**Step 2: Run tests to verify failure**

Run:

```bash
cd /home/wwwroot/article_sentinel/web/admin && npm test -- --runInBand results logs rectify
```

Expected: FAIL.

**Step 3: Implement pages**

Implement:

- result list with risk tags and highlighted snippets
- batch action confirmation modals
- rectify form with HTML textarea for body
- operation log table and detail modal

**Step 4: Re-run tests**

Run:

```bash
cd /home/wwwroot/article_sentinel/web/admin && npm test -- --runInBand results logs rectify
```

Expected: PASS.

**Step 5: Commit**

```bash
git add web/admin
git commit -m "feat: add result handling and audit pages"
```

### Task 15: Add docs, API references, demo seed, and end-to-end verification

**Files:**
- Create: `/home/wwwroot/article_sentinel/docs/article-inspection-api.md`
- Create: `/home/wwwroot/article_sentinel/docs/article-inspection-pages.md`
- Create: `/home/wwwroot/article_sentinel/scripts/article_inspection_seed.sql`
- Modify: `/home/wwwroot/article_sentinel/README.md`
- Test: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Write the failing documentation presence test**

Cover:

- API doc exists
- page doc exists
- seed script exists

**Step 2: Run tests to verify failure**

Run:

```bash
go test ./internal/modules/articleinspect -run TestInspectionDocsArtifactsExist -v
```

Expected: FAIL.

**Step 3: Write docs and seed artifacts**

Document:

- route list
- request / response examples
- page descriptions
- local startup and acceptance flow
- demo seed instructions

**Step 4: Run full verification**

Run:

```bash
go test ./...
cd /home/wwwroot/article_sentinel/web/admin && npm test -- --runInBand
```

Expected: all backend and frontend tests pass.

**Step 5: Run starter verification commands**

Run:

```bash
make verify
```

Expected: project verification passes, or remaining gaps are documented before completion.

**Step 6: Commit**

```bash
git add docs scripts README.md
git commit -m "docs: add article inspection docs and demo assets"
```
