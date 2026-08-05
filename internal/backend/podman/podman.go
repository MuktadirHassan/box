package podman

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/MuktadirHassan/box/internal/box"
)

type Backend struct {
	runner Runner
	env    func(string) string
}

type Options struct {
	Runner Runner
	Env    func(string) string
}

func New(options Options) *Backend {
	runner := options.Runner
	if runner == nil {
		runner = NewCommandRunner("")
	}
	env := options.Env
	if env == nil {
		env = os.Getenv
	}

	return &Backend{runner: runner, env: env}
}

func (b *Backend) Name() box.Backend { return box.PodmanBackend }

func (b *Backend) Validate(ctx context.Context) error {
	output, err := b.runner.Output(ctx, "info", "--format", "{{.Host.Security.Rootless}}")
	if err != nil {
		return fmt.Errorf("validate Podman: %w", err)
	}
	if strings.TrimSpace(output) != "true" {
		return fmt.Errorf("Podman is not running rootlessly")
	}
	return nil
}

func (b *Backend) Create(ctx context.Context, definition box.Definition) (box.RuntimeMetadata, error) {
	if definition.Backend != box.PodmanBackend {
		return box.RuntimeMetadata{}, fmt.Errorf("create Podman runtime for backend %q", definition.Backend)
	}
	if err := box.ValidateName(definition.Name); err != nil {
		return box.RuntimeMetadata{}, err
	}
	arguments, err := b.createArguments(definition)
	if err != nil {
		return box.RuntimeMetadata{}, err
	}
	output, err := b.runner.Output(ctx, arguments...)
	if err != nil {
		return box.RuntimeMetadata{}, fmt.Errorf("create Podman runtime: %w", err)
	}
	id := strings.TrimSpace(output)
	if !validIdentifier(id) {
		return box.RuntimeMetadata{}, fmt.Errorf("create Podman runtime: invalid container identifier %q", id)
	}
	return box.RuntimeMetadata{Backend: box.PodmanBackend, ID: id, State: box.RuntimeCreated}, nil
}

func (b *Backend) Start(ctx context.Context, metadata box.RuntimeMetadata) error {
	return b.run(ctx, metadata, "start", metadata.ID)
}

func (b *Backend) Stop(ctx context.Context, metadata box.RuntimeMetadata) error {
	return b.run(ctx, metadata, "stop", metadata.ID)
}

func (b *Backend) Inspect(ctx context.Context, metadata box.RuntimeMetadata) (box.RuntimeStatus, error) {
	if err := b.validateMetadata(metadata); err != nil {
		return box.RuntimeStatus{}, err
	}
	output, err := b.runner.Output(ctx, "inspect", "--format", "{{.State.Status}}", metadata.ID)
	if err != nil {
		return box.RuntimeStatus{}, fmt.Errorf("inspect Podman runtime: %w", err)
	}
	state, err := runtimeState(strings.TrimSpace(output))
	if err != nil {
		return box.RuntimeStatus{}, err
	}
	return box.RuntimeStatus{State: state}, nil
}

func (b *Backend) Delete(ctx context.Context, metadata box.RuntimeMetadata) error {
	return b.run(ctx, metadata, "rm", "--force", metadata.ID)
}

func (b *Backend) Enter(ctx context.Context, metadata box.RuntimeMetadata) error {
	status, err := b.Inspect(ctx, metadata)
	if err != nil {
		return err
	}
	switch status.State {
	case box.RuntimeRunning:
		return b.run(ctx, metadata, "exec", "--interactive", "--tty", metadata.ID, "/bin/sh")
	case box.RuntimeCreated, box.RuntimeStopped:
		return b.run(ctx, metadata, "start", "--attach", "--interactive", metadata.ID)
	default:
		return fmt.Errorf("enter Podman runtime: runtime is %s", status.State)
	}
}

func (b *Backend) Exec(ctx context.Context, metadata box.RuntimeMetadata, command []string) error {
	if len(command) == 0 {
		return fmt.Errorf("command cannot be empty")
	}
	if err := b.validateMetadata(metadata); err != nil {
		return err
	}
	arguments := append([]string{"exec", metadata.ID}, command...)
	if err := b.runner.Run(ctx, arguments...); err != nil {
		return fmt.Errorf("run command in Podman runtime: %w", err)
	}
	return nil
}

func (b *Backend) run(ctx context.Context, metadata box.RuntimeMetadata, arguments ...string) error {
	if err := b.validateMetadata(metadata); err != nil {
		return err
	}
	if err := b.runner.Run(ctx, arguments...); err != nil {
		return fmt.Errorf("run Podman command: %w", err)
	}
	return nil
}

func (b *Backend) validateMetadata(metadata box.RuntimeMetadata) error {
	if metadata.Backend != box.PodmanBackend {
		return fmt.Errorf("use Podman runtime with backend %q", metadata.Backend)
	}
	if !validIdentifier(metadata.ID) {
		return fmt.Errorf("invalid Podman runtime identifier %q", metadata.ID)
	}
	return nil
}
