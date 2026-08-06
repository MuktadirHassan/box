package main

import (
	"os"

	"github.com/MuktadirHassan/box/internal/app"
)

func main() {
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}
