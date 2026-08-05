package box

import (
	"fmt"
	"regexp"
)

const CurrentDefinitionVersion = 1

type State string

const (
	CreatedState State = "created"
	ReadyState   State = "ready"
)

var namePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Definition struct {
	Version       int           `toml:"version"`
	Name          string        `toml:"name"`
	State         State         `toml:"state"`
	Backend       Backend       `toml:"backend"`
	Configuration Configuration `toml:"configuration,omitempty"`
}

type Configuration struct {
	Image        string       `toml:"image"`
	Mounts       []Mount      `toml:"mounts"`
	Home         Persistence  `toml:"home"`
	Caches       Persistence  `toml:"caches"`
	Limits       Limits       `toml:"limits"`
	Network      string       `toml:"network"`
	Integrations Integrations `toml:"integrations"`
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
	Wayland  bool `toml:"wayland"`
	SSHAgent bool `toml:"ssh_agent"`
}

func NewDefinition(name string) Definition {
	return Definition{
		Version: CurrentDefinitionVersion,
		Name:    name,
		State:   CreatedState,
		Backend: PodmanBackend,
	}
}

func DefaultConfiguration() Configuration {
	return Configuration{
		Image:   "ubuntu:24.04",
		Home:    Persistence{Enabled: true},
		Caches:  Persistence{Enabled: true},
		Network: "outbound",
	}
}

func ValidateName(name string) error {
	if !namePattern.MatchString(name) {
		return fmt.Errorf("%w %q: use lowercase letters, numbers, and single hyphens only", ErrInvalidName, name)
	}

	return nil
}
