package app

import (
	"io"

	"github.com/MuktadirHassan/box/internal/backend"
	"github.com/MuktadirHassan/box/internal/backend/podman"
	"github.com/MuktadirHassan/box/internal/cli"
	"github.com/MuktadirHassan/box/internal/store"
	"github.com/MuktadirHassan/box/internal/templates"
	"github.com/MuktadirHassan/box/internal/terminal"
	"github.com/MuktadirHassan/box/internal/ui"
	assets "github.com/MuktadirHassan/box/templates"
	"github.com/spf13/cobra"
)

func Execute() error {
	catalog := templates.NewEmbeddedCatalog(assets.FS())
	command, err := newCommand(catalog, terminal.NewPresenter(catalog), nil)
	if err != nil {
		return err
	}
	return command.Execute()
}

// NewCommand constructs an independent command instance using the default local
// store and runtime registry. Callers may safely configure its arguments and I/O
// without sharing Cobra command state with another invocation.
func NewCommand(presenter ui.Presenter) (*cobra.Command, error) {
	return newCommand(templates.NewEmbeddedCatalog(assets.FS()), presenter, nil)
}

// NewCommandWithOutput constructs an independent command instance whose runtime
// command output is written to output.
func NewCommandWithOutput(presenter ui.Presenter, output io.Writer) (*cobra.Command, error) {
	return newCommand(templates.NewEmbeddedCatalog(assets.FS()), presenter, output)
}

func newCommand(catalog templates.Catalog, presenter ui.Presenter, output io.Writer) (*cobra.Command, error) {
	definitions, err := store.Default()
	if err != nil {
		return nil, err
	}

	// Register Lima here when it is implemented: backend.NewRegistry(podman.New(podman.Options{}), lima.New(lima.Options{})).
	podmanOptions := podman.Options{Catalog: catalog}
	if output != nil {
		podmanOptions.Runner = podman.NewCommandRunnerWithWriters("", output, output)
	}
	runtimes, err := backend.NewRegistry(podman.New(podmanOptions))
	if err != nil {
		return nil, err
	}

	return cli.NewRootCommandWithCatalog(definitions, runtimes, catalog, presenter), nil
}
