package store_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/MuktadirHassan/box/internal/box"
	"github.com/MuktadirHassan/box/internal/store"
)

func newTestStore(t *testing.T) store.Store {
	t.Helper()
	return store.New(filepath.Join(t.TempDir(), "boxes"))
}

func definition(name string) box.Definition {
	return box.NewDefinition(name)
}

func TestCreateWritesReadableTOML(t *testing.T) {
	root := filepath.Join(t.TempDir(), "boxes")
	definitions := store.New(root)
	if err := definitions.Create(definition("demo")); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "demo", "box.toml"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	want := "version = 1\nname = 'demo'\nstate = 'created'\nbackend = 'podman'\n"
	if string(data) != want {
		t.Errorf("box.toml = %q, want %q", data, want)
	}
}

func TestCreateDuplicateDoesNotAlterDefinition(t *testing.T) {
	definitions := newTestStore(t)
	original := definition("demo")
	if err := definitions.Create(original); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	duplicate := original
	duplicate.State = "changed"
	if err := definitions.Create(duplicate); err == nil {
		t.Fatal("second Create() error = nil, want duplicate error")
	}

	loaded, err := definitions.Load("demo")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !reflect.DeepEqual(loaded, original) {
		t.Errorf("Load() = %#v, want %#v", loaded, original)
	}
}

func TestCreateRejectsInvalidNames(t *testing.T) {
	for _, name := range []string{"", "Uppercase", "two words", "-start", "end-", "two--hyphens", "a/b", "a..b"} {
		t.Run(name, func(t *testing.T) {
			definitions := newTestStore(t)
			err := definitions.Create(definition(name))
			if err == nil {
				t.Fatal("Create() error = nil, want invalid name error")
			}
			if !strings.Contains(err.Error(), "invalid box name") {
				t.Errorf("Create() error = %q, want invalid name error", err)
			}
		})
	}
}

func TestListReturnsDefinitionsInNameOrder(t *testing.T) {
	definitions := newTestStore(t)
	for _, name := range []string{"zebra", "amber", "middle"} {
		if err := definitions.Create(definition(name)); err != nil {
			t.Fatalf("Create(%q) error = %v", name, err)
		}
	}

	got, err := definitions.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	want := []box.Definition{definition("amber"), definition("middle"), definition("zebra")}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("List() = %#v, want %#v", got, want)
	}
}

func TestLoadReturnsExpectedDefinition(t *testing.T) {
	definitions := newTestStore(t)
	want := definition("demo")
	if err := definitions.Create(want); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := definitions.Load("demo")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Load() = %#v, want %#v", got, want)
	}
}

func TestDeleteRemovesOnlyTargetDirectory(t *testing.T) {
	root := filepath.Join(t.TempDir(), "boxes")
	definitions := store.New(root)
	for _, name := range []string{"keep", "remove"} {
		if err := definitions.Create(definition(name)); err != nil {
			t.Fatalf("Create(%q) error = %v", name, err)
		}
	}

	if err := definitions.Delete("remove"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "remove")); !os.IsNotExist(err) {
		t.Errorf("removed directory stat error = %v, want not exist", err)
	}
	if _, err := os.Stat(filepath.Join(root, "keep", "box.toml")); err != nil {
		t.Errorf("unrelated definition stat error = %v", err)
	}
}
