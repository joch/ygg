package cli

import (
	"fmt"
	"os"

	"github.com/joch/ygg/internal/shell"
	"github.com/joch/ygg/internal/worktree"
	"github.com/joch/ygg/internal/zellij"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new worktree and enter it",
	Long: `Create a new git worktree with the specified name.

This will:
1. Fetch and pull the default branch (main/master)
2. Create a new worktree with a branch named <name>
3. Enter a subshell in the new worktree directory

Exit the subshell to return to your original directory.`,
	Args: cobra.ExactArgs(1),
	RunE: runNew,
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
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

	// Fetch latest
	info("Fetching from origin...")
	if err := wm.Fetch(); err != nil {
		// Non-fatal, continue anyway
		info("Could not fetch (offline?)")
	}

	// Pull default branch in main repo
	defaultBranch, err := wm.DefaultBranch()
	if err != nil {
		errorMsg("Could not detect default branch: %v", err)
		return err
	}

	info("Creating worktree: %s (based on %s)", name, defaultBranch)

	wt, err := wm.Create(name)
	if err != nil {
		errorMsg("Failed to create worktree: %v", err)
		return err
	}

	success("Created worktree at %s", wt.Path)

	// Report on copied files
	if wt.CopyError != nil {
		info("Warning: failed to copy some files: %v", wt.CopyError)
	} else if wt.CopiedFiles > 0 {
		info("Copied %d untracked file(s) from main worktree", wt.CopiedFiles)
	}

	// If already in a ygg shell, just output cd command for the wrapper to eval
	if InYggShell() {
		fmt.Printf("cd %s\n", wt.Path)
		return nil
	}

	// Use zellij tab if inside zellij, otherwise spawn a subshell
	if zellij.InZellij() {
		info("Opening zellij tab...")
		if err := zellij.OpenTab(wt.Path, wm.RepoName(), wt.Name); err != nil {
			info("Zellij failed, falling back to subshell: %v", err)
			return shell.Spawn(wt.Path, wt.Name)
		}
		return nil
	}

	info("Entering worktree (exit to return)...")
	return shell.Spawn(wt.Path, wt.Name)
}
