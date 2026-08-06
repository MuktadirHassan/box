package box

import "testing"

func TestTemplatePackages(t *testing.T) {
	packages := TemplatePackages(TerminalToolsTemplate)
	want := []string{"fish", "jq", "neovim", "tmux", "ripgrep", "starship", "wl-clipboard"}
	if len(packages) != len(want) {
		t.Fatalf("TemplatePackages() = %v, want %v", packages, want)
	}
	for index, value := range want {
		if packages[index] != value {
			t.Errorf("TemplatePackages()[%d] = %q, want %q", index, packages[index], value)
		}
	}
	if err := ValidateTemplate("unknown"); err == nil {
		t.Error("ValidateTemplate() error = nil for an unknown template")
	}
}

func TestNewDefinitionUsesInitialValues(t *testing.T) {
	definition := NewDefinition("demo")

	if definition.Version != CurrentDefinitionVersion {
		t.Errorf("Version = %d, want %d", definition.Version, CurrentDefinitionVersion)
	}
	if definition.Name != "demo" {
		t.Errorf("Name = %q, want %q", definition.Name, "demo")
	}
	if definition.State != CreatedState {
		t.Errorf("State = %q, want %q", definition.State, CreatedState)
	}
	if definition.Backend != PodmanBackend {
		t.Errorf("Backend = %q, want %q", definition.Backend, PodmanBackend)
	}
}
