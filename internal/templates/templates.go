package templates

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

//go:embed all:terminal-tools
var files embed.FS

var namePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Template struct {
	Name        string `toml:"name"`
	Version     int    `toml:"version"`
	Description string `toml:"description"`
	Family      string `toml:"family"`

	root          string
	containerfile string
	dotfiles      string
}

func Resolve(name, image string) (Template, error) {
	if !namePattern.MatchString(name) {
		return Template{}, fmt.Errorf("invalid template name %q", name)
	}
	family := imageFamily(image)
	template, err := load(name, family)
	if err != nil {
		return Template{}, fmt.Errorf("resolve template %q for %s: %w", name, image, err)
	}
	return template, nil
}

// ValidateCompatibility verifies that a template supports the selected base image.
func ValidateCompatibility(name, image string) error {
	if name == "" {
		return nil
	}
	_, err := Resolve(name, image)
	return err
}

func load(name, family string) (Template, error) {
	root := path.Join(name, family)
	data, err := fs.ReadFile(files, path.Join(root, "template.toml"))
	if err != nil {
		return Template{}, fmt.Errorf("template does not support image family %q", family)
	}

	var template Template
	if err := toml.Unmarshal(data, &template); err != nil {
		return Template{}, fmt.Errorf("decode manifest: %w", err)
	}
	if template.Name != name || template.Version < 1 || template.Family != family {
		return Template{}, fmt.Errorf("manifest is invalid")
	}
	template.root = root
	template.containerfile = path.Join(root, "Containerfile")
	template.dotfiles = path.Join(root, "dotfiles")
	if _, err := fs.Stat(files, template.containerfile); err != nil {
		return Template{}, fmt.Errorf("Containerfile: %w", err)
	}
	if _, err := fs.Stat(files, template.dotfiles); err != nil {
		return Template{}, fmt.Errorf("dotfiles: %w", err)
	}
	return template, nil
}

func All() ([]Template, error) {
	entries, err := fs.ReadDir(files, ".")
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	available := make([]Template, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		variants, err := fs.ReadDir(files, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("list template %q variants: %w", entry.Name(), err)
		}
		if len(variants) == 0 || !variants[0].IsDir() {
			return nil, fmt.Errorf("template %q has no variants", entry.Name())
		}
		template, err := load(entry.Name(), variants[0].Name())
		if err != nil {
			return nil, fmt.Errorf("load template %q: %w", entry.Name(), err)
		}
		available = append(available, template)
	}
	return available, nil
}

func (t Template) BuildContext(destination string) error {
	return fs.WalkDir(files, t.root, func(source string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative := strings.TrimPrefix(source, t.root)
		relative = strings.TrimPrefix(relative, "/")
		target := path.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(files, source)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

func imageFamily(image string) string {
	image = strings.Split(image, "@")[0]
	name := path.Base(image)
	family, _, _ := strings.Cut(name, ":")
	return family
}
