package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/MuktadirHassan/box/internal/backend"
	"github.com/MuktadirHassan/box/internal/store"
	"github.com/spf13/cobra"
)

func newDeleteCommand(definitions definitionStore, runtimes *backend.Registry) *cobra.Command {
	var purge bool
	command := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a box",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, arguments []string) error {
			if !purge {
				return ErrPurgeRequired
			}
			metadata, err := definitions.LoadMetadata(arguments[0])
			if err == nil {
				if runtimes == nil {
					return fmt.Errorf("runtime deletion is unavailable")
				}
				runtime, err := runtimes.Get(metadata.Runtime.Backend)
				if err != nil {
					return err
				}
				definition, err := definitions.Load(arguments[0])
				if err != nil {
					return err
				}
				if err := runtime.Delete(context.Background(), definition, metadata.Runtime, backend.DeleteOptions{RemovePersistentData: true}); err != nil {
					return err
				}
			} else if !errors.Is(err, store.ErrMetadataNotFound) {
				return err
			}
			if err := definitions.Delete(arguments[0]); err != nil {
				return err
			}
			_, err = fmt.Fprintf(command.OutOrStdout(), "Deleted box %q.\n", arguments[0])
			return err
		},
	}
	command.Flags().BoolVar(&purge, "purge", false, "Remove the runtime, managed persistent data, and definition")
	return command
}
