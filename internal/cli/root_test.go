package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

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
	if output.String() != "demo\tcreated\n" {
		t.Errorf("list output = %q", output.String())
	}

	output.Reset()
	command.SetArgs([]string{"inspect", "demo"})
	if err := command.Execute(); err != nil {
		t.Fatalf("inspect command error = %v", err)
	}
	if output.String() != "name: demo\nstate: created\nversion: 1\n" {
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

func TestDeleteRequiresPurge(t *testing.T) {
	command := newRootCommand(store.New(filepath.Join(t.TempDir(), "boxes")))
	command.SetArgs([]string{"delete", "demo"})

	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "--purge") {
		t.Errorf("delete error = %v, want --purge requirement", err)
	}
}
