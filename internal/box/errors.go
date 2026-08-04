package box

import "errors"

var (
	ErrInvalidName    = errors.New("invalid box name")
	ErrNameGeneration = errors.New("generate box name")
)
