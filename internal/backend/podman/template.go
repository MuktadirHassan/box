package podman

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/MuktadirHassan/box/internal/box"
	"github.com/MuktadirHassan/box/internal/templates"
)

func (b *Backend) buildTemplate(ctx context.Context, definition box.Definition) (box.Definition, error) {
	if definition.Configuration.Template == "" {
		return definition, nil
	}
	if err := box.ValidateTemplate(definition.Configuration.Template); err != nil {
		return box.Definition{}, err
	}
	template, err := templates.Resolve(definition.Configuration.Template, definition.Configuration.Image)
	if err != nil {
		return box.Definition{}, err
	}
	directory, err := os.MkdirTemp("", "box-template-")
	if err != nil {
		return box.Definition{}, fmt.Errorf("create template build directory: %w", err)
	}
	defer os.RemoveAll(directory)
	if err := template.BuildContext(directory); err != nil {
		return box.Definition{}, fmt.Errorf("prepare template build context: %w", err)
	}

	image := templateImageName(definition.Name)
	if _, err := b.runner.Output(ctx, "build", "--quiet", "--build-arg", "BASE_IMAGE="+definition.Configuration.Image, "--file", filepath.Join(directory, "Containerfile"), "--tag", image, directory); err != nil {
		return box.Definition{}, fmt.Errorf("build template image: %w", err)
	}
	definition.Configuration.Image = image
	return definition, nil
}

func templateImageName(name string) string { return "box-" + name + "-template" }
