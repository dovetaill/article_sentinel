# Admin IA Redesign Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Rebuild the admin app around organization-scoped master data, task-centric inspection workflows, and a `shadcn-admin`-style shell while fixing the current `/api` dev 404s and queue startup gaps.

**Architecture:** Introduce organization and category master data in the Go backend, convert keyword rules and task creation to use controlled organization/category references, and restore `/articles` as a real article center backed by `xt_article` and `xt_article_info`. On the frontend, replace the current simplified shell with a collapsible sidebar + fixed header frame, centralize active-organization state, move risk operations into `/tasks/:id/results`, and proxy `/api` requests from Vite to the backend during local development.

**Tech Stack:** Go, Huma, Gorm, MySQL migrations, React 18, React Router 7, Ant Design 5, Vitest, Vite, Asynq, Make

---

### Task 1: Lock the backend route and data-shape contracts

**Files:**
- Modify: `internal/api/register/router_test.go`
- Modify: `internal/modules/articleinspect/articleinspect_test.go`
- Reference: `internal/api/register/router.go`
- Reference: `internal/modules/articleinspect/handler.go`
- Reference: `internal/modules/articleinspect/model.go`

**Step 1: Write failing router-contract tests**

Update `internal/api/register/router_test.go` so the starter router must register these additional paths:

```go
wantPaths := []string{
  "/api/v1/article-inspect/articles",
  "/api/v1/article-inspect/articles/{article_id}",
  "/api/v1/article-inspect/categories",
  "/api/v1/article-inspect/categories/{id}",
  "/api/v1/article-inspect/categories/{id}/status",
  "/api/v1/article-inspect/orgs",
  // keep existing inspect routes here too
}
```

Also add operation assertions for:

```go
assertOperation(t, doc.Paths, "/api/v1/article-inspect/orgs", http.MethodGet)
assertOperation(t, doc.Paths, "/api/v1/article-inspect/categories", http.MethodGet)
assertOperation(t, doc.Paths, "/api/v1/article-inspect/categories", http.MethodPost)
assertOperation(t, doc.Paths, "/api/v1/article-inspect/categories/{id}", http.MethodGet)
assertOperation(t, doc.Paths, "/api/v1/article-inspect/categories/{id}", http.MethodPut)
assertOperation(t, doc.Paths, "/api/v1/article-inspect/categories/{id}", http.MethodDelete)
assertOperation(t, doc.Paths, "/api/v1/article-inspect/categories/{id}/status", http.MethodPatch)
assertOperation(t, doc.Paths, "/api/v1/article-inspect/articles", http.MethodGet)
assertOperation(t, doc.Paths, "/api/v1/article-inspect/articles/{article_id}", http.MethodGet)
```

**Step 2: Write failing article-inspect behavior tests**

In `internal/modules/articleinspect/articleinspect_test.go`, add coverage for:

- listing organizations returns seeded org `29 / 一县一端`
- categories are scoped by `orgid`
- category CRUD rejects missing `orgid`
- keyword create/update payloads use `category_id`
- article list endpoint reads real articles, not aggregated inspect results
- article detail endpoint includes basic article data plus recent inspect summary hooks

Keep tests concrete and table-driven around the existing request helpers in the file.

**Step 3: Run the focused Go tests to verify they fail**

Run: `go test ./internal/api/register ./internal/modules/articleinspect`
Expected: FAIL because the new routes, DTOs, and handlers do not exist yet.

**Step 4: Capture the minimum implementation targets in comments or TODO notes only if needed**

Do not implement yet; only note the intended shape if a test needs a small clarifying comment, for example:

```go
// org list is read-only for this phase; create/edit org management is out of scope
```

**Step 5: Re-run the focused Go tests without implementation**

Run: `go test ./internal/api/register ./internal/modules/articleinspect`
Expected: FAIL for the same missing-route and missing-handler reasons.

**Step 6: Commit**

```bash
git add internal/api/register/router_test.go internal/modules/articleinspect/articleinspect_test.go
git commit -m "test(api): lock admin IA backend contracts"
```

### Task 2: Add organization/category master data and real article-center APIs to the backend

