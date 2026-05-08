package articleinspect

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

func TestRouteRegistrationRegistersArticleInspectPaths(t *testing.T) {
	db := newArticleInspectTestDB(t)
	dispatcher := &articleInspectTaskDispatcherStub{}

	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))
	RegisterRoutes(api, NewRoutes(db, dispatcher))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("openapi status = %d, want %d", rec.Code, http.StatusOK)
	}

	var doc struct {
		Paths map[string]map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("decode openapi: %v", err)
	}

	requiredPaths := []string{
		"/api/v1/article-inspect/orgs",
		"/api/v1/article-inspect/categories",
		"/api/v1/article-inspect/categories/{id}",
		"/api/v1/article-inspect/categories/{id}/status",
		"/api/v1/article-inspect/articles",
		"/api/v1/article-inspect/articles/{article_id}",
		"/api/v1/article-inspect/keywords",
		"/api/v1/article-inspect/keywords/{id}",
		"/api/v1/article-inspect/tasks",
		"/api/v1/article-inspect/tasks/{id}",
		"/api/v1/article-inspect/results",
		"/api/v1/article-inspect/results/{id}",
		"/api/v1/article-inspect/actions/batch-offline",
		"/api/v1/article-inspect/actions/batch-ignore",
		"/api/v1/article-inspect/actions/batch-process",
		"/api/v1/article-inspect/articles/{article_id}/rectify",
		"/api/v1/article-inspect/articles/{article_id}/republish",
		"/api/v1/article-inspect/logs/operations",
		"/api/v1/article-inspect/logs/field-changes",
	}

	for _, path := range requiredPaths {
		if _, ok := doc.Paths[path]; !ok {
			t.Fatalf("openapi missing path %s", path)
		}
	}

	assertArticleInspectOperationID(t, doc.Paths, "/api/v1/article-inspect/categories", http.MethodPost, "article-inspect-category-create")
	assertArticleInspectOperationID(t, doc.Paths, "/api/v1/article-inspect/tasks", http.MethodPost, "article-inspect-task-create")
	assertArticleInspectOperationID(t, doc.Paths, "/api/v1/article-inspect/keywords/{id}", http.MethodGet, "article-inspect-keyword-detail")

	if !articleInspectHasResponseStatus(doc.Paths, "/api/v1/article-inspect/categories", http.MethodPost, "201") {
		t.Fatal("category create must document 201 response")
	}
	if !articleInspectHasResponseStatus(doc.Paths, "/api/v1/article-inspect/tasks", http.MethodPost, "201") {
		t.Fatal("task create must document 201 response")
	}

	if got := articleInspectParameterSchemaType(t, doc.Paths, "/api/v1/article-inspect/keywords/{id}", http.MethodGet, "id"); got != "integer" {
		t.Fatalf("keyword detail path id schema type = %q, want %q", got, "integer")
	}
	if got := articleInspectParameterSchemaType(t, doc.Paths, "/api/v1/article-inspect/categories", http.MethodGet, "enabled"); got != "boolean" {
		t.Fatalf("category list enabled schema type = %q, want %q", got, "boolean")
	}
	if got := articleInspectParameterSchemaType(t, doc.Paths, "/api/v1/article-inspect/results", http.MethodGet, "orgid"); got != "integer" {
		t.Fatalf("result list orgid schema type = %q, want %q", got, "integer")
	}
}

func TestNewRoutesBuildsModuleDependencies(t *testing.T) {
	db := newArticleInspectTestDB(t)
	dispatcher := &articleInspectTaskDispatcherStub{}

	routes := NewRoutes(db, dispatcher)
	if routes.Categories == nil {
		t.Fatal("NewRoutes().Categories = nil")
	}
	if routes.Keywords == nil {
		t.Fatal("NewRoutes().Keywords = nil")
	}
	if routes.Tasks == nil {
		t.Fatal("NewRoutes().Tasks = nil")
	}
	if routes.Results == nil {
		t.Fatal("NewRoutes().Results = nil")
	}
	if routes.Actions == nil {
		t.Fatal("NewRoutes().Actions = nil")
	}
	if routes.Lifecycle == nil {
		t.Fatal("NewRoutes().Lifecycle = nil")
	}
	if routes.Logs == nil {
		t.Fatal("NewRoutes().Logs = nil")
	}
	if routes.Articles == nil {
		t.Fatal("NewRoutes().Articles = nil")
	}
	if routes.Dispatcher != dispatcher {
		t.Fatalf("NewRoutes().Dispatcher = %#v, want %#v", routes.Dispatcher, dispatcher)
	}
}
