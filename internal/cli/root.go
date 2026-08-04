package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

func Execute() error {
	return newRootCommand().Execute()
}

func newRootCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "box",
		Short: "Manage development boxes",
	}

	command.AddCommand(
		newBoxCommand("create", "[name]", "Create a box", cobra.MaximumNArgs(1), createBox),
		newBoxCommand("setup", "<box-name>", "Set up a box", cobra.ExactArgs(1), setupBox),
		newBoxCommand("edit", "<box-name>", "Edit a box", cobra.ExactArgs(1), editBox),
		newBoxCommand("delete", "<box-name>", "Delete a box", cobra.ExactArgs(1), deleteBox),
		newBoxCommand("enter", "<box-name>", "Enter a box", cobra.ExactArgs(1), enterBox),
	)

	return command
}

func newBoxCommand(name, usage, description string, arguments cobra.PositionalArgs, run func([]string)) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s %s", name, usage),
		Short: description,
		Args:  arguments,
		Run: func(command *cobra.Command, arguments []string) {
			run(arguments)
		},
	}
}

func createBox(arguments []string) {
	logCommand("create", arguments)
}

func setupBox(arguments []string) {
	logCommand("setup", arguments)
}

func editBox(arguments []string) {
	logCommand("edit", arguments)
}

func deleteBox(arguments []string) {
	logCommand("delete", arguments)
}

func enterBox(arguments []string) {
	logCommand("enter", arguments)
}

func logCommand(name string, arguments []string) {
	slog.New(slog.NewTextHandler(os.Stderr, nil)).Info("command called", "command", name, "arguments", arguments)
}
