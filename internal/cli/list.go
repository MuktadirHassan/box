package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newListCommand(definitions definitionStore) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List boxes",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, arguments []string) error {
			items, err := definitions.List()
			if err != nil {
				return err
			}
			for _, definition := range items {
				if _, err := fmt.Fprintf(command.OutOrStdout(), "%s\t%s\n", definition.Name, definition.State); err != nil {
					return err
				}
			}

			return nil
		},
	}
}
