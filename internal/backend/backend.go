package backend

import (
	"context"
	"fmt"

	"github.com/MuktadirHassan/box/internal/box"
)

type Backend interface {
	Name() box.Backend
	Validate(context.Context) error
	Create(context.Context, box.Definition) (box.RuntimeMetadata, error)
	Start(context.Context, box.RuntimeMetadata) error
	Stop(context.Context, box.RuntimeMetadata) error
	Inspect(context.Context, box.RuntimeMetadata) (box.RuntimeStatus, error)
	Delete(context.Context, box.Definition, box.RuntimeMetadata) error
	Enter(context.Context, box.RuntimeMetadata) error
	Exec(context.Context, box.RuntimeMetadata, []string) error
}

type Registry struct {
	backends map[box.Backend]Backend
}

func NewRegistry(backends ...Backend) (*Registry, error) {
	registry := &Registry{backends: make(map[box.Backend]Backend, len(backends))}
	for _, backend := range backends {
		if backend == nil {
			return nil, fmt.Errorf("register backend: backend is nil")
		}
		name := backend.Name()
		if err := box.ValidateBackend(name); err != nil {
			return nil, err
		}
		if _, exists := registry.backends[name]; exists {
			return nil, fmt.Errorf("register backend %q: already registered", name)
		}
		registry.backends[name] = backend
	}

	return registry, nil
}

func (r *Registry) Get(name box.Backend) (Backend, error) {
	backend, ok := r.backends[name]
	if !ok {
		return nil, fmt.Errorf("backend %q is not available", name)
	}

	return backend, nil
}
