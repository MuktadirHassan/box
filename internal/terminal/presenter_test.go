package terminal

import (
	"bytes"
	"testing"

	"github.com/MuktadirHassan/box/internal/box"
	"github.com/MuktadirHassan/box/internal/templates"
)

type presenterCatalog struct{}

func (presenterCatalog) List() ([]templates.Descriptor, error) {
	return []templates.Descriptor{{ID: "opaque-id", Name: "opaque-id", Description: "Manifest label"}}, nil
}
func (presenterCatalog) Resolve(string) (templates.Resolved, error) { return nil, nil }

func TestEnvironmentTemplateOptionsUseCatalogLabelsAndCanonicalValues(t *testing.T) {
	options, err := environmentTemplateOptions(presenterCatalog{})
	if err != nil {
		t.Fatal(err)
	}
	if len(options) != 2 || options[1].Value != "opaque-id" || options[1].Key != "Manifest label" {
		t.Fatalf("options = %#v", options)
	}
}

func TestShowDefinitionAndRuntimeAlignInspectFields(t *testing.T) {
	definition := box.Definition{Name: "demo", State: box.ReadyState, Backend: box.PodmanBackend, Version: 1, Configuration: box.Configuration{Image: "ubuntu:24.04", User: "swift-maple", Network: "outbound", Home: box.Persistence{Enabled: true}, Caches: box.Persistence{Enabled: true}}}
	output := &bytes.Buffer{}
	presenter := Presenter{}
	if err := presenter.ShowDefinition(output, definition); err != nil {
		t.Fatal(err)
	}
	if err := presenter.ShowRuntime(output, box.RuntimeStopped, "container-id"); err != nil {
		t.Fatal(err)
	}
	want := "Box demo\n  State              ready\n  Backend            podman\n  Version            1\n  Image              ubuntu:24.04\n  User               swift-maple\n  Network            outbound\n  Template           \n  Shell              \n  Prompt             \n  Persistent home    true\n  Persistent caches  true\n  Clipboard          false\n  SSH agent          false\n  Runtime            stopped\n  Runtime ID         container-id\n"
	if output.String() != want {
		t.Errorf("inspection output = %q, want %q", output.String(), want)
	}
}

func TestShowListAlignsColumnsAndUsesImagePlaceholder(t *testing.T) {
	definitions := []box.Definition{
		{Name: "demo", State: box.ReadyState, Configuration: box.Configuration{Image: "ubuntu:24.04"}},
		{Name: "swift-maple", State: box.CreatedState},
	}
	output := &bytes.Buffer{}
	if err := (Presenter{}).ShowList(output, definitions); err != nil {
		t.Fatal(err)
	}
	want := "Boxes\n  NAME         STATE    IMAGE\n  demo         ready    ubuntu:24.04\n  swift-maple  created  -\n"
	if output.String() != want {
		t.Errorf("ShowList() output = %q, want %q", output.String(), want)
	}
}
