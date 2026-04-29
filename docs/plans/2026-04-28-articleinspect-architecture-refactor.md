# Article Inspect Architecture Refactor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Refactor `internal/modules/articleinspect` into a feature-organized, Huma-idiomatic module with clearer route ownership, module-local assembly, and stronger OpenAPI contract accuracy while preserving current business behavior.

**Architecture:** Keep `internal/modules/articleinspect` as a single Go package, but split the monolithic `handler.go` into feature-scoped route files plus shared route helpers. Introduce a module-local route assembly constructor, use a base Huma group for `/api/v1/article-inspect`, preserve explicit `OperationID`s, and tighten OpenAPI/test coverage around success statuses and typed parameters.

**Tech Stack:** Go, Huma v2, net/http, GORM, SQLite-backed Go tests, OpenAPI JSON from Huma

---

### Task 1: Add design-guard tests for route contract drift

**Files:**
- Modify: `internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Write the failing test**

Expand `TestRouteRegistrationRegistersArticleInspectPaths` so it asserts more than path existence.

Add checks for:
- stable `operationId` on representative routes
- `201` success response docs for create endpoints
- typed parameter schema for representative fields like `id`, `orgid`, and `enabled`

```go
var doc struct {
	Paths map[string]map[string]struct {
		OperationID string                 `json:"operationId"`
		Parameters  []map[string]any       `json:"parameters"`
		Responses   map[string]map[string]any `json:"responses"`
	} `json:"paths"`
}

