package worktree

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestShouldSkip(t *testing.T) {
	skipDirs := map[string]bool{
		"node_modules": true,
		"vendor":       true,
		".cache":       true,
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"skip node_modules", "node_modules/package/index.js", true},
		{"skip nested node_modules", "frontend/node_modules/pkg/file.js", true},
		{"skip vendor", "vendor/github.com/pkg/file.go", true},
		{"skip .cache", ".cache/data", true},
		{"allow .env", ".env", false},
		{"allow .env.local", ".env.local", false},
		{"allow config dir", "config/settings.json", false},
		{"allow src files", "src/main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSkip(tt.path, skipDirs)
			if result != tt.expected {
				t.Errorf("shouldSkip(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ygg-copy-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source file with content
	srcPath := filepath.Join(tmpDir, "source.txt")
	content := "test content\nline 2"
	if err := os.WriteFile(srcPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Copy to destination
	dstPath := filepath.Join(tmpDir, "dest.txt")
	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	// Verify content
	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != content {
		t.Errorf("content = %q, want %q", string(got), content)
	}

	// Verify permissions preserved
	srcInfo, _ := os.Stat(srcPath)
	dstInfo, _ := os.Stat(dstPath)
	if srcInfo.Mode() != dstInfo.Mode() {
		t.Errorf("mode = %v, want %v", dstInfo.Mode(), srcInfo.Mode())
	}
}

func TestCopyFileCreatesDirectories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ygg-copy-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source file
	srcPath := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(srcPath, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Copy to nested destination that doesn't exist
	dstPath := filepath.Join(tmpDir, "nested", "deep", "dest.txt")
	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(dstPath); err != nil {
		t.Errorf("destination file not created: %v", err)
	}
}

func TestCopyUntrackedFiles(t *testing.T) {
	// Create temp directory for test repo
	tmpDir, err := os.MkdirTemp("", "ygg-copy-integration")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcRepo := filepath.Join(tmpDir, "repo")
	dstWorktree := filepath.Join(tmpDir, "worktree")

	// Initialize git repo
	if err := os.MkdirAll(srcRepo, 0755); err != nil {
		t.Fatal(err)
	}
	runGit(t, srcRepo, "init")
	runGit(t, srcRepo, "config", "user.email", "test@test.com")
	runGit(t, srcRepo, "config", "user.name", "Test")

	// Create .gitignore
	gitignore := ".env\n.env.local\nconfig/secrets.json\nnode_modules/\n"
	if err := os.WriteFile(filepath.Join(srcRepo, ".gitignore"), []byte(gitignore), 0644); err != nil {
		t.Fatal(err)
	}

	// Create ignored files
	if err := os.WriteFile(filepath.Join(srcRepo, ".env"), []byte("SECRET=value"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcRepo, ".env.local"), []byte("LOCAL=value"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(srcRepo, "config"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcRepo, "config", "secrets.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create node_modules (should be skipped)
	if err := os.MkdirAll(filepath.Join(srcRepo, "node_modules", "pkg"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcRepo, "node_modules", "pkg", "index.js"), []byte("// code"), 0644); err != nil {
		t.Fatal(err)
	}

	// Commit .gitignore so git tracks it
	runGit(t, srcRepo, "add", ".gitignore")
	runGit(t, srcRepo, "commit", "-m", "initial")

	// Create destination worktree directory
	if err := os.MkdirAll(dstWorktree, 0755); err != nil {
		t.Fatal(err)
	}

	// Run copy
	copied, err := CopyUntrackedFiles(srcRepo, dstWorktree)
	if err != nil {
		t.Fatalf("CopyUntrackedFiles failed: %v", err)
	}

	// Should have copied 3 files (.env, .env.local, config/secrets.json)
	// NOT node_modules contents
	if copied != 3 {
		t.Errorf("copied = %d, want 3", copied)
	}

	// Verify .env was copied
	content, err := os.ReadFile(filepath.Join(dstWorktree, ".env"))
	if err != nil {
		t.Errorf(".env not copied: %v", err)
	} else if string(content) != "SECRET=value" {
		t.Errorf(".env content = %q, want %q", string(content), "SECRET=value")
	}

	// Verify .env.local was copied
	if _, err := os.Stat(filepath.Join(dstWorktree, ".env.local")); err != nil {
		t.Errorf(".env.local not copied: %v", err)
	}

	// Verify config/secrets.json was copied
	if _, err := os.Stat(filepath.Join(dstWorktree, "config", "secrets.json")); err != nil {
		t.Errorf("config/secrets.json not copied: %v", err)
	}

	// Verify node_modules was NOT copied
	if _, err := os.Stat(filepath.Join(dstWorktree, "node_modules")); !os.IsNotExist(err) {
		t.Error("node_modules should not have been copied")
	}
}

func TestCopyUntrackedFilesEmptyRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ygg-copy-empty")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcRepo := filepath.Join(tmpDir, "repo")
	dstWorktree := filepath.Join(tmpDir, "worktree")

	// Initialize git repo with no ignored files
	if err := os.MkdirAll(srcRepo, 0755); err != nil {
		t.Fatal(err)
	}
	runGit(t, srcRepo, "init")

	if err := os.MkdirAll(dstWorktree, 0755); err != nil {
		t.Fatal(err)
	}

	// Should succeed with 0 files copied
	copied, err := CopyUntrackedFiles(srcRepo, dstWorktree)
	if err != nil {
		t.Fatalf("CopyUntrackedFiles failed: %v", err)
	}
	if copied != 0 {
		t.Errorf("copied = %d, want 0", copied)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}
