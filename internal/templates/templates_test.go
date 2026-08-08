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
	template, err := Resolve("ubuntu-24.04-terminal-tools", "ubuntu:24.04")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if template.Family != "ubuntu" || template.Version != 1 {
		t.Errorf("template = %#v", template)
	}
	if _, err := Resolve("missing", "ubuntu:24.04"); err == nil {
		t.Error("Resolve() error = nil for missing template")
	}
	if _, err := Resolve("ubuntu-24.04-terminal-tools", "archlinux:latest"); err == nil {
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
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("descriptors are not deterministic: %#v / %#v", first, second)
	}
	if len(first) == 0 {
		t.Fatal("catalog is empty")
	}

	ids := make(map[string]struct{}, len(first))
	for index, descriptor := range first {
		if _, exists := ids[descriptor.ID]; exists {
			t.Errorf("duplicate template ID %q", descriptor.ID)
		}
		ids[descriptor.ID] = struct{}{}
		if index > 0 && first[index-1].ID >= descriptor.ID {
			t.Errorf("template IDs are not sorted: %q before %q", first[index-1].ID, descriptor.ID)
		}

		resolved, err := catalog.Resolve(descriptor.ID)
		if err != nil {
			t.Errorf("Resolve(%q) error = %v", descriptor.ID, err)
			continue
		}
		if got := resolved.Descriptor(); !reflect.DeepEqual(got, descriptor) {
			t.Errorf("Resolve(%q) descriptor = %#v, want %#v", descriptor.ID, got, descriptor)
		}

		request := Request{
			Image:  descriptor.ImageFamily + ":" + descriptor.ImageVersion,
			Shell:  descriptor.Shells[0],
			Prompt: descriptor.Prompts[0],
		}
		if err := resolved.Validate(request); err != nil {
			t.Errorf("Validate(%q, %#v) error = %v", descriptor.ID, request, err)
		}
		request.Image += "@sha256:" + strings.Repeat("a", 64)
		if err := resolved.Validate(request); err != nil {
			t.Errorf("Validate(%q, digest-qualified image) error = %v", descriptor.ID, err)
		}

		destination := t.TempDir()
		if err := resolved.BuildContext(destination); err != nil {
			t.Errorf("BuildContext(%q) error = %v", descriptor.ID, err)
			continue
		}
		for _, asset := range []string{"Containerfile", "initialize-home", "dotfiles"} {
			if _, err := os.Stat(filepath.Join(destination, asset)); err != nil {
				t.Errorf("template %q is missing materialized asset %q: %v", descriptor.ID, asset, err)
			}
		}
	}

	missing := "missing-template"
	for {
		if _, exists := ids[missing]; !exists {
			break
		}
		missing += "-x"
	}
	if _, err := catalog.Resolve(missing); err == nil {
		t.Errorf("Resolve(%q) error = nil", missing)
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
	template, err := Resolve("ubuntu-24.04-terminal-tools", "ubuntu:24.04")
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
	if strings.Contains(string(containerfile), "rm -rf /var/lib/apt/lists/*") {
		t.Error("Containerfile removes APT package indexes needed by mutable boxes")
	}
	for _, packageName := range []string{"ca-certificates", "curl", "fish", "git", "iproute2", "iputils-ping", "jq", "neovim", "procps", "ripgrep", "sudo", "tmux", "wl-clipboard"} {
		if !strings.Contains(string(containerfile), packageName) {
			t.Errorf("Containerfile does not install %q", packageName)
		}
	}
	for _, setup := range []string{"ARG BOX_USER", "ARG BOX_UID", "ARG BOX_GID", "NOPASSWD: ALL", "chmod 0440"} {
		if !strings.Contains(string(containerfile), setup) {
			t.Errorf("Containerfile does not configure %q", setup)
		}
	}
	for _, installerDetail := range []string{"https://starship.rs/install.sh", "--bin-dir /usr/local/bin"} {
		if !strings.Contains(string(containerfile), installerDetail) {
			t.Errorf("Containerfile does not install Starship using %q", installerDetail)
		}
	}
	if _, err := os.Stat(filepath.Join(directory, "dotfiles", ".config", "fish", "config.fish")); err != nil {
		t.Errorf("dotfiles are missing from build context: %v", err)
	}
}
