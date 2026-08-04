package cli

import (
	"github.com/spf13/cobra"
)

func newInspectCommand(definitions definitionStore) *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <name>",
		Short: "Inspect a box",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, arguments []string) error {
			definition, err := definitions.Load(arguments[0])
			if err != nil {
				return err
			}

			return writeDefinition(command.OutOrStdout(), definition)
		},
	}
}
