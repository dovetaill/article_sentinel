package register_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"sync"
	"testing"

	"github.com/dovetaill/article-sentinel/internal/api/register"
	"github.com/dovetaill/article-sentinel/internal/api/response"
	"github.com/dovetaill/article-sentinel/internal/app/bootstrap"
	"github.com/dovetaill/article-sentinel/internal/identity"
	"github.com/dovetaill/article-sentinel/pkg/config"
)

type openAPIDocument struct {
	Paths map[string]map[string]any `json:"paths"`
}

func TestRouterRegistersStarterOpenAPIPaths(t *testing.T) {
	rt := newRouterTestRuntime(true)
	handler := register.NewRouter(rt)
	doc := fetchOpenAPIDocument(t, handler, signedRouterSessionCookie(t, rt))

	wantPaths := []string{
		"/api/v1/demo/me",
		"/healthz",
		"/readyz",
	}
	gotPaths := sortedPathKeys(doc.Paths)
	if !reflect.DeepEqual(gotPaths, wantPaths) {
		t.Fatalf("openapi paths = %v, want %v", gotPaths, wantPaths)
	}

	assertOperation(t, doc.Paths, "/api/v1/demo/me", http.MethodGet)
	assertOperation(t, doc.Paths, "/healthz", http.MethodGet)
	assertOperation(t, doc.Paths, "/readyz", http.MethodGet)
	assertPathAbsent(t, doc.Paths, "/api/v1/article-inspect/orgs")
	assertPathAbsent(t, doc.Paths, "/api/v1/posts")
	assertPathAbsent(t, doc.Paths, "/api/v1/auth/login")
	assertPathAbsent(t, doc.Paths, "/auth/login")
	assertPathAbsent(t, doc.Paths, "/api/v1/member/auth/login")
	assertPathAbsent(t, doc.Paths, "/api/v1/admin/users")
}

func TestRouterServesStarterHealthAndReadyEndpoints(t *testing.T) {
	handler := register.NewRouter(newRouterTestRuntime(true))

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{name: "healthz", path: "/healthz", wantStatus: http.StatusOK},
		{name: "readyz", path: "/readyz", wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("%s status = %d, want %d", tt.path, rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestRouterDemoMeRequiresSessionEnvelope(t *testing.T) {
	handler := register.NewRouter(newRouterTestRuntime(true))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/demo/me", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("/api/v1/demo/me status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	got := decodeRouterEnvelope(t, rec)
	if got.Code != http.StatusUnauthorized {
		t.Fatalf("code = %d, want %d", got.Code, http.StatusUnauthorized)
	}
	if got.Message != "unauthorized" {
		t.Fatalf("message = %q, want %q", got.Message, "unauthorized")
	}
}

func TestRouterDemoMeReturnsActorFromSignedSessionCookie(t *testing.T) {
	rt := newRouterTestRuntime(true)
	handler := register.NewRouter(rt)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/demo/me", nil)
	req.AddCookie(signedRouterSessionCookie(t, rt))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("/api/v1/demo/me status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	got := decodeRouterEnvelope(t, rec)
	if got.Code != 0 {
		t.Fatalf("code = %d, want %d", got.Code, 0)
	}
	if got.Message != "me" {
		t.Fatalf("message = %q, want %q", got.Message, "me")
	}
	data, ok := got.Data.(map[string]any)
	if !ok {
		t.Fatalf("data type = %T, want map[string]any", got.Data)
	}
	if got := data["id"]; got != float64(1) {
		t.Fatalf("data.id = %v, want %d", got, 1)
	}
	if got := data["username"]; got != "Admin" {
		t.Fatalf("data.username = %v, want %q", got, "Admin")
	}
	if got := data["role"]; got != "admin" {
		t.Fatalf("data.role = %v, want %q", got, "admin")
	}
	if got := data["status"]; got != "active" {
		t.Fatalf("data.status = %v, want %q", got, "active")
	}
}

func TestRouterProtectsDocumentationEndpointsWhenDocsEnabled(t *testing.T) {
	handler := register.NewRouter(newRouterTestRuntime(true))
	protectedPaths := []string{
		"/docs",
		"/openapi.json",
		"/openapi.yaml",
		"/openapi-3.0.json",
		"/openapi-3.0.yaml",
		"/schemas/ErrorModel.json",
	}

	for _, path := range protectedPaths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("%s status = %d, want %d", path, rec.Code, http.StatusUnauthorized)
			}
			var got response.Envelope
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got.Code != http.StatusUnauthorized || got.Message != "unauthorized" {
				t.Fatalf("body = %+v, want unauthorized auth envelope", got)
			}
		})
	}
}

