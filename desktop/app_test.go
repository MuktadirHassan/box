package main

import (
	"bytes"
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
)

func TestCommandRunnerCreatesCommandAndCapturesOutput(t *testing.T) {
	ctx := context.WithValue(context.Background(), "request", "desktop")
	var created int32
	runner := commandRunner{newCommand: func(output *bytes.Buffer) (*cobra.Command, error) {
		atomic.AddInt32(&created, 1)
		return &cobra.Command{RunE: func(command *cobra.Command, arguments []string) error {
			if command.Context().Value("request") != "desktop" {
				t.Fatal("command context was not preserved")
			}
			if _, err := fmt.Fprint(output, "runtime "); err != nil {
				return err
			}
			_, err := fmt.Fprint(command.OutOrStdout(), arguments[0])
			return err
		}}, nil
	}}

	output, err := runner.run(ctx, "first")
	if err != nil {
		t.Fatalf("first run error = %v", err)
	}
	if output != "runtime first" {
		t.Errorf("first output = %q", output)
	}
	output, err = runner.run(ctx, "second")
	if err != nil {
		t.Fatalf("second run error = %v", err)
	}
	if output != "runtime second" {
		t.Errorf("second output = %q", output)
	}
	if got := atomic.LoadInt32(&created); got != 2 {
		t.Errorf("commands created = %d, want 2", got)
	}
}

func TestAppNamedRejectsMissingBox(t *testing.T) {
	application := NewApp()
	if _, err := application.Inspect(""); err == nil {
		t.Fatal("Inspect(\"\") error = nil")
	}
}
