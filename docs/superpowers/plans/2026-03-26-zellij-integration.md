# Zellij Tab Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** When running inside zellij, use zellij tabs instead of subshells for `ygg new` and `ygg switch`.

**Architecture:** New `internal/zellij` package with detection and tab management. CLI commands check `zellij.InZellij()` before falling back to `shell.Spawn()`. The `Manager` exposes `RepoName()` for tab naming as `<repo>/<worktree>`.

**Tech Stack:** Go, `os/exec` for zellij CLI commands, no new dependencies.

---

## File Structure

| File | Action | Responsibility |
|------|--------|---------------|
| `internal/zellij/zellij.go` | Create | Zellij detection and tab operations |
| `internal/zellij/zellij_test.go` | Create | Tests for detection and tab name logic |
| `internal/worktree/worktree.go` | Modify | Add `RepoName()` getter |
| `internal/cli/new.go` | Modify | Add zellij check before `shell.Spawn()` |
| `internal/cli/switch.go` | Modify | Add zellij check before `shell.Spawn()` |
| `README.md` | Modify | Add Zellij section |

---

### Task 1: Add `RepoName()` getter to Manager

**Files:**
- Modify: `internal/worktree/worktree.go:47-55`

- [ ] **Step 1: Add the RepoName method**

Add after the existing `BaseDir()` method at line 55 in `internal/worktree/worktree.go`:

```go
// RepoName returns the repository name (base directory name).
func (m *Manager) RepoName() string {
	return m.repoName
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/joch/dev/joch/ygg && go build ./...`
Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add internal/worktree/worktree.go
git commit -m "feat: expose RepoName() on worktree Manager"
```

---

### Task 2: Create `internal/zellij` package with tests

**Files:**
- Create: `internal/zellij/zellij.go`
- Create: `internal/zellij/zellij_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/zellij/zellij_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/joch/dev/joch/ygg && go test ./internal/zellij/...`
Expected: Compilation error — package doesn't exist yet.

- [ ] **Step 3: Write the implementation**

Create `internal/zellij/zellij.go`:

```go
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

	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if strings.TrimSpace(line) == name {
			// Tab exists, focus it
			focus := exec.Command("zellij", "action", "go-to-tab-name", name)
			if err := focus.Run(); err != nil {
				return fmt.Errorf("failed to focus zellij tab %q: %w", name, err)
			}
			return nil
		}
	}

	// Tab doesn't exist, create it
	create := exec.Command("zellij", "action", "new-tab", "--name", name, "--cwd", dir)
	if err := create.Run(); err != nil {
		return fmt.Errorf("failed to create zellij tab %q: %w", name, err)
	}
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/joch/dev/joch/ygg && go test ./internal/zellij/... -v`
Expected: All 3 tests pass (the `OpenTab` function isn't unit-tested since it shells out to zellij).

- [ ] **Step 5: Commit**

```bash
git add internal/zellij/zellij.go internal/zellij/zellij_test.go
git commit -m "feat: add zellij package with detection and tab management"
```

---

### Task 3: Integrate zellij into `ygg new`

**Files:**
- Modify: `internal/cli/new.go`

- [ ] **Step 1: Add zellij import and check**

In `internal/cli/new.go`, add `"github.com/joch/ygg/internal/zellij"` to imports.

Then replace lines 82-84:

```go
	// Otherwise spawn a subshell
	info("Entering worktree (exit to return)...")
	return shell.Spawn(wt.Path, wt.Name)
```

With:

```go
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
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/joch/dev/joch/ygg && go build ./...`
Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add internal/cli/new.go
git commit -m "feat: use zellij tabs in ygg new when inside zellij"
```

---

### Task 4: Integrate zellij into `ygg switch`

**Files:**
- Modify: `internal/cli/switch.go`

- [ ] **Step 1: Add zellij import and check**

In `internal/cli/switch.go`, add `"github.com/joch/ygg/internal/zellij"` to imports.

Then replace lines 55-57:

```go
	// Otherwise spawn a subshell
	info("Entering %s (exit to return)...", wt.Name)
	return shell.Spawn(wt.Path, wt.Name)
```

With:

```go
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
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/joch/dev/joch/ygg && go build ./...`
Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add internal/cli/switch.go
git commit -m "feat: use zellij tabs in ygg switch when inside zellij"
```

---

### Task 5: Update README

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Add Zellij section**

Add the following after the "Prompt Integration" section (after line 109) in `README.md`:

```markdown
## Zellij Integration

When running inside a [zellij](https://zellij.dev/) session, ygg automatically creates named tabs instead of spawning subshells. No configuration needed — it detects zellij via the `ZELLIJ` environment variable.

- `ygg new my-feature` creates a tab named `<repo>/my-feature` with the worktree as the working directory
- `ygg switch my-feature` focuses the existing tab, or creates one if it doesn't exist

If zellij commands fail for any reason, ygg falls back to the normal subshell behavior.
```

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add Zellij integration section to README"
```

---

### Task 6: Run all tests and verify

- [ ] **Step 1: Run full test suite**

Run: `cd /Users/joch/dev/joch/ygg && go test ./... -v`
Expected: All tests pass.

- [ ] **Step 2: Run build**

Run: `cd /Users/joch/dev/joch/ygg && go build -o /dev/null ./cmd/ygg`
Expected: Clean build, no errors.

- [ ] **Step 3: Run vet and lint**

Run: `cd /Users/joch/dev/joch/ygg && go vet ./...`
Expected: No issues.
