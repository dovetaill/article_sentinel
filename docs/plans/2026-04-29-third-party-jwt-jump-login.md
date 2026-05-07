# Third-Party JWT Jump Login Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add same-domain third-party jump login that validates the incoming JWT with the legacy secret, re-signs a one-day session JWT into a cookie, and makes the backend trust only the cookie session `orgid` for admin tenancy.

**Architecture:** Introduce a dedicated admin-session JWT model and cookie middleware beside the existing generic identity helpers, then add three HTTP endpoints for login bridge, session lookup, and logout. Keep the current `articleinspect` HTTP surface mostly stable for one transition pass, but override all tenant selection from request context so frontend-provided `orgid` stops being authoritative while the React admin app switches to bootstrapping from `/api/v1/auth/session` and globally redirects on `401`.

**Tech Stack:** Go, net/http, Huma router shell, `github.com/golang-jwt/jwt/v5`, React 18, TypeScript, Vite, Vitest, Testing Library

---

### Task 1: Add failing config coverage for admin session settings

**Files:**
- Modify: `pkg/config/config.go`
- Modify: `pkg/config/config_test.go`
- Modify: `configs/config.example.yaml`
- Modify: `README.md`

**Step 1: Write the failing test**

Add a config test that asserts `config.AuthConfig` exposes a dedicated session config with env-backed fields for legacy secret, session secret, cookie name, issuer, TTL hours, and secure-cookie behavior.

```go
func TestAuthConfigExposesSessionFields(t *testing.T) {
    typ := reflect.TypeOf(config.AuthConfig{})
    field, ok := typ.FieldByName("Session")
    if !ok {
        t.Fatal("AuthConfig.Session missing")
    }
    if field.Type != reflect.TypeOf(config.SessionConfig{}) {
        t.Fatalf("AuthConfig.Session type = %v", field.Type)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/config -run Session -v`
Expected: FAIL because `SessionConfig` does not exist yet.

**Step 3: Write minimal implementation**

Add a new config block under `AuthConfig`:

```go
type SessionConfig struct {
    LegacySecret string `yaml:"legacy_secret" env:"AUTH_SESSION_LEGACY_SECRET"`
    Secret       string `yaml:"secret" env:"AUTH_SESSION_SECRET"`
    CookieName   string `yaml:"cookie_name" env:"AUTH_SESSION_COOKIE_NAME" env-default:"as_admin_session"`
    Issuer       string `yaml:"issuer" env:"AUTH_SESSION_ISSUER" env-default:"article-sentinel-admin"`
    TTLHours     int    `yaml:"ttl_hours" env:"AUTH_SESSION_TTL_HOURS" env-default:"24"`
    SecureCookie bool   `yaml:"secure_cookie" env:"AUTH_SESSION_SECURE_COOKIE" env-default:"true"`
}
```

