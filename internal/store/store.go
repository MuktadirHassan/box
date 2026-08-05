package store

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/MuktadirHassan/box/internal/box"
	"github.com/pelletier/go-toml/v2"
)

const definitionFile = "box.toml"

type Store struct {
	root string
}

func New(root string) Store {
	return Store{root: root}
}

func Default() (Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Store{}, fmt.Errorf("find home directory: %w", err)
	}

	return New(filepath.Join(home, ".local", "share", "box", "boxes")), nil
}

func (s Store) Create(definition box.Definition) error {
	if err := box.ValidateName(definition.Name); err != nil {
		return err
	}

	if err := os.MkdirAll(s.root, 0o755); err != nil {
		return fmt.Errorf("create boxes directory: %w", err)
	}
	if err := os.Mkdir(s.boxPath(definition.Name), 0o755); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return fmt.Errorf("box %q: %w", definition.Name, ErrAlreadyExists)
		}
		return fmt.Errorf("create box directory: %w", err)
	}

	return s.save(definition)
}

func (s Store) Update(definition box.Definition) error {
	if err := box.ValidateName(definition.Name); err != nil {
		return err
	}

	path := s.definitionPath(definition.Name)
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("load box %q before update: %w", definition.Name, err)
	}

	return s.save(definition)
}

func (s Store) save(definition box.Definition) error {
	path := s.definitionPath(definition.Name)
	data, err := toml.Marshal(definition)
	if err != nil {
		return fmt.Errorf("encode definition: %w", err)
	}

	temporary, err := os.CreateTemp(filepath.Dir(path), ".box.toml-*")
	if err != nil {
		return fmt.Errorf("create temporary definition: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	if err := temporary.Chmod(0o644); err != nil {
		temporary.Close()
		return fmt.Errorf("set temporary definition permissions: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return fmt.Errorf("write temporary definition: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary definition: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("save definition: %w", err)
	}

	return nil
}

func (s Store) List() ([]box.Definition, error) {
	entries, err := os.ReadDir(s.root)
	if errors.Is(err, fs.ErrNotExist) {
		return []box.Definition{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list boxes: %w", err)
	}

	definitions := make([]box.Definition, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		definition, err := s.Load(entry.Name())
		if err != nil {
			return nil, err
		}
		definitions = append(definitions, definition)
	}
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].Name < definitions[j].Name })

	return definitions, nil
}

func (s Store) Load(name string) (box.Definition, error) {
	if err := box.ValidateName(name); err != nil {
		return box.Definition{}, err
	}

	data, err := os.ReadFile(s.definitionPath(name))
	if err != nil {
		return box.Definition{}, fmt.Errorf("load box %q: %w", name, err)
	}

	var definition box.Definition
	if err := toml.Unmarshal(data, &definition); err != nil {
		return box.Definition{}, fmt.Errorf("decode box %q: %w", name, err)
	}

	return definition, nil
}

func (s Store) Delete(name string) error {
	if err := box.ValidateName(name); err != nil {
		return err
	}
	if err := os.RemoveAll(s.boxPath(name)); err != nil {
		return fmt.Errorf("delete box %q: %w", name, err)
	}

	return nil
}

func (s Store) boxPath(name string) string {
	return filepath.Join(s.root, name)
}

func (s Store) definitionPath(name string) string {
	return filepath.Join(s.boxPath(name), definitionFile)
}
