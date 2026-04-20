package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dovetaill/article-sentinel/pkg/config"
)

func TestNewReturnsJSONLogger(t *testing.T) {
	logDir := t.TempDir()
	cfg := config.LogConfig{
		Level:      "info",
		Output:     "file",
		Dir:        logDir,
		Filename:   "app.log",
		MaxSizeMB:  10,
		MaxBackups: 3,
		MaxAgeDays: 7,
	}

	logger, cleanup, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() {
		if err := cleanup(); err != nil {
			t.Fatalf("cleanup() error = %v", err)
		}
	}()

	logger.Info("hello", "component", "logger")

	data, err := os.ReadFile(filepath.Join(logDir, "app.log"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if payload["msg"] != "hello" {
		t.Fatalf("payload[msg] = %v, want %q", payload["msg"], "hello")
	}
	if payload["component"] != "logger" {
		t.Fatalf("payload[component] = %v, want %q", payload["component"], "logger")
	}
}

func TestNewSupportsStdoutOnly(t *testing.T) {
	logDir := t.TempDir()
	stdoutFile, err := os.CreateTemp(t.TempDir(), "stdout-*.log")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer stdoutFile.Close()

	oldStdout := os.Stdout
	os.Stdout = stdoutFile
	defer func() {
		os.Stdout = oldStdout
	}()

	cfg := config.LogConfig{
		Level:    "info",
		Output:   "stdout",
		Dir:      logDir,
		Filename: "app.log",
	}

	logger, cleanup, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() {
		if err := cleanup(); err != nil {
			t.Fatalf("cleanup() error = %v", err)
		}
	}()

	logger.Info("stdout-only")

	if _, err := os.Stat(filepath.Join(logDir, "app.log")); !os.IsNotExist(err) {
		t.Fatalf("Stat() error = %v, want not exists", err)
	}
}

func TestDailyRotatorCallsRotateAfterDayChange(t *testing.T) {
	base := time.Date(2026, time.March, 18, 23, 59, 0, 0, time.Local)
	ticker := &fakeTicker{ch: make(chan time.Time, 2)}
	rotator := &fakeRotator{called: make(chan struct{}, 1)}

	cleanup := startDailyRotation(func() time.Time {
		return base
	}, func(time.Duration) rotationTicker {
		return ticker
	}, rotator)
	defer func() {
		if err := cleanup(); err != nil {
			t.Fatalf("cleanup() error = %v", err)
		}
	}()

	ticker.ch <- base.Add(30 * time.Second)
	select {
	case <-rotator.called:
		t.Fatal("Rotate() called before day change")
	case <-time.After(50 * time.Millisecond):
	}

	ticker.ch <- base.Add(2 * time.Minute)
	select {
	case <-rotator.called:
	case <-time.After(time.Second):
		t.Fatal("Rotate() was not called after day change")
	}
}

type fakeTicker struct {
	ch chan time.Time
}

func (t *fakeTicker) C() <-chan time.Time {
	return t.ch
}

func (t *fakeTicker) Stop() {}

type fakeRotator struct {
	count  int
	called chan struct{}
}

func (r *fakeRotator) Rotate() error {
	r.count++
	if r.count == 1 {
		r.called <- struct{}{}
	}
	return nil
}
