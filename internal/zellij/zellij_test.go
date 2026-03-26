package zellij

import (
	"os"
	"testing"
)

func TestInZellij_WhenSet(t *testing.T) {
	os.Setenv("ZELLIJ", "0")
	defer os.Unsetenv("ZELLIJ")

	if !InZellij() {
		t.Error("expected InZellij() to return true when ZELLIJ is set")
	}
}

func TestInZellij_WhenUnset(t *testing.T) {
	os.Unsetenv("ZELLIJ")

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
