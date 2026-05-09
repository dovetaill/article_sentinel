# ArticleInspect Thin Root Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Collapse `internal/modules/articleinspect/` into a thin module-entry package while preserving API contracts, DB schema behavior, queue behavior, OpenAPI output, and existing article inspection semantics.

**Architecture:** Execute the cleanup in a dedicated worktree. First move reusable test support and feature tests toward their owning packages, then consolidate canonical ownership in `articles`, `worker`, and `outbox`, then rewrite the root package to be only `module.go` plus `routes.go`, and only then delete wrappers and aliases. Treat the root package as a composition boundary, not a business package.

**Tech Stack:** Go, Huma v2, GORM, Asynq, SQLite-backed Go tests, OpenAPI contract tests, git worktrees

---

### Task 1: Create the isolated worktree and capture a clean baseline

**Files:**
- Create worktree: `.worktrees/articleinspect-thin-root`
- Check: `docs/plans/2026-05-08-articleinspect-thin-root-design.md`
- Check: `internal/modules/articleinspect/module.go`
- Check: `internal/modules/articleinspect/routes.go`

**Step 1: Create the dedicated worktree**

Use `@superpowers:using-git-worktrees`.

Run:

```bash
git worktree add .worktrees/articleinspect-thin-root -b articleinspect-thin-root
```

Expected: a new worktree exists on branch `articleinspect-thin-root`

**Step 2: Verify the refactor baseline before touching code**

Run:

```bash
go test ./internal/modules/articleinspect/...
go test ./internal/api/register
go test ./internal/queue/asynq
```

Expected: PASS for all three commands

**Step 3: Capture the current root-package sprawl as the success baseline**

Run:

```bash
find internal/modules/articleinspect -maxdepth 1 -type f -name '*.go' | sort
find internal/modules/articleinspect -maxdepth 1 -type f -name '*.go' | wc -l
```

Expected: the list shows the current root Go files and the count is much larger than the final target

**Step 4: Review the approved design before implementation**

Read:

- `docs/plans/2026-05-08-articleinspect-thin-root-design.md`

Expected: implementation starts from the approved thin-root target instead of improvising a new structure

**Step 5: Commit only if a worktree setup note was intentionally added**

Run:

```bash
git status --short
```

Expected: clean worktree, or only an intentional setup note; do not create a meaningless baseline commit

### Task 2: Create shared test support and move runtime tests to owning packages

**Files:**
- Create: `internal/modules/articleinspect/testutil/testdb.go`
- Create: `internal/modules/articleinspect/testutil/seeds.go`
- Create: `internal/modules/articleinspect/testutil/http.go`
- Create: `internal/modules/articleinspect/testutil/time.go`
- Create: `internal/modules/articleinspect/scan/scanner_test.go`
- Create: `internal/modules/articleinspect/lifecycle/diff_test.go`
- Create: `internal/modules/articleinspect/worker/executor_test.go`
- Modify: `internal/modules/articleinspect/fixtures_test.go`
- Modify: `internal/modules/articleinspect/scanner_test.go`
- Modify: `internal/modules/articleinspect/worker_test.go`

**Step 1: Write the failing moved tests first**

Use `@superpowers:test-driven-development`.

Create the new package-local tests with imports from `scan`, `lifecycle`, `worker`, and `testutil`.

Example snippets:

```go
func TestKeywordScanner(t *testing.T) {
	scanner := NewKeywordScanner()
	article := CandidateArticle{Title: "Breaking spam headline"}
	rules := []KeywordRule{{Name: "spam", MatchType: domainpkg.MatchTypeContains, Scopes: []string{domainpkg.KeywordScopeTitle}}}
	hits, err := scanner.ScanArticle(context.Background(), article, rules)
	if err != nil || len(hits) != 1 {
		t.Fatalf("ScanArticle() = %v, %v; want 1 hit", hits, err)
	}
}
```

```go
func TestFieldDiff(t *testing.T) {
	before := EditableArticleFields{Title: "Old"}
	after := EditableArticleFields{Title: "New"}
	if len(DiffEditableFields(before, after)) != 1 {
		t.Fatal("DiffEditableFields() should report one change")
	}
}
```

```go
func TestExecutorExecuteTask(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedCandidateArticles(t, db)
	executor := NewWorker(db)
	if err := executor.ExecuteTask(context.Background(), queuetasks.ArticleInspectTaskPayload{TaskID: 1, OrgID: 100}); err != nil {
		t.Fatalf("ExecuteTask() error = %v", err)
	}
}
```

**Step 2: Run the moved-package tests and verify they fail**

