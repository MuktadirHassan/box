package cli

import (
	"github.com/MuktadirHassan/box/internal/backend"
	"github.com/MuktadirHassan/box/internal/ui"
	"github.com/spf13/cobra"
)

func NewRootCommand(definitions definitionStore, runtimes *backend.Registry, presenters ...ui.Presenter) *cobra.Command {
	var presenter ui.Presenter
	if len(presenters) > 0 {
		presenter = presenters[0]
	}
	command := &cobra.Command{
		Use:   "box",
		Short: "Manage development boxes",
	}

	command.AddCommand(
		newCreateCommand(definitions, presenter),
		newSetupCommand(definitions, runtimes, presenter),
		newListCommand(definitions, presenter),
		newInspectCommand(definitions, runtimes, presenter),
		newEnterCommand(definitions, runtimes),
		newExecCommand(definitions, runtimes),
		newStopCommand(definitions, runtimes),
		newDeleteCommand(definitions, runtimes),
	)

	return command
}
