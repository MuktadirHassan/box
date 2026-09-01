package podman

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/MuktadirHassan/box/internal/box"
)

func (b *Backend) createArguments(definition box.Definition) ([]string, error) {
	configuration := definition.Configuration
	if err := validateConfiguration(configuration); err != nil {
		return nil, err
	}
	shell, err := shellPath(configuration.Shell)
	if err != nil {
		return nil, err
	}

	home := box.ContainerHome(configuration.User)
	uid, gid := b.identity()
	runtimeDirectory := home
	arguments := []string{
		"create", "--tty", "--name", containerName(definition.Name),
		"--userns", "keep-id", "--user", containerUser(uid, gid),
	}
	if configuration.Template == "" {
		arguments = append(arguments, "--passwd-entry", passwdEntry(configuration.User, home, shell, uid, gid))
	}
	arguments = append(arguments,
		"--env", "HOME="+home, "--env", "USER="+configuration.User, "--env", "LOGNAME="+configuration.User,
		"--env", "XDG_RUNTIME_DIR="+runtimeDirectory,
		"--env", "BOX_TEMPLATE="+configuration.Template, "--env", "BOX_TEMPLATE_REVISION="+strconv.Itoa(configuration.TemplateRevision), "--env", "BOX_PROMPT="+configuration.Prompt, "--env", "SHELL="+shell,
		"--workdir", home, "--hostname", definition.Name,
		"--network", networkMode(configuration.Network),
	)
	if configuration.Limits.PIDsLimit > 0 {
		arguments = append(arguments, "--pids-limit", strconv.Itoa(configuration.Limits.PIDsLimit))
	}
	if configuration.Limits.Memory != "" {
		arguments = append(arguments, "--memory", configuration.Limits.Memory)
	}
	if configuration.Limits.CPUs != "" {
		arguments = append(arguments, "--cpus", configuration.Limits.CPUs)
	}
	if configuration.Home.Enabled {
		arguments = append(arguments, "--mount", "type=volume,src="+homeVolumeName(definition.Name)+",dst="+home+",rw,U=true")
	}
	if configuration.Caches.Enabled {
		arguments = append(arguments, "--mount", "type=volume,src="+cacheVolumeName(definition.Name)+",dst="+home+"/.cache,rw,U=true")
	}
	arguments = append(arguments, "--tmpfs", "/tmp:rw,nosuid,nodev")
	for _, mount := range configuration.Mounts {
		arguments = append(arguments, "--mount", "type=bind,src="+mount.Source+",dst="+mount.Destination+",rw,nosuid,nodev")
	}
	if configuration.Integrations.Clipboard {
		var err error
		arguments, err = b.withClipboard(arguments)
		if err != nil {
			return nil, err
		}
	}
	if configuration.Integrations.SSHAgent {
		var err error
		arguments, err = b.withSSHAgent(arguments)
		if err != nil {
			return nil, err
		}
	}
	if configuration.Integrations.InsecureMode {
		var err error
		arguments, err = b.withHostPodmanSocket(arguments)
		if err != nil {
			return nil, err
		}
	}

	return append(arguments, configuration.Image, shell), nil
}

func (b *Backend) withClipboard(arguments []string) ([]string, error) {
	runtimeDirectory := b.env("XDG_RUNTIME_DIR")
	display := b.env("WAYLAND_DISPLAY")
	if display == "" {
		display = "wayland-0"
	}
	if !filepath.IsAbs(runtimeDirectory) || !validSocketName(display) {
		return nil, fmt.Errorf("enable clipboard integration: invalid Wayland runtime path")
	}
	socket, err := secureSocket(filepath.Join(runtimeDirectory, display))
	if err != nil {
		return nil, fmt.Errorf("enable clipboard integration: %w", err)
	}
	return append(arguments,
		"--env", "WAYLAND_DISPLAY="+display,
		"--mount", "type=bind,src="+socket+",dst=/tmp/"+display+",rw,nosuid,nodev",
	), nil
}

func (b *Backend) withSSHAgent(arguments []string) ([]string, error) {
	socket, err := secureSocket(b.env("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, fmt.Errorf("enable SSH agent integration: %w", err)
	}
	return append(arguments,
		"--env", "SSH_AUTH_SOCK=/tmp/ssh-agent.sock",
		"--mount", "type=bind,src="+socket+",dst=/tmp/ssh-agent.sock,rw,nosuid,nodev",
	), nil
}

func (b *Backend) withHostPodmanSocket(arguments []string) ([]string, error) {
	uid, _ := b.identity()
	socket, err := secureHostPodmanSocket(b.env("XDG_RUNTIME_DIR"), uid)
	if err != nil {
		return nil, fmt.Errorf("enable insecure mode: host rootless Podman socket unavailable: %w; activate it with systemctl --user enable --now podman.socket", err)
	}
	return append(arguments,
		"--env", "DOCKER_HOST=unix:///tmp/podman.sock",
		"--env", "CONTAINER_HOST=unix:///tmp/podman.sock",
		"--mount", "type=bind,src="+socket+",dst=/tmp/podman.sock,rw,nosuid,nodev",
	), nil
}
