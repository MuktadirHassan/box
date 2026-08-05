package cli

import (
	"fmt"
	"strings"

	"github.com/MuktadirHassan/box/internal/box"
	"github.com/spf13/cobra"
)

type setupOptions struct {
	backend   string
	image     string
	mounts    []string
	cpus      string
	memory    string
	pids      int
	network   string
	clipboard bool
	sshAgent  bool
	yes       bool
}

func newSetupCommand(definitions definitionStore) *cobra.Command {
	options := setupOptions{}
	command := &cobra.Command{
		Use:   "setup <name>",
		Short: "Configure a box",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, arguments []string) error {
			definition, err := definitions.Load(arguments[0])
			if err != nil {
				return err
			}

			configuration, err := resolveConfiguration(command, definition.Configuration, options)
			if err != nil {
				return err
			}
			if definition.Backend == "" {
				definition.Backend = box.PodmanBackend
			}
			if command.Flags().Changed("backend") {
				definition.Backend = box.Backend(options.backend)
			}
			if err := box.ValidateBackend(definition.Backend); err != nil {
				return err
			}
			definition.Configuration = configuration
			definition.State = box.ReadyState

			if err := writeDefinition(command.OutOrStdout(), definition); err != nil {
				return err
			}
			if !options.yes {
				return ErrSetupConfirmation
			}
			if err := definitions.Update(definition); err != nil {
				return err
			}

			_, err = fmt.Fprintf(command.OutOrStdout(), "Configured box %q. Runtime creation is not implemented yet.\n", definition.Name)
			return err
		},
	}

	flags := command.Flags()
	flags.StringVar(&options.backend, "backend", "", "runtime backend: podman or lima")
	flags.StringVar(&options.image, "image", "", "base image")
	flags.StringArrayVar(&options.mounts, "mount", nil, "writable host mount in source:destination form")
	flags.StringVar(&options.cpus, "cpus", "", "CPU limit")
	flags.StringVar(&options.memory, "memory", "", "memory limit")
	flags.IntVar(&options.pids, "pids-limit", 0, "process limit")
	flags.StringVar(&options.network, "network", "", "network policy: outbound or none")
	flags.BoolVar(&options.clipboard, "clipboard", false, "enable host clipboard integration")
	flags.BoolVar(&options.sshAgent, "ssh-agent", false, "enable SSH agent forwarding")
	flags.BoolVar(&options.yes, "yes", false, "save the displayed configuration")

	return command
}

func resolveConfiguration(command *cobra.Command, current box.Configuration, options setupOptions) (box.Configuration, error) {
	configuration := current
	if configuration.Image == "" {
		configuration = box.DefaultConfiguration()
	}

	flags := command.Flags()
	if flags.Changed("image") {
		configuration.Image = options.image
	}
	if flags.Changed("mount") {
		mounts, err := parseMounts(options.mounts)
		if err != nil {
			return box.Configuration{}, err
		}
		configuration.Mounts = mounts
	}
	if flags.Changed("cpus") {
		configuration.Limits.CPUs = options.cpus
	}
	if flags.Changed("memory") {
		configuration.Limits.Memory = options.memory
	}
	if flags.Changed("pids-limit") {
		configuration.Limits.PIDsLimit = options.pids
	}
	if flags.Changed("network") {
		configuration.Network = options.network
	}
	if flags.Changed("clipboard") {
		configuration.Integrations.Clipboard = options.clipboard
	}
	if flags.Changed("ssh-agent") {
		configuration.Integrations.SSHAgent = options.sshAgent
	}

	if configuration.Image == "" {
		return box.Configuration{}, fmt.Errorf("base image cannot be empty")
	}
	if configuration.Limits.PIDsLimit < 0 {
		return box.Configuration{}, fmt.Errorf("process limit cannot be negative")
	}
	if configuration.Network != "outbound" && configuration.Network != "none" {
		return box.Configuration{}, fmt.Errorf("network policy %q is not supported; use outbound or none", configuration.Network)
	}

	return configuration, nil
}

func parseMounts(values []string) ([]box.Mount, error) {
	mounts := make([]box.Mount, 0, len(values))
	for _, value := range values {
		source, destination, ok := strings.Cut(value, ":")
		if !ok || source == "" || destination == "" {
			return nil, fmt.Errorf("invalid mount %q: use source:destination", value)
		}
		mounts = append(mounts, box.Mount{Source: source, Destination: destination})
	}

	return mounts, nil
}
