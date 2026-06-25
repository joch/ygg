package worktree

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitOut runs a git command and returns trimmed stdout, failing the test on
// error. It uses Output() (not CombinedOutput) so stderr can never pollute the
// returned value, which callers compare against commit SHAs.
func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v failed: %v", args, err)
	}
	return strings.TrimSpace(string(out))
}

// setupRepoWithStaleMain builds a local clone whose origin/main is one commit
// ahead of the local main branch (simulating "forgot to pull main"). It returns
// the local repo path, the stale local main SHA (A), and the fresh origin SHA (B).
func setupRepoWithStaleMain(t *testing.T) (localDir, staleA, freshB string) {
	t.Helper()
	tmp := t.TempDir()

	originDir := filepath.Join(tmp, "origin.git")
	localDir = filepath.Join(tmp, "local")
	otherDir := filepath.Join(tmp, "other")

	// Bare origin
	runGit(t, tmp, "init", "-q", "--bare", "--initial-branch=main", originDir)

	// Local clone: commit A on main, push, record origin HEAD
	runGit(t, tmp, "clone", "-q", originDir, localDir)
	runGit(t, localDir, "config", "user.email", "test@test.com")
	runGit(t, localDir, "config", "user.name", "Test")
	runGit(t, localDir, "checkout", "-q", "-b", "main")
	runGit(t, localDir, "commit", "-q", "--allow-empty", "-m", "commit A")
	runGit(t, localDir, "push", "-q", "-u", "origin", "main")
	runGit(t, localDir, "remote", "set-head", "origin", "-a")
	staleA = gitOut(t, localDir, "rev-parse", "HEAD")

	// Teammate pushes commit B to origin from a separate clone
	runGit(t, tmp, "clone", "-q", originDir, otherDir)
	runGit(t, otherDir, "config", "user.email", "other@test.com")
	runGit(t, otherDir, "config", "user.name", "Other")
	runGit(t, otherDir, "commit", "-q", "--allow-empty", "-m", "commit B")
	runGit(t, otherDir, "push", "-q", "origin", "main")
	freshB = gitOut(t, otherDir, "rev-parse", "HEAD")

	if staleA == freshB {
		t.Fatal("setup error: stale and fresh SHAs are identical")
	}
	return localDir, staleA, freshB
}

