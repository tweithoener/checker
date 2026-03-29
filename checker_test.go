package checker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	c := New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
	if c.name == "" {
		t.Error("New() set an empty name")
	}
	if c.interval != 5*time.Minute {
		t.Errorf("New() default interval mismatch: got %v, want %v", c.interval, 5*time.Minute)
	}
	if c.running {
		t.Error("New() returned a running checker")
	}
}

func TestChecker_AddCheck(t *testing.T) {
	c := New()
	mockCheck := func(ctx context.Context, cs CheckState) (State, string) {
		return OK, ""
	}

	err := c.AddCheck("test-check", mockCheck)
	if err != nil {
		t.Errorf("AddCheck() returned error: %v", err)
	}

	if len(c.checks) != 1 {
		t.Errorf("AddCheck() did not add check, count: %d", len(c.checks))
	}

	if c.checks[0].Name != "test-check" {
		t.Errorf("AddCheck() name mismatch: got %s, want %s", c.checks[0].Name, "test-check")
	}

	// Test duplicate check
	err = c.AddCheck("test-check", mockCheck)
	if err == nil {
		t.Error("AddCheck() did not return error for duplicate name")
	}

	// Test adding check while running
	c.running = true
	err = c.AddCheck("another-check", mockCheck)
	if err == nil {
		t.Error("AddCheck() did not return error when running")
	}
}

func TestChecker_AddNotifier(t *testing.T) {
	c := New()
	mockNotifier := func(ctx context.Context, cs CheckState) {}

	err := c.AddNotifier(mockNotifier)
	if err != nil {
		t.Errorf("AddNotifier() returned error: %v", err)
	}

	if len(c.notifiers) != 1 {
		t.Errorf("AddNotifier() did not add notifier, count: %d", len(c.notifiers))
	}

	// Test adding notifier while running
	c.running = true
	err = c.AddNotifier(mockNotifier)
	if err == nil {
		t.Error("AddNotifier() did not return error when running")
	}
}

func TestChecker_SetNameAndInterval(t *testing.T) {
	c := New()
	c.SetName("custom-name")
	if c.name != "custom-name" {
		t.Errorf("SetName() failed: got %s, want %s", c.name, "custom-name")
	}

	c.SetInterval(10 * time.Second)
	if c.interval != 10*time.Second {
		t.Errorf("SetInterval() failed: got %v, want %v", c.interval, 10*time.Second)
	}
}

func TestChecker_StartAndShutdown(t *testing.T) {
	c := New()
	mockCheck := func(ctx context.Context, cs CheckState) (State, string) {
		return OK, ""
	}

	// Start without checks
	err := c.Start()
	if err == nil {
		t.Error("Start() should return error with no checks")
	}

	_ = c.AddCheck("test", mockCheck)
	c.SetInterval(100 * time.Millisecond)

	err = c.Start()
	if err != nil {
		t.Errorf("Start() failed: %v", err)
	}

	if !c.running {
		t.Error("Checker not running after Start()")
	}

	err = c.Start()
	if err == nil {
		t.Error("Start() should return error if already running")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = c.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}

	if c.running {
		t.Error("Checker still running after Shutdown()")
	}
}

func TestChecker_RunCheck(t *testing.T) {
	c := New()
	var checkCalled int32
	mockCheck := func(ctx context.Context, cs CheckState) (State, string) {
		atomic.AddInt32(&checkCalled, 1)
		return Warn, "warning message"
	}

	var notifierCalled int32
	mockNotifier := func(ctx context.Context, cs CheckState) {
		atomic.AddInt32(&notifierCalled, 1)
	}

	_ = c.AddCheck("test-check", mockCheck)
	_ = c.AddNotifier(mockNotifier)

	meta := c.checks[0]
	c.runCheck(meta)

	// wait for async runCheck and notifier
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&checkCalled) != 1 {
		t.Errorf("Check function not called: got %d, want 1", checkCalled)
	}

	if atomic.LoadInt32(&notifierCalled) != 1 {
		t.Errorf("Notifier not called on warning: got %d, want 1", notifierCalled)
	}

	if meta.State != Warn {
		t.Errorf("Meta state not updated: got %s, want %s", meta.State, Warn)
	}

	if meta.Message != "warning message" {
		t.Errorf("Meta message not updated: got %s, want %s", meta.Message, "warning message")
	}

	if meta.Streak != 1 {
		t.Errorf("Meta streak mismatch: got %d, want 1", meta.Streak)
	}
}
