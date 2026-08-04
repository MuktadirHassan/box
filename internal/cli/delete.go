package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDeleteCommand(definitions definitionStore) *cobra.Command {
	var purge bool
	command := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a box",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, arguments []string) error {
			if !purge {
				return ErrPurgeRequired
			}
			if err := definitions.Delete(arguments[0]); err != nil {
				return err
			}
			_, err := fmt.Fprintf(command.OutOrStdout(), "Deleted box %q.\n", arguments[0])
			return err
		},
	}
	command.Flags().BoolVar(&purge, "purge", false, "Remove the box definition directory")

	return command
}
