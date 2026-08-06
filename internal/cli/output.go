package cli

import (
	"fmt"
	"io"

	"github.com/MuktadirHassan/box/internal/box"
)

func writeDefinition(writer io.Writer, definition box.Definition) error {
	if _, err := fmt.Fprintf(writer, "name: %s\nstate: %s\nbackend: %s\nversion: %d\n", definition.Name, definition.State, definition.Backend, definition.Version); err != nil {
		return err
	}
	if definition.State != box.ReadyState {
		return nil
	}
	configuration := definition.Configuration
	if _, err := fmt.Fprintf(writer, "image: %s\nuser: %s\nnetwork: %s\ntemplate: %s\npersistent home: %t\npersistent caches: %t\n", configuration.Image, configuration.User, configuration.Network, configuration.Template, configuration.Home.Enabled, configuration.Caches.Enabled); err != nil {
		return err
	}
	if configuration.Limits.CPUs != "" {
		if _, err := fmt.Fprintf(writer, "cpus: %s\n", configuration.Limits.CPUs); err != nil {
			return err
		}
	}
	if configuration.Limits.Memory != "" {
		if _, err := fmt.Fprintf(writer, "memory: %s\n", configuration.Limits.Memory); err != nil {
			return err
		}
	}
	if configuration.Limits.PIDsLimit != 0 {
		if _, err := fmt.Fprintf(writer, "pids limit: %d\n", configuration.Limits.PIDsLimit); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(writer, "clipboard: %t\nssh agent: %t\n", configuration.Integrations.Clipboard, configuration.Integrations.SSHAgent); err != nil {
		return err
	}
	for _, mount := range configuration.Mounts {
		if _, err := fmt.Fprintf(writer, "mount: %s:%s\n", mount.Source, mount.Destination); err != nil {
			return err
		}
	}
	return nil
}
