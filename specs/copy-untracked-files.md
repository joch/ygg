# Copy Untracked Files on Worktree Creation

## Summary

Add automatic copying of untracked/gitignored files (like `.env`) from the main worktree to new worktrees when running `ygg new`.

## Status

- [x] Design complete
- [x] Tests written
- [x] Implementation complete
- [x] Integration complete

## Implementation Steps

### 1. Create `internal/worktree/copy.go` with copy logic
### 2. Create `internal/worktree/copy_test.go` with tests
### 3. Integrate into `Manager.Create()` in worktree.go
### 4. Update CLI messaging in new.go (optional)

## Files

| File | Change |
|------|--------|
| `internal/worktree/copy.go` | **Create** - copy logic |
| `internal/worktree/copy_test.go` | **Create** - tests |
| `internal/worktree/worktree.go:146` | Add `CopyUntrackedFiles()` call |
| `internal/cli/new.go` | Show copy count |

## Design

### Functions

```go
// CopyUntrackedFiles copies untracked/ignored files from srcRepo to dstWorktree
// Returns the number of files copied
func CopyUntrackedFiles(srcRepo, dstWorktree string) (int, error)

// getIgnoredFiles returns list of ignored files in repo
func getIgnoredFiles(repoPath string) ([]string, error)

// shouldSkip returns true if path should be skipped (large generated dirs)
func shouldSkip(relPath string, skipDirs map[string]bool) bool

// copyFile copies a single file preserving permissions
func copyFile(src, dst string) error
```

### Skip Directories

```go
var defaultSkipDirs = map[string]bool{
    "node_modules": true,
    "vendor":       true,
    "__pycache__":  true,
    "dist":         true,
    "build":        true,
    "target":       true,
    ".gradle":      true,
    "venv":         true,
    ".venv":        true,
    ".next":        true,
    ".nuxt":        true,
    ".turbo":       true,
    ".terraform":   true,
    "coverage":     true,
    ".cache":       true,
    "tmp":          true,
}
```

## Notes

- Copy failures warn but don't fail `ygg new`
- Always enabled (no flag needed initially)
- Copy, not symlink - keeps worktrees independent
