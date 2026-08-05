package cli

import "github.com/MuktadirHassan/box/internal/box"

type definitionStore interface {
	Create(box.Definition) error
	Update(box.Definition) error
	List() ([]box.Definition, error)
	Load(string) (box.Definition, error)
	Delete(string) error
}
