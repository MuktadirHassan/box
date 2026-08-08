package cli

import (
	"bytes"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/MuktadirHassan/box/internal/backend"
	"github.com/MuktadirHassan/box/internal/box"
	"github.com/MuktadirHassan/box/internal/store"
	"github.com/MuktadirHassan/box/internal/templates"
	"github.com/MuktadirHassan/box/internal/version"
)

type setupCatalog struct{ descriptor templates.Descriptor }

func (c setupCatalog) List() ([]templates.Descriptor, error) {
	return []templates.Descriptor{c.descriptor}, nil
}
func (c setupCatalog) Resolve(string) (templates.Resolved, error) {
	return setupResolved{descriptor: c.descriptor}, nil
}

type setupResolved struct{ descriptor templates.Descriptor }

func (r setupResolved) Descriptor() templates.Descriptor { return r.descriptor }
func (setupResolved) Validate(templates.Request) error   { return nil }
func (setupResolved) BuildContext(string) error          { return nil }

func TestRootCommandReportsVersion(t *testing.T) {
	command := NewRootCommand(store.New(filepath.Join(t.TempDir(), "boxes")), nil)
	output := &bytes.Buffer{}
	command.SetOut(output)
	command.SetArgs([]string{"--version"})

	if err := command.Execute(); err != nil {
		t.Fatalf("version command error = %v", err)
	}
	want := "box version " + version.Version + "\n"
	if output.String() != want {
		t.Errorf("version output = %q, want %q", output.String(), want)
	}
}

func TestBoxLifecycleCommandsUseInjectedStore(t *testing.T) {
	definitions := store.New(filepath.Join(t.TempDir(), "boxes"))
	command := NewRootCommand(definitions, nil)
	output := &bytes.Buffer{}
	command.SetOut(output)
	command.SetErr(output)

	command.SetArgs([]string{"create", "demo"})
	if err := command.Execute(); err != nil {
		t.Fatalf("create command error = %v", err)
	}
	if !strings.Contains(output.String(), "Created box \"demo\"") {
		t.Errorf("create output = %q", output.String())
	}

	output.Reset()
	command.SetArgs([]string{"list"})
	if err := command.Execute(); err != nil {
		t.Fatalf("list command error = %v", err)
	}
	if output.String() != "demo\tcreated\t-\n" {
		t.Errorf("list output = %q", output.String())
	}

	output.Reset()
	command.SetArgs([]string{"inspect", "demo"})
	if err := command.Execute(); err != nil {
		t.Fatalf("inspect command error = %v", err)
	}
	if output.String() != "name: demo\nstate: created\nbackend: podman\nversion: 1\nruntime: not created\n" {
		t.Errorf("inspect output = %q", output.String())
	}

	output.Reset()
	command.SetArgs([]string{"delete", "demo", "--purge"})
	if err := command.Execute(); err != nil {
		t.Fatalf("delete command error = %v", err)
	}
	if !strings.Contains(output.String(), "Deleted box \"demo\"") {
		t.Errorf("delete output = %q", output.String())
	}
	if _, err := definitions.Load("demo"); err == nil {
		t.Error("Load() error = nil after purge")
	}
}

func TestSetupRequiresConfirmationBeforeSaving(t *testing.T) {
	definitions := store.New(filepath.Join(t.TempDir(), "boxes"))
	if err := definitions.Create(box.NewDefinition("demo")); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	command := NewRootCommand(definitions, nil)
	output := &bytes.Buffer{}
	command.SetOut(output)
	command.SetErr(output)
	command.SetArgs([]string{"setup", "demo", "--image", "archlinux:latest", "--user", "dev"})

	if err := command.Execute(); err != ErrSetupConfirmation {
		t.Fatalf("setup error = %v, want confirmation error", err)
	}
	if !strings.Contains(output.String(), "image: archlinux:latest") {
		t.Errorf("setup output = %q, want resolved image", output.String())
	}

	definition, err := definitions.Load("demo")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if definition.State != box.CreatedState {
		t.Errorf("State = %q, want %q before confirmation", definition.State, box.CreatedState)
	}
}

func TestSetupGeneratesUserWithoutHostIdentity(t *testing.T) {
	definitions := store.New(filepath.Join(t.TempDir(), "boxes"))
	if err := definitions.Create(box.NewDefinition("demo")); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	command := NewRootCommand(definitions, nil)
	output := &bytes.Buffer{}
	command.SetOut(output)
	command.SetArgs([]string{"setup", "demo"})
	if err := command.Execute(); err != ErrSetupConfirmation {
		t.Fatalf("setup error = %v, want confirmation error", err)
	}
	if !strings.Contains(output.String(), "user: ") {
		t.Errorf("setup output = %q, want generated user", output.String())
	}
}

