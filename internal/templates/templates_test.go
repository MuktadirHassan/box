package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadTerminalTools(t *testing.T) {
	template, err := Resolve("terminal-tools", "ubuntu:24.04")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if template.Family != "ubuntu" || template.Version != 1 {
		t.Errorf("template = %#v", template)
	}
	if _, err := Resolve("missing", "ubuntu:24.04"); err == nil {
		t.Error("Resolve() error = nil for missing template")
	}
	if _, err := Resolve("terminal-tools", "archlinux:latest"); err == nil {
		t.Error("Resolve() error = nil for unsupported image family")
	}
}

func TestImageFamilyUsesImageRepositoryName(t *testing.T) {
	for image, want := range map[string]string{"ubuntu:24.04": "ubuntu", "docker.io/library/ubuntu:24.04": "ubuntu", "quay.io/example/custom@sha256:abc": "custom"} {
		if got := imageFamily(image); got != want {
			t.Errorf("imageFamily(%q) = %q, want %q", image, got, want)
		}
	}
}

func TestBuildContextIncludesTemplateAssets(t *testing.T) {
	template, err := Resolve("terminal-tools", "ubuntu:24.04")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if err := template.BuildContext(directory); err != nil {
		t.Fatalf("BuildContext() error = %v", err)
	}
	containerfile, err := os.ReadFile(filepath.Join(directory, "Containerfile"))
	if err != nil {
		t.Fatal(err)
	}
	for _, packageName := range []string{"fish", "jq", "neovim", "tmux", "ripgrep", "starship", "wl-clipboard"} {
		if !strings.Contains(string(containerfile), packageName) {
			t.Errorf("Containerfile does not install %q", packageName)
		}
	}
	if _, err := os.Stat(filepath.Join(directory, "dotfiles", ".config", "fish", "config.fish")); err != nil {
		t.Errorf("dotfiles are missing from build context: %v", err)
	}
}
