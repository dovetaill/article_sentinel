package register_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"testing"

	"github.com/dovetaill/article-sentinel/internal/api/register"
	"github.com/dovetaill/article-sentinel/internal/app/bootstrap"
	"github.com/dovetaill/article-sentinel/pkg/config"
	"github.com/dovetaill/article-sentinel/pkg/database"
	"gorm.io/gorm"
)

type openAPIDocument struct {
	Paths map[string]map[string]any `json:"paths"`
}

func TestRouterRegistersArticleInspectRoutes(t *testing.T) {
	handler := register.NewRouter(newRouterTestRuntime(true))
	doc := fetchOpenAPIDocument(t, handler)

	wantPaths := []string{
		"/api/v1/article-inspect/actions/batch-ignore",
		"/api/v1/article-inspect/actions/batch-process",
		"/api/v1/article-inspect/articles",
		"/api/v1/article-inspect/articles/{article_id}",
		"/api/v1/article-inspect/articles/{article_id}/change-logs",
		"/api/v1/article-inspect/articles/{article_id}/offline",
		"/api/v1/article-inspect/articles/{article_id}/operation-logs",
		"/api/v1/article-inspect/articles/{article_id}/rectify",
		"/api/v1/article-inspect/articles/{article_id}/republish",
		"/api/v1/article-inspect/categories",
		"/api/v1/article-inspect/categories/{id}",
		"/api/v1/article-inspect/categories/{id}/status",
		"/api/v1/article-inspect/keywords",
		"/api/v1/article-inspect/keywords/{id}",
		"/api/v1/article-inspect/keywords/{id}/status",
		"/api/v1/article-inspect/logs/field-changes",
		"/api/v1/article-inspect/logs/operations",
		"/api/v1/article-inspect/orgs",
		"/api/v1/article-inspect/results",
		"/api/v1/article-inspect/results/{id}",
		"/api/v1/article-inspect/tasks",
		"/api/v1/article-inspect/tasks/{id}",
		"/api/v1/posts",
		"/api/v1/posts/{id}",
		"/healthz",
		"/readyz",
	}
	gotPaths := sortedPathKeys(doc.Paths)
	if !reflect.DeepEqual(gotPaths, wantPaths) {
		t.Fatalf("openapi paths = %v, want %v", gotPaths, wantPaths)
	}

	assertOperation(t, doc.Paths, "/healthz", http.MethodGet)
	assertOperation(t, doc.Paths, "/readyz", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/orgs", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/categories", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/categories", http.MethodPost)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/categories/{id}", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/categories/{id}", http.MethodPut)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/categories/{id}", http.MethodDelete)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/categories/{id}/status", http.MethodPatch)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/keywords", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/keywords", http.MethodPost)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/keywords/{id}", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/keywords/{id}", http.MethodPut)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/keywords/{id}", http.MethodDelete)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/keywords/{id}/status", http.MethodPatch)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/articles", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/articles/{article_id}", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/tasks", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/tasks", http.MethodPost)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/tasks/{id}", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/results", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/results/{id}", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/actions/batch-ignore", http.MethodPost)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/actions/batch-process", http.MethodPost)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/articles/{article_id}/offline", http.MethodPost)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/articles/{article_id}/rectify", http.MethodPut)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/articles/{article_id}/republish", http.MethodPost)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/logs/operations", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/article-inspect/logs/field-changes", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/posts", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/posts", http.MethodPost)
	assertOperation(t, doc.Paths, "/api/v1/posts/{id}", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/posts/{id}", http.MethodPatch)
	assertOperation(t, doc.Paths, "/api/v1/posts/{id}", http.MethodDelete)

	assertPathAbsent(t, doc.Paths, "/api/v1/auth/login")
	assertPathAbsent(t, doc.Paths, "/api/v1/member/auth/login")
	assertPathAbsent(t, doc.Paths, "/api/v1/admin/users")
}

func TestRouterServesStarterHealthAndDocsEndpoints(t *testing.T) {
	handler := register.NewRouter(newRouterTestRuntime(true))

	tests := []struct {
		name       string
		path       string
		wantStatus int
		allow3xx   bool
	}{
		{name: "healthz", path: "/healthz", wantStatus: http.StatusOK},
		{name: "readyz", path: "/readyz", wantStatus: http.StatusOK},
		{name: "openapi", path: "/openapi.json", wantStatus: http.StatusOK},
		{name: "docs", path: "/docs", wantStatus: http.StatusOK, allow3xx: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if tt.allow3xx {
				if rec.Code >= http.StatusBadRequest {
					t.Fatalf("%s status = %d, want < %d", tt.path, rec.Code, http.StatusBadRequest)
				}
				return
			}

			if rec.Code != tt.wantStatus {
				t.Fatalf("%s status = %d, want %d", tt.path, rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestRouterDisablesDocsEndpointsWhenDocsDisabled(t *testing.T) {
	handler := register.NewRouter(newRouterTestRuntime(false))

	for _, path := range []string{"/openapi.json", "/docs"} {
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
				JWT: config.JWTConfig{
					Secret:     "test-secret",
					Issuer:     "article-sentinel-test",
					TTLMinutes: 120,
				},
			},
		},
		Logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		Resources: &database.Resources{DB: &gorm.DB{}},
	}
}

func fetchOpenAPIDocument(t *testing.T, handler http.Handler) openAPIDocument {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
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
