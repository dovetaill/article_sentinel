# Main Starter / Showcase Split Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Reposition `main` as a minimal starter with one CRUD demo module while preserving the current multi-surface application in a dedicated showcase branch.

**Architecture:** Keep shared runtime, middleware, OpenAPI, and response conventions on `main`, but move the current business-heavy module graph and role model into a showcase branch. Refactor identity so starter infrastructure is generic and no longer imports `internal/modules/auth`.

**Tech Stack:** Go, `net/http`, `http.ServeMux`, Huma v2, GORM, Redis, slog, Go test

---

### Task 1: Preserve the current app as a showcase branch

**Files:**
- Reference: `README.md`
- Reference: `internal/api/register/router.go`
- Reference: `internal/modules/auth/handler.go`
- Reference: `internal/modules/article/admin_handler.go`

**Step 1: Create the showcase branch from the current HEAD**

Run: `git branch showcase/multisurface`
Expected: command succeeds with no output

**Step 2: Verify the showcase branch points to the current business-rich state**

Run: `git log --oneline -n 1 showcase/multisurface`
Expected: the commit hash matches current `HEAD`

**Step 3: Add a short branch note to the starter README draft**

Update `README.md` so the future starter README links to `showcase/multisurface` as the full example app.

**Step 4: Commit the branch-preservation prep once README drafting starts**

```bash
git add README.md
git commit -m "docs: point starter to showcase branch"
```

### Task 2: Decouple identity primitives from `internal/modules/auth`

**Files:**
- Modify: `internal/identity/claims.go`
- Modify: `internal/identity/token.go`
- Modify: `internal/identity/password.go`
- Modify: `internal/middleware/auth.go`
- Modify: `internal/middleware/authorize.go`
- Test: `internal/middleware/auth_test.go`
- Create: `internal/identity/actor.go`
- Create: `internal/identity/jwt.go`

**Step 1: Write the failing tests for generic identity behavior**

Add tests that verify middleware can store an authenticated actor in context without importing the current admin/member business roles.

**Step 2: Run the middleware tests to confirm the starter refactor is not implemented yet**

Run: `go test ./internal/middleware -run 'TestAuthenticate|TestRequire'`
Expected: FAIL because the new generic actor API does not exist yet

**Step 3: Introduce generic actor and JWT primitives in `internal/identity`**

Implement an `Actor` or `Principal` type that does not depend on `internal/modules/auth`, plus token helpers that live entirely in the identity package.

**Step 4: Update middleware to depend on the generic identity API**

Refactor `internal/middleware/auth.go` and `internal/middleware/authorize.go` so `RequireAuthenticated()` remains, while starter defaults no longer hardcode `RequireAdmin()` or `RequireMember()`.

**Step 5: Re-run the focused middleware tests**

Run: `go test ./internal/middleware -run 'TestAuthenticate|TestRequire'`
Expected: PASS

**Step 6: Commit the identity extraction**

```bash
git add internal/identity internal/middleware
git commit -m "refactor: extract generic identity primitives"
```

### Task 3: Add a single starter CRUD demo module

**Files:**
- Create: `internal/modules/post/model.go`
- Create: `internal/modules/post/repository.go`
- Create: `internal/modules/post/service.go`
- Create: `internal/modules/post/handler.go`
- Create: `internal/modules/post/post_test.go`
- Modify: `internal/app/bootstrap/schema.go`

**Step 1: Write the failing test for the starter demo module**

Add a focused test that covers one list/read path and one write path for the new `post` module.

**Step 2: Run the new module test to verify it fails first**

Run: `go test ./internal/modules/post -v`
Expected: FAIL because the module files are not implemented yet

**Step 3: Implement the minimal `post` model, repository, service, and Huma handler**

Keep the module intentionally small. Use it only to demonstrate the project's preferred layering and envelope style.

**Step 4: Register the `post` model in schema bootstrap for starter auto-migrate**

Update `internal/app/bootstrap/schema.go` or a nearby starter-owned registration file so the demo module can boot cleanly.

