package podman

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MuktadirHassan/box/internal/box"
)

func (b *Backend) buildTemplate(ctx context.Context, definition box.Definition) (box.Definition, error) {
	if definition.Configuration.Template == "" {
		return definition, nil
	}
	if err := box.ValidateTemplate(definition.Configuration.Template); err != nil {
		return box.Definition{}, err
	}
	if !templateSupportedImage(definition.Configuration.Image) {
		return box.Definition{}, fmt.Errorf("template %q requires an Ubuntu, Debian, or Arch Linux base image", definition.Configuration.Template)
	}

	directory, err := os.MkdirTemp("", "box-template-")
	if err != nil {
		return box.Definition{}, fmt.Errorf("create template build directory: %w", err)
	}
	defer os.RemoveAll(directory)

	containerfile := filepath.Join(directory, "Containerfile")
	contents := templateContainerfile(definition.Configuration.Image, definition.Configuration.Template)
	if err := os.WriteFile(containerfile, []byte(contents), 0o600); err != nil {
		return box.Definition{}, fmt.Errorf("write template Containerfile: %w", err)
	}

	image := templateImageName(definition.Name)
	if _, err := b.runner.Output(ctx, "build", "--quiet", "--file", containerfile, "--tag", image, directory); err != nil {
		return box.Definition{}, fmt.Errorf("build template image: %w", err)
	}
	definition.Configuration.Image = image
	return definition, nil
}

func templateImageName(name string) string { return "box-" + name + "-template" }

func templateContainerfile(image, template string) string {
	packages := strings.Join(box.TemplatePackages(template), " ")
	if isArchImage(image) {
		return fmt.Sprintf("FROM %s\nRUN pacman -Syu --noconfirm --needed %s && pacman -Scc --noconfirm\n", image, packages)
	}
	return fmt.Sprintf("FROM %s\nRUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install --yes --no-install-recommends %s && rm -rf /var/lib/apt/lists/*\n", image, packages)
}

func templateSupportedImage(image string) bool {
	return isUbuntuOrDebianImage(image) || isArchImage(image)
}

func isUbuntuOrDebianImage(image string) bool {
	return image == "ubuntu" || image == "debian" || strings.HasPrefix(image, "ubuntu:") || strings.HasPrefix(image, "debian:") || strings.Contains(image, "/ubuntu:") || strings.Contains(image, "/debian:")
}

func isArchImage(image string) bool {
	return image == "archlinux" || strings.HasPrefix(image, "archlinux:") || strings.Contains(image, "/archlinux:")
}
