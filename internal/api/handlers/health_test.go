package handlers_test

import (
	"bufio"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
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
	db := mustOpenSQLite(t)
	fakeRedisAddr := startFakeRedisPingServer(t)

	rt := &bootstrap.Runtime{
		Config: &config.Config{
			App:  config.AppConfig{Name: "article-sentinel"},
			Docs: config.DocsConfig{Enabled: true, OpenAPIPath: "/openapi.json", UIPath: "/docs"},
			HTTP: config.HTTPConfig{ReadTimeoutSeconds: 15},
		},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Resources: &database.Resources{
			DB:    db,
			Redis: redis.NewClient(&redis.Options{Addr: fakeRedisAddr, Protocol: 2}),
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

func TestReadyzReportsUnhealthyConfiguredDependencies(t *testing.T) {
	db := mustOpenSQLite(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("sqlDB.Close() error = %v", err)
	}

	rt := &bootstrap.Runtime{
		Config: &config.Config{
			App:  config.AppConfig{Name: "article-sentinel"},
			Docs: config.DocsConfig{Enabled: true, OpenAPIPath: "/openapi.json", UIPath: "/docs"},
			HTTP: config.HTTPConfig{ReadTimeoutSeconds: 15},
		},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Resources: &database.Resources{
			DB: db,
			Redis: redis.NewClient(&redis.Options{
				Addr:         "127.0.0.1:1",
				DialTimeout:  200 * time.Millisecond,
				ReadTimeout:  200 * time.Millisecond,
				WriteTimeout: 200 * time.Millisecond,
			}),
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

	assertDependencyStatus(t, deps, "database", true, false)
	assertDependencyStatus(t, deps, "redis", true, false)
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

func mustOpenSQLite(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err != nil {
			return
		}
		_ = sqlDB.Close()
	})

	return db
}

func startFakeRedisPingServer(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen fake redis: %v", err)
	}

	t.Cleanup(func() {
		_ = ln.Close()
	})

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}

			go func(conn net.Conn) {
				defer conn.Close()

				reader := bufio.NewReader(conn)
				for {
					command, err := readRESPCommand(reader)
					if err != nil {
						return
					}

					var response []byte
					switch command {
					case "HELLO":
						response = []byte("-ERR unknown command 'hello'\r\n")
					case "PING":
						response = []byte("+PONG\r\n")
					default:
						response = []byte("+OK\r\n")
					}

					if _, err := conn.Write(response); err != nil {
						return
					}
				}
			}(conn)
		}
	}()

	return ln.Addr().String()
}

func readRESPCommand(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	if len(line) == 0 || line[0] != '*' {
		return "", io.ErrUnexpectedEOF
	}

	argCount, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "*")))
	if err != nil {
		return "", err
	}

	var command string
	for i := 0; i < argCount; i++ {
		bulkLen, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		if len(bulkLen) == 0 || bulkLen[0] != '$' {
			return "", io.ErrUnexpectedEOF
		}

		size, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(bulkLen, "$")))
		if err != nil {
			return "", err
		}

		payload := make([]byte, size+2)
		if _, err := io.ReadFull(reader, payload); err != nil {
			return "", err
		}
		if i == 0 {
			command = strings.ToUpper(strings.TrimSpace(string(payload[:size])))
		}
	}

	return command, nil
}
