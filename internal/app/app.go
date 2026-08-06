package app

import (
	"github.com/MuktadirHassan/box/internal/backend"
	"github.com/MuktadirHassan/box/internal/backend/podman"
	"github.com/MuktadirHassan/box/internal/cli"
	"github.com/MuktadirHassan/box/internal/store"
)

func Execute() error {
	definitions, err := store.Default()
	if err != nil {
		return err
	}

	// Register Lima here when it is implemented: backend.NewRegistry(podman.New(podman.Options{}), lima.New(lima.Options{})).
	runtimes, err := backend.NewRegistry(podman.New(podman.Options{}))
	if err != nil {
		return err
	}

	return cli.NewRootCommand(definitions, runtimes).Execute()
}
