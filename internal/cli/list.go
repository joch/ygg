package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/joch/ygg/internal/worktree"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all worktrees",
	Aliases: []string{"ls"},
	RunE:    runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
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

	if len(worktrees) == 0 {
		info("No worktrees found. Use 'ygg new <name>' to create one.")
		return nil
	}

	current, _ := wm.Current()

	yellow := color.New(color.FgYellow)
	green := color.New(color.FgGreen)

	for _, wt := range worktrees {
		marker := "  "
		if current != nil && wt.Path == current.Path {
			marker = green.Sprint("* ")
		}

		status := ""
		if wm.HasUncommittedChanges(wt.Path) {
			status = yellow.Sprint(" [modified]")
		}

		fmt.Printf("%s%s%s\n", marker, wt.Name, status)
	}

	return nil
}

// ListNames returns just the worktree names (for completion).
func ListNames() ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	wm, err := worktree.NewManager(cwd)
	if err != nil {
		return nil, err
	}

	worktrees, err := wm.List()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(worktrees))
	for _, wt := range worktrees {
		if !strings.HasPrefix(wt.Name, ".") {
			names = append(names, wt.Name)
		}
	}
	return names, nil
}