**Step 5: Re-run the module tests**

Run: `go test ./internal/modules/post -v`
Expected: PASS

**Step 6: Commit the starter demo module**

```bash
git add internal/modules/post internal/app/bootstrap/schema.go
git commit -m "feat: add starter post module"
```

### Task 4: Slim router and bootstrap to starter scope

**Files:**
- Modify: `internal/api/register/router.go`
- Modify: `internal/app/bootstrap/server.go`
- Modify: `cmd/server/main.go`
- Test: `internal/api/handlers/health_test.go`

**Step 1: Write or update a failing router-level test for the starter surface**

Adjust the existing router tests so they assert the starter still serves `/healthz`, `/readyz`, docs endpoints, and the new `post` routes.

**Step 2: Run the router tests before refactoring**

Run: `go test ./internal/api/... ./cmd/server/... -v`
Expected: FAIL once the test expectations mention the new starter wiring

**Step 3: Refactor `router.go` to register only starter-owned routes**

Keep Huma setup, health/readiness, shared middleware, and the starter module registration. Remove direct assembly of `auth`, `user`, `member`, `category`, `article`, and `engagement` from `main`.

**Step 4: Remove business-specific bootstrap defaults**

Update `internal/app/bootstrap/server.go` so starter boot does not seed a default admin account or assume the richer business module graph.

**Step 5: Re-run router and server tests**

Run: `go test ./internal/api/... ./internal/app/bootstrap/... ./cmd/server/... -v`
Expected: PASS

**Step 6: Commit the starter bootstrap cleanup**

```bash
git add internal/api/register/router.go internal/app/bootstrap/server.go cmd/server/main.go internal/api/handlers/health_test.go
git commit -m "refactor: slim starter router and bootstrap"
```

### Task 5: Rewrite documentation for starter-first onboarding

**Files:**
- Modify: `README.md`
- Modify: `internal/modules/example/README.md`
- Create: `docs/showcase/multisurface.md`
- Reference: `docs/plans/2026-04-03-main-starter-showcase-split-design.md`

**Step 1: Draft failing documentation expectations as a checklist**

Create a short checklist in your working notes: starter value proposition, quickstart, demo module replacement guidance, showcase branch link, and removal of business-heavy API inventory from `main`.

**Step 2: Rewrite `README.md` around the starter story**

Describe infrastructure, the single CRUD demo module, and how to replace it. Add a prominent link to `showcase/multisurface`.

**Step 3: Update `internal/modules/example/README.md`**

Explain that `post` is the minimal reference module on `main`, while the richer module set lives in the showcase branch.

**Step 4: Add showcase migration notes**

Create `docs/showcase/multisurface.md` with a short summary of what lives in the showcase branch and why it was split from `main`.

**Step 5: Verify docs references are internally consistent**

Run: `rg -n 'member auth|member self|engagement|admin users|showcase/multisurface|post module' README.md docs internal/modules/example/README.md`
Expected: results reflect the new branch split and no longer sell `main` as the multi-surface productized app

**Step 6: Commit the documentation rewrite**

```bash
git add README.md internal/modules/example/README.md docs/showcase/multisurface.md
git commit -m "docs: reposition main as starter"
```

### Task 6: Final verification and cleanup

**Files:**
- Reference: `go.mod`
- Reference: `README.md`
- Reference: `internal/api/register/router.go`

**Step 1: Run the full test suite**

Run: `go test ./...`
Expected: PASS

**Step 2: Smoke-check the starter server locally**

Run: `go run ./cmd/server -config configs/config.example.yaml`
Expected: server starts, serves `/healthz`, `/readyz`, `/openapi.json`, `/docs`, and starter `post` endpoints after configuration is adjusted for local dependencies

**Step 3: Review git diff for accidental business leftovers on `main`**

Run: `git diff --stat`
Expected: changes are concentrated in starter docs, identity, router/bootstrap, and the new demo module

**Step 4: Commit the verification pass if any last-minute fixes were needed**

```bash
git add -A
git commit -m "chore: verify starter branch split"
```
