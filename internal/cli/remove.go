package cli

import (
	"fmt"
	"os"

	"github.com/joch/ygg/internal/shell"
	"github.com/joch/ygg/internal/worktree"
	"github.com/spf13/cobra"
)

var forceRemove bool

var removeCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a worktree",
	Long: `Remove a worktree and return to main.

If no name is given and you're inside a worktree, removes the current one.
Use --force to remove even with uncommitted changes or unmerged branches.`,
	Aliases:           []string{"rm"},
	Args:              cobra.MaximumNArgs(1),
	RunE:              runRemove,
	ValidArgsFunction: completeWorktreeNames,
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().BoolVarP(&forceRemove, "force", "f", false, "Force removal even with uncommitted changes")
}

func runRemove(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	wm, err := worktree.NewManager(cwd)
	if err != nil {
		errorMsg("Not in a git repository")
		return err
	}

	var worktreeName string
	var needsCd bool

	if len(args) > 0 {
		// Remove by name
		worktreeName = args[0]
		wt, err := wm.Get(worktreeName)
		if err != nil {
			errorMsg("Worktree %q not found", worktreeName)
			return err
		}
		if wt.IsPrimary {
			errorMsg("Cannot remove the main worktree")
			return fmt.Errorf("cannot remove primary worktree")
		}
		if !forceRemove && wm.HasUncommittedChanges(wt.Path) {
			errorMsg("Worktree %s has uncommitted changes", worktreeName)
			info("Commit or stash your changes, or use --force")
			return fmt.Errorf("uncommitted changes")
		}
		if !forceRemove {
			merged, err := wm.IsBranchMerged(wt.Branch)
			if err == nil && !merged {
				errorMsg("Branch %s has not been merged", wt.Branch)
				info("Merge your changes first, or use --force to remove anyway")
				return fmt.Errorf("unmerged branch")
			}
		}
		// Check if we're inside this worktree
		current, _ := wm.Current()
		needsCd = current != nil && current.Name == worktreeName
	} else {
		// Remove current worktree
		current, err := wm.Current()
		if err != nil {
			errorMsg("Not in a worktree. Specify a name: ygg remove <name>")
			return err
		}
		if current.IsPrimary {
			errorMsg("Cannot remove the main worktree")
			info("Specify a worktree name: ygg remove <name>")
			return fmt.Errorf("cannot remove primary worktree")
		}
		if !forceRemove && wm.HasUncommittedChanges(current.Path) {
			errorMsg("Worktree %s has uncommitted changes", current.Name)
			info("Commit or stash your changes, or use --force")
			return fmt.Errorf("uncommitted changes")
		}
		if !forceRemove {
			merged, err := wm.IsBranchMerged(current.Branch)
			if err == nil && !merged {
				errorMsg("Branch %s has not been merged", current.Branch)
				info("Merge your changes first, or use --force to remove anyway")
				return fmt.Errorf("unmerged branch")
			}
		}
		worktreeName = current.Name
		needsCd = true
	}

	info("Removing worktree: %s", worktreeName)

	if err := wm.Remove(worktreeName); err != nil {
		errorMsg("Failed to remove worktree: %v", err)
		return err
	}

	success("Removed worktree: %s", worktreeName)

	if needsCd {
		mainPath := wm.RepoPath()
		if InYggShell() {
			fmt.Printf("cd %s\n", mainPath)
			return nil
		}
		info("Returning to main...")
		return shell.Spawn(mainPath, "main")
	}

	return nil
}
