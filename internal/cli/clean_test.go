package cli

import (
	"testing"

	"github.com/joch/ygg/internal/worktree"
)

func TestSelectMergedWorktreesExcludesDefaultBranch(t *testing.T) {
	worktrees := []*worktree.Worktree{
		{Name: "main", Branch: "main", IsPrimary: true},
		{Name: "mainwt", Branch: "main", IsPrimary: false}, // non-primary, on default branch
		{Name: "feature", Branch: "feature", IsPrimary: false},
		{Name: "wip", Branch: "wip", IsPrimary: false},
	}
	// MergedBranches now measures against a fully-qualified ref, so the default
	// branch itself appears in the merged list (it no longer self-filters).
	merged := []string{"main", "feature"}

	got := selectMergedWorktrees(worktrees, merged, "main")

	// Only "feature" qualifies: primary is skipped, the non-primary default-branch
	// worktree must be skipped, and "wip" is not merged.
	var names []string
	for _, wt := range got {
		names = append(names, wt.Name)
	}
	if len(names) != 1 || names[0] != "feature" {
		t.Fatalf("selected %v, want exactly [feature]", names)
	}
}
