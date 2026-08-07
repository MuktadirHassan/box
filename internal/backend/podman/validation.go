package podman

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MuktadirHassan/box/internal/box"
	"github.com/MuktadirHassan/box/internal/templates"
)

func (b *Backend) ValidateConfiguration(configuration box.Configuration) error {
	configuration = box.NormalizeConfiguration(configuration)
	if err := validateConfiguration(configuration); err != nil {
		return err
	}
	if configuration.TemplateRevision < 0 {
		return fmt.Errorf("template revision cannot be negative")
	}
	if err := box.ValidateTemplate(configuration.Template); err != nil {
		return err
	}
	if configuration.Template != "" {
		resolved, err := b.catalog.Resolve(configuration.Template)
		if err != nil {
			return err
		}
		if err := resolved.Validate(templates.Request{Image: configuration.Image, Shell: configuration.Shell, Prompt: configuration.Prompt}); err != nil {
			return err
		}
	}
	if configuration.Template == "" && (configuration.Shell != "sh" || configuration.Prompt != "none") {
		return fmt.Errorf("a non-default shell or prompt requires an environment template")
	}
	if configuration.Prompt == "starship" && configuration.Shell == "sh" {
		return fmt.Errorf("prompt %q requires bash, fish, or zsh", configuration.Prompt)
	}
	if _, err := shellPath(configuration.Shell); err != nil {
		return err
	}
	if configuration.Prompt != "none" && configuration.Prompt != "starship" {
		return fmt.Errorf("unsupported prompt %q", configuration.Prompt)
	}
	return nil
}

func validateConfiguration(configuration box.Configuration) error {
	if err := box.ValidateUser(configuration.User); err != nil {
		return err
	}
	if !validImage(configuration.Image) {
		return fmt.Errorf("invalid image reference %q", configuration.Image)
	}
	if configuration.Network != "outbound" && configuration.Network != "none" {
		return fmt.Errorf("unsupported network policy %q", configuration.Network)
	}
	if configuration.Limits.PIDsLimit < 0 {
		return fmt.Errorf("process limit cannot be negative")
	}
	if !validLimit(configuration.Limits.CPUs) || !validLimit(configuration.Limits.Memory) {
		return fmt.Errorf("resource limits must be non-empty Podman quantities without whitespace or a leading hyphen")
	}
	for _, mount := range configuration.Mounts {
		if err := validateMount(mount); err != nil {
			return err
		}
	}
	return nil
}

func validateMount(mount box.Mount) error {
	if !filepath.IsAbs(mount.Source) || !filepath.IsAbs(mount.Destination) || filepath.Clean(mount.Destination) != mount.Destination || mount.Destination == "/" {
		return fmt.Errorf("invalid mount %q:%q: paths must be clean absolute paths and destination cannot be /", mount.Source, mount.Destination)
	}
	info, err := os.Stat(mount.Source)
	if err != nil {
		return fmt.Errorf("inspect mount source %q: %w", mount.Source, err)
	}
	if !info.IsDir() && !info.Mode().IsRegular() {
		return fmt.Errorf("mount source %q must be a file or directory", mount.Source)
	}
	return nil
}

func secureSocket(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("socket path must be absolute")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 || info.Mode()&os.ModeSocket == 0 {
		return "", fmt.Errorf("%q is not a Unix socket", path)
	}
	return path, nil
}

func validImage(image string) bool {
	if image == "" || strings.TrimSpace(image) != image || strings.ContainsAny(image, " \t\r\n") || strings.HasPrefix(image, "-") {
		return false
	}
	if strings.Count(image, "@") > 1 {
		return false
	}
	_, digest, hasDigest := strings.Cut(image, "@")
	if hasDigest {
		algorithm, value, valid := strings.Cut(digest, ":")
		if !valid || algorithm == "" || value == "" || strings.Trim(value, "0123456789abcdefABCDEF") != "" || len(value) != 64 {
			return false
		}
	}
	return true
}

func validLimit(limit string) bool {
	return limit == "" || (strings.TrimSpace(limit) == limit && !strings.HasPrefix(limit, "-") && !strings.ContainsAny(limit, " \t\r\n"))
}

func validSocketName(name string) bool {
	return name != "" && filepath.Base(name) == name && name != "." && name != ".."
}

func validIdentifier(id string) bool {
	return id != "" && !strings.HasPrefix(id, "-") && !strings.ContainsAny(id, " \t\r\n")
}

func containerName(name string) string { return "box-" + name }

func homeVolumeName(name string) string { return containerName(name) + "-home" }

func cacheVolumeName(name string) string { return containerName(name) + "-cache" }

func containerHome(user string) string { return "/home/" + user }

func containerUser(uid, gid int) string { return fmt.Sprintf("%d:%d", uid, gid) }

func passwdEntry(user, home, shell string, uid, gid int) string {
	return fmt.Sprintf("%s:x:%d:%d::%s:%s", user, uid, gid, home, shell)
}

func shellPath(shell string) (string, error) {
	switch shell {
	case "", "sh":
		return "/bin/sh", nil
	case "bash":
		return "/bin/bash", nil
	case "fish":
		return "/usr/bin/fish", nil
	case "zsh":
		return "/usr/bin/zsh", nil
	default:
		return "", fmt.Errorf("unsupported shell %q", shell)
	}
}

func networkMode(policy string) string {
	if policy == "none" {
		return "none"
	}
	return "pasta"
}

func runtimeState(status string) (box.RuntimeState, error) {
	switch status {
	case "created", "configured":
		return box.RuntimeCreated, nil
	case "running":
		return box.RuntimeRunning, nil
	case "stopped", "exited":
		return box.RuntimeStopped, nil
	default:
		return box.RuntimeMissing, fmt.Errorf("unexpected Podman runtime state %q", status)
	}
}
