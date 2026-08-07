package app

import (
	"github.com/MuktadirHassan/box/internal/backend"
	"github.com/MuktadirHassan/box/internal/backend/podman"
	"github.com/MuktadirHassan/box/internal/cli"
	"github.com/MuktadirHassan/box/internal/store"
	"github.com/MuktadirHassan/box/internal/templates"
	"github.com/MuktadirHassan/box/internal/terminal"
	assets "github.com/MuktadirHassan/box/templates"
)

func Execute() error {
	definitions, err := store.Default()
	if err != nil {
		return err
	}

	catalog := templates.NewEmbeddedCatalog(assets.FS())
	// Register Lima here when it is implemented: backend.NewRegistry(podman.New(podman.Options{}), lima.New(lima.Options{})).
	runtimes, err := backend.NewRegistry(podman.New(podman.Options{Catalog: catalog}))
	if err != nil {
		return err
	}

	return cli.NewRootCommandWithCatalog(definitions, runtimes, catalog, terminal.NewPresenter(catalog)).Execute()
}
