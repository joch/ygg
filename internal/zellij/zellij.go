package zellij

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// InZellij returns true if running inside a zellij session.
func InZellij() bool {
	return os.Getenv("ZELLIJ") != ""
}

// TabName returns the zellij tab name for a worktree: "<repo>/<worktree>".
func TabName(repoName, worktreeName string) string {
	return repoName + "/" + worktreeName
}

// OpenTab opens or focuses a zellij tab for the given worktree.
// If a tab with the name already exists, it focuses it.
// Otherwise, it creates a new tab with the given working directory.
func OpenTab(dir, repoName, worktreeName string) error {
	name := TabName(repoName, worktreeName)

	// Check if tab already exists
	cmd := exec.Command("zellij", "action", "query-tab-names")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to query zellij tabs: %w", err)
	}

	tabs := strings.TrimSpace(string(output))
	if tabs != "" {
		for _, line := range strings.Split(tabs, "\n") {
			if strings.TrimSpace(line) == name {
				// Tab exists, focus it
				focus := exec.Command("zellij", "action", "go-to-tab-name", name)
				if err := focus.Run(); err != nil {
					return fmt.Errorf("failed to focus zellij tab %q: %w", name, err)
				}
				return nil
			}
		}
	}

	// Tab doesn't exist, create it
	create := exec.Command("zellij", "action", "new-tab", "--name", name, "--cwd", dir)
	if err := create.Run(); err != nil {
		return fmt.Errorf("failed to create zellij tab %q: %w", name, err)
	}
	return nil
}

// CloseTab closes a zellij tab matching the given worktree, if it exists.
// This is best-effort — errors are returned but callers may choose to ignore them.
func CloseTab(repoName, worktreeName string) error {
	name := TabName(repoName, worktreeName)

	// Check if tab exists
	cmd := exec.Command("zellij", "action", "query-tab-names")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to query zellij tabs: %w", err)
	}

	tabs := strings.TrimSpace(string(output))
	if tabs == "" {
		return nil
	}

	for _, line := range strings.Split(tabs, "\n") {
		if strings.TrimSpace(line) == name {
			// Focus the tab first, then close it
			focus := exec.Command("zellij", "action", "go-to-tab-name", name)
			if err := focus.Run(); err != nil {
				return fmt.Errorf("failed to focus zellij tab %q: %w", name, err)
			}
			close := exec.Command("zellij", "action", "close-tab")
			if err := close.Run(); err != nil {
				return fmt.Errorf("failed to close zellij tab %q: %w", name, err)
			}
			return nil
		}
	}

	return nil
}
