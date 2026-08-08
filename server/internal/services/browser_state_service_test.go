package services

import (
	"errors"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/zhitu/server/internal/models"
	"gorm.io/gorm"
)

func newBrowserStateTestService(t *testing.T) *BrowserStateService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.BrowserState{}); err != nil {
		t.Fatal(err)
	}
	return NewBrowserStateService(db)
}

func TestBrowserStateServiceIsolatesWorkspaces(t *testing.T) {
	svc := newBrowserStateTestService(t)
	if err := svc.Put(1, "copilot_sessions", []byte(`[{"id":"a"}]`)); err != nil {
		t.Fatal(err)
	}
	if err := svc.Put(2, "copilot_sessions", []byte(`[{"id":"b"}]`)); err != nil {
		t.Fatal(err)
	}
	a, err := svc.Get(1, "copilot_sessions")
	if err != nil {
		t.Fatal(err)
	}
	b, err := svc.Get(2, "copilot_sessions")
	if err != nil {
		t.Fatal(err)
	}
	if string(a) != `[{"id":"a"}]` || string(b) != `[{"id":"b"}]` {
		t.Fatalf("unexpected isolated values: %s / %s", a, b)
	}
	if err := svc.Delete(1, "copilot_sessions"); err != nil {
		t.Fatal(err)
	}
	a, _ = svc.Get(1, "copilot_sessions")
	b, _ = svc.Get(2, "copilot_sessions")
	if string(a) != "[]" || string(b) != `[{"id":"b"}]` {
		t.Fatalf("delete crossed workspace boundary: %s / %s", a, b)
	}
}

func TestBrowserStateServiceValidation(t *testing.T) {
	svc := newBrowserStateTestService(t)
	if err := svc.Put(1, "other", []byte(`[]`)); !errors.Is(err, ErrBrowserStateKey) {
		t.Fatalf("expected key error, got %v", err)
	}
	if err := svc.Put(1, "copilot_sessions", []byte(`not-json`)); !errors.Is(err, ErrBrowserStateInvalid) {
		t.Fatalf("expected JSON error, got %v", err)
	}
	tooLarge := []byte(`"` + strings.Repeat("a", MaxBrowserStateBytes) + `"`)
	if err := svc.Put(1, "copilot_sessions", tooLarge); !errors.Is(err, ErrBrowserStateTooLarge) {
		t.Fatalf("expected size error, got %v", err)
	}
}
