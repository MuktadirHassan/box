package main

import (
	"os"

	"github.com/MuktadirHassan/box/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
