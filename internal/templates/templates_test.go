package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	assets "github.com/MuktadirHassan/box/templates"
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

type catalogStub struct {
	descriptors []Descriptor
	resolved    map[string]Resolved
}

func (c catalogStub) List() ([]Descriptor, error) {
	return append([]Descriptor(nil), c.descriptors...), nil
}
func (c catalogStub) Resolve(id string) (Resolved, error) {
	if resolved, ok := c.resolved[id]; ok {
		return resolved, nil
	}
	return nil, fmt.Errorf("unsupported template %q", id)
}

type resolvedStub struct {
	descriptor  Descriptor
	validateErr error
	built       bool
}

func (r *resolvedStub) Descriptor() Descriptor { return r.descriptor }
func (r *resolvedStub) Validate(Request) error { return r.validateErr }
func (r *resolvedStub) BuildContext(destination string) error {
	r.built = true
	return os.WriteFile(filepath.Join(destination, "Containerfile"), []byte("FROM scratch\n"), 0644)
}

func TestCatalogContractEmbedded(t *testing.T) {
	catalog := NewEmbeddedCatalog(assets.FS())
	first, err := catalog.List()
	if err != nil {
		t.Fatal(err)
	}
	second, err := catalog.List()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) || len(first) != 1 || first[0].ID != "ubuntu-24.04-terminal-tools" {
		t.Fatalf("descriptors are not deterministic and canonical: %#v / %#v", first, second)
	}
	if first[0].Description != "Ubuntu 24.04 — Terminal tools" || !contains(first[0].Shells, "fish") || !contains(first[0].Prompts, "starship") {
		t.Errorf("descriptor metadata = %#v", first[0])
	}
	for _, id := range []string{"ubuntu-22.04-terminal-tools", "missing"} {
		if _, err := catalog.Resolve(id); err == nil {
			t.Errorf("Resolve(%q) error = nil", id)
		}
	}
	canonical, err := catalog.Resolve("ubuntu-24.04-terminal-tools")
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := catalog.Resolve("terminal-tools")
	if err != nil {
		t.Fatal(err)
	}
	if canonical.Descriptor().ID != legacy.Descriptor().ID {
		t.Errorf("legacy ID = %q, canonical = %q", legacy.Descriptor().ID, canonical.Descriptor().ID)
	}
	for _, request := range []Request{{Image: "ubuntu:22.04", Shell: "fish", Prompt: "starship"}, {Image: "ubuntu:24.04", Shell: "bash", Prompt: "unknown"}} {
		if err := canonical.Validate(request); err == nil {
			t.Errorf("Validate(%#v) error = nil", request)
		}
	}
	if err := canonical.Validate(Request{Image: "ubuntu:24.04@sha256:" + strings.Repeat("a", 64), Shell: "fish", Prompt: "starship"}); err != nil {
		t.Fatal(err)
	}
	destination := t.TempDir()
	if err := canonical.BuildContext(destination); err != nil {
		t.Fatal(err)
	}
	for _, asset := range []string{"Containerfile", "initialize-home", "dotfiles"} {
		if _, err := os.Stat(filepath.Join(destination, asset)); err != nil {
			t.Errorf("missing materialized asset %q: %v", asset, err)
		}
	}
}

func TestRegistryRejectsNilAndDuplicateProviders(t *testing.T) {
	var nilCatalog *catalogStub
	if _, err := NewRegistry(nilCatalog); err == nil {
		t.Error("nil provider accepted")
	}
	provider := catalogStub{descriptors: []Descriptor{{ID: "same"}}}
	if _, err := NewRegistry(provider, provider); err == nil {
		t.Error("duplicate provider accepted")
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