Run:

```bash
go test ./internal/modules/articleinspect/scan -run TestKeywordScanner -v
go test ./internal/modules/articleinspect/lifecycle -run TestFieldDiff -v
go test ./internal/modules/articleinspect/worker -run TestExecutorExecuteTask -v
```

Expected: FAIL because `testutil` does not exist yet and the moved tests are not wired up

**Step 3: Implement `testutil` and move the helpers with minimal behavior change**

Create focused helpers instead of a new giant fixture file.

Suggested split:

```go
// testutil/testdb.go
func NewArticleInspectTestDB(t *testing.T) *gorm.DB

// testutil/seeds.go
func SeedCandidateArticles(t *testing.T, db *gorm.DB)
func SeedInspectionTaskForWorker(t *testing.T, db *gorm.DB, rules []scanpkg.KeywordRule) domainpkg.InspectionTask

// testutil/time.go
func MustTime(t *testing.T, value string) time.Time
```

Then move the scanner, diff, and worker tests into their owning packages and either delete or shrink the root test files.

**Step 4: Run focused regressions and then the whole module package tree**

Run:

```bash
go test ./internal/modules/articleinspect/scan -v
go test ./internal/modules/articleinspect/lifecycle -v
go test ./internal/modules/articleinspect/worker -v
go test ./internal/modules/articleinspect/... 
```

Expected: PASS, with runtime tests now living beside runtime code

**Step 5: Commit**

```bash
git add internal/modules/articleinspect/testutil \
  internal/modules/articleinspect/scan/scanner_test.go \
  internal/modules/articleinspect/lifecycle/diff_test.go \
  internal/modules/articleinspect/worker/executor_test.go \
  internal/modules/articleinspect/fixtures_test.go \
  internal/modules/articleinspect/scanner_test.go \
  internal/modules/articleinspect/worker_test.go
git commit -m "test(articleinspect): move runtime tests into owning packages"
```

### Task 3: Move workflow and HTTP contract tests to feature packages

**Files:**
- Create: `internal/modules/articleinspect/rules/category_routes_test.go`
- Create: `internal/modules/articleinspect/rules/keyword_routes_test.go`
- Create: `internal/modules/articleinspect/tasks/service_test.go`
- Create: `internal/modules/articleinspect/outbox/relay_test.go`
- Create: `internal/modules/articleinspect/lifecycle/service_test.go`
- Create: `internal/modules/articleinspect/articles/routes_test.go`
- Create: `internal/modules/articleinspect/results/routes_test.go`
- Create: `internal/modules/articleinspect/audit/routes_test.go`
- Modify: `internal/modules/articleinspect/http_routes_test.go`
- Modify: `internal/modules/articleinspect/task_service_test.go`
- Modify: `internal/modules/articleinspect/task_outbox_test.go`
- Modify: `internal/modules/articleinspect/lifecycle_test.go`

**Step 1: Write the failing feature-local tests first**

Use `@superpowers:test-driven-development`.

Split the root HTTP and workflow coverage into package-local tests that import `testutil`.

Example snippets:

```go
func TestTaskCreation(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedOrgCategoryFixtures(t, db)
	service := NewTaskService(db, rulespkg.NewKeywordRepository(db), articlespkg.NewArticleRepository(db))
	_, err := service.Create(context.Background(), CreateInspectionTaskInput{OrgID: 100, KeywordIDs: []uint64{1}})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}
```

```go
func TestTaskOutboxRelayDispatchesPendingMessage(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	relay := NewTaskOutboxRelay(db, dispatcherStub{}, nil)
	if _, err := relay.DispatchPending(context.Background(), 10); err != nil {
		t.Fatalf("DispatchPending() error = %v", err)
	}
}
```

**Step 2: Run the moved tests and verify they fail before the helpers/imports are complete**

Run:

```bash
go test ./internal/modules/articleinspect/tasks -run TestTaskCreation -v
go test ./internal/modules/articleinspect/outbox -run TestTaskOutboxRelayDispatchesPendingMessage -v
go test ./internal/modules/articleinspect/rules -run 'Test(Category|Keyword)' -v
```

Expected: FAIL until the tests stop depending on root aliases and root-only fixtures

**Step 3: Finish the test move and keep the root package focused on module-level coverage**

After the moved tests pass, reduce the root package to:

- `openapi_test.go`
- any small module wiring test that still makes sense at the root

Delete or shrink the old root files so they no longer hold feature-specific test logic.

**Step 4: Run all articleinspect tests plus router contract checks**

Run:

