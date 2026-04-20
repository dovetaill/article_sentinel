package handlers_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/dovetaill/article-sentinel/internal/api/register"
	"github.com/dovetaill/article-sentinel/internal/api/response"
	"github.com/dovetaill/article-sentinel/internal/app/bootstrap"
	"github.com/dovetaill/article-sentinel/pkg/config"
	"github.com/dovetaill/article-sentinel/pkg/database"
)

func TestHealthzReturnsAlive(t *testing.T) {
	rt := &bootstrap.Runtime{
		Config: &config.Config{
			App:  config.AppConfig{Name: "article-sentinel"},
			Docs: config.DocsConfig{Enabled: true, OpenAPIPath: "/openapi.json", UIPath: "/docs"},
			HTTP: config.HTTPConfig{ReadTimeoutSeconds: 15},
		},
		Logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		Resources: &database.Resources{},
	}

	handler := register.NewRouter(rt)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got response.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Message != "alive" {
		t.Fatalf("message = %q, want %q", got.Message, "alive")
	}
}

func TestReadyzReturnsDependencyStatus(t *testing.T) {
	rt := &bootstrap.Runtime{
		Config: &config.Config{
			App:  config.AppConfig{Name: "article-sentinel"},
			Docs: config.DocsConfig{Enabled: true, OpenAPIPath: "/openapi.json", UIPath: "/docs"},
			HTTP: config.HTTPConfig{ReadTimeoutSeconds: 15},
		},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	handler := register.NewRouter(rt)
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got response.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	data, ok := got.Data.(map[string]any)
	if !ok {
		t.Fatalf("data type = %T, want map[string]any", got.Data)
	}
	deps, ok := data["dependencies"].(map[string]any)
	if !ok {
		t.Fatalf("dependencies type = %T, want map[string]any", data["dependencies"])
	}

	assertDependencyStatus(t, deps, "database", false, false)
	assertDependencyStatus(t, deps, "redis", false, false)
}

func TestReadyzReportsConfiguredAndHealthyDependencies(t *testing.T) {
	rt := &bootstrap.Runtime{
		Config: &config.Config{
			App:  config.AppConfig{Name: "article-sentinel"},
			Docs: config.DocsConfig{Enabled: true, OpenAPIPath: "/openapi.json", UIPath: "/docs"},
			HTTP: config.HTTPConfig{ReadTimeoutSeconds: 15},
		},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Resources: &database.Resources{
			DB:    &gorm.DB{},
			Redis: &redis.Client{},
		},
	}

	handler := register.NewRouter(rt)
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got response.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	data, ok := got.Data.(map[string]any)
	if !ok {
		t.Fatalf("data type = %T, want map[string]any", got.Data)
	}
	deps, ok := data["dependencies"].(map[string]any)
	if !ok {
		t.Fatalf("dependencies type = %T, want map[string]any", data["dependencies"])
	}

	assertDependencyStatus(t, deps, "database", true, true)
	assertDependencyStatus(t, deps, "redis", true, true)
}

func TestResponseHelpersReturnStandardShape(t *testing.T) {
	ok := response.OK("ok", map[string]any{"name": "article-sentinel"})
	if ok.Code != 0 {
		t.Fatalf("OK code = %d, want %d", ok.Code, 0)
	}
	if ok.Message != "ok" {
		t.Fatalf("OK message = %q, want %q", ok.Message, "ok")
	}
	if ok.Data == nil {
		t.Fatal("OK data = nil, want non-nil")
	}

	fail := response.Fail(1001, "bad request")
	if fail.Code != 1001 {
		t.Fatalf("Fail code = %d, want %d", fail.Code, 1001)
	}
	if fail.Message != "bad request" {
		t.Fatalf("Fail message = %q, want %q", fail.Message, "bad request")
	}
}

func assertDependencyStatus(t *testing.T, deps map[string]any, name string, configured bool, healthy bool) {
	t.Helper()

	dependency, ok := deps[name].(map[string]any)
	if !ok {
		t.Fatalf("dependencies.%s type = %T, want map[string]any", name, deps[name])
	}
	if got := dependency["configured"]; got != configured {
		t.Fatalf("dependencies.%s.configured = %v, want %v", name, got, configured)
	}
	if got := dependency["healthy"]; got != healthy {
		t.Fatalf("dependencies.%s.healthy = %v, want %v", name, got, healthy)
	}
}