**Files:**
- Modify: `migrations/20260420_01_article_inspection.sql`
- Modify: `internal/modules/articleinspect/model.go`
- Create: `internal/modules/articleinspect/dto_categories.go`
- Create: `internal/modules/articleinspect/repository_categories.go`
- Create: `internal/modules/articleinspect/service_categories.go`
- Create: `internal/modules/articleinspect/dto_articles.go`
- Modify: `internal/modules/articleinspect/repository_articles.go`
- Create: `internal/modules/articleinspect/service_articles.go`
- Modify: `internal/modules/articleinspect/dto_keywords.go`
- Modify: `internal/modules/articleinspect/repository_keywords.go`
- Modify: `internal/modules/articleinspect/service_keywords.go`
- Modify: `internal/modules/articleinspect/validator_keywords.go`
- Modify: `internal/modules/articleinspect/handler.go`
- Modify: `internal/api/register/router.go`
- Test: `internal/api/register/router_test.go`
- Test: `internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Write the migration and model changes**

Extend `migrations/20260420_01_article_inspection.sql` to create and seed:

- `xt_chuangqi_org`
- `xt_article_inspect_categories`
- seed org `29 / 一县一端`
- seed a small category set for org `29`

Also update `internal/modules/articleinspect/model.go` with matching Gorm models.

Recommended minimal Gorm shapes:

```go
type ChuangqiOrg struct {
    ID        uint64    `gorm:"column:id;primaryKey" json:"id"`
    Name      string    `gorm:"column:name;size:128;not null" json:"name"`
    CateID    uint64    `gorm:"column:cateid;not null;default:0" json:"cateid"`
    Enabled   bool      `gorm:"column:enabled;not null;default:true" json:"enabled"`
    Sort      int64     `gorm:"column:sort;not null;default:0" json:"sort"`
    CreateAt  time.Time `gorm:"column:create_at;not null;autoCreateTime" json:"create_at"`
    UpdateAt  time.Time `gorm:"column:update_at;not null;autoUpdateTime" json:"update_at"`
}
```

```go
type InspectionCategory struct {
    ID          uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
    OrgID       uint64 `gorm:"column:orgid;not null;index" json:"orgid"`
    Name        string `gorm:"column:name;size:128;not null" json:"name"`
    Code        string `gorm:"column:code;size:64;not null" json:"code"`
    Enabled     bool   `gorm:"column:enabled;not null;default:true" json:"enabled"`
    Sort        int64  `gorm:"column:sort;not null;default:0" json:"sort"`
    CreatorID   uint64 `gorm:"column:creator_id;not null;default:0" json:"creator_id"`
    CreatorName string `gorm:"column:creator_name;size:128;not null;default:''" json:"creator_name"`
    UpdaterID   uint64 `gorm:"column:updater_id;not null;default:0" json:"updater_id"`
    UpdaterName string `gorm:"column:updater_name;size:128;not null;default:''" json:"updater_name"`
    InspectionTimestamps
}
```

**Step 2: Implement repository/service/handler support for organizations and categories**

Add organization list support and category CRUD support under `internal/modules/articleinspect`, then wire them in `internal/api/register/router.go` and `internal/modules/articleinspect/handler.go`.

Use the existing module conventions for:

- request DTOs
- paging
- envelope responses
- validation error style

**Step 3: Convert keyword DTOs to category references**

Replace free-form category request usage with controlled references:

- keyword mutation DTOs accept `category_id`
- keyword list/detail payloads return `category_id` and `category_name`
- validation rejects zero or missing `category_id`
- repository joins categories by `orgid`

Keep backward-compatible field removal minimal and explicit; do not leave ambiguous dual-write behavior in place.

**Step 4: Add real article-center APIs backed by `xt_article`**

Create DTOs and service methods for:

- `GET /api/v1/article-inspect/articles`
- `GET /api/v1/article-inspect/articles/{article_id}`

Requirements:

- article list returns real article rows scoped by `orgid`
- article list is rooted in `xt_article`, not `xt_article_inspect_results`
- article detail returns article body from `xt_article_info`
- article detail can include latest inspect summary fields when inspect rows exist

Recommended response shape for the list:

```go
type ArticleListItem struct {
    ID              uint64     `json:"id"`
    OrgID           uint64     `json:"orgid"`
    Title           string     `json:"title"`
    State           int8       `json:"state"`
    PublishAtTime   *time.Time `json:"publish_at_time"`
    LatestRiskLevel string     `json:"latest_risk_level,omitempty"`
    LatestTaskID    uint64     `json:"latest_task_id,omitempty"`
}
```

**Step 5: Run focused Go tests**

Run: `go test ./internal/api/register ./internal/modules/articleinspect`
Expected: PASS.

**Step 6: Run the full backend test suite**

Run: `go test ./...`
Expected: PASS.

**Step 7: Commit**

```bash
git add migrations/20260420_01_article_inspection.sql internal/modules/articleinspect/model.go internal/modules/articleinspect/dto_categories.go internal/modules/articleinspect/repository_categories.go internal/modules/articleinspect/service_categories.go internal/modules/articleinspect/dto_articles.go internal/modules/articleinspect/repository_articles.go internal/modules/articleinspect/service_articles.go internal/modules/articleinspect/dto_keywords.go internal/modules/articleinspect/repository_keywords.go internal/modules/articleinspect/service_keywords.go internal/modules/articleinspect/validator_keywords.go internal/modules/articleinspect/handler.go internal/api/register/router.go internal/api/register/router_test.go internal/modules/articleinspect/articleinspect_test.go
git commit -m "feat(api): add org/category master data and article center"
```

### Task 3: Backend article-center task folded into Task 2

During execution on `2026-04-23`, Task 1 contract tests made `/api/v1/article-inspect/articles` and `/api/v1/article-inspect/articles/{article_id}` a prerequisite for the Task 2 verification command. To keep the implementation order and verification steps coherent, the original Task 3 backend work is executed inside Task 2 and does not require a separate implementation pass.

### Task 4: Lock the new admin shell and organization-context expectations

**Files:**
- Modify: `web/admin/src/App.test.tsx`
- Create: `web/admin/src/services/orgs.ts`
- Create: `web/admin/src/services/categories.ts`
- Create: `web/admin/src/context/org-context.tsx`
- Modify: `web/admin/src/routes.tsx`
- Reference: `web/admin/src/components/layout/admin-shell.tsx`
- Reference: `web/admin/src/components/layout/sidebar-nav.tsx`

**Step 1: Write failing frontend tests for the new IA and shell**

Update `web/admin/src/App.test.tsx` so it asserts:

```tsx
expect(screen.getByRole('navigation', { name: /主导航/i })).toBeInTheDocument();
expect(screen.getByRole('link', { name: /规则中心/i })).toBeInTheDocument();
expect(screen.getByRole('link', { name: /检测任务/i })).toBeInTheDocument();
expect(screen.getByRole('link', { name: /文稿中心/i })).toBeInTheDocument();
expect(screen.getByRole('link', { name: /操作日志/i })).toBeInTheDocument();
expect(screen.queryByRole('link', { name: /风险结果/i })).not.toBeInTheDocument();
expect(screen.getByRole('button', { name: /一县一端/i })).toBeInTheDocument();
expect(screen.getByRole('button', { name: /当前用户|退出登录/i })).toBeInTheDocument();
```

Also add a route assertion that `/articles` no longer depends on the aggregated inspect-articles view.

**Step 2: Run the focused frontend test**

Run: `cd web/admin && npm test -- src/App.test.tsx`
Expected: FAIL because the shell still uses the old sidebar-only frame and routes still include the old results/article semantics.

**Step 3: Add the minimum frontend plumbing without updating page implementations yet**

Create:

- `web/admin/src/services/orgs.ts`
- `web/admin/src/services/categories.ts`
- `web/admin/src/context/org-context.tsx`

The org context should centralize:

- active org ID
- active org name
- available org list
- setter for selected org

Default should resolve to org `29` once orgs load.

**Step 4: Re-run the focused test before shell implementation**

Run: `cd web/admin && npm test -- src/App.test.tsx`
Expected: FAIL, still because layout and route structure are not implemented yet.

**Step 5: Commit**

```bash
git add web/admin/src/App.test.tsx web/admin/src/services/orgs.ts web/admin/src/services/categories.ts web/admin/src/context/org-context.tsx
git commit -m "test(admin): lock IA shell redesign expectations"
```

### Task 5: Rebuild the shell around a fixed header, collapsible sidebar, and dev proxy

**Files:**
- Modify: `web/admin/src/App.tsx`
- Modify: `web/admin/src/routes.tsx`
- Modify: `web/admin/src/components/layout/admin-shell.tsx`
- Modify: `web/admin/src/components/layout/sidebar-nav.tsx`
- Create: `web/admin/src/components/layout/header-bar.tsx`
- Create: `web/admin/src/components/layout/org-switcher.tsx`
- Create: `web/admin/src/components/layout/user-menu.tsx`
- Modify: `web/admin/src/styles/layout.css`
- Modify: `web/admin/vite.config.ts`
- Test: `web/admin/src/App.test.tsx`
- Test: `web/admin/src/styles/visual-tokens.test.ts`

**Step 1: Implement the shell frame and route regrouping**

Refactor routes so the top-level IA is:

- `rules/categories`
- `rules/keywords`
- `tasks`
- `tasks/new`
- `tasks/:taskId/results`
- `articles`
- `articles/:articleId`
- `articles/:articleId/rectify`
- `logs`

Use the shell to render:

- sidebar with grouped nav
- fixed header with sidebar trigger, page title, org switcher, and user menu
- inset content container

**Step 2: Add the Vite `/api` proxy**

Update `web/admin/vite.config.ts` so local frontend requests to `/api` proxy to the backend service, for example via an env-driven target like:

```ts
server: {
  host: '0.0.0.0',
  port: 5173,
  proxy: {
    '/api': {
      target: process.env.ADMIN_API_BASE_URL ?? 'http://127.0.0.1:8080',
      changeOrigin: true,
    },
  },
},
```

Keep the final exact target aligned with the existing local backend config.

**Step 3: Run focused frontend tests**

Run: `cd web/admin && npm test -- src/App.test.tsx src/styles/visual-tokens.test.ts`
Expected: PASS.

**Step 4: Run a frontend build smoke test**

Run: `cd web/admin && npm run build`
Expected: PASS.

**Step 5: Review the shell diff for accidental scope creep**

Run: `git diff -- web/admin/src/App.tsx web/admin/src/routes.tsx web/admin/src/components/layout/admin-shell.tsx web/admin/src/components/layout/sidebar-nav.tsx web/admin/src/components/layout/header-bar.tsx web/admin/src/components/layout/org-switcher.tsx web/admin/src/components/layout/user-menu.tsx web/admin/src/styles/layout.css web/admin/vite.config.ts`
Expected: only shell/proxy/IA changes.

**Step 6: Commit**

```bash
git add web/admin/src/App.tsx web/admin/src/routes.tsx web/admin/src/components/layout/admin-shell.tsx web/admin/src/components/layout/sidebar-nav.tsx web/admin/src/components/layout/header-bar.tsx web/admin/src/components/layout/org-switcher.tsx web/admin/src/components/layout/user-menu.tsx web/admin/src/styles/layout.css web/admin/vite.config.ts web/admin/src/App.test.tsx web/admin/src/styles/visual-tokens.test.ts
git commit -m "refactor(admin): rebuild shell and dev proxy"
```

### Task 6: Add category management UI and convert keyword rules to controlled category selection

**Files:**
- Create: `web/admin/src/pages/categories/index.tsx`
- Create: `web/admin/src/pages/categories/index.test.tsx`
- Modify: `web/admin/src/pages/keywords/index.tsx`
- Modify: `web/admin/src/pages/keywords/index.test.tsx`
- Modify: `web/admin/src/services/categories.ts`
- Modify: `web/admin/src/services/keywords.ts`
- Modify: `web/admin/src/routes.tsx`
- Reference: `web/admin/src/context/org-context.tsx`

**Step 1: Write failing tests for category management and keyword category selection**

Add category-page tests that assert:

- category list loads for current org
- create/edit modal works through the category service seam
- status toggle and delete actions are wired

Update keyword-page tests so they assert:

- rule form renders `规则分类` as a searchable select
- selected value posts `category_id`
- raw category text input no longer exists

**Step 2: Run the focused frontend tests**

Run: `cd web/admin && npm test -- src/pages/categories/index.test.tsx src/pages/keywords/index.test.tsx`
Expected: FAIL because the category page and controlled select behavior do not exist yet.

**Step 3: Implement the category page and keyword form changes**

For categories:

- list page with search, status filter, and create/edit modal
- organization shown from context, not manually entered

For keywords:

- fetch categories by active org
- use select options with `value=id`, `label=name`
- submit `category_id`
- display `category_name` in the table

**Step 4: Run focused frontend tests**

Run: `cd web/admin && npm test -- src/pages/categories/index.test.tsx src/pages/keywords/index.test.tsx`
Expected: PASS.

**Step 5: Run a broader admin regression slice**

Run: `cd web/admin && npm test -- src/App.test.tsx src/pages/categories/index.test.tsx src/pages/keywords/index.test.tsx`
Expected: PASS.

**Step 6: Commit**

```bash
git add web/admin/src/pages/categories/index.tsx web/admin/src/pages/categories/index.test.tsx web/admin/src/pages/keywords/index.tsx web/admin/src/pages/keywords/index.test.tsx web/admin/src/services/categories.ts web/admin/src/services/keywords.ts web/admin/src/routes.tsx
git commit -m "feat(admin): add category management and controlled keyword categories"
```

### Task 7: Simplify task creation and add the dedicated task-results page

**Files:**
- Create: `web/admin/src/pages/tasks/results.tsx`
- Create: `web/admin/src/pages/tasks/results.test.tsx`
- Modify: `web/admin/src/pages/tasks/index.tsx`
- Modify: `web/admin/src/pages/tasks/index.test.tsx`
- Modify: `web/admin/src/pages/tasks/new.tsx`
- Modify: `web/admin/src/pages/tasks/new.test.tsx`
- Modify: `web/admin/src/services/tasks.ts`
- Modify: `web/admin/src/services/results.ts`
- Modify: `web/admin/src/services/logs.ts`
- Modify: `web/admin/src/routes.tsx`
- Optionally modify: `web/admin/src/pages/tasks/detail.tsx`

**Step 1: Write failing tests for the new task flow**

Add coverage for:

- task list row shows `运行结果` and links to `/tasks/501/results`
- top-level nav no longer links to `/results`
- new task form no longer renders `文章编号` or `标题检索`
- new task form shows read-only org name and searchable multi-select rules
- task-results page renders task summary, result list, hit-article actions, bulk actions, and log section

**Step 2: Run the focused task tests to verify they fail**

Run: `cd web/admin && npm test -- src/pages/tasks/index.test.tsx src/pages/tasks/new.test.tsx src/pages/tasks/results.test.tsx`
Expected: FAIL because the results page route and simplified task form do not exist yet.

**Step 3: Implement the simplified task form and results workspace**

Frontend requirements:

- remove article ID and title filters from task creation
- always submit current `orgid`
- keep `keyword_ids`
- add `/tasks/:taskId/results` page that composes:
  - task detail
  - `listResults({ orgid, task_id })`
  - `listOperationLogs({ orgid, task_id })`
- expose actions for view detail, rectify, offline, ignore/mark processed, and batch operations

If existing service seams are insufficient, extend them minimally instead of inventing unrelated APIs.

**Step 4: Run focused task tests**

Run: `cd web/admin && npm test -- src/pages/tasks/index.test.tsx src/pages/tasks/new.test.tsx src/pages/tasks/results.test.tsx`
Expected: PASS.

**Step 5: Run a wider regression sweep**

Run: `cd web/admin && npm test -- src/App.test.tsx src/pages/tasks/index.test.tsx src/pages/tasks/new.test.tsx src/pages/tasks/results.test.tsx src/pages/logs/index.test.tsx`
Expected: PASS.

**Step 6: Commit**

```bash
git add web/admin/src/pages/tasks/results.tsx web/admin/src/pages/tasks/results.test.tsx web/admin/src/pages/tasks/index.tsx web/admin/src/pages/tasks/index.test.tsx web/admin/src/pages/tasks/new.tsx web/admin/src/pages/tasks/new.test.tsx web/admin/src/services/tasks.ts web/admin/src/services/results.ts web/admin/src/services/logs.ts web/admin/src/routes.tsx
git commit -m "feat(admin): add task results workspace"
```

### Task 8: Replace `/articles` with the real article center and align logs/actions

**Files:**
- Modify: `web/admin/src/services/articles.ts`
- Modify: `web/admin/src/services/articles.test.ts`
- Modify: `web/admin/src/pages/articles/index.tsx`
- Modify: `web/admin/src/pages/articles/index.test.tsx`
- Modify: `web/admin/src/pages/articles/detail.tsx`
- Modify: `web/admin/src/pages/articles/detail.test.tsx`
- Modify: `web/admin/src/pages/articles/rectify.tsx`
- Modify: `web/admin/src/pages/articles/rectify.test.tsx`
- Modify: `web/admin/src/pages/logs/index.tsx`
- Modify: `web/admin/src/pages/logs/index.test.tsx`
- Reference: `web/admin/src/services/logs.ts`
- Reference: `web/admin/src/services/results.ts`

**Step 1: Write failing article-center tests**

Update article service and page tests so they assert:

- article list uses real article payloads, not `summarizeArticles()` over results
- article detail renders real article body and latest inspect summary
- article detail still exposes rectify and lifecycle actions
- logs page links correctly into the real article center
- empty-state copy refers to real articles, not inspect aggregations

**Step 2: Run the focused article/log tests to verify they fail**

Run: `cd web/admin && npm test -- src/services/articles.test.ts src/pages/articles/index.test.tsx src/pages/articles/detail.test.tsx src/pages/articles/rectify.test.tsx src/pages/logs/index.test.tsx`
Expected: FAIL because `/articles` still uses the old aggregated inspect service shape.

**Step 3: Implement the real article-center service and pages**

Requirements:

- `services/articles.ts` calls the new backend article list/detail endpoints
- article list shows real article metadata plus latest inspect enrichments when available
- article detail shows article basics, body snapshot, inspect summary, hit records, and operation history
- rectify page keeps its existing lifecycle role but now sits under the corrected article-center semantics

Do not preserve the old `巡检稿件` page identity under `/articles`.

**Step 4: Run focused article/log tests**

Run: `cd web/admin && npm test -- src/services/articles.test.ts src/pages/articles/index.test.tsx src/pages/articles/detail.test.tsx src/pages/articles/rectify.test.tsx src/pages/logs/index.test.tsx`
Expected: PASS.

**Step 5: Run the full frontend test suite**

Run: `cd web/admin && npm test`
Expected: PASS.

**Step 6: Commit**

```bash
git add web/admin/src/services/articles.ts web/admin/src/services/articles.test.ts web/admin/src/pages/articles/index.tsx web/admin/src/pages/articles/index.test.tsx web/admin/src/pages/articles/detail.tsx web/admin/src/pages/articles/detail.test.tsx web/admin/src/pages/articles/rectify.tsx web/admin/src/pages/articles/rectify.test.tsx web/admin/src/pages/logs/index.tsx web/admin/src/pages/logs/index.test.tsx
git commit -m "refactor(admin): restore real article center"
```

### Task 9: Fix the local runtime workflow and run full verification

**Files:**
- Modify: `Makefile`
- Create: `scripts/dev.sh`
- Optionally create: `scripts/dev-admin.sh`
- Optionally create: `scripts/dev-worker.sh`
- Optionally create: `scripts/dev-scheduler.sh`
- Reference: `cmd/server/main.go`
- Reference: `cmd/worker/main.go`
- Reference: `cmd/scheduler/main.go`
- Reference: `web/admin/package.json`

**Step 1: Write a failing workflow check as a shell-script smoke target or documented assertion**

Add or update a lightweight smoke script expectation so local dev startup must cover:

- backend server
- worker
- scheduler
- admin dev server

At minimum, document and encode the intended `make` behavior in a script or smoke assertion rather than relying on manual memory.

**Step 2: Run the current dev target to observe the gap**

Run: `make -n dev`
Expected: shows only `go run ./cmd/server ...`, proving the worker/scheduler/admin processes are not yet included.

**Step 3: Implement the new dev orchestration**

Update `Makefile` so there is a clean daily-driver entry point, for example:

- `make dev` -> starts server + worker + scheduler + admin
- `make dev-api`
- `make dev-worker`
- `make dev-scheduler`
- `make dev-admin`

Use a script such as `scripts/dev.sh` if needed to keep the Makefile readable.

**Step 4: Run backend, frontend, and build verification**

Run: `go test ./...`
Expected: PASS.

Run: `cd web/admin && npm test`
Expected: PASS.

Run: `cd web/admin && npm run build`
Expected: PASS.

**Step 5: Run final smoke checks for routing and startup assumptions**

Run: `make -n dev`
Expected: prints the full multi-process startup flow.

Run: `git status --short`
Expected: only planned files are modified.

**Step 6: Commit**

```bash
git add Makefile scripts/dev.sh
git commit -m "chore(dev): start admin stack with queue services"
```
