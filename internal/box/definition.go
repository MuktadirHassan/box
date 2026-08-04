package box

import (
	"fmt"
	"regexp"
)

const CurrentDefinitionVersion = 1

type State string

const CreatedState State = "created"

var namePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Definition struct {
	Version int    `toml:"version"`
	Name    string `toml:"name"`
	State   State  `toml:"state"`
}

func NewDefinition(name string) Definition {
	return Definition{
		Version: CurrentDefinitionVersion,
		Name:    name,
		State:   CreatedState,
	}
}

func ValidateName(name string) error {
	if !namePattern.MatchString(name) {
		return fmt.Errorf("%w %q: use lowercase letters, numbers, and single hyphens only", ErrInvalidName, name)
	}

	return nil
}
