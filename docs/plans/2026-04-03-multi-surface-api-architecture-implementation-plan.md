# Multi-Surface API Architecture Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Refactor article-sentinel from a single mixed API surface into clear admin, public, and member-oriented surfaces with stable route grouping and maintainable module boundaries.

**Architecture:** Keep the existing `handler -> service -> repository` direction, but change API composition so `internal/api/register/router.go` only performs wiring and Huma group registration. Split mixed modules such as `article` and `category` by surface, introduce a dedicated member identity domain, and place content interactions in a separate engagement domain.

**Tech Stack:** Go 1.25+, Huma v2, net/http, GORM, existing middleware chain, existing response envelope helpers, Huma `Group` middleware, `humatest` or current Huma test harness style.

---

### Task 1: Lock in surface boundaries with router-level tests

**Files:**
- Modify: `internal/api/register/router_test.go`
- Reference: `internal/api/register/router.go`
- Reference: `internal/modules/article/article_test.go`

**Step 1: Write the failing tests**

Add tests that assert the desired boundary behavior:

```go
func TestPublicArticleRoutesAreAccessibleWithoutAuth(t *testing.T)
func TestMemberRoutesRequireMemberAuth(t *testing.T)
func TestAdminRoutesRejectMemberPrincipal(t *testing.T)
```

Cover these cases:

- `GET /api/v1/articles` returns non-`401`
- `GET /api/v1/me` returns `401` without member auth
- `GET /api/v1/admin/articles` returns `401` without auth
- `GET /api/v1/admin/articles` returns `403` for non-admin auth

**Step 2: Run the targeted test command and confirm failure**

Run: `go test ./internal/api/register -run 'Surface|Public|Member|Admin' -v`
Expected: FAIL because current router still protects article routes with path prefix rules and lacks member route coverage.

**Step 3: Add minimal test helpers only**

If the tests need reusable helpers for admin or member tokens, add them inside `internal/api/register/router_test.go` or reuse existing auth helpers from module tests.

**Step 4: Run the targeted test command again**

Run: `go test ./internal/api/register -run 'Surface|Public|Member|Admin' -v`
Expected: FAIL remains, but now failing for the intended architecture reasons instead of compile errors.

**Step 5: Commit the red test state if working in an isolated implementation branch**

```bash
git add internal/api/register/router_test.go
git commit -m "test: define api surface boundaries"
```

### Task 2: Introduce shared identity primitives for admin and member principals

**Files:**
- Create: `internal/identity/claims.go`
- Create: `internal/identity/context.go`
- Create: `internal/identity/password.go`
- Create: `internal/identity/token.go`
- Modify: `internal/middleware/auth.go`
- Modify: `internal/middleware/authorize.go`
- Test: `internal/middleware/auth_test.go`

**Step 1: Write the failing tests**

Add tests proving middleware can distinguish:

```go
func TestAuthenticateStoresAdminPrincipal(t *testing.T)
func TestAuthenticateStoresMemberPrincipal(t *testing.T)
func TestRequireAdminRejectsMemberPrincipal(t *testing.T)
func TestRequireMemberRejectsAnonymous(t *testing.T)
```

**Step 2: Run the middleware test command and confirm failure**

Run: `go test ./internal/middleware -run 'Authenticate|RequireAdmin|RequireMember' -v`
Expected: FAIL because member principal support and distinct authorization middleware do not exist yet.

**Step 3: Implement the minimal identity layer**

Add explicit principal claims and context helpers, then refit middleware so:

```go
type PrincipalKind string
const (
    PrincipalAdmin  PrincipalKind = "admin"
    PrincipalMember PrincipalKind = "member"
)
```

Keep this change narrowly scoped. Do not refactor every module yet.

**Step 4: Run the middleware tests again**

Run: `go test ./internal/middleware -run 'Authenticate|RequireAdmin|RequireMember' -v`
Expected: PASS for the new identity tests.

**Step 5: Commit**

```bash
git add internal/identity internal/middleware/auth.go internal/middleware/authorize.go internal/middleware/auth_test.go
git commit -m "refactor: add shared identity primitives"
```

### Task 3: Convert router composition to Huma groups

**Files:**
- Modify: `internal/api/register/router.go`
- Modify: `internal/api/register/router_test.go`
- Reference: `internal/api/handlers/health.go`
- Reference: `internal/api/handlers/ready.go`

**Step 1: Write the failing router composition assertions**

Extend tests so they assert route registration is attached to groups instead of `rootMux.Handle` path guards. Test desired public/admin/member behavior through actual HTTP calls.

**Step 2: Run the router package tests and confirm failure**

Run: `go test ./internal/api/register -v`
Expected: FAIL because current composition still depends on `rootMux` path matching.

**Step 3: Implement Huma groups and shrink router responsibilities**

Refactor `NewRouter` so it:

```go
publicV1 := huma.NewGroup(api, "/api/v1")
memberAuth := huma.NewGroup(api, "/api/v1/member/auth")
memberSelf := huma.NewGroup(api, "/api/v1/me")
adminV1 := huma.NewGroup(api, "/api/v1/admin")
```

Then apply group middleware rather than path-prefix `rootMux.Handle` wrappers.

**Step 4: Run the router package tests again**

Run: `go test ./internal/api/register -v`
Expected: PASS with no path-prefix drift left in `router.go`.

**Step 5: Commit**

```bash
git add internal/api/register/router.go internal/api/register/router_test.go
git commit -m "refactor: group routes by api surface"
```

### Task 4: Split article routes into public and admin handlers

**Files:**
- Create: `internal/modules/article/public_handler.go`
- Create: `internal/modules/article/admin_handler.go`
- Modify: `internal/modules/article/service.go`
- Modify: `internal/modules/article/repository.go`
- Modify: `internal/modules/article/article_test.go`
- Modify: `internal/api/register/router.go`

