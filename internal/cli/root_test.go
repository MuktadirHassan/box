package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MuktadirHassan/box/internal/box"
	"github.com/MuktadirHassan/box/internal/store"
)

func TestBoxLifecycleCommandsUseInjectedStore(t *testing.T) {
	definitions := store.New(filepath.Join(t.TempDir(), "boxes"))
	command := newRootCommand(definitions)
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
	if output.String() != "name: demo\nstate: created\nbackend: podman\nversion: 1\n" {
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

	command := newRootCommand(definitions)
	output := &bytes.Buffer{}
	command.SetOut(output)
	command.SetErr(output)
	command.SetArgs([]string{"setup", "demo", "--image", "archlinux:latest"})

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

func TestSetupSavesResolvedConfiguration(t *testing.T) {
	definitions := store.New(filepath.Join(t.TempDir(), "boxes"))
	if err := definitions.Create(box.NewDefinition("demo")); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	command := newRootCommand(definitions)
	command.SetArgs([]string{"setup", "demo", "--backend", "lima", "--image", "archlinux:latest", "--mount", "/work:/workspace", "--memory", "4g", "--pids-limit", "512", "--clipboard", "--yes"})
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
	if definition.Backend != box.LimaBackend {
		t.Errorf("Backend = %q, want %q", definition.Backend, box.LimaBackend)
	}
	if definition.Configuration.Image != "archlinux:latest" {
		t.Errorf("Image = %q, want archlinux:latest", definition.Configuration.Image)
	}
	if definition.Configuration.Limits.Memory != "4g" || definition.Configuration.Limits.PIDsLimit != 512 {
		t.Errorf("Limits = %#v, want memory 4g and pids 512", definition.Configuration.Limits)
	}
	if !definition.Configuration.Integrations.Clipboard {
		t.Error("Clipboard = false, want true")
	}
	if mounts := definition.Configuration.Mounts; len(mounts) != 1 || mounts[0] != (box.Mount{Source: "/work", Destination: "/workspace"}) {
		t.Errorf("Mounts = %#v, want /work:/workspace", mounts)
	}
}

func TestDeleteRequiresPurge(t *testing.T) {
	command := newRootCommand(store.New(filepath.Join(t.TempDir(), "boxes")))
	command.SetArgs([]string{"delete", "demo"})

	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "--purge") {
		t.Errorf("delete error = %v, want --purge requirement", err)
	}
}