if got := doc.Paths["/api/v1/article-inspect/categories"]["post"].OperationID; got != "article-inspect-category-create" {
	t.Fatalf("category create operationId = %q", got)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/modules/articleinspect -run TestRouteRegistrationRegistersArticleInspectPaths -v`
Expected: FAIL because the current OpenAPI still documents create routes as `200` and some parameters are still string-shaped.

**Step 3: Write minimal implementation**

Do not implement behavior changes yet. Only add the failing assertions and any tiny test helpers needed to inspect the OpenAPI structure cleanly.

**Step 4: Run test to verify it still fails for the intended reasons**

Run: `go test ./internal/modules/articleinspect -run TestRouteRegistrationRegistersArticleInspectPaths -v`
Expected: FAIL with contract mismatches rather than JSON decode errors.

**Step 5: Commit**

```bash
git add internal/modules/articleinspect/articleinspect_test.go
git commit -m "test: lock article inspect route contract"
```

### Task 2: Introduce module-local route assembly and shared registration entrypoint

**Files:**
- Create: `internal/modules/articleinspect/module.go`
- Modify: `internal/api/register/router.go`
- Modify: `internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Write the failing test**

Add or update a test helper path so both runtime code and tests can build routes through the module rather than manually assembling every service inline.

```go
func TestNewRoutesBuildsModuleDependencies(t *testing.T) {
	db := newArticleInspectTestDB(t)
	routes := NewRoutes(db, &articleInspectTaskDispatcherStub{})
	if routes.Categories == nil || routes.Tasks == nil || routes.Dispatcher == nil {
		t.Fatalf("routes not fully wired: %#v", routes)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/modules/articleinspect -run TestNewRoutesBuildsModuleDependencies -v`
Expected: FAIL because `NewRoutes` does not exist yet.

**Step 3: Write minimal implementation**

Create `module.go` with a constructor that assembles repositories and services.

```go
func NewRoutes(db *gorm.DB, dispatcher TaskDispatcher) Routes {
	if db == nil {
		return Routes{}
	}
	categoryRepo := NewCategoryRepository(db)
	keywordRepo := NewKeywordRepository(db)
	articleRepo := NewArticleRepository(db)
	return Routes{
		Categories: NewCategoryService(categoryRepo),
		Keywords:   NewKeywordService(keywordRepo),
		Tasks:      NewTaskService(db, keywordRepo, articleRepo),
		Results:    NewResultService(db),
		Actions:    NewActionService(db, NewActionRepository(db)),
		Lifecycle:  NewLifecycleService(db),
		Logs:       NewLogService(db),
		Articles:   NewArticleService(articleRepo),
		Dispatcher: dispatcher,
	}
}
```

Update `internal/api/register/router.go` and local tests to call `NewRoutes(...)`.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/modules/articleinspect -run TestNewRoutesBuildsModuleDependencies -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/modules/articleinspect/module.go \
  internal/api/register/router.go \
  internal/modules/articleinspect/articleinspect_test.go
git commit -m "refactor: centralize article inspect module wiring"
```

### Task 3: Split route registration into feature files with a module base group

**Files:**
- Create: `internal/modules/articleinspect/routes.go`
- Create: `internal/modules/articleinspect/routes_common.go`
- Create: `internal/modules/articleinspect/category_routes.go`
- Create: `internal/modules/articleinspect/keyword_routes.go`
- Create: `internal/modules/articleinspect/task_routes.go`
- Create: `internal/modules/articleinspect/result_routes.go`
- Create: `internal/modules/articleinspect/action_routes.go`
- Create: `internal/modules/articleinspect/lifecycle_routes.go`
- Create: `internal/modules/articleinspect/log_routes.go`
- Create: `internal/modules/articleinspect/article_routes.go`
- Delete or shrink: `internal/modules/articleinspect/handler.go`

**Step 1: Write the failing test**

Do not add new behavior assertions. Reuse the existing route behavior tests plus the contract test from Task 1 as the safety net for the file split.

If needed, add one regression assertion that `RegisterRoutes` still exposes the full `/api/v1/article-inspect/...` paths after group introduction.

```go
if _, ok := doc.Paths["/api/v1/article-inspect/keywords"]; !ok {
	t.Fatal("keyword path missing after group split")
}
```

**Step 2: Run test to verify the current baseline**

Run: `go test ./internal/modules/articleinspect -run 'TestRouteRegistrationRegistersArticleInspectPaths|TestHandlerKeywordTaskAndResultsRoutes|TestHandlerOrgCategoryAndArticleCenterContracts' -v`
Expected: current baseline passes except for the Task 1 contract assertions that intentionally fail.

**Step 3: Write minimal implementation**

Move route code out of `handler.go` into the new files.

Implementation rules:
- `routes.go` owns `RegisterRoutes`
- create `inspect := huma.NewGroup(api, "/api/v1/article-inspect")`
- feature files register relative paths like `"/categories"`, `"/tasks"`, `"/articles/{article_id}"`
- keep all current explicit `OperationID`s unchanged
- keep request structs next to the feature registrar that uses them
- move shared envelope/status helpers into `routes_common.go`

```go
func RegisterRoutes(api huma.API, routes Routes) {
	if api == nil {
		return
	}
	inspect := huma.NewGroup(api, "/api/v1/article-inspect")
	if routes.Categories != nil {
		registerCategoryRoutes(inspect, routes.Categories)
	}
	// ...
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/modules/articleinspect -run 'TestHandlerKeywordTaskAndResultsRoutes|TestHandlerOrgCategoryAndArticleCenterContracts' -v`
Expected: PASS with unchanged runtime behavior.

**Step 5: Commit**

```bash
git add internal/modules/articleinspect/routes.go \
  internal/modules/articleinspect/routes_common.go \
  internal/modules/articleinspect/category_routes.go \
  internal/modules/articleinspect/keyword_routes.go \
  internal/modules/articleinspect/task_routes.go \
  internal/modules/articleinspect/result_routes.go \
  internal/modules/articleinspect/action_routes.go \
  internal/modules/articleinspect/lifecycle_routes.go \
  internal/modules/articleinspect/log_routes.go \
  internal/modules/articleinspect/article_routes.go \
  internal/modules/articleinspect/handler.go
git commit -m "refactor: split article inspect route registration"
```

### Task 4: Replace string parsing with Huma-typed inputs where low-risk

**Files:**
- Modify: `internal/modules/articleinspect/routes_common.go`
- Modify: `internal/modules/articleinspect/category_routes.go`
- Modify: `internal/modules/articleinspect/keyword_routes.go`
- Modify: `internal/modules/articleinspect/task_routes.go`
- Modify: `internal/modules/articleinspect/result_routes.go`
- Modify: `internal/modules/articleinspect/lifecycle_routes.go`
- Modify: `internal/modules/articleinspect/log_routes.go`
- Modify: `internal/modules/articleinspect/article_routes.go`
- Modify: `internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Write the failing test**

Use the OpenAPI assertions from Task 1 to require typed parameter schemas for representative fields.

Representative target shapes:
- path IDs use `uint64`
- optional `enabled` uses `*bool`
- optional state uses `*int8`
- time filters use `*time.Time` where supported by current clients

```go
if got := parameterSchemaType(t, doc, "/api/v1/article-inspect/categories", "get", "orgid"); got != "integer" {
	t.Fatalf("orgid schema type = %q, want integer", got)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/modules/articleinspect -run TestRouteRegistrationRegistersArticleInspectPaths -v`
Expected: FAIL because the current request fields still advertise strings for several parameters.

**Step 3: Write minimal implementation**

Convert route input structs away from string parsing where safe.

Examples:

```go
type keywordDetailRequest struct {
	ID    uint64 `path:"id"`
	OrgID uint64 `query:"orgid"`
}

type categoryQueryRequest struct {
	OrgID    uint64 `query:"orgid"`
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
	Query    string `query:"name"`
	Enabled  *bool  `query:"enabled"`
}
```

Delete obsolete parse helpers that are no longer needed, but keep any remaining helpers for values that Huma cannot replace cleanly in this pass.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/modules/articleinspect -run 'TestRouteRegistrationRegistersArticleInspectPaths|TestHandlerOrgCategoryAndArticleCenterContracts' -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/modules/articleinspect/routes_common.go \
  internal/modules/articleinspect/category_routes.go \
  internal/modules/articleinspect/keyword_routes.go \
  internal/modules/articleinspect/task_routes.go \
  internal/modules/articleinspect/result_routes.go \
  internal/modules/articleinspect/lifecycle_routes.go \
  internal/modules/articleinspect/log_routes.go \
  internal/modules/articleinspect/article_routes.go \
  internal/modules/articleinspect/articleinspect_test.go
git commit -m "refactor: use typed huma inputs for article inspect routes"
```

### Task 5: Correct documented success statuses and shared route outputs

**Files:**
- Modify: `internal/modules/articleinspect/routes_common.go`
- Modify: `internal/modules/articleinspect/category_routes.go`
- Modify: `internal/modules/articleinspect/keyword_routes.go`
- Modify: `internal/modules/articleinspect/task_routes.go`
- Modify: `internal/modules/articleinspect/action_routes.go`
- Modify: `internal/modules/articleinspect/lifecycle_routes.go`
- Modify: `internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Write the failing test**

Require create endpoints to document `201` instead of only `200`.

Representative routes:
- `POST /api/v1/article-inspect/categories`
- `POST /api/v1/article-inspect/keywords`
- `POST /api/v1/article-inspect/tasks`

```go
if _, ok := doc.Paths["/api/v1/article-inspect/tasks"]["post"].Responses["201"]; !ok {
	t.Fatal("task create must document 201 response")
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/modules/articleinspect -run TestRouteRegistrationRegistersArticleInspectPaths -v`
Expected: FAIL because the shared output still documents `200`.

**Step 3: Write minimal implementation**

Introduce status-specific success outputs such as `okEnvelopeOutput` and `createdEnvelopeOutput`, and use them on the relevant route handlers while preserving the same runtime envelope body.

```go
type okEnvelopeOutput struct {
	Status int `status:"200"`
	Body   response.Envelope
}

type createdEnvelopeOutput struct {
	Status int `status:"201"`
	Body   response.Envelope
}
```

Use the `createdEnvelopeOutput` return type on create handlers and keep the response body unchanged.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/modules/articleinspect -run 'TestRouteRegistrationRegistersArticleInspectPaths|TestHandlerKeywordTaskAndResultsRoutes' -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/modules/articleinspect/routes_common.go \
  internal/modules/articleinspect/category_routes.go \
  internal/modules/articleinspect/keyword_routes.go \
  internal/modules/articleinspect/task_routes.go \
  internal/modules/articleinspect/action_routes.go \
  internal/modules/articleinspect/lifecycle_routes.go \
  internal/modules/articleinspect/articleinspect_test.go
git commit -m "refactor: correct article inspect success response docs"
```

### Task 6: Normalize route-local type ownership and remove now-dead transport helpers

**Files:**
- Modify: `internal/modules/articleinspect/dto_categories.go`
- Modify: `internal/modules/articleinspect/dto_keywords.go`
- Modify: `internal/modules/articleinspect/dto_tasks.go`
- Modify: `internal/modules/articleinspect/dto_articles.go`
- Modify: `internal/modules/articleinspect/repository_results.go`
- Modify: `internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Write the failing test**

Add a small internal-structure regression test only if needed, otherwise rely on compilation plus existing route and service tests.

Potential assertion: route files compile without referencing removed transport structs from the old `handler.go` layout.

**Step 2: Run test to verify the baseline**

Run: `go test ./internal/modules/articleinspect -run 'TestKeywordService|TestHandlerOrgCategoryAndArticleCenterContracts' -v`
Expected: PASS or baseline pass after earlier tasks.

**Step 3: Write minimal implementation**

Clean up remaining ownership drift.

Rules:
- delete unused route-only structs from shared DTO files
- keep service input/output structs in DTO files
- keep repository-only query structs in repo files unless a cleaner local type split is required
- remove parsing helpers that are now dead code

**Step 4: Run test to verify it passes**

Run: `go test ./internal/modules/articleinspect -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/modules/articleinspect/dto_categories.go \
  internal/modules/articleinspect/dto_keywords.go \
  internal/modules/articleinspect/dto_tasks.go \
  internal/modules/articleinspect/dto_articles.go \
  internal/modules/articleinspect/repository_results.go \
  internal/modules/articleinspect/articleinspect_test.go
git commit -m "refactor: clarify article inspect type ownership"
```

### Task 7: Final verification and diff review

**Files:**
- Review: `internal/modules/articleinspect/*.go`
- Review: `internal/api/register/router.go`
- Review: `docs/plans/2026-04-28-articleinspect-architecture-refactor-design.md`
- Review: `docs/plans/2026-04-28-articleinspect-architecture-refactor.md`

**Step 1: Run focused module verification**

Run: `go test ./internal/modules/articleinspect -v`
Expected: PASS.

**Step 2: Run broader API registration verification**

Run: `go test ./internal/api/... ./internal/modules/... -run 'TestRouteRegistrationRegistersArticleInspectPaths|TestHandlerKeywordTaskAndResultsRoutes|TestHandlerOrgCategoryAndArticleCenterContracts' -v`
Expected: PASS.

**Step 3: Inspect the final diff**

Run: `git diff --stat`
Expected: route split, module wiring, contract test, and route-type cleanup only.

**Step 4: Sanity-check for old file leftovers**

Run: `rg -n "parseUint64ID|parseOptionalBool|parseOptionalInt8|parseOptionalTime" internal/modules/articleinspect`
Expected: no dead helper references remain except any intentionally retained helper still in active use.

**Step 5: Commit**

```bash
git add internal/modules/articleinspect \
  internal/api/register/router.go \
  docs/plans/2026-04-28-articleinspect-architecture-refactor-design.md \
  docs/plans/2026-04-28-articleinspect-architecture-refactor.md
git commit -m "refactor: reorganize article inspect module"
```
