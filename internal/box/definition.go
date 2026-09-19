package box

import (
	"fmt"
	"regexp"
	"strings"
)

const CurrentDefinitionVersion = 1

type State string

const (
	CreatedState State = "created"
	ReadyState   State = "ready"
)

var (
	namePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	userPattern = regexp.MustCompile(`^[a-z_][a-z0-9_-]*$`)
)

type Definition struct {
	Version       int           `toml:"version"`
	Name          string        `toml:"name"`
	State         State         `toml:"state"`
	Backend       Backend       `toml:"backend"`
	Configuration Configuration `toml:"configuration,omitempty"`
}

type Configuration struct {
	Image            string       `toml:"image"`
	User             string       `toml:"user"`
	Mounts           []Mount      `toml:"mounts"`
	Home             Persistence  `toml:"home"`
	Caches           Persistence  `toml:"caches"`
	Limits           Limits       `toml:"limits"`
	Network          string       `toml:"network"`
	Template         string       `toml:"template,omitempty"`
	TemplateRevision int          `toml:"template_revision,omitempty"`
	Shell            string       `toml:"shell"`
	Prompt           string       `toml:"prompt"`
	Integrations     Integrations `toml:"integrations"`
}

type Mount struct {
	Source      string `toml:"source"`
	Destination string `toml:"destination"`
}

type Persistence struct {
	Enabled bool `toml:"enabled"`
}

type Limits struct {
	CPUs      string `toml:"cpus"`
	Memory    string `toml:"memory"`
	PIDsLimit int    `toml:"pids_limit"`
}

type Integrations struct {
	Clipboard    bool `toml:"clipboard"`
	SSHAgent     bool `toml:"ssh_agent"`
	InsecureMode bool `toml:"insecure_mode"`
}

func NewDefinition(name string) Definition {
	return Definition{
		Version: CurrentDefinitionVersion,
		Name:    name,
		State:   CreatedState,
		Backend: PodmanBackend,
	}
}

func ContainerHome(user string) string { return "/home/" + user }

func ResolveMountDestination(destination, user string) string {
	if destination == "~" {
		return ContainerHome(user)
	}
	if strings.HasPrefix(destination, "~/") {
		return ContainerHome(user) + "/" + strings.TrimPrefix(destination, "~/")
	}
	return destination
}

func DefaultConfiguration() Configuration {
	return Configuration{
		Image:    "ubuntu:24.04",
		Home:     Persistence{Enabled: true},
		Caches:   Persistence{Enabled: true},
		Network:  "outbound",
		Template: "ubuntu-24.04-terminal-tools",
		Shell:    "sh",
		Prompt:   "none",
	}
}

// NormalizeConfiguration fills values whose empty form has a portable default.
func NormalizeConfiguration(configuration Configuration) Configuration {
	if configuration.Shell == "" {
		configuration.Shell = "sh"
	}
	if configuration.Prompt == "" {
		configuration.Prompt = "none"
	}
	return configuration
}

func ValidateName(name string) error {
	if !namePattern.MatchString(name) {
		return fmt.Errorf("%w %q: use lowercase letters, numbers, and single hyphens only", ErrInvalidName, name)
	}

	return nil
}

func ValidateUser(user string) error {
	if !userPattern.MatchString(user) {
		return fmt.Errorf("invalid user %q: use a lowercase Linux username", user)
	}

	return nil
}