func TestSetupSavesResolvedConfiguration(t *testing.T) {
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
	command.SetArgs([]string{"setup", "demo", "--image", "ubuntu:24.04", "--user", "tamim", "--mount", "/work:~/workspace", "--mount", "/data:/workspace/data", "--memory", "4g", "--pids-limit", "512", "--template", "ubuntu-24.04-terminal-tools", "--shell", "fish", "--prompt", "starship", "--clipboard", "--yes"})
	if err := command.Execute(); err != nil {
		t.Fatalf("setup command error = %v", err)
	}

	definition, err := definitions.Load("demo")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if definition.State != box.ReadyState {
		t.Errorf("State = %q, want %q", definition.State, box.ReadyState)
	}
	if definition.Backend != box.PodmanBackend {
		t.Errorf("Backend = %q, want %q", definition.Backend, box.PodmanBackend)
	}
	if definition.Configuration.User != "tamim" {
		t.Errorf("User = %q, want tamim", definition.Configuration.User)
	}
	if definition.Configuration.Image != "ubuntu:24.04" {
		t.Errorf("Image = %q, want ubuntu:24.04", definition.Configuration.Image)
	}
	if definition.Configuration.Shell != "fish" || definition.Configuration.Prompt != "starship" {
		t.Errorf("shell configuration = %q/%q, want fish/starship", definition.Configuration.Shell, definition.Configuration.Prompt)
	}
	if definition.Configuration.Limits.Memory != "4g" || definition.Configuration.Limits.PIDsLimit != 512 {
		t.Errorf("Limits = %#v, want memory 4g and pids 512", definition.Configuration.Limits)
	}
	if !definition.Configuration.Integrations.Clipboard {
		t.Error("Clipboard = false, want true")
	}
	if definition.Configuration.Template != "ubuntu-24.04-terminal-tools" {
		t.Errorf("Template = %q, want %q", definition.Configuration.Template, "ubuntu-24.04-terminal-tools")
	}
	wantMounts := []box.Mount{{Source: "/work", Destination: "/home/tamim/workspace"}, {Source: "/data", Destination: "/workspace/data"}}
	if mounts := definition.Configuration.Mounts; !reflect.DeepEqual(mounts, wantMounts) {
		t.Errorf("Mounts = %#v, want %#v", mounts, wantMounts)
	}
}

func TestSetupPersistsCatalogCanonicalIDAndRecreatesOnTemplateChange(t *testing.T) {
	definitions := store.New(filepath.Join(t.TempDir(), "boxes"))
	if err := definitions.Create(box.NewDefinition("demo")); err != nil {
		t.Fatal(err)
	}
	runtime := &recordingBackend{}
	registry, err := backend.NewRegistry(runtime)
	if err != nil {
		t.Fatal(err)
	}
	catalog := setupCatalog{descriptor: templates.Descriptor{ID: "canonical-template", Description: "Canonical label"}}
	command := NewRootCommandWithCatalog(definitions, registry, catalog, nil)
	command.SetArgs([]string{"setup", "demo", "--template", "legacy-template", "--image", "ubuntu:24.04", "--user", "dev", "--yes"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	definition, err := definitions.Load("demo")
	if err != nil {
		t.Fatal(err)
	}
	if definition.Configuration.Template != "canonical-template" {
		t.Errorf("Template = %q", definition.Configuration.Template)
	}
	if runtime.created.Name != "demo" {
		t.Fatalf("created definition = %#v", runtime.created)
	}

	command = NewRootCommandWithCatalog(definitions, registry, catalog, nil)
	command.SetArgs([]string{"setup", "demo", "--template", "legacy-template", "--image", "ubuntu:24.04@sha256:" + strings.Repeat("a", 64), "--user", "dev", "--yes"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if runtime.deleted.Name != "demo" || runtime.created.Configuration.Template != "canonical-template" {
		t.Errorf("recreation definitions deleted=%#v created=%#v", runtime.deleted, runtime.created)
	}
}

func TestDeleteRequiresPurge(t *testing.T) {
	command := NewRootCommand(store.New(filepath.Join(t.TempDir(), "boxes")), nil)
	command.SetArgs([]string{"delete", "demo"})

	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "--purge") {
		t.Errorf("delete error = %v, want --purge requirement", err)
	}
}
