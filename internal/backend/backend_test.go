package backend

import (
	"context"
	"strings"
	"testing"

	"github.com/MuktadirHassan/box/internal/box"
)

type testBackend struct {
	name box.Backend
}

func (b testBackend) Name() box.Backend            { return b.name }
func (testBackend) Validate(context.Context) error { return nil }
func (testBackend) Create(context.Context, box.Definition) (box.RuntimeMetadata, error) {
	return box.RuntimeMetadata{}, nil
}
func (testBackend) Start(context.Context, box.RuntimeMetadata) error { return nil }
func (testBackend) Stop(context.Context, box.RuntimeMetadata) error  { return nil }
func (testBackend) Inspect(context.Context, box.RuntimeMetadata) (box.RuntimeStatus, error) {
	return box.RuntimeStatus{}, nil
}
func (testBackend) Delete(context.Context, box.Definition, box.RuntimeMetadata, DeleteOptions) error {
	return nil
}
func (testBackend) Enter(context.Context, box.RuntimeMetadata) error          { return nil }
func (testBackend) Exec(context.Context, box.RuntimeMetadata, []string) error { return nil }

func TestRegistryReturnsRegisteredBackend(t *testing.T) {
	podman := testBackend{name: box.PodmanBackend}
	registry, err := NewRegistry(podman)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	got, err := registry.Get(box.PodmanBackend)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != podman {
		t.Errorf("Get() = %#v, want %#v", got, podman)
	}
}

func TestRegistryRejectsDuplicateBackends(t *testing.T) {
	_, err := NewRegistry(testBackend{name: box.PodmanBackend}, testBackend{name: box.PodmanBackend})
	if err == nil || !strings.Contains(err.Error(), "already registered") {
		t.Errorf("NewRegistry() error = %v, want duplicate backend error", err)
	}
}
