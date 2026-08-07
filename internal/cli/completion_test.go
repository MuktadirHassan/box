package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MuktadirHassan/box/internal/store"
)

func TestCompletionGeneratesSupportedShellScripts(t *testing.T) {
	for _, shell := range []string{"bash", "fish", "zsh"} {
		t.Run(shell, func(t *testing.T) {
			command := NewRootCommand(store.New(filepath.Join(t.TempDir(), "boxes")), nil)
			output := &bytes.Buffer{}
			command.SetOut(output)
			command.SetArgs([]string{"completion", shell})

			if err := command.Execute(); err != nil {
				t.Fatalf("completion command error = %v", err)
			}
			if output.Len() == 0 || !strings.Contains(output.String(), "box") {
				t.Errorf("completion output = %q", output.String())
			}
		})
	}
}

func TestCompletionRejectsUnsupportedShell(t *testing.T) {
	command := NewRootCommand(store.New(filepath.Join(t.TempDir(), "boxes")), nil)
	command.SetArgs([]string{"completion", "powershell"})
	if err := command.Execute(); err == nil {
		t.Fatal("completion command succeeded for an unsupported shell")
	}
}
