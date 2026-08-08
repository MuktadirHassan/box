package cli

import (
	"fmt"

	"github.com/MuktadirHassan/box/internal/backend"
	"github.com/MuktadirHassan/box/internal/box"
	"github.com/spf13/cobra"
)

func runtimeFor(definitions definitionStore, runtimes *backend.Registry, name string) (backend.Backend, box.Metadata, error) {
	if runtimes == nil {
		return nil, box.Metadata{}, fmt.Errorf("runtime commands are unavailable")
	}
	metadata, err := definitions.LoadMetadata(name)
	if err != nil {
		return nil, box.Metadata{}, err
	}
	runtime, err := runtimes.Get(metadata.Runtime.Backend)
	if err != nil {
		return nil, box.Metadata{}, err
	}
	return runtime, metadata, nil
}

func newEnterCommand(definitions definitionStore, runtimes *backend.Registry) *cobra.Command {
	return &cobra.Command{Use: "enter <name>", Short: "Enter a box", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, arguments []string) error {
		runtime, metadata, err := runtimeFor(definitions, runtimes, arguments[0])
		if err != nil {
			return err
		}
		definition, err := definitions.Load(arguments[0])
		if err != nil {
			return err
		}
		return runtime.Enter(command.Context(), definition, metadata.Runtime)
	}}
}

func newExecCommand(definitions definitionStore, runtimes *backend.Registry) *cobra.Command {
	return &cobra.Command{Use: "exec <name> -- <command>", Short: "Run a command in a box", Args: cobra.MinimumNArgs(2), RunE: func(command *cobra.Command, arguments []string) error {
		runtime, metadata, err := runtimeFor(definitions, runtimes, arguments[0])
		if err != nil {
			return err
		}
		return runtime.Exec(command.Context(), metadata.Runtime, arguments[1:])
	}}
}

func newStopCommand(definitions definitionStore, runtimes *backend.Registry) *cobra.Command {
	return &cobra.Command{Use: "stop <name>", Short: "Stop a box", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, arguments []string) error {
		runtime, metadata, err := runtimeFor(definitions, runtimes, arguments[0])
		if err != nil {
			return err
		}
		if err := runtime.Stop(command.Context(), metadata.Runtime); err != nil {
			return err
		}
		metadata.Runtime.State = box.RuntimeStopped
		return definitions.SaveMetadata(arguments[0], metadata)
	}}
}
