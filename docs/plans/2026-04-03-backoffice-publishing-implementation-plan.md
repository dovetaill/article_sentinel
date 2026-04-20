# Backoffice Publishing System Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 在 article-sentinel 当前 skeleton 上实现一个可运行的后台多用户文稿发布系统，包含登录、JWT 鉴权、用户/分类/文稿 CRUD 与基础权限控制。

**Architecture:** 继续采用“统一 bootstrap + Huma 路由注册 + 模块化 handler/service/repository”方式。权限控制拆成 JWT 认证中间件、角色校验 helper 与 service 层 ownership 校验，持久化优先使用 GORM + AutoMigrate + seed admin 让首版业务先跑通。

**Tech Stack:** Go 1.25+、http.ServeMux、Huma v2、GORM、MySQL/PostgreSQL、Redis、JWT、bcrypt

---

### Task 1: Expand Runtime Config For Auth And Seed

**Files:**
- Modify: `pkg/config/config.go`
- Modify: `pkg/config/config_test.go`
- Modify: `configs/config.example.yaml`

**Step 1: Write the failing tests**

Add tests covering:

```go
func TestLoadReadsJWTConfig(t *testing.T) {}
func TestLoadReadsSeedAdminConfig(t *testing.T) {}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/config -run 'JWT|SeedAdmin' -v`
Expected: FAIL with missing `Auth` / `SeedAdmin` fields or nested config symbols

**Step 3: Write minimal implementation**

Implement config structs for:

- `AuthConfig`
- `JWTConfig`
- `SeedAdminConfig`

