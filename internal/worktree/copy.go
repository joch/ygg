package worktree

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// defaultSkipDirs contains directories that should not be copied
// (large generated directories that can be regenerated)
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

// CopyUntrackedFiles copies untracked/ignored files from srcRepo to dstWorktree.
// It skips large generated directories like node_modules.
// Returns the number of files copied.
func CopyUntrackedFiles(srcRepo, dstWorktree string) (int, error) {
	files, err := getIgnoredFiles(srcRepo)
	if err != nil {
		return 0, err
	}

	copied := 0
	for _, relPath := range files {
		if shouldSkip(relPath, defaultSkipDirs) {
			continue
		}

		src := filepath.Join(srcRepo, relPath)
		dst := filepath.Join(dstWorktree, relPath)

		if err := copyFile(src, dst); err != nil {
			// Non-fatal: continue copying other files
			continue
		}
		copied++
	}

	return copied, nil
}

// getIgnoredFiles returns a list of files that are ignored by git
// (untracked files that match .gitignore patterns)
func getIgnoredFiles(repoPath string) ([]string, error) {
	// Get files that are ignored by git (in .gitignore but exist on disk)
	cmd := exec.Command("git", "ls-files", "--others", "--ignored", "--exclude-standard")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var files []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}

	return files, nil
}

// shouldSkip returns true if the path should be skipped based on skipDirs.
// A path is skipped if any of its directory components match a skip directory.
func shouldSkip(relPath string, skipDirs map[string]bool) bool {
	parts := strings.Split(relPath, string(filepath.Separator))
	for _, part := range parts {
		if skipDirs[part] {
			return true
		}
	}
	return false
}

// copyFile copies a file from src to dst, preserving permissions.
// It creates any necessary parent directories.
func copyFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Skip directories (we only copy files)
	if srcInfo.IsDir() {
		return nil
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file with same permissions
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy contents
	_, err = io.Copy(dstFile, srcFile)
	return err
}