```bash
go test ./internal/modules/articleinspect/...
go test ./internal/api/register -run TestRouterRegistersArticleInspectRoutes -v
```

Expected: PASS, with feature and workflow tests now living beside the feature packages

**Step 5: Commit**

```bash
git add internal/modules/articleinspect/rules/*_test.go \
  internal/modules/articleinspect/tasks/service_test.go \
  internal/modules/articleinspect/outbox/relay_test.go \
  internal/modules/articleinspect/lifecycle/service_test.go \
  internal/modules/articleinspect/articles/routes_test.go \
  internal/modules/articleinspect/results/routes_test.go \
  internal/modules/articleinspect/audit/routes_test.go \
  internal/modules/articleinspect/http_routes_test.go \
  internal/modules/articleinspect/task_service_test.go \
  internal/modules/articleinspect/task_outbox_test.go \
  internal/modules/articleinspect/lifecycle_test.go
git commit -m "test(articleinspect): move feature tests into feature packages"
```

### Task 4: Consolidate canonical ownership in `articles`, `tasks`, `worker`, and `outbox`

**Files:**
- Create: `internal/modules/articleinspect/articles/repository_candidates.go`
- Create: `internal/modules/articleinspect/articles/repository_candidates_test.go`
- Modify: `internal/modules/articleinspect/articles/repository.go`
- Modify: `internal/modules/articleinspect/tasks/service.go`
- Modify: `internal/modules/articleinspect/tasks/routes.go`
- Modify: `internal/modules/articleinspect/worker/executor.go`
- Modify: `internal/modules/articleinspect/module.go`
- Modify: `internal/modules/articleinspect/outbox/relay.go`
- Modify: `internal/modules/articleinspect/outbox/outbox.go`
- Modify: `internal/modules/articleinspect/outbox/relay_test.go`
- Modify: `internal/modules/articleinspect/tasks/service_test.go`
- Modify: `internal/modules/articleinspect/worker/executor_test.go`

**Step 1: Write the failing canonical-ownership tests first**

Add one focused repository test and tighten the worker/task tests to use the canonical packages directly.

Example snippets:

```go
func TestArticleRepositoryListCandidateArticles(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedCandidateArticles(t, db)
	repo := NewArticleRepository(db)
	items, nextID, err := repo.ListCandidateArticles(context.Background(), taskspkg.CandidateArticleFilter{OrgID: 100, Limit: 10})
	if err != nil || len(items) == 0 || nextID == 0 {
		t.Fatalf("ListCandidateArticles() = %v, %d, %v", items, nextID, err)
	}
}
```

```go
func TestRegisterTaskRoutesUsesOutboxDispatcher(t *testing.T) {
	var _ outboxpkg.TaskDispatcher = dispatcherStub{}
}
```

**Step 2: Run the focused tests and verify they fail**

Run:

```bash
go test ./internal/modules/articleinspect/articles -run TestArticleRepositoryListCandidateArticles -v
go test ./internal/modules/articleinspect/tasks -run TestRegisterTaskRoutesUsesOutboxDispatcher -v
go test ./internal/modules/articleinspect/worker -run TestExecutorExecuteTask -v
```

Expected: FAIL because candidate-article loading still lives in the root package and task-route dispatcher ownership is still duplicated

**Step 3: Move the canonical implementations with the smallest possible code changes**

Implement these ownership rules:

```go
// articles/repository_candidates.go
func (r *ArticleRepository) ListCandidateArticles(ctx context.Context, filter taskspkg.CandidateArticleFilter) ([]scanpkg.CandidateArticle, uint64, error)
```

```go
// tasks/routes.go
func RegisterTaskRoutes(api huma.API, service *TaskService, dispatcher outboxpkg.TaskDispatcher, logger *slog.Logger, outboxSettings outboxpkg.TaskOutboxSettings)
```

```go
// worker/executor.go
func NewWorker(db *gorm.DB) *Executor {
	return NewExecutorWithDeps(db, scanpkg.NewKeywordScanner(), articlespkg.NewArticleRepository(db), 100)
}
```

Keep behavior identical; this task is ownership cleanup, not feature redesign.

**Step 4: Run focused tests and then the full module tests**

Run:

```bash
go test ./internal/modules/articleinspect/articles -v
go test ./internal/modules/articleinspect/tasks -v
go test ./internal/modules/articleinspect/worker -v
go test ./internal/modules/articleinspect/...
```

Expected: PASS, with `articles`, `worker`, and `outbox` now serving as the only real owners

**Step 5: Commit**

