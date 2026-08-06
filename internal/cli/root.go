package cli

import (
	"github.com/MuktadirHassan/box/internal/backend"
	"github.com/spf13/cobra"
)

func NewRootCommand(definitions definitionStore, runtimes *backend.Registry) *cobra.Command {
	command := &cobra.Command{
		Use:   "box",
		Short: "Manage development boxes",
	}

	command.AddCommand(
		newCreateCommand(definitions),
		newSetupCommand(definitions, runtimes),
		newListCommand(definitions),
		newInspectCommand(definitions, runtimes),
		newEnterCommand(definitions, runtimes),
		newExecCommand(definitions, runtimes),
		newStopCommand(definitions, runtimes),
		newDeleteCommand(definitions, runtimes),
	)

	return command
}
