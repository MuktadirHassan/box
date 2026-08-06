package store

import "errors"

var (
	ErrAlreadyExists    = errors.New("box already exists")
	ErrMetadataNotFound = errors.New("runtime metadata not found")
)
