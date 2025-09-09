package cli

import (
	"os"

	"github.com/joch/ygg/internal/worktree"
	"github.com/spf13/cobra"
)

// completeWorktreeNames provides shell completion for worktree names
func completeWorktreeNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	wm, err := worktree.NewManager(cwd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	worktrees, err := wm.List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var names []string
	for _, wt := range worktrees {
		names = append(names, wt.Name)
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}