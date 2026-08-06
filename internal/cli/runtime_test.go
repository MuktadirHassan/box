package cli

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/MuktadirHassan/box/internal/backend"
	"github.com/MuktadirHassan/box/internal/box"
	"github.com/MuktadirHassan/box/internal/store"
)

type recordingBackend struct {
	created box.Definition
}

func (b *recordingBackend) Name() box.Backend            { return box.PodmanBackend }
func (*recordingBackend) Validate(context.Context) error { return nil }
func (b *recordingBackend) Create(_ context.Context, definition box.Definition) (box.RuntimeMetadata, error) {
	b.created = definition
	return box.RuntimeMetadata{Backend: box.PodmanBackend, ID: "runtime-id", State: box.RuntimeCreated}, nil
}
func (*recordingBackend) Start(context.Context, box.RuntimeMetadata) error { return nil }
func (*recordingBackend) Stop(context.Context, box.RuntimeMetadata) error  { return nil }
func (*recordingBackend) Inspect(context.Context, box.RuntimeMetadata) (box.RuntimeStatus, error) {
	return box.RuntimeStatus{}, nil
}
func (*recordingBackend) Delete(context.Context, box.Definition, box.RuntimeMetadata) error {
	return nil
}
func (*recordingBackend) Enter(context.Context, box.RuntimeMetadata) error          { return nil }
func (*recordingBackend) Exec(context.Context, box.RuntimeMetadata, []string) error { return nil }

func TestSetupCreatesAndPersistsRuntimeMetadata(t *testing.T) {
	definitions := store.New(filepath.Join(t.TempDir(), "boxes"))
	if err := definitions.Create(box.NewDefinition("demo")); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	runtime := &recordingBackend{}
	registry, err := backend.NewRegistry(runtime)
	if err != nil {
		t.Fatal(err)
	}
	command := NewRootCommand(definitions, registry)
	command.SetArgs([]string{"setup", "demo", "--user", "dev", "--yes"})
	if err := command.Execute(); err != nil {
		t.Fatalf("setup error = %v", err)
	}
	if runtime.created.Name != "demo" || runtime.created.Configuration.User != "dev" {
		t.Errorf("created definition = %#v", runtime.created)
	}
	metadata, err := definitions.LoadMetadata("demo")
	if err != nil {
		t.Fatalf("LoadMetadata() error = %v", err)
	}
	if metadata.Runtime.ID != "runtime-id" {
		t.Errorf("runtime ID = %q, want runtime-id", metadata.Runtime.ID)
	}
}
