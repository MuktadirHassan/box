package cli

import (
	"fmt"
	"io"

	"github.com/MuktadirHassan/box/internal/box"
)

func writeDefinition(writer io.Writer, definition box.Definition) error {
	_, err := fmt.Fprintf(writer, "name: %s\nstate: %s\nversion: %d\n", definition.Name, definition.State, definition.Version)
	return err
}
