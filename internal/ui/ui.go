package ui

import (
	"io"

	"github.com/MuktadirHassan/box/internal/box"
)

type Step interface {
	Success()
	Fail()
}

type Presenter interface {
	ConfigureInitial(box.Definition) (box.Definition, error)
	ConfirmSetup() error
	StartStep(io.Writer, string) (Step, error)
	ShowDefinition(io.Writer, box.Definition) error
	ShowRuntime(io.Writer, box.RuntimeState, string) error
	ShowList(io.Writer, []box.Definition) error
	ShowWarning(io.Writer, string) error
	ShowSuccess(io.Writer, string) error
}
