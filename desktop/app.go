package main

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/MuktadirHassan/box/internal/app"
	"github.com/spf13/cobra"
)

type commandFactory func(*bytes.Buffer) (*cobra.Command, error)

type commandRunner struct {
	mu         sync.Mutex
	newCommand commandFactory
}

func (r *commandRunner) run(ctx context.Context, arguments ...string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	output := &bytes.Buffer{}
	command, err := r.newCommand(output)
	if err != nil {
		return "", err
	}

	command.SetOut(output)
	command.SetErr(output)
	command.SetArgs(arguments)
	command.SetContext(ctx)
	command.SilenceUsage = true
	command.SilenceErrors = true
	if err := command.Execute(); err != nil {
		return output.String(), err
	}
	return output.String(), nil
}

// App exposes only predefined noninteractive desktop operations. Arguments are
// passed directly to Cobra; it never invokes a shell or an installed box binary.
type App struct {
	context context.Context
	runner  commandRunner
}

func NewApp() *App {
	return &App{
		context: context.Background(),
		runner: commandRunner{
			newCommand: func(output *bytes.Buffer) (*cobra.Command, error) {
				return app.NewCommandWithOutput(nil, output)
			},
		},
	}
}

func (a *App) startup(ctx context.Context) {
	a.context = ctx
}

func (a *App) List() (string, error) {
	return a.run("list")
}

func (a *App) Create(name string) (string, error) {
	if name == "" {
		return a.run("create")
	}
	return a.run("create", name)
}

func (a *App) Inspect(name string) (string, error) {
	return a.named("inspect", name)
}

func (a *App) Setup(name string) (string, error) {
	return a.named("setup", name, "--yes")
}

func (a *App) Stop(name string) (string, error) {
	return a.named("stop", name)
}

func (a *App) Delete(name string) (string, error) {
	return a.named("delete", name, "--purge")
}

func (a *App) named(operation, name string, extra ...string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("select a box first")
	}
	arguments := append([]string{operation, name}, extra...)
	return a.run(arguments...)
}

func (a *App) run(arguments ...string) (string, error) {
	return a.runner.run(a.context, arguments...)
}
