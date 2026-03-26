package zellij

import (
	"testing"
)

func TestInZellij_WhenSet(t *testing.T) {
	t.Setenv("ZELLIJ", "0")

	if !InZellij() {
		t.Error("expected InZellij() to return true when ZELLIJ is set")
	}
}

func TestInZellij_WhenUnset(t *testing.T) {
	t.Setenv("ZELLIJ", "")

	if InZellij() {
		t.Error("expected InZellij() to return false when ZELLIJ is unset")
	}
}

func TestTabName(t *testing.T) {
	tests := []struct {
		repoName      string
		worktreeName  string
		expected      string
	}{
		{"ygg", "my-feature", "ygg/my-feature"},
		{"my-repo", "fix-bug", "my-repo/fix-bug"},
	}

	for _, tt := range tests {
		got := TabName(tt.repoName, tt.worktreeName)
		if got != tt.expected {
			t.Errorf("TabName(%q, %q) = %q, want %q", tt.repoName, tt.worktreeName, got, tt.expected)
		}
	}
}