// TestCreateBasesOnFreshOrigin verifies that a new worktree is based on the
// freshly fetched origin/<default> tip, not the stale local default branch.
func TestCreateBasesOnFreshOrigin(t *testing.T) {
	localDir, staleA, freshB := setupRepoWithStaleMain(t)

	wm, err := NewManager(localDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Mirror "ygg new": fetch origin (updates origin/main, not local main).
	if err := wm.Fetch(); err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Sanity: local main is still stale, origin/main is fresh.
	if got := gitOut(t, localDir, "rev-parse", "main"); got != staleA {
		t.Fatalf("precondition: local main = %s, want stale %s", got, staleA)
	}
	if got := gitOut(t, localDir, "rev-parse", "origin/main"); got != freshB {
		t.Fatalf("precondition: origin/main = %s, want fresh %s", got, freshB)
	}

	wt, err := wm.Create("feature")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	head := gitOut(t, wt.Path, "rev-parse", "HEAD")
	if head != freshB {
		t.Errorf("worktree HEAD = %s, want fresh origin/main %s (stale local main was %s)",
			head, freshB, staleA)
	}
	if wt.Base != "origin/main" {
		t.Errorf("wt.Base = %q, want %q", wt.Base, "origin/main")
	}
}

// TestCreateUsesLocalWhenAheadOfOrigin verifies that when the local default
// branch has commits not yet on origin (committed directly without pushing, or
// fetch failed offline), the new worktree includes them rather than silently
// basing on an older origin/<default> tip.
func TestCreateUsesLocalWhenAheadOfOrigin(t *testing.T) {
	tmp := t.TempDir()
	originDir := filepath.Join(tmp, "origin.git")
	localDir := filepath.Join(tmp, "local")

	runGit(t, tmp, "init", "-q", "--bare", "--initial-branch=main", originDir)
	runGit(t, tmp, "clone", "-q", originDir, localDir)
	runGit(t, localDir, "config", "user.email", "test@test.com")
	runGit(t, localDir, "config", "user.name", "Test")
	runGit(t, localDir, "checkout", "-q", "-b", "main")
	runGit(t, localDir, "commit", "-q", "--allow-empty", "-m", "commit A")
	runGit(t, localDir, "push", "-q", "-u", "origin", "main")
	runGit(t, localDir, "remote", "set-head", "origin", "-a")

	// Commit B locally on main without pushing: local main ahead of origin/main.
	runGit(t, localDir, "commit", "-q", "--allow-empty", "-m", "commit B (unpushed)")
	localAhead := gitOut(t, localDir, "rev-parse", "main")

	wm, err := NewManager(localDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	if err := wm.Fetch(); err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	wt, err := wm.Create("feature")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if head := gitOut(t, wt.Path, "rev-parse", "HEAD"); head != localAhead {
		t.Errorf("worktree HEAD = %s, want local main %s (must not drop unpushed commits)", head, localAhead)
	}
	if wt.Base != "main" {
		t.Errorf("wt.Base = %q, want %q", wt.Base, "main")
	}
}

// TestCreateDoesNotTrackOrigin verifies the new branch is not set up to track
// origin/<default> as its upstream when based on the remote-tracking ref.
func TestCreateDoesNotTrackOrigin(t *testing.T) {
	localDir, _, _ := setupRepoWithStaleMain(t)

	wm, err := NewManager(localDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	if err := wm.Fetch(); err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if _, err := wm.Create("feature"); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// No upstream configured -> rev-parse for the upstream fails (exit non-zero).
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "feature@{upstream}")
	cmd.Dir = localDir
	if out, err := cmd.CombinedOutput(); err == nil {
		t.Errorf("feature branch unexpectedly tracks %q; want no upstream", strings.TrimSpace(string(out)))
	}
}

// TestCreateWithoutRemoteUsesLocalBranch verifies that in a repo with no remote,
// Create falls back to the local default branch and still succeeds.
func TestCreateWithoutRemoteUsesLocalBranch(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")

	runGit(t, tmp, "init", "-q", "--initial-branch=main", repo)
	runGit(t, repo, "config", "user.email", "test@test.com")
	runGit(t, repo, "config", "user.name", "Test")
	runGit(t, repo, "commit", "-q", "--allow-empty", "-m", "commit A")
	localMain := gitOut(t, repo, "rev-parse", "HEAD")

	wm, err := NewManager(repo)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	wt, err := wm.Create("feature")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if head := gitOut(t, wt.Path, "rev-parse", "HEAD"); head != localMain {
		t.Errorf("worktree HEAD = %s, want local main %s", head, localMain)
	}
}

// TestCreateChecksOutExistingBranch verifies that calling Create with a name
// that already exists as a local branch checks that branch out into the new
// worktree (at its own tip), rather than failing or re-pointing it.
func TestCreateChecksOutExistingBranch(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")

	runGit(t, tmp, "init", "-q", "--initial-branch=main", repo)
	runGit(t, repo, "config", "user.email", "test@test.com")
	runGit(t, repo, "config", "user.name", "Test")
	runGit(t, repo, "commit", "-q", "--allow-empty", "-m", "commit A")

	// Pre-create branch "feature" pointing at its own commit.
	runGit(t, repo, "branch", "feature")
	runGit(t, repo, "checkout", "-q", "feature")
	runGit(t, repo, "commit", "-q", "--allow-empty", "-m", "feature commit")
	featureTip := gitOut(t, repo, "rev-parse", "feature")
	runGit(t, repo, "checkout", "-q", "main")

	wm, err := NewManager(repo)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	wt, err := wm.Create("feature")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if head := gitOut(t, wt.Path, "rev-parse", "HEAD"); head != featureTip {
		t.Errorf("worktree HEAD = %s, want existing feature tip %s", head, featureTip)
	}
	if wt.Base != "" {
		t.Errorf("wt.Base = %q, want empty for existing-branch checkout", wt.Base)
	}
}
