package cli

import (
	"github.com/MuktadirHassan/box/internal/backend"
	"github.com/MuktadirHassan/box/internal/templates"
	"github.com/MuktadirHassan/box/internal/ui"
	"github.com/MuktadirHassan/box/internal/version"
	"github.com/spf13/cobra"
)

func NewRootCommand(definitions definitionStore, runtimes *backend.Registry, presenters ...ui.Presenter) *cobra.Command {
	return newRootCommand(definitions, runtimes, nil, presenters...)
}

func NewRootCommandWithCatalog(definitions definitionStore, runtimes *backend.Registry, catalog templates.Catalog, presenter ui.Presenter) *cobra.Command {
	return newRootCommand(definitions, runtimes, catalog, presenter)
}

func newRootCommand(definitions definitionStore, runtimes *backend.Registry, catalog templates.Catalog, presenters ...ui.Presenter) *cobra.Command {
	var presenter ui.Presenter
	if len(presenters) > 0 {
		presenter = presenters[0]
	}
	command := &cobra.Command{
		Use:     "box",
		Short:   "Manage development boxes",
		Version: version.Version,
	}

	command.AddCommand(
		newCreateCommand(definitions, presenter),
		newSetupCommand(definitions, runtimes, presenter, catalog),
		newListCommand(definitions, presenter),
		newInspectCommand(definitions, runtimes, presenter),
		newEnterCommand(definitions, runtimes),
		newExecCommand(definitions, runtimes),
		newStopCommand(definitions, runtimes),
		newDeleteCommand(definitions, runtimes),
		newCompletionCommand(),
	)

	return command
}
