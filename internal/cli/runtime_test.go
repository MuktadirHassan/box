package cli

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"slices"
	"testing"

	"github.com/MuktadirHassan/box/internal/backend"
	"github.com/MuktadirHassan/box/internal/box"
	"github.com/MuktadirHassan/box/internal/store"
	"github.com/MuktadirHassan/box/internal/ui"
)

type recordingBackend struct {
	created            box.Definition
	createErr          error
	deleted            box.Definition
	deletedRuntime     box.RuntimeMetadata
	deleteOptions      backend.DeleteOptions
	deleteErr          error
	configurationError error
}

func (b *recordingBackend) Name() box.Backend            { return box.PodmanBackend }
func (*recordingBackend) Validate(context.Context) error { return nil }
func (b *recordingBackend) ValidateConfiguration(box.Configuration) error {
	return b.configurationError
}
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
func (*recordingBackend) Enter(context.Context, box.Definition, box.RuntimeMetadata) error {
	return nil
}
func (*recordingBackend) Exec(context.Context, box.RuntimeMetadata, []string) error { return nil }

type setupPresenter struct {
	configure func(box.Definition) box.Definition
	confirmed bool
	labels    []string
	steps     []*setupStep
}

type setupStep struct{ result string }

func (s *setupStep) Success() { s.result = "success" }
func (s *setupStep) Fail()    { s.result = "failed" }

func (p *setupPresenter) ConfigureInitial(definition box.Definition) (box.Definition, error) {
	return p.configure(definition), nil
}
func (p *setupPresenter) ConfirmSetup() error { p.confirmed = true; return nil }
func (p *setupPresenter) StartStep(_ io.Writer, label string) (ui.Step, error) {
	step := &setupStep{}
	p.labels = append(p.labels, label)
	p.steps = append(p.steps, step)
	return step, nil
}
func (*setupPresenter) ShowDefinition(io.Writer, box.Definition) error        { return nil }
func (*setupPresenter) ShowRuntime(io.Writer, box.RuntimeState, string) error { return nil }
func (*setupPresenter) ShowList(io.Writer, []box.Definition) error            { return nil }
func (*setupPresenter) ShowWarning(io.Writer, string) error                   { return nil }
func (*setupPresenter) ShowSuccess(io.Writer, string) error                   { return nil }

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

func TestSetupReportsOrderedSuccessfulSteps(t *testing.T) {
	definitions := store.New(filepath.Join(t.TempDir(), "boxes"))
	if err := definitions.Create(box.NewDefinition("demo")); err != nil {
		t.Fatal(err)
	}
	runtime := &recordingBackend{}
	registry, err := backend.NewRegistry(runtime)
	if err != nil {
		t.Fatal(err)
	}
	presenter := &setupPresenter{}
	command := NewRootCommand(definitions, registry, presenter)
	command.SetArgs([]string{"setup", "demo", "--image", "ubuntu:24.04", "--user", "dev", "--template", "ubuntu-24.04-terminal-tools", "--yes"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}

	want := []string{"Checking Podman", "Building template and creating box runtime", "Saving box configuration"}
	if !slices.Equal(presenter.labels, want) {
		t.Fatalf("setup steps = %q, want %q", presenter.labels, want)
	}
	for index, step := range presenter.steps {
		if step.result != "success" {
			t.Errorf("step %q result = %q, want success", presenter.labels[index], step.result)
		}
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

func TestSetupValidatesInteractiveConfigurationBeforeConfirmation(t *testing.T) {
	definitions := store.New(filepath.Join(t.TempDir(), "boxes"))
	if err := definitions.Create(box.NewDefinition("demo")); err != nil {
		t.Fatal(err)
	}
	configurationErr := errors.New("a non-default shell requires an environment template")
	runtime := &recordingBackend{configurationError: configurationErr}
	registry, err := backend.NewRegistry(runtime)
	if err != nil {
		t.Fatal(err)
	}
	presenter := &setupPresenter{configure: func(definition box.Definition) box.Definition {
		definition.Configuration.Shell = "fish"
		return definition
	}}
	command := NewRootCommand(definitions, registry, presenter)
	command.SetArgs([]string{"setup", "demo"})
	if err := command.Execute(); !errors.Is(err, configurationErr) {
		t.Fatalf("setup error = %v, want configuration validation error", err)
	}
	if presenter.confirmed {
		t.Error("setup asked for confirmation after configuration validation failed")
	}
	if runtime.created.Name != "" {
		t.Errorf("runtime was created for invalid interactive configuration: %#v", runtime.created)
	}
}

func TestSetupRejectsUnsupportedTemplateImageBeforeRuntimeChanges(t *testing.T) {
	definitions := store.New(filepath.Join(t.TempDir(), "boxes"))
	if err := definitions.Create(box.NewDefinition("demo")); err != nil {
		t.Fatal(err)
	}
	runtime := &recordingBackend{configurationError: errors.New("template does not support image family")}
	registry, err := backend.NewRegistry(runtime)
	if err != nil {
		t.Fatal(err)
	}
	command := NewRootCommand(definitions, registry)
	command.SetArgs([]string{"setup", "demo", "--image", "archlinux:latest", "--template", "ubuntu-24.04-terminal-tools", "--yes"})
	if err := command.Execute(); err == nil {
		t.Fatal("setup error = nil for an unsupported template image")
	}
	if runtime.created.Name != "" {
		t.Errorf("runtime was created for unsupported template image: %#v", runtime.created)
	}
}

func TestSetupRefreshTemplateRecreatesRuntime(t *testing.T) {
	definitions := store.New(filepath.Join(t.TempDir(), "boxes"))
	definition := box.NewDefinition("demo")
	definition.State = box.ReadyState
	definition.Configuration = box.DefaultConfiguration()
	definition.Configuration.User = "dev"
	definition.Configuration.Template = "ubuntu-24.04-terminal-tools"
	if err := definitions.Create(definition); err != nil {
		t.Fatal(err)
	}
	if err := definitions.SaveMetadata("demo", box.Metadata{Runtime: box.RuntimeMetadata{Backend: box.PodmanBackend, ID: "old-runtime", State: box.RuntimeRunning}}); err != nil {
		t.Fatal(err)
	}
	runtime := &recordingBackend{}
	registry, err := backend.NewRegistry(runtime)
	if err != nil {
		t.Fatal(err)
	}
	command := NewRootCommand(definitions, registry)
	command.SetArgs([]string{"setup", "demo", "--refresh-template", "--yes"})
	if err := command.Execute(); err != nil {
		t.Fatalf("setup refresh error = %v", err)
	}
	if runtime.deleted.Name != "demo" || runtime.created.Configuration.TemplateRevision != 1 {
		t.Errorf("refresh did not recreate with a new revision: deleted=%#v created=%#v", runtime.deleted, runtime.created)
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
	presenter := &setupPresenter{}
	command := NewRootCommand(definitions, registry, presenter)
	command.SetArgs([]string{"setup", "demo", "--image", "archlinux:latest", "--yes"})
	if err := command.Execute(); !errors.Is(err, createErr) {
		t.Fatalf("setup error = %v, want create error", err)
	}
	wantSteps := []string{"Checking Podman", "Removing existing runtime", "Creating box runtime"}
	if !slices.Equal(presenter.labels, wantSteps) {
		t.Fatalf("setup steps = %q, want %q", presenter.labels, wantSteps)
	}
	if presenter.steps[0].result != "success" || presenter.steps[1].result != "success" || presenter.steps[2].result != "failed" {
		t.Errorf("setup step results = %q, %q, %q", presenter.steps[0].result, presenter.steps[1].result, presenter.steps[2].result)
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
