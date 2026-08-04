package cli

import (
	"github.com/MuktadirHassan/box/internal/store"
	"github.com/spf13/cobra"
)

func Execute() error {
	definitions, err := store.Default()
	if err != nil {
		return err
	}

	return newRootCommand(definitions).Execute()
}

func newRootCommand(definitions definitionStore) *cobra.Command {
	command := &cobra.Command{
		Use:   "box",
		Short: "Manage development boxes",
	}

	command.AddCommand(
		newCreateCommand(definitions),
		newListCommand(definitions),
		newInspectCommand(definitions),
		newDeleteCommand(definitions),
	)

	return command
}
