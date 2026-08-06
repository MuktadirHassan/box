package cli

import (
	"errors"
	"fmt"

	"github.com/MuktadirHassan/box/internal/box"
	"github.com/MuktadirHassan/box/internal/store"
	"github.com/MuktadirHassan/box/internal/ui"
	"github.com/spf13/cobra"
)

const generatedNameAttempts = 16

func newCreateCommand(definitions definitionStore, presenter ui.Presenter) *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Short: "Create a box",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(command *cobra.Command, arguments []string) error {
			name, err := createDefinition(definitions, arguments)
			if err != nil {
				return err
			}

			message := fmt.Sprintf("Created box %q. Next: box setup %s", name, name)
			if presenter != nil {
				return presenter.ShowSuccess(command.OutOrStdout(), message)
			}
			_, err = fmt.Fprintln(command.OutOrStdout(), message)
			return err
		},
	}
}

func createDefinition(definitions definitionStore, arguments []string) (string, error) {
	if len(arguments) == 1 {
		name := arguments[0]
		return name, definitions.Create(box.NewDefinition(name))
	}

	var err error
	for range generatedNameAttempts {
		name, err := generateName()
		if err != nil {
			return "", err
		}
		err = definitions.Create(box.NewDefinition(name))
		if !errors.Is(err, store.ErrAlreadyExists) {
			return name, err
		}
	}

	return "", err
}