Update `config.example.yaml` to include JWT secret / TTL and default admin seed settings.

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/config -run 'JWT|SeedAdmin' -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/config/config.go pkg/config/config_test.go configs/config.example.yaml
git commit -m "feat: add auth and seed runtime config"
```

### Task 2: Add Bootstrap Support For AutoMigrate And Seed Admin

**Files:**
- Create: `internal/app/bootstrap/schema.go`
- Create: `internal/app/bootstrap/schema_test.go`
- Modify: `internal/app/bootstrap/runtime.go`

**Step 1: Write the failing tests**

Add tests covering:

```go
func TestAutoMigrateRegistersAllBusinessModels(t *testing.T) {}
func TestSeedAdminCreatesDefaultAdminWhenMissing(t *testing.T) {}
func TestSeedAdminSkipsWhenDisabled(t *testing.T) {}
```

Use seams/helpers instead of a real database connection.

**Step 2: Run test to verify it fails**

Run: `go test ./internal/app/bootstrap -run 'AutoMigrate|SeedAdmin' -v`
Expected: FAIL with missing schema/seed helpers

**Step 3: Write minimal implementation**

Implement:

- business model registration list
- `AutoMigrateBusinessTables(...)`
- `SeedDefaultAdmin(...)`
- integrate schema bootstrap into server runtime build path

**Step 4: Run test to verify it passes**

Run: `go test ./internal/app/bootstrap -run 'AutoMigrate|SeedAdmin' -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/app/bootstrap/schema.go internal/app/bootstrap/schema_test.go internal/app/bootstrap/runtime.go
git commit -m "feat: bootstrap schema and seed admin"
```

### Task 3: Add Auth Module With JWT Login And Current User Context

**Files:**
- Create: `internal/modules/auth/model.go`
- Create: `internal/modules/auth/password.go`
- Create: `internal/modules/auth/jwt.go`
- Create: `internal/modules/auth/repository.go`
- Create: `internal/modules/auth/service.go`
- Create: `internal/modules/auth/handler.go`
- Create: `internal/middleware/auth.go`
- Create: `internal/middleware/authorize.go`
- Create: `internal/modules/auth/auth_test.go`

**Step 1: Write the failing tests**

Add tests covering:

```go
func TestLoginReturnsJWTForValidCredentials(t *testing.T) {}
func TestLoginRejectsInvalidPassword(t *testing.T) {}
func TestAuthMiddlewareLoadsCurrentUserFromBearerToken(t *testing.T) {}
func TestAuthMiddlewareRejectsDisabledUser(t *testing.T) {}
func TestRequireAdminRejectsNonAdmin(t *testing.T) {}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/modules/auth ./internal/middleware -run 'Login|AuthMiddleware|RequireAdmin' -v`
Expected: FAIL with missing auth/jwt/middleware symbols

**Step 3: Write minimal implementation**

Implement:

- password hash + verify helpers
- JWT sign + parse helpers
- auth repository query by username / id
- login service
- current user context accessors
- auth middleware + admin guard helper
- `/api/v1/auth/login` and `/api/v1/auth/me`

**Step 4: Run test to verify it passes**

Run: `go test ./internal/modules/auth ./internal/middleware -run 'Login|AuthMiddleware|RequireAdmin' -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/modules/auth internal/middleware/auth.go internal/middleware/authorize.go
git commit -m "feat: add jwt auth module"
```

### Task 4: Add Admin User Management Module

**Files:**
- Create: `internal/modules/user/model.go`
- Create: `internal/modules/user/repository.go`
- Create: `internal/modules/user/service.go`
- Create: `internal/modules/user/handler.go`
- Create: `internal/modules/user/user_test.go`

**Step 1: Write the failing tests**

Add tests covering:

```go
func TestAdminCanCreateUser(t *testing.T) {}
func TestAdminCanListUsers(t *testing.T) {}
func TestNonAdminCannotAccessUserAdminEndpoints(t *testing.T) {}
func TestCreateUserRejectsDuplicateUsername(t *testing.T) {}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/modules/user -v`
Expected: FAIL with missing user symbols

**Step 3: Write minimal implementation**

Implement:

- user model + role/status constants
- repository CRUD + paged list
- service validation for duplicate username / password hashing
- admin-only handler registration under `/api/v1/admin/users`

**Step 4: Run test to verify it passes**

Run: `go test ./internal/modules/user -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/modules/user
git commit -m "feat: add admin user management"
```

### Task 5: Add Admin Category Management Module

**Files:**
- Create: `internal/modules/category/model.go`
- Create: `internal/modules/category/repository.go`
- Create: `internal/modules/category/service.go`
- Create: `internal/modules/category/handler.go`
- Create: `internal/modules/category/category_test.go`

**Step 1: Write the failing tests**

Add tests covering:

```go
func TestAdminCanCreateCategory(t *testing.T) {}
func TestAdminCanListCategories(t *testing.T) {}
func TestNonAdminCannotAccessCategoryAdminEndpoints(t *testing.T) {}
func TestCreateCategoryRejectsDuplicateSlug(t *testing.T) {}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/modules/category -v`
Expected: FAIL with missing category symbols

**Step 3: Write minimal implementation**

Implement:

- category model
- repository CRUD + list
- duplicate slug validation
- admin-only category handlers under `/api/v1/admin/categories`

**Step 4: Run test to verify it passes**

Run: `go test ./internal/modules/category -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/modules/category
git commit -m "feat: add admin category management"
```

### Task 6: Add Article Module With Ownership And Publish Flow

**Files:**
- Create: `internal/modules/article/model.go`
- Create: `internal/modules/article/repository.go`
- Create: `internal/modules/article/service.go`
- Create: `internal/modules/article/handler.go`
- Create: `internal/modules/article/article_test.go`

**Step 1: Write the failing tests**

Add tests covering:

```go
func TestUserCanCreateOwnArticle(t *testing.T) {}
func TestUserCanOnlyListOwnArticles(t *testing.T) {}
func TestUserCannotUpdateOtherUsersArticle(t *testing.T) {}
func TestAdminCanManageAnyArticle(t *testing.T) {}
func TestPublishAndUnpublishTransitionsStatus(t *testing.T) {}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/modules/article -v`
Expected: FAIL with missing article symbols

**Step 3: Write minimal implementation**

Implement:

- article model + status constants
- repository CRUD + filtered list
- ownership checks in service
- publish/unpublish state transition logic
- handlers under `/api/v1/articles`

**Step 4: Run test to verify it passes**

Run: `go test ./internal/modules/article -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/modules/article
git commit -m "feat: add article publishing module"
```

### Task 7: Wire Router And Standardize Pagination / Error Responses

**Files:**
- Modify: `internal/api/register/router.go`
- Modify: `internal/api/response/response.go`
- Create: `internal/api/register/router_test.go`

**Step 1: Write the failing tests**

Add tests covering:

```go
func TestRouterRegistersAuthAndBusinessRoutes(t *testing.T) {}
func TestListEndpointsReturnPagedEnvelope(t *testing.T) {}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/api/register ./internal/api/response -v`
Expected: FAIL with missing business route registration or paged response helpers

**Step 3: Write minimal implementation**

Implement:

- shared paged response helper
- auth + user + category + article route registration
- dependency assembly from runtime resources

**Step 4: Run test to verify it passes**

Run: `go test ./internal/api/register ./internal/api/response -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/api/register/router.go internal/api/response/response.go internal/api/register/router_test.go
git commit -m "feat: wire business modules into router"
```

### Task 8: Full Verification And Docs Update

**Files:**
- Modify: `README.md`
- Modify: `verification.md`
- Modify: `internal/modules/example/README.md`

**Step 1: Update docs**

Document:

- auth/login usage
- role/ownership rules
- example business module mapping to real modules
- deferred items after first business release

**Step 2: Run full verification**

Run:

```bash
go test ./... -v
go mod tidy
go fmt ./...
go vet ./...
go build ./...
```

Expected: PASS

**Step 3: Update verification.md**

Append:

- verification command list
- result summary
- remaining deferred items

**Step 4: Commit**

```bash
git add README.md verification.md internal/modules/example/README.md
git commit -m "docs: record backoffice publishing system progress"
```
