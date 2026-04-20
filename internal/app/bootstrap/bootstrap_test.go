package bootstrap

import (
	"io"
	"log/slog"
	"testing"

	"github.com/dovetaill/article-sentinel/internal/app/lifecycle"
	"github.com/dovetaill/article-sentinel/pkg/config"
	"github.com/dovetaill/article-sentinel/pkg/database"
)

func TestBuildServerRuntimeReturnsSharedResources(t *testing.T) {
	origLoadConfig := loadConfigFn
	origNewLogger := newLoggerFn
	origBootstrapDatabase := bootstrapDatabaseFn
	t.Cleanup(func() {
		loadConfigFn = origLoadConfig
		newLoggerFn = origNewLogger
		bootstrapDatabaseFn = origBootstrapDatabase
	})

	wantConfig := &config.Config{App: config.AppConfig{Name: "article-sentinel"}}
	wantLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	wantResources := &database.Resources{}

	loadConfigFn = func(path string) (*config.Config, error) {
		return wantConfig, nil
	}
	newLoggerFn = func(cfg config.LogConfig) (*slog.Logger, func() error, error) {
		return wantLogger, func() error { return nil }, nil
	}
	bootstrapDatabaseFn = func(cfg *config.Config) (*database.Resources, error) {
		return wantResources, nil
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
	if rt.Resources != wantResources {
		t.Fatal("runtime.Resources does not match bootstrapped resources")
	}
}

func TestBuildWorkerRuntimeReturnsSharedResources(t *testing.T) {
	origLoadConfig := loadConfigFn
	origNewLogger := newLoggerFn
	origBootstrapDatabase := bootstrapDatabaseFn
	t.Cleanup(func() {
		loadConfigFn = origLoadConfig
		newLoggerFn = origNewLogger
		bootstrapDatabaseFn = origBootstrapDatabase
	})

	wantConfig := &config.Config{App: config.AppConfig{Name: "article-sentinel"}}
	wantLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	wantResources := &database.Resources{}

	loadConfigFn = func(path string) (*config.Config, error) {
		return wantConfig, nil
	}
	newLoggerFn = func(cfg config.LogConfig) (*slog.Logger, func() error, error) {
		return wantLogger, func() error { return nil }, nil
	}
	bootstrapDatabaseFn = func(cfg *config.Config) (*database.Resources, error) {
		return wantResources, nil
	}

	rt, err := BuildWorkerRuntime("configs/config.yaml")
	if err != nil {
		t.Fatalf("BuildWorkerRuntime() error = %v", err)
	}

	if rt.Config != wantConfig {
		t.Fatal("runtime.Config does not match loaded config")
	}
	if rt.Logger != wantLogger {
		t.Fatal("runtime.Logger does not match created logger")
	}
	if rt.Resources != wantResources {
		t.Fatal("runtime.Resources does not match bootstrapped resources")
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