func TestRouterProtectsConfiguredDocumentationEndpointsWhenDocsEnabled(t *testing.T) {
	rt := newRouterTestRuntime(true)
	rt.Config.Docs.OpenAPIPath = "/schema.json"
	rt.Config.Docs.UIPath = "/reference"

	handler := register.NewRouter(rt)
	protectedPaths := []string{
		"/reference",
		"/reference/",
		"/reference/index.html",
		"/schema.json",
		"/schema.yaml",
		"/schema-3.0.json",
		"/schema-3.0.yaml",
		"/schemas/ErrorModel.json",
	}

	for _, path := range protectedPaths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("%s status = %d, want %d", path, rec.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestRouterAllowsConfiguredDocumentationEndpointsWithSession(t *testing.T) {
	rt := newRouterTestRuntime(true)
	rt.Config.Docs.OpenAPIPath = "/schema.json"
	rt.Config.Docs.UIPath = "/reference"

	handler := register.NewRouter(rt)
	cookie := signedRouterSessionCookie(t, rt)

	for _, path := range []string{"/schema.json", "/reference"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.AddCookie(cookie)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("%s status = %d, want %d, body=%s", path, rec.Code, http.StatusOK, rec.Body.String())
			}
		})
	}
}

func TestRouterNormalizesYAMLOpenAPIPathBeforeRegisteringAndProtectingDocs(t *testing.T) {
	rt := newRouterTestRuntime(true)
	rt.Config.Docs.OpenAPIPath = "/schema.yaml"

	handler := register.NewRouter(rt)
	openAPIPaths := []string{
		"/schema.json",
		"/schema.yaml",
		"/schema-3.0.json",
		"/schema-3.0.yaml",
	}

	for _, path := range openAPIPaths {
		t.Run("anonymous "+path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("%s status = %d, want %d", path, rec.Code, http.StatusUnauthorized)
			}
		})
	}

	cookie := signedRouterSessionCookie(t, rt)
	for _, path := range openAPIPaths {
		t.Run("authenticated "+path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.AddCookie(cookie)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("%s status = %d, want %d, body=%s", path, rec.Code, http.StatusOK, rec.Body.String())
			}
		})
	}

	req := httptest.NewRequest(http.MethodGet, "/schema.yaml.json", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("/schema.yaml.json status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestRouterServesOpenAPIWhenDocumentationRequestHasSession(t *testing.T) {
	rt := newRouterTestRuntime(true)
	handler := register.NewRouter(rt)

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	req.AddCookie(signedRouterSessionCookie(t, rt))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("/openapi.json status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var doc openAPIDocument
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("decode openapi: %v", err)
	}
	if _, ok := doc.Paths["/healthz"]; !ok {
		t.Fatal("openapi document missing /healthz path")
	}
}

func TestRouterLogsRejectedDocumentationRequests(t *testing.T) {
	memory := &memorySlogHandler{}
	rt := newRouterTestRuntime(true)
	rt.Logger = slog.New(memory)

	handler := register.NewRouter(rt)
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if len(memory.records) != 1 {
		t.Fatalf("record count = %d, want %d", len(memory.records), 1)
	}

	record := memory.records[0]
	if record["msg"] != "http_access" {
		t.Fatalf("msg = %v, want %q", record["msg"], "http_access")
	}
	if record["path"] != "/openapi.json" {
		t.Fatalf("path = %v, want %q", record["path"], "/openapi.json")
	}
	if got, ok := record["status_code"].(int64); !ok || got != http.StatusUnauthorized {
		t.Fatalf("status_code = %v, want %d", record["status_code"], http.StatusUnauthorized)
	}
	if _, ok := record["request_id"].(string); !ok {
		t.Fatalf("request_id type = %T, want string", record["request_id"])
	}
}

func TestRouterServesAuthBridgeAndSessionRoutesAtRuntime(t *testing.T) {
	handler := register.NewRouter(newRouterTestRuntime(true))

	loginReq := httptest.NewRequest(http.MethodGet, "/auth/login?jwt=bad-token", nil)
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusFound {
		t.Fatalf("/auth/login status = %d, want %d", loginRec.Code, http.StatusFound)
	}

	sessionReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
	sessionRec := httptest.NewRecorder()
	handler.ServeHTTP(sessionRec, sessionReq)
	if sessionRec.Code != http.StatusUnauthorized {
		t.Fatalf("/api/v1/auth/session status = %d, want %d", sessionRec.Code, http.StatusUnauthorized)
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logoutRec := httptest.NewRecorder()
	handler.ServeHTTP(logoutRec, logoutReq)
	if logoutRec.Code != http.StatusOK {
		t.Fatalf("/api/v1/auth/logout status = %d, want %d", logoutRec.Code, http.StatusOK)
	}
}

func TestRouterDisablesDocsEndpointsWhenDocsDisabled(t *testing.T) {
	handler := register.NewRouter(newRouterTestRuntime(false))

	for _, path := range []string{"/docs", "/openapi.json", "/openapi.yaml", "/openapi-3.0.json", "/openapi-3.0.yaml", "/schemas/ErrorModel.json"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("%s status = %d, want %d", path, rec.Code, http.StatusNotFound)
			}
		})
	}
}

