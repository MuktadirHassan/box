package box

import "testing"

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
