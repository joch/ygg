package cli

import (
	"fmt"
	"os"

	"github.com/joch/ygg/internal/shell"
	"github.com/joch/ygg/internal/worktree"
	"github.com/joch/ygg/internal/zellij"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "Switch to a worktree",
	Long: `Switch to an existing worktree by name.

This spawns a subshell in the worktree directory.
Exit the subshell to return to your original directory.`,
	Args:              cobra.ExactArgs(1),
	Aliases:           []string{"sw"},
	RunE:              runSwitch,
	ValidArgsFunction: completeWorktreeNames,
}

func init() {
	rootCmd.AddCommand(switchCmd)
}

func runSwitch(cmd *cobra.Command, args []string) error {
	name := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	wm, err := worktree.NewManager(cwd)
	if err != nil {
		errorMsg("Not in a git repository")
		return err
	}

	wt, err := wm.Get(name)
	if err != nil {
		errorMsg("Worktree %q not found", name)
		return err
	}

	// If already in a ygg shell, just output cd command for the wrapper to eval
	if InYggShell() {
		fmt.Printf("cd %s\n", wt.Path)
		return nil
	}

	// Use zellij tab if inside zellij, otherwise spawn a subshell
	if zellij.InZellij() {
		info("Switching to zellij tab...")
		if err := zellij.OpenTab(wt.Path, wm.RepoName(), wt.Name); err != nil {
			info("Zellij failed, falling back to subshell: %v", err)
			return shell.Spawn(wt.Path, wt.Name)
		}
		return nil
	}

	info("Entering %s (exit to return)...", wt.Name)
	return shell.Spawn(wt.Path, wt.Name)
}