Also add matching sample config and README deployment notes. Do not commit the real legacy/new secrets to the repo; keep placeholders in docs/example config.

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/config -run Session -v`
Expected: PASS

### Task 2: Lock the session JWT exchange behavior with tests

**Files:**
- Create: `internal/identity/admin_session.go`
- Create: `internal/identity/admin_session_token.go`
- Create: `internal/identity/admin_session_token_test.go`

**Step 1: Write the failing test**

Add focused tests for:

- parsing a legacy third-party JWT signed with the legacy secret
- mapping it into an internal `AdminSession`
- signing a new cookie JWT with one-day expiry
- rejecting expired or wrong-signature legacy JWTs
- parsing the new cookie JWT back into an `AdminSession`

```go
func TestAdminSessionManagerExchangesLegacyJWT(t *testing.T) {
    mgr := NewAdminSessionManager(config.SessionConfig{
        LegacySecret: "legacy-secret",
        Secret:       "session-secret",
        Issuer:       "article-sentinel-admin",
        TTLHours:     24,
    })

    legacy := signedLegacyJWT(t, "legacy-secret", map[string]any{
        "id": "90525",
        "orgid": "29",
        "orgname": "一县一端测试机构",
        "platform": "chuangqi",
        "priv": "super",
        "roleid": "1",
        "nickname": "用户A",
        "avatar": "https://example.com/a.png",
    })

    token, session, expiresAt, err := mgr.ExchangeLegacyJWT(legacy)
    if err != nil {
        t.Fatalf("ExchangeLegacyJWT() error = %v", err)
    }
    if session.OrgID != 29 || session.Nickname != "用户A" {
        t.Fatalf("session = %+v", session)
    }
    if expiresAt.Sub(mgr.now()).Round(time.Hour) != 24*time.Hour {
        t.Fatalf("ttl = %v", expiresAt.Sub(mgr.now()))
    }
    parsed, err := mgr.ParseSessionJWT(token)
    if err != nil || parsed.OrgID != 29 {
        t.Fatalf("ParseSessionJWT() = %+v, %v", parsed, err)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/identity -run AdminSession -v`
Expected: FAIL because the admin-session manager does not exist.

**Step 3: Write minimal implementation**

Implement:

- `AdminSession` struct with current-user and current-org fields
- context helpers `ContextWithAdminSession` / `AdminSessionFromContext`
- a manager that can:
  - parse legacy JWTs signed with the legacy secret
  - validate `exp`
  - map legacy claims into internal session fields
  - sign a new JWT with registered claims and one-day expiry
  - parse the new JWT back into `AdminSession`
- a helper that derives `identity.Actor` from session data for audit logs (`Username <- Nickname`, `Role <- Priv`, `ID <- session user id`)

**Step 4: Run test to verify it passes**

Run: `go test ./internal/identity -run AdminSession -v`
Expected: PASS

### Task 3: Add failing middleware coverage for cookie session hydration and cleanup

**Files:**
- Create: `internal/middleware/session.go`
- Create: `internal/middleware/session_test.go`

**Step 1: Write the failing test**

Cover three cases:

- valid `as_admin_session` cookie hydrates `AdminSession`, `Actor`, and `Principal`
- invalid/expired cookie causes a clearing `Set-Cookie` header
- anonymous requests keep moving when no cookie is present

```go
func TestSessionContextStoresAdminSessionFromCookie(t *testing.T) {
    mgr := newTestAdminSessionManager(t)
    token := signedSessionToken(t, mgr, identity.AdminSession{UserID: 90525, OrgID: 29, Nickname: "用户A", Priv: "super"})

    handler := SessionContext(mgr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        session, ok := identity.AdminSessionFromContext(r.Context())
        if !ok || session.OrgID != 29 {
            t.Fatalf("session = %+v, ok=%v", session, ok)
        }
        actor, ok := identity.ActorFromContext(r.Context())
        if !ok || actor.Username != "用户A" {
            t.Fatalf("actor = %+v, ok=%v", actor, ok)
        }
        w.WriteHeader(http.StatusNoContent)
    }))

    req := httptest.NewRequest(http.MethodGet, "/api/v1/article-inspect/tasks", nil)
    req.AddCookie(&http.Cookie{Name: "as_admin_session", Value: token})
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusNoContent {
        t.Fatalf("status = %d", rec.Code)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/middleware -run SessionContext -v`
Expected: FAIL because no session middleware exists.

**Step 3: Write minimal implementation**

Implement a dedicated middleware that:

- reads the configured session cookie
- parses it with the admin-session manager
- writes session data to context on success
- also seeds `identity.Actor` and `identity.Principal`
- clears the cookie when a malformed/expired session cookie is encountered

Keep this middleware separate from the existing bearer/header middleware in `internal/middleware/auth.go`.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/middleware -run SessionContext -v`
Expected: PASS

### Task 4: Add failing HTTP tests for login bridge, session lookup, and logout

**Files:**
- Create: `internal/api/handlers/auth.go`
- Create: `internal/api/handlers/auth_test.go`
- Modify: `internal/api/register/router.go`
- Modify: `internal/api/register/router_test.go`

**Step 1: Write the failing test**

Add handler tests for:

- `GET /auth/login?jwt=...` sets the session cookie and redirects to `/`
- invalid legacy JWT clears cookie and redirects to `https://appadmin.cq.qiludev.com/cq-admin/index.html`
- `GET /api/v1/auth/session` returns the current session envelope when cookie/session context exists
- `GET /api/v1/auth/session` returns `401` and clears cookie when the cookie is invalid
- `POST /api/v1/auth/logout` always clears the cookie

Also extend router tests to assert these paths are actually served at runtime.

```go
func TestAuthLoginBridgesLegacyJWTIntoSessionCookie(t *testing.T) {
    handler := newAuthHandlerForTest(t)
    req := httptest.NewRequest(http.MethodGet, "/auth/login?jwt="+url.QueryEscape(validLegacyJWT), nil)
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusFound {
        t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
    }
    if got := rec.Header().Get("Location"); got != "/" {
        t.Fatalf("Location = %q, want %q", got, "/")
    }
    if !strings.Contains(rec.Header().Get("Set-Cookie"), "as_admin_session=") {
        t.Fatalf("Set-Cookie = %q", rec.Header().Get("Set-Cookie"))
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/api/handlers ./internal/api/register -run 'Auth|Router' -v`
Expected: FAIL because auth handlers and routes do not exist.

**Step 3: Write minimal implementation**

Implement plain `net/http` handlers registered on the same mux as the current Huma adapter:

- `GET /auth/login`
- `GET /api/v1/auth/session`
- `POST /api/v1/auth/logout`

Use the admin-session manager for exchange/parse and `response.OK/Fail` envelopes for JSON endpoints. Add the session middleware to the router chain before the request reaches the handlers so `/api/v1/auth/session` can read the hydrated context.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/api/handlers ./internal/api/register -run 'Auth|Router' -v`
Expected: PASS

### Task 5: Lock tenant scoping with failing request-scope tests

**Files:**
- Create: `internal/modules/articleinspect/request_scope.go`
- Modify: `internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Write the failing test**

Add regression coverage proving that business behavior follows the session `orgid`, not the request `orgid`:

- send a request with session org `29` and query/body `orgid=100`
- expect only org `29` data back or only org `29` mutation to occur
- keep one anonymous request test to assert missing session is rejected with `401`

Update `sendArticleInspectRequest` to inject a valid `AdminSession` into request context by default so the suite stays readable.

```go
func TestArticleInspectRoutesPreferSessionOrgIDOverQueryOrgID(t *testing.T) {
    handler := newArticleInspectHandler(t)

    result := sendArticleInspectRequest(t, handler, http.MethodGet,
        "/api/v1/article-inspect/articles?orgid=100&page=1&page_size=20", nil,
    )

    if result.status != http.StatusOK {
        t.Fatalf("status = %d", result.status)
    }
    items := articleInspectItems(t, result.body)
    for _, item := range items {
        if articleInspectUint64Field(t, item, "orgid") != 29 {
            t.Fatalf("orgid = %d, want session org 29", articleInspectUint64Field(t, item, "orgid"))
        }
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/modules/articleinspect -run 'SessionOrgID|anonymous' -v`
Expected: FAIL because routes still trust request `orgid` and do not require session context.

**Step 3: Write minimal implementation**

Create request-scope helpers that read `AdminSession` from context and use them across the route layer.

Example helper:

```go
func currentOrgID(ctx context.Context) (uint64, error) {
    session, ok := identity.AdminSessionFromContext(ctx)
    if !ok || session.OrgID == 0 {
        return 0, identity.ErrUnauthorized
    }
    return session.OrgID, nil
}
```

Use these helpers to override incoming org values in every route group under `internal/modules/articleinspect/*_routes.go`.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/modules/articleinspect -run 'SessionOrgID|anonymous' -v`
Expected: PASS

### Task 6: Convert all articleinspect route groups to session-backed `orgid`

**Files:**
- Modify: `internal/modules/articleinspect/category_routes.go`
- Modify: `internal/modules/articleinspect/keyword_routes.go`
- Modify: `internal/modules/articleinspect/task_routes.go`
- Modify: `internal/modules/articleinspect/result_routes.go`
- Modify: `internal/modules/articleinspect/article_routes.go`
- Modify: `internal/modules/articleinspect/action_routes.go`
- Modify: `internal/modules/articleinspect/lifecycle_routes.go`
- Modify: `internal/modules/articleinspect/log_routes.go`
- Modify: `internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Write the failing test**

Add or extend focused tests for each route family to cover one mutation and one list/detail path where a mismatched request `orgid` must still behave as the session tenant.

Examples:

- category create ignores body `orgid`
- keyword list ignores query `orgid`
- task detail ignores query `orgid`
- result detail ignores query `orgid`
- article rectify/offline/republish ignore body `orgid`
- logs endpoints ignore query `orgid`

**Step 2: Run test to verify it fails**

Run: `go test ./internal/modules/articleinspect -run 'category|keyword|task|result|article|log' -v`
Expected: FAIL in newly added session-scoping assertions.

**Step 3: Write minimal implementation**

For each route:

- still parse legacy request params if needed for backward compatibility of request shape
- overwrite `OrgID` before calling service methods
- for create/update inputs, inject `OrgID` from current session
- surface `401` when no session is present

Representative change:

```go
orgID, err := currentOrgID(ctx)
if err != nil {
    return failureOKEnvelope(http.StatusUnauthorized, "unauthorized"), nil
}
result, err := service.List(ctx, CategoryListInput{
    OrgID: orgID,
    Page: page,
    PageSize: pageSize,
    Enabled: enabled,
    Query: input.Query,
})
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/modules/articleinspect -run 'category|keyword|task|result|article|log' -v`
Expected: PASS

### Task 7: Add failing frontend request/session bootstrap tests

**Files:**
- Modify: `web/admin/src/lib/request.ts`
- Modify: `web/admin/src/lib/request.test.ts`
- Create: `web/admin/src/services/auth.ts`
- Create: `web/admin/src/context/session-context.tsx`
- Modify: `web/admin/src/App.tsx`
- Modify: `web/admin/src/App.test.tsx`

**Step 1: Write the failing test**

Add tests for:

- `apiRequest` sends same-origin credentials
- `apiRequest` redirects on `401`
- app bootstrap loads `/api/v1/auth/session` before rendering shell
- failed session bootstrap redirects to the fixed login URL instead of rendering the app

```ts
it('sends same-origin credentials', async () => {
  const fetchMock = vi.fn().mockResolvedValue({
    ok: true,
    status: 200,
    json: async () => ({ code: 0, message: 'ok', data: { ok: true } })
  });
  vi.stubGlobal('fetch', fetchMock);

  await apiRequest('/api/v1/auth/session');

  expect(fetchMock).toHaveBeenCalledWith('/api/v1/auth/session', expect.objectContaining({
    credentials: 'same-origin'
  }));
});
```

**Step 2: Run test to verify it fails**

Run: `cd web/admin && npm test -- --runInBand src/lib/request.test.ts src/App.test.tsx`
Expected: FAIL because credentials, session bootstrap, and redirect logic do not exist.

**Step 3: Write minimal implementation**

Implement:

- `web/admin/src/services/auth.ts` with `getSession()` and `logout()`
- `web/admin/src/context/session-context.tsx` that loads session once and exposes user/org info
- `apiRequest()` with `credentials: 'same-origin'` and a fixed-login redirect on `401`
- `App.tsx` so the shell renders only after session bootstrap succeeds

**Step 4: Run test to verify it passes**

Run: `cd web/admin && npm test -- --runInBand src/lib/request.test.ts src/App.test.tsx`
Expected: PASS

### Task 8: Replace org switching with read-only org display and real user info

**Files:**
- Modify: `web/admin/src/context/org-context.tsx`
- Modify: `web/admin/src/components/layout/org-switcher.tsx`
- Modify: `web/admin/src/components/layout/user-menu.tsx`
- Modify: `web/admin/src/components/layout/header-bar.tsx`
- Modify: `web/admin/src/workbench/provider.tsx`
- Modify: `web/admin/src/components/layout/header-bar.test.tsx`
- Modify: `web/admin/src/App.test.tsx`

**Step 1: Write the failing test**

Add UI tests that assert:

- current org is displayed but no dropdown menu is available
- user menu shows `nickname`
- logout calls the auth service and redirects
- workbench still keys tabs by the current session org id

**Step 2: Run test to verify it fails**

Run: `cd web/admin && npm test -- --runInBand src/App.test.tsx src/components/layout/header-bar.test.tsx`
Expected: FAIL because the UI still expects org lists and a fake “当前用户”.

**Step 3: Write minimal implementation**

Refactor `org-context.tsx` into a thin adapter over the session context so existing consumers can keep reading `activeOrgId` and `activeOrgName`, but remove mutable org switching state.

Representative value shape:

```ts
return {
  activeOrgId: session.orgid,
  activeOrgName: session.orgname,
  isLoading,
};
```

Update `OrgSwitcher` to a read-only button/tag and `UserMenu` to use `nickname/avatar` plus the real logout action.

**Step 4: Run test to verify it passes**

Run: `cd web/admin && npm test -- --runInBand src/App.test.tsx src/components/layout/header-bar.test.tsx`
Expected: PASS

### Task 9: Remove frontend reliance on explicit `orgid` service params

**Files:**
- Modify: `web/admin/src/services/articles.ts`
- Modify: `web/admin/src/services/categories.ts`
- Modify: `web/admin/src/services/keywords.ts`
- Modify: `web/admin/src/services/tasks.ts`
- Modify: `web/admin/src/services/results.ts`
- Modify: `web/admin/src/services/logs.ts`
- Modify: `web/admin/src/services/articles.test.ts`
- Modify: `web/admin/src/pages/**/*.test.tsx`

**Step 1: Write the failing test**

Update service and page tests so API URLs no longer include `orgid=` in query strings for standard admin requests.

```ts
expect(mockedApiRequest).toHaveBeenCalledWith('/api/v1/article-inspect/articles?page=1&page_size=20');
```

**Step 2: Run test to verify it fails**

Run: `cd web/admin && npm test -- --runInBand src/services/articles.test.ts src/pages/articles/index.test.tsx src/pages/categories/index.test.tsx src/pages/logs/index.test.tsx`
Expected: FAIL because the services and pages still require `orgid` parameters.

**Step 3: Write minimal implementation**

Remove `orgid` from service method signatures and URL construction where the backend now derives tenancy from cookie session. Keep non-tenant filters like `page`, `page_size`, `status`, `task_id`, `article_id`, and `title` intact.

**Step 4: Run test to verify it passes**

Run: `cd web/admin && npm test -- --runInBand src/services/articles.test.ts src/pages/articles/index.test.tsx src/pages/categories/index.test.tsx src/pages/logs/index.test.tsx`
Expected: PASS

### Task 10: Run full verification and capture docs-only commit

**Files:**
- Modify: `docs/plans/2026-04-29-third-party-jwt-jump-login-design.md`
- Modify: `docs/plans/2026-04-29-third-party-jwt-jump-login.md`
- Modify: files from Tasks 1-9

**Step 1: Run backend verification**

Run: `go test ./...`
Expected: PASS

**Step 2: Run frontend verification**

Run: `cd web/admin && npm test -- --runInBand`
Expected: PASS

**Step 3: Run optional smoke check**

Run: `bash scripts/verify.sh`
Expected: PASS if the local environment is configured; otherwise document the blocking dependency precisely.

**Step 4: Commit**

```bash
git add \
  docs/plans/2026-04-29-third-party-jwt-jump-login-design.md \
  docs/plans/2026-04-29-third-party-jwt-jump-login.md \
  pkg/config/config.go pkg/config/config_test.go configs/config.example.yaml README.md \
  internal/identity/admin_session.go internal/identity/admin_session_token.go internal/identity/admin_session_token_test.go \
  internal/middleware/session.go internal/middleware/session_test.go \
  internal/api/handlers/auth.go internal/api/handlers/auth_test.go internal/api/register/router.go internal/api/register/router_test.go \
  internal/modules/articleinspect/request_scope.go internal/modules/articleinspect/articleinspect_test.go \
  internal/modules/articleinspect/category_routes.go internal/modules/articleinspect/keyword_routes.go internal/modules/articleinspect/task_routes.go \
  internal/modules/articleinspect/result_routes.go internal/modules/articleinspect/article_routes.go internal/modules/articleinspect/action_routes.go \
  internal/modules/articleinspect/lifecycle_routes.go internal/modules/articleinspect/log_routes.go \
  web/admin/src/lib/request.ts web/admin/src/lib/request.test.ts \
  web/admin/src/services/auth.ts web/admin/src/context/session-context.tsx web/admin/src/context/org-context.tsx \
  web/admin/src/App.tsx web/admin/src/App.test.tsx web/admin/src/components/layout/org-switcher.tsx \
  web/admin/src/components/layout/user-menu.tsx web/admin/src/components/layout/header-bar.tsx \
  web/admin/src/components/layout/header-bar.test.tsx web/admin/src/workbench/provider.tsx \
  web/admin/src/services/articles.ts web/admin/src/services/categories.ts web/admin/src/services/keywords.ts \
  web/admin/src/services/tasks.ts web/admin/src/services/results.ts web/admin/src/services/logs.ts \
  web/admin/src/services/articles.test.ts

git commit -m "feat(auth): add third-party jwt jump login"
```
