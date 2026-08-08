package podman

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

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
	template, err := b.catalog.Resolve(definition.Configuration.Template)
	if err != nil {
		return box.Definition{}, err
	}
	if err := template.Validate(templates.Request{Image: definition.Configuration.Image, Shell: definition.Configuration.Shell, Prompt: definition.Configuration.Prompt}); err != nil {
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

	shell := definition.Configuration.Shell
	if shell == "" {
		shell = "sh"
	}
	prompt := definition.Configuration.Prompt
	if prompt == "" {
		prompt = "none"
	}
	image := templateImageName(definition.Name)
	uid, gid := b.identity()
	if _, err := b.runner.Output(ctx, "build", "--quiet",
		"--build-arg", "BASE_IMAGE="+definition.Configuration.Image,
		"--build-arg", "BOX_USER="+definition.Configuration.User,
		"--build-arg", "BOX_UID="+strconv.Itoa(uid),
		"--build-arg", "BOX_GID="+strconv.Itoa(gid),
		"--build-arg", "BOX_SHELL="+shell,
		"--build-arg", "BOX_PROMPT="+prompt,
		"--build-arg", "BOX_TEMPLATE_REVISION="+strconv.Itoa(definition.Configuration.TemplateRevision),
		"--file", filepath.Join(directory, "Containerfile"), "--tag", image, directory); err != nil {
		return box.Definition{}, fmt.Errorf("build template image: %w", err)
	}
	definition.Configuration.Image = image
	return definition, nil
}

func templateImageName(name string) string { return "box-" + name + "-template" }
