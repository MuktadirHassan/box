package cli

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/MuktadirHassan/box/internal/backend"
	"github.com/MuktadirHassan/box/internal/box"
	"github.com/MuktadirHassan/box/internal/store"
)

type recordingBackend struct {
	created        box.Definition
	createErr      error
	deleted        box.Definition
	deletedRuntime box.RuntimeMetadata
	deleteOptions  backend.DeleteOptions
	deleteErr      error
}

func (b *recordingBackend) Name() box.Backend            { return box.PodmanBackend }
func (*recordingBackend) Validate(context.Context) error { return nil }
func (b *recordingBackend) Create(_ context.Context, definition box.Definition) (box.RuntimeMetadata, error) {
	b.created = definition
	if b.createErr != nil {
		return box.RuntimeMetadata{}, b.createErr
	}
	return box.RuntimeMetadata{Backend: box.PodmanBackend, ID: "runtime-id", State: box.RuntimeCreated}, nil
}
func (*recordingBackend) Start(context.Context, box.RuntimeMetadata) error { return nil }
func (*recordingBackend) Stop(context.Context, box.RuntimeMetadata) error  { return nil }
func (*recordingBackend) Inspect(context.Context, box.RuntimeMetadata) (box.RuntimeStatus, error) {
	return box.RuntimeStatus{}, nil
}
func (b *recordingBackend) Delete(_ context.Context, definition box.Definition, metadata box.RuntimeMetadata, options backend.DeleteOptions) error {
	b.deleted = definition
	b.deletedRuntime = metadata
	b.deleteOptions = options
	return b.deleteErr
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

func TestSetupRecreatesChangedRuntimeAndPreservesPersistentData(t *testing.T) {
	definitions := store.New(filepath.Join(t.TempDir(), "boxes"))
	definition := box.NewDefinition("demo")
	definition.State = box.ReadyState
	definition.Configuration = box.DefaultConfiguration()
	definition.Configuration.User = "dev"
	if err := definitions.Create(definition); err != nil {
		t.Fatal(err)
	}
	oldMetadata := box.Metadata{Runtime: box.RuntimeMetadata{Backend: box.PodmanBackend, ID: "old-runtime", State: box.RuntimeRunning}}
	if err := definitions.SaveMetadata("demo", oldMetadata); err != nil {
		t.Fatal(err)
	}

	runtime := &recordingBackend{}
	registry, err := backend.NewRegistry(runtime)
	if err != nil {
		t.Fatal(err)
	}
	command := NewRootCommand(definitions, registry)
	command.SetArgs([]string{"setup", "demo", "--image", "archlinux:latest", "--yes"})
	if err := command.Execute(); err != nil {
		t.Fatalf("setup command error = %v", err)
	}
	if runtime.deleted.Name != definition.Name || runtime.deleted.Backend != definition.Backend || runtime.deleted.Configuration.Image != definition.Configuration.Image || runtime.deletedRuntime != oldMetadata.Runtime {
		t.Errorf("deleted runtime = %#v, %#v", runtime.deleted, runtime.deletedRuntime)
	}
	if runtime.deleteOptions.RemovePersistentData {
		t.Error("recreation removed managed persistent data")
	}
	updated, err := definitions.Load("demo")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Configuration.Image != "archlinux:latest" || updated.State != box.ReadyState {
		t.Errorf("updated definition = %#v", updated)
	}
}

func TestSetupRecordsMissingRuntimeWhenReplacementCreationFails(t *testing.T) {
	definitions := store.New(filepath.Join(t.TempDir(), "boxes"))
	definition := box.NewDefinition("demo")
	definition.State = box.ReadyState
	definition.Configuration = box.DefaultConfiguration()
	definition.Configuration.User = "dev"
	if err := definitions.Create(definition); err != nil {
		t.Fatal(err)
	}
	oldMetadata := box.Metadata{Runtime: box.RuntimeMetadata{Backend: box.PodmanBackend, ID: "old-runtime", State: box.RuntimeRunning}}
	if err := definitions.SaveMetadata("demo", oldMetadata); err != nil {
		t.Fatal(err)
	}

	createErr := errors.New("create failed")
	runtime := &recordingBackend{createErr: createErr}
	registry, err := backend.NewRegistry(runtime)
	if err != nil {
		t.Fatal(err)
	}
	command := NewRootCommand(definitions, registry)
	command.SetArgs([]string{"setup", "demo", "--image", "archlinux:latest", "--yes"})
	if err := command.Execute(); !errors.Is(err, createErr) {
		t.Fatalf("setup error = %v, want create error", err)
	}
	unchanged, err := definitions.Load("demo")
	if err != nil {
		t.Fatal(err)
	}
	if unchanged.Backend != definition.Backend || unchanged.Configuration.Image != definition.Configuration.Image || unchanged.Configuration.User != definition.Configuration.User || unchanged.State != definition.State {
		t.Errorf("definition changed after failed recreation: %#v", unchanged)
	}
	metadata, err := definitions.LoadMetadata("demo")
	if err != nil {
		t.Fatal(err)
	}
	if metadata.Runtime.ID != oldMetadata.Runtime.ID || metadata.Runtime.State != box.RuntimeMissing {
		t.Errorf("metadata after failed recreation = %#v", metadata)
	}
}
