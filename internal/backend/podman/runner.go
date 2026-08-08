package podman

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
)

type Runner interface {
	Output(context.Context, ...string) (string, error)
	Run(context.Context, ...string) error
}

type CommandRunner struct {
	binary string
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func NewCommandRunner(binary string) CommandRunner {
	return NewCommandRunnerWithWriters(binary, os.Stdout, os.Stderr)
}

func NewCommandRunnerWithWriters(binary string, stdout, stderr io.Writer) CommandRunner {
	if binary == "" {
		binary = "podman"
	}

	return CommandRunner{
		binary: binary,
		stdin:  os.Stdin,
		stdout: stdout,
		stderr: stderr,
	}
}

func (r CommandRunner) Output(ctx context.Context, arguments ...string) (string, error) {
	command := exec.CommandContext(ctx, r.binary, arguments...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		return "", &commandError{err: err, stderr: stderr.String()}
	}

	return string(output), nil
}

func (r CommandRunner) Run(ctx context.Context, arguments ...string) error {
	command := exec.CommandContext(ctx, r.binary, arguments...)
	command.Stdin = r.stdin
	command.Stdout = r.stdout
	var stderr bytes.Buffer
	command.Stderr = io.MultiWriter(r.stderr, &stderr)
	if err := command.Run(); err != nil {
		return &commandError{err: err, stderr: stderr.String()}
	}

	return nil
}

type commandError struct {
	err    error
	stderr string
}

func (e *commandError) Error() string {
	if e.stderr == "" {
		return e.err.Error()
	}

	return e.err.Error() + ": " + e.stderr
}

func (e *commandError) Unwrap() error {
	return e.err
}