**Step 1: Write the failing tests**

Add tests for:

```go
func TestPublicArticleListOnlyReturnsPublishedItems(t *testing.T)
func TestPublicArticleDetailLoadsBySlug(t *testing.T)
func TestAdminArticleRoutesPreserveDraftManagement(t *testing.T)
```

**Step 2: Run the article test command and confirm failure**

Run: `go test ./internal/modules/article -v`
Expected: FAIL because public/admin handler split and slug-based public detail are not implemented yet.

**Step 3: Implement the minimal split**

- Move public read endpoints into `public_handler.go`
- Move admin CRUD and publish endpoints into `admin_handler.go`
- Add repository/service helpers for published-only list/detail and slug lookup
- Keep existing domain rules in `service.go`

**Step 4: Run article tests again**

Run: `go test ./internal/modules/article -v`
Expected: PASS with both public and admin routes covered.

**Step 5: Commit**

```bash
git add internal/modules/article internal/api/register/router.go
git commit -m "refactor: split article public and admin routes"
```

### Task 5: Split category routes into public and admin handlers

**Files:**
- Create: `internal/modules/category/public_handler.go`
- Create: `internal/modules/category/admin_handler.go`
- Modify: `internal/modules/category/service.go`
- Modify: `internal/modules/category/category_test.go`
- Modify: `internal/api/register/router.go`

**Step 1: Write the failing tests**

Add tests for:

```go
func TestPublicCategoryListIsAccessibleWithoutAuth(t *testing.T)
func TestAdminCategoryCrudStillRequiresAdmin(t *testing.T)
```

**Step 2: Run the category test command and confirm failure**

Run: `go test ./internal/modules/category -v`
Expected: FAIL because there is no public/admin split yet.

**Step 3: Implement the minimal split**

- Public handler only exposes read endpoints
- Admin handler keeps CRUD
- Router registers each set into the correct group

**Step 4: Run category tests again**

Run: `go test ./internal/modules/category -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/modules/category internal/api/register/router.go
git commit -m "refactor: split category public and admin routes"
```

### Task 6: Add a dedicated member module for frontend identity

**Files:**
- Create: `internal/modules/member/model.go`
- Create: `internal/modules/member/repository.go`
- Create: `internal/modules/member/service.go`
- Create: `internal/modules/member/public_handler.go`
- Create: `internal/modules/member/self_handler.go`
- Create: `internal/modules/member/member_test.go`
- Modify: `internal/api/register/router.go`
- Modify: `internal/app/bootstrap/runtime.go` if extra resources/config are needed

**Step 1: Write the failing tests**

Add tests for:

```go
func TestMemberRegister(t *testing.T)
func TestMemberLogin(t *testing.T)
func TestMemberCanFetchSelfProfile(t *testing.T)
```

**Step 2: Run the member test command and confirm failure**

Run: `go test ./internal/modules/member -v`
Expected: FAIL because the member module does not exist yet.

**Step 3: Implement the minimal member slice**

Support only:

- register
- login
- refresh if already needed by the chosen token flow
- fetch self profile

Do not add likes/favorites logic here.

**Step 4: Run member tests again**

Run: `go test ./internal/modules/member -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/modules/member internal/api/register/router.go
 git commit -m "feat: add member identity module"
```

### Task 7: Add engagement module for likes and favorites

**Files:**
- Create: `internal/modules/engagement/model.go`
- Create: `internal/modules/engagement/repository.go`
- Create: `internal/modules/engagement/service.go`
- Create: `internal/modules/engagement/handler.go`
- Create: `internal/modules/engagement/engagement_test.go`
- Modify: `internal/api/register/router.go`

**Step 1: Write the failing tests**

Add tests for:

```go
func TestMemberCanLikeArticle(t *testing.T)
func TestMemberCanFavoriteArticle(t *testing.T)
func TestDuplicateFavoriteReturnsConflict(t *testing.T)
func TestAnonymousCannotFavoriteArticle(t *testing.T)
func TestMyFavoritesReturnsMemberScopedList(t *testing.T)
```

**Step 2: Run the engagement test command and confirm failure**

Run: `go test ./internal/modules/engagement -v`
Expected: FAIL because the module does not exist yet.

**Step 3: Implement the minimal engagement slice**

Keep the first version intentionally small:

- add like
- remove like
- add favorite
- remove favorite
- list my favorites

Avoid adding history, feeds, counters, or recommendation logic in this task.

**Step 4: Run engagement tests again**

Run: `go test ./internal/modules/engagement -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/modules/engagement internal/api/register/router.go
git commit -m "feat: add engagement module"
```

### Task 8: Update docs and verify the whole architecture slice

**Files:**
- Modify: `README.md`
- Modify: `internal/modules/example/README.md`
- Modify: `verification.md`
- Reference: `docs/plans/2026-04-03-multi-surface-api-architecture-design.md`

**Step 1: Write the failing verification checklist**

Create a checklist in the PR description or local notes covering:

- admin/public/member surfaces exist
- router has no path-prefix authorization drift
- article/category are split by surface
- member and engagement modules are documented

**Step 2: Run the full verification commands**

Run: `go test ./internal/api/register ./internal/middleware ./internal/modules/article ./internal/modules/category ./internal/modules/member ./internal/modules/engagement -v`
Expected: PASS

Run: `go test ./...`
Expected: PASS

**Step 3: Update docs minimally**

Document:

- new API surfaces
- new module map
- recommended extension rules

**Step 4: Re-run the full verification**

Run: `go test ./...`
Expected: PASS

**Step 5: Commit**

```bash
git add README.md verification.md internal/modules/example/README.md
git commit -m "docs: describe multi-surface api architecture"
```
