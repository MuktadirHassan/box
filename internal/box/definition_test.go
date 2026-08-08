package box

import "testing"

func TestValidateTemplate(t *testing.T) {
	for _, name := range []string{"terminal-tools", "ubuntu-24.04-terminal-tools"} {
		if err := ValidateTemplate(name); err != nil {
			t.Errorf("ValidateTemplate(%q) error = %v", name, err)
		}
	}
	if err := ValidateTemplate("../template"); err == nil {
		t.Error("ValidateTemplate() error = nil for an invalid name")
	}
}

func TestResolveMountDestinationUsesContainerHome(t *testing.T) {
	for _, test := range []struct {
		destination string
		want        string
	}{
		{destination: "~", want: "/home/dev"},
		{destination: "~/.claude", want: "/home/dev/.claude"},
		{destination: "/workspace", want: "/workspace"},
		{destination: "~dev/project", want: "~dev/project"},
	} {
		if got := ResolveMountDestination(test.destination, "dev"); got != test.want {
			t.Errorf("ResolveMountDestination(%q, dev) = %q, want %q", test.destination, got, test.want)
		}
	}
}

func TestDefaultConfigurationUsesUbuntuDevelopmentEnvironment(t *testing.T) {
	configuration := DefaultConfiguration()
	if configuration.Image != "ubuntu:24.04" || configuration.Template != "ubuntu-24.04-terminal-tools" {
		t.Errorf("DefaultConfiguration() = %#v", configuration)
	}
	if !configuration.Home.Enabled || !configuration.Caches.Enabled || configuration.Network != "outbound" {
		t.Errorf("DefaultConfiguration() persistence/network = %#v", configuration)
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