func newRouterTestRuntime(docsEnabled bool) *bootstrap.Runtime {
	return &bootstrap.Runtime{
		Config: &config.Config{
			App:  config.AppConfig{Name: "article-sentinel"},
			Docs: config.DocsConfig{Enabled: docsEnabled, OpenAPIPath: "/openapi.json", UIPath: "/docs"},
			HTTP: config.HTTPConfig{RequestTimeoutSeconds: 15, ReadTimeoutSeconds: 15},
			Auth: config.AuthConfig{
				Session: config.SessionConfig{
					LegacySecret: "legacy-secret",
					Secret:       "session-secret",
					Issuer:       "article-sentinel-admin",
					TTLHours:     24,
					LoginURL:     "https://appadmin.cq.qiludev.com/cq-admin/index.html#/home",
					RedirectURL:  "http://127.0.0.1:5173/",
				},
			},
		},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func fetchOpenAPIDocument(t *testing.T, handler http.Handler, cookie *http.Cookie) openAPIDocument {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var doc openAPIDocument
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("decode openapi: %v", err)
	}

	return doc
}

func signedRouterSessionCookie(t *testing.T, rt *bootstrap.Runtime) *http.Cookie {
	t.Helper()

	manager := identity.NewAdminSessionManager(rt.Config.Auth.Session)
	token, _, err := manager.SignSessionJWT(identity.AdminSession{
		UserID:   1,
		OrgID:    10,
		OrgName:  "Test Org",
		Priv:     "admin",
		Nickname: "Admin",
		Status:   "active",
	})
	if err != nil {
		t.Fatalf("SignSessionJWT() error = %v", err)
	}

	return &http.Cookie{Name: manager.CookieName(), Value: token}
}

func decodeRouterEnvelope(t *testing.T, rec *httptest.ResponseRecorder) response.Envelope {
	t.Helper()

	var got response.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return got
}

type memorySlogHandler struct {
	mu      sync.Mutex
	records []map[string]any
}

func (h *memorySlogHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *memorySlogHandler) Handle(_ context.Context, record slog.Record) error {
	entry := map[string]any{"msg": record.Message}
	record.Attrs(func(attr slog.Attr) bool {
		entry[attr.Key] = attr.Value.Any()
		return true
	})

	h.mu.Lock()
	h.records = append(h.records, entry)
	h.mu.Unlock()
	return nil
}

func (h *memorySlogHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *memorySlogHandler) WithGroup(_ string) slog.Handler      { return h }

func sortedPathKeys(paths map[string]map[string]any) []string {
	keys := make([]string, 0, len(paths))
	for path := range paths {
		keys = append(keys, path)
	}
	slices.Sort(keys)
	return keys
}

func assertOperation(t *testing.T, paths map[string]map[string]any, path string, method string) {
	t.Helper()

	operations, ok := paths[path]
	if !ok {
		t.Fatalf("missing path %s", path)
	}
	if _, ok := operations[httpMethodKey(method)]; !ok {
		t.Fatalf("missing %s %s operation", method, path)
	}
}

func assertPathAbsent(t *testing.T, paths map[string]map[string]any, path string) {
	t.Helper()
	if _, ok := paths[path]; ok {
		t.Fatalf("path %s should not be registered on starter router", path)
	}
}

func httpMethodKey(method string) string {
	switch method {
	case http.MethodGet:
		return "get"
	case http.MethodPost:
		return "post"
	case http.MethodPut:
		return "put"
	case http.MethodPatch:
		return "patch"
	case http.MethodDelete:
		return "delete"
	default:
		return ""
	}
}
