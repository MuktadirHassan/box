package box

import "testing"

func TestValidateTemplate(t *testing.T) {
	if err := ValidateTemplate("terminal-tools"); err != nil {
		t.Fatalf("ValidateTemplate() error = %v", err)
	}
	if err := ValidateTemplate("../template"); err == nil {
		t.Error("ValidateTemplate() error = nil for an invalid name")
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
