package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Worktree struct {
	Name        string
	Path        string
	Branch      string
	IsBare      bool
	IsLocked    bool
	IsPrimary   bool
	CopiedFiles int   // Number of untracked files copied during creation
	CopyError   error // Error during file copy (non-fatal)
}

type Manager struct {
	repoPath string
	repoName string
	baseDir  string
}

// NewManager creates a worktree manager for the repository at the given path.
// Worktrees are stored in <repo-root>/.worktrees/
func NewManager(path string) (*Manager, error) {
	repoRoot, err := FindRepoRoot(path)
	if err != nil {
		return nil, err
	}

	repoName := filepath.Base(repoRoot)
	// Place worktrees inside repo root: <repo>/.worktrees/<branch-name>/
	baseDir := filepath.Join(repoRoot, ".worktrees")

	return &Manager{
		repoPath: repoRoot,
		repoName: repoName,
		baseDir:  baseDir,
	}, nil
}

// RepoPath returns the path to the main repository.
func (m *Manager) RepoPath() string {
	return m.repoPath
}

// BaseDir returns the base directory where worktrees are stored.
func (m *Manager) BaseDir() string {
	return m.baseDir
}

// RepoName returns the repository name (base directory name).
func (m *Manager) RepoName() string {
	return m.repoName
}

// DefaultBranch returns the default branch (main, master, etc).
func (m *Manager) DefaultBranch() (string, error) {
	// Try to get from remote HEAD
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd.Dir = m.repoPath
	output, err := cmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(output))
		branch = strings.TrimPrefix(branch, "refs/remotes/origin/")
		return branch, nil
	}

	// Fallback: check if main or master exists
	for _, branch := range []string{"main", "master"} {
		cmd := exec.Command("git", "rev-parse", "--verify", branch)
		cmd.Dir = m.repoPath
		if err := cmd.Run(); err == nil {
			return branch, nil
		}
	}

	return "", fmt.Errorf("could not detect default branch")
}

// Fetch fetches from origin.
func (m *Manager) Fetch() error {
	cmd := exec.Command("git", "fetch", "origin")
	cmd.Dir = m.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch failed: %w\n%s", err, output)
	}
	return nil
}

// Pull pulls the current branch from origin.
func (m *Manager) Pull() error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = m.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %w\n%s", err, output)
	}
	return nil
}

// HasUncommittedChanges returns true if there are uncommitted changes.
func (m *Manager) HasUncommittedChanges(path string) bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

// Create creates a new worktree with the given name, based on the default branch.
func (m *Manager) Create(name string) (*Worktree, error) {
	if err := os.MkdirAll(m.baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create worktrees directory: %w", err)
	}

	worktreePath := filepath.Join(m.baseDir, name)

	if _, err := os.Stat(worktreePath); err == nil {
		return nil, fmt.Errorf("worktree %s already exists at %s", name, worktreePath)
	}

	defaultBranch, err := m.DefaultBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to detect default branch: %w", err)
	}

	// Create worktree with new branch based on default branch
	cmd := exec.Command("git", "worktree", "add", "-b", name, worktreePath, defaultBranch)
	cmd.Dir = m.repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Branch might already exist, try without -b
		if strings.Contains(string(output), "already exists") {
			cmd = exec.Command("git", "worktree", "add", worktreePath, name)
			cmd.Dir = m.repoPath
			output, err = cmd.CombinedOutput()
		}

		if err != nil {
			return nil, fmt.Errorf("failed to create worktree: %w\n%s", err, output)
		}
	}

	// Copy untracked/ignored files from main worktree
	copied, copyErr := CopyUntrackedFiles(m.repoPath, worktreePath)

	return &Worktree{
		Name:        name,
		Path:        worktreePath,
		Branch:      name,
		CopiedFiles: copied,
		CopyError:   copyErr,
	}, nil
}

// List returns all worktrees for this repository.
func (m *Manager) List() ([]*Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = m.repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	return parseWorktreeList(string(output))
}

// Get returns a specific worktree by name.
func (m *Manager) Get(name string) (*Worktree, error) {
	worktrees, err := m.List()
	if err != nil {
		return nil, err
	}

	for _, wt := range worktrees {
		if wt.Name == name || wt.Branch == name {
			return wt, nil
		}
	}

	return nil, fmt.Errorf("worktree %q not found", name)
}

// Remove removes a worktree by name.
func (m *Manager) Remove(name string) error {
	wt, err := m.Get(name)
	if err != nil {
		return err
	}

	if wt.IsPrimary {
		return fmt.Errorf("cannot remove the primary worktree")
	}

	cmd := exec.Command("git", "worktree", "remove", wt.Path)
	cmd.Dir = m.repoPath

	if output, err := cmd.CombinedOutput(); err != nil {
		// Try with force if needed
		if strings.Contains(string(output), "force") {
			cmd = exec.Command("git", "worktree", "remove", "--force", wt.Path)
			cmd.Dir = m.repoPath
			if output, err = cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to remove worktree: %w\n%s", err, output)
			}
		} else {
			return fmt.Errorf("failed to remove worktree: %w\n%s", err, output)
		}
	}

	return nil
}

// MergedBranches returns branches that have been merged into the given branch.
func (m *Manager) MergedBranches(into string) ([]string, error) {
	cmd := exec.Command("git", "branch", "--merged", into)
	cmd.Dir = m.repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get merged branches: %w", err)
	}

	var branches []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "* ") // Remove current branch marker
		if line != "" && line != into {
			branches = append(branches, line)
		}
	}
	return branches, nil
}

// IsBranchMerged checks if a branch has been merged into the default branch.
func (m *Manager) IsBranchMerged(branch string) (bool, error) {
	defaultBranch, err := m.DefaultBranch()
	if err != nil {
		return false, err
	}

	// The branch itself is considered "merged" if it's the default branch
	if branch == defaultBranch {
		return true, nil
	}

	merged, err := m.MergedBranches(defaultBranch)
	if err != nil {
		return false, err
	}

	for _, b := range merged {
		if b == branch {
			return true, nil
		}
	}
	return false, nil
}

// Current returns the worktree for the current working directory.
func (m *Manager) Current() (*Worktree, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	worktrees, err := m.List()
	if err != nil {
		return nil, err
	}

	for _, wt := range worktrees {
		if strings.HasPrefix(cwd, wt.Path) {
			return wt, nil
		}
	}

	return nil, fmt.Errorf("not in a worktree")
}

// FindRepoRoot finds the git repository root from any path.
func FindRepoRoot(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = absPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository: %s", absPath)
	}

	return strings.TrimSpace(string(output)), nil
}

func parseWorktreeList(output string) ([]*Worktree, error) {
	var worktrees []*Worktree
	lines := strings.Split(output, "\n")

	var current *Worktree
	firstWorktree := true
	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			if current != nil {
				worktrees = append(worktrees, current)
			}
			path := strings.TrimPrefix(line, "worktree ")
			name := filepath.Base(path)
			isPrimary := firstWorktree
			if firstWorktree {
				name = "main"
				firstWorktree = false
			}
			current = &Worktree{
				Path:      path,
				Name:      name,
				IsPrimary: isPrimary,
			}
		} else if current != nil {
			if strings.HasPrefix(line, "branch ") {
				current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
			} else if strings.HasPrefix(line, "bare") {
				current.IsBare = true
			} else if strings.HasPrefix(line, "locked") {
				current.IsLocked = true
			}
		}
	}

	if current != nil {
		worktrees = append(worktrees, current)
	}

	return worktrees, nil
}