```bash
git add internal/modules/articleinspect/articles/repository*.go \
  internal/modules/articleinspect/articles/repository_candidates_test.go \
  internal/modules/articleinspect/tasks/service.go \
  internal/modules/articleinspect/tasks/routes.go \
  internal/modules/articleinspect/worker/executor.go \
  internal/modules/articleinspect/outbox/*.go \
  internal/modules/articleinspect/module.go
git commit -m "refactor(articleinspect): consolidate canonical package ownership"
```

### Task 5: Thin the root package, switch external wiring, and delete wrappers

**Files:**
- Modify: `internal/modules/articleinspect/module.go`
- Modify: `internal/modules/articleinspect/routes.go`
- Modify: `internal/modules/articleinspect/openapi_test.go`
- Modify: `internal/api/register/router.go`
- Modify: `internal/app/bootstrap/schema.go`
- Modify: `internal/queue/asynq/handlers.go`
- Modify: `internal/queue/asynq/articleinspect_dispatcher.go`
- Modify: `cmd/scheduler/main.go`
- Create: `internal/modules/articleinspect/structure_test.go`
- Delete: `internal/modules/articleinspect/action_routes.go`
- Delete: `internal/modules/articleinspect/action_service.go`
- Delete: `internal/modules/articleinspect/article_routes.go`
- Delete: `internal/modules/articleinspect/category_routes.go`
- Delete: `internal/modules/articleinspect/constants.go`
- Delete: `internal/modules/articleinspect/diff.go`
- Delete: `internal/modules/articleinspect/dto_articles.go`
- Delete: `internal/modules/articleinspect/dto_categories.go`
- Delete: `internal/modules/articleinspect/dto_keywords.go`
- Delete: `internal/modules/articleinspect/dto_results.go`
- Delete: `internal/modules/articleinspect/dto_tasks.go`
- Delete: `internal/modules/articleinspect/keyword_routes.go`
- Delete: `internal/modules/articleinspect/lifecycle_routes.go`
- Delete: `internal/modules/articleinspect/lifecycle_service.go`
- Delete: `internal/modules/articleinspect/log_routes.go`
- Delete: `internal/modules/articleinspect/model.go`
- Delete: `internal/modules/articleinspect/model_article_source.go`
- Delete: `internal/modules/articleinspect/model_inspection.go`
- Delete: `internal/modules/articleinspect/operator.go`
- Delete: `internal/modules/articleinspect/operator_audit.go`
- Delete: `internal/modules/articleinspect/paging.go`
- Delete: `internal/modules/articleinspect/repository_actions.go`
- Delete: `internal/modules/articleinspect/repository_article_candidates.go`
- Delete: `internal/modules/articleinspect/repository_articles.go`
- Delete: `internal/modules/articleinspect/repository_categories.go`
- Delete: `internal/modules/articleinspect/repository_keywords.go`
- Delete: `internal/modules/articleinspect/repository_results.go`
- Delete: `internal/modules/articleinspect/request_scope.go`
- Delete: `internal/modules/articleinspect/result_routes.go`
- Delete: `internal/modules/articleinspect/routes_common.go`
- Delete: `internal/modules/articleinspect/scanner.go`
- Delete: `internal/modules/articleinspect/service_articles.go`
- Delete: `internal/modules/articleinspect/service_categories.go`
- Delete: `internal/modules/articleinspect/service_keywords.go`
- Delete: `internal/modules/articleinspect/service_logs.go`
- Delete: `internal/modules/articleinspect/service_results.go`
- Delete: `internal/modules/articleinspect/service_tasks.go`
- Delete: `internal/modules/articleinspect/subpackage_compat_test.go`
- Delete: `internal/modules/articleinspect/task_outbox.go`
- Delete: `internal/modules/articleinspect/task_outbox_relay.go`
- Delete: `internal/modules/articleinspect/task_outbox_settings.go`
- Delete: `internal/modules/articleinspect/task_routes.go`
- Delete: `internal/modules/articleinspect/transport_envelope.go`
- Delete: `internal/modules/articleinspect/transport_errors.go`
- Delete: `internal/modules/articleinspect/transport_params.go`
- Delete: `internal/modules/articleinspect/util_ids.go`
- Delete: `internal/modules/articleinspect/worker.go`
- Delete: `internal/modules/articleinspect/worker_rules.go`

**Step 1: Write the failing root-structure guard first**

Use `@superpowers:test-driven-development`.

Create a test that encodes the target root shape.

Example snippet:

```go
func TestArticleInspectRootStaysThin(t *testing.T) {
	entries, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}
	want := map[string]struct{}{
		"module.go": {},
		"routes.go": {},
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry, "_test.go") {
			continue
		}
		if _, ok := want[filepath.Base(entry)]; !ok {
			t.Fatalf("unexpected root go file: %s", entry)
		}
	}
}
```

