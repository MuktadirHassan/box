package podman

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MuktadirHassan/box/internal/box"
)

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
	return image != "" && strings.TrimSpace(image) == image && !strings.ContainsAny(image, " \t\r\n") && !strings.HasPrefix(image, "-")
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

func passwdEntry(user, home string, uid, gid int) string {
	return fmt.Sprintf("%s:x:%d:%d::%s:/bin/sh", user, uid, gid, home)
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
