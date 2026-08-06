package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/MuktadirHassan/box/internal/backend"
	"github.com/MuktadirHassan/box/internal/store"
	"github.com/MuktadirHassan/box/internal/ui"
	"github.com/spf13/cobra"
)

func newInspectCommand(definitions definitionStore, runtimes *backend.Registry, presenter ui.Presenter) *cobra.Command {
	return &cobra.Command{
		Use: "inspect <name>", Short: "Inspect a box", Args: cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, arguments []string) error {
			definition, err := definitions.Load(arguments[0])
			if err != nil {
				return err
			}
			if presenter != nil {
				if err := presenter.ShowDefinition(command.OutOrStdout(), definition); err != nil {
					return err
				}
			} else if err := writeDefinition(command.OutOrStdout(), definition); err != nil {
				return err
			}
			metadata, err := definitions.LoadMetadata(definition.Name)
			if errors.Is(err, store.ErrMetadataNotFound) {
				_, err = fmt.Fprintln(command.OutOrStdout(), "runtime: not created")
				return err
			}
			if err != nil {
				return err
			}
			if runtimes == nil {
				_, err = fmt.Fprintf(command.OutOrStdout(), "runtime: %s\n", metadata.Runtime.State)
				return err
			}
			runtime, err := runtimes.Get(metadata.Runtime.Backend)
			if err != nil {
				return err
			}
			status, err := runtime.Inspect(context.Background(), metadata.Runtime)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(command.OutOrStdout(), "runtime: %s\nruntime id: %s\n", status.State, metadata.Runtime.ID)
			return err
		},
	}
}