**Step 2: Run the new guard and verify it fails**

Run:

```bash
go test ./internal/modules/articleinspect -run TestArticleInspectRootStaysThin -v
```

Expected: FAIL because the root package still contains many non-test `.go` files

**Step 3: Rewrite root wiring and external imports, then delete the wrappers**

The end state should look like this:

```go
type Routes struct {
	Categories *rulespkg.CategoryService
	Keywords   *rulespkg.KeywordService
	Tasks      *taskspkg.TaskService
	Results    *resultspkg.ResultService
	Actions    *actionspkg.ActionService
	Lifecycle  *lifecyclepkg.LifecycleService
	Logs       *auditpkg.LogService
	Articles   *articlespkg.ArticleService
	Dispatcher outboxpkg.TaskDispatcher
	Logger     *slog.Logger
	Outbox     outboxpkg.TaskOutboxSettings
}
```

```go
func init() {
	RegisterBusinessModels(
		postmodule.Post{},
		domainpkg.InspectionKeyword{},
		domainpkg.InspectionKeywordScope{},
		domainpkg.InspectionTask{},
	)
}
```

```go
var newArticleInspectExecutorFn = func(rt *bootstrap.Runtime) articleInspectExecutor {
	if rt == nil || rt.Resources == nil || rt.Resources.DB == nil {
		return nil
	}
	return workerpkg.NewWorker(rt.Resources.DB)
}
```

Delete the root wrappers only after the imports and constructors have been switched.

**Step 4: Run targeted regressions, then full verification**

Run:

```bash
go test ./internal/modules/articleinspect/...
go test ./internal/api/register -run TestRouterRegistersArticleInspectRoutes -v
go test ./internal/queue/asynq -v
go test ./internal/app/bootstrap -v
```

Expected: PASS, with the root package now reduced to module assembly and route registration

**Step 5: Commit**

```bash
git add internal/modules/articleinspect/module.go \
  internal/modules/articleinspect/routes.go \
  internal/modules/articleinspect/structure_test.go \
  internal/api/register/router.go \
  internal/app/bootstrap/schema.go \
  internal/queue/asynq/handlers.go \
  internal/queue/asynq/articleinspect_dispatcher.go \
  cmd/scheduler/main.go
git add -u internal/modules/articleinspect
git commit -m "refactor(articleinspect): collapse module root to thin entrypoints"
```

### Task 6: Refresh maintainer docs and run final verification

**Files:**
- Modify: `README.md`
- Modify: `docs/README.md`
- Modify: `docs/maintainer-development-flow.md`
- Check: `docs/plans/2026-05-08-articleinspect-thin-root-design.md`
- Check: `docs/plans/2026-05-08-articleinspect-thin-root.md`

**Step 1: Search for stale guidance that still treats the root package as the implementation package**

Run:

```bash
rg -n "internal/modules/articleinspect/" README.md docs/README.md docs/maintainer-development-flow.md
```

Expected: several references that need to be rewritten to describe the new thin-root structure accurately

**Step 2: Update docs to point maintainers at the owning packages**

Preferred wording patterns:

```md
- `internal/modules/articleinspect/module.go` and `internal/modules/articleinspect/routes.go` are the module entrypoints.
- Feature behavior lives under `internal/modules/articleinspect/{rules,tasks,results,actions,lifecycle,articles,audit}`.
- Runtime behavior lives under `internal/modules/articleinspect/{scan,worker,outbox}`.
```

**Step 3: Run the full verification suite from the worktree**

Use `@superpowers:verification-before-completion`.

Run:

```bash
gofmt -w $(find internal/modules/articleinspect internal/api/register internal/app/bootstrap internal/queue/asynq cmd/scheduler -name '*.go')
go test ./internal/modules/articleinspect/...
go test ./internal/api/register
go test ./internal/queue/asynq
go test ./internal/app/bootstrap
```

Expected: PASS for format-sensitive code and all targeted package suites

**Step 4: Verify the root package is actually thin**

Run:

```bash
find internal/modules/articleinspect -maxdepth 1 -type f -name '*.go' | sort
find internal/modules/articleinspect -maxdepth 1 -type f -name '*.go' | wc -l
```

Expected: only `module.go`, `routes.go`, and root test files remain; the non-test root file count is 2

**Step 5: Commit**

```bash
git add README.md docs/README.md docs/maintainer-development-flow.md
git commit -m "docs(articleinspect): document thin-root package ownership"
```
