package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:       "completion [bash|fish|zsh]",
		Short:     "Generate shell completion scripts",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "fish", "zsh"},
		RunE: func(command *cobra.Command, arguments []string) error {
			root := command.Root()
			switch arguments[0] {
			case "bash":
				return root.GenBashCompletion(command.OutOrStdout())
			case "fish":
				return root.GenFishCompletion(command.OutOrStdout(), true)
			case "zsh":
				return root.GenZshCompletion(command.OutOrStdout())
			default:
				return fmt.Errorf("unsupported shell %q: use bash, fish, or zsh", arguments[0])
			}
		},
	}
}
