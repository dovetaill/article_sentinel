package bootstrap

import (
	"io"
	"log/slog"
	"reflect"
	"testing"

	"github.com/dovetaill/article-sentinel/internal/app/lifecycle"
	"github.com/dovetaill/article-sentinel/pkg/config"
)

func TestBuildServerRuntimeReturnsConfigAndLoggerOnly(t *testing.T) {
	origLoadConfig := loadConfigFn
	origNewLogger := newLoggerFn
	t.Cleanup(func() {
		loadConfigFn = origLoadConfig
		newLoggerFn = origNewLogger
	})

	wantConfig := &config.Config{App: config.AppConfig{Name: "article-sentinel"}}
	wantLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	loggerClosed := false

	loadConfigFn = func(path string) (*config.Config, error) {
		return wantConfig, nil
	}
	newLoggerFn = func(cfg config.LogConfig) (*slog.Logger, func() error, error) {
		return wantLogger, func() error {
			loggerClosed = true
			return nil
		}, nil
	}

	rt, err := BuildServerRuntime("configs/config.yaml")
	if err != nil {
		t.Fatalf("BuildServerRuntime() error = %v", err)
	}

	if rt.Config != wantConfig {
		t.Fatal("runtime.Config does not match loaded config")
	}
	if rt.Logger != wantLogger {
		t.Fatal("runtime.Logger does not match created logger")
	}
	if _, ok := reflect.TypeOf(Runtime{}).FieldByName("Resources"); ok {
		t.Fatal("Runtime.Resources is present, want reduced runtime without resources")
	}
	if err := rt.Shutdown(); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if !loggerClosed {
		t.Fatal("logger closer was not registered")
	}
}

func TestShutdownRunsClosersInReverseOrder(t *testing.T) {
	calls := make([]string, 0, 2)
	closers := []lifecycle.Closer{
		func() error {
			calls = append(calls, "first")
			return nil
		},
		func() error {
			calls = append(calls, "second")
			return nil
		},
	}

	if err := lifecycle.Shutdown(closers...); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("close call count = %d, want %d", len(calls), 2)
	}
	if calls[0] != "second" || calls[1] != "first" {
		t.Fatalf("close order = %v, want [second first]", calls)
	}
}
