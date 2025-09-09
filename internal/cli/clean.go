package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/joch/ygg/internal/worktree"
	"github.com/spf13/cobra"
)

var (
	cleanForce bool
	cleanDry   bool
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove worktrees with merged branches",
	Long: `Remove worktrees whose branches have been merged to the default branch.

By default, prompts for confirmation before removing each worktree.
Use --force to skip confirmation.
Use --dry-run to see what would be removed without actually removing.`,
	RunE: runClean,
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().BoolVarP(&cleanForce, "force", "f", false, "Skip confirmation prompts")
	cleanCmd.Flags().BoolVarP(&cleanDry, "dry-run", "n", false, "Show what would be removed without removing")
}

func runClean(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	wm, err := worktree.NewManager(cwd)
	if err != nil {
		errorMsg("Not in a git repository")
		return err
	}

	worktrees, err := wm.List()
	if err != nil {
		errorMsg("Failed to list worktrees: %v", err)
		return err
	}

	defaultBranch, err := wm.DefaultBranch()
	if err != nil {
		errorMsg("Could not detect default branch: %v", err)
		return err
	}

	// Find merged branches
	mergedBranches, err := wm.MergedBranches(defaultBranch)
	if err != nil {
		errorMsg("Failed to get merged branches: %v", err)
		return err
	}

	mergedSet := make(map[string]bool)
	for _, b := range mergedBranches {
		mergedSet[b] = true
	}

	var toRemove []*worktree.Worktree
	for _, wt := range worktrees {
		if wt.IsPrimary {
			continue
		}
		if mergedSet[wt.Branch] {
			toRemove = append(toRemove, wt)
		}
	}

	if len(toRemove) == 0 {
		info("No merged worktrees to clean up")
		return nil
	}

	info("Found %d merged worktree(s):", len(toRemove))
	for _, wt := range toRemove {
		fmt.Printf("  %s (branch: %s)\n", wt.Name, wt.Branch)
	}

	if cleanDry {
		info("Dry run - no worktrees removed")
		return nil
	}

	if !cleanForce {
		fmt.Print("\nRemove these worktrees? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			info("Aborted")
			return nil
		}
	}

	for _, wt := range toRemove {
		if err := wm.Remove(wt.Name); err != nil {
			errorMsg("Failed to remove %s: %v", wt.Name, err)
		} else {
			success("Removed %s", wt.Name)
		}
	}

	return nil
}
