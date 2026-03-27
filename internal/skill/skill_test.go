package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallCreatesFiles(t *testing.T) {
	home := t.TempDir()
	results := Install(home)

	if len(results) != len(Targets) {
		t.Fatalf("expected %d results, got %d", len(Targets), len(results))
	}

	for i, r := range results {
		if r.Err != nil {
			t.Errorf("target %s: unexpected error: %v", Targets[i].Name, r.Err)
		}
		if r.Updated {
			t.Errorf("target %s: expected new install, got update", Targets[i].Name)
		}

		path := filepath.Join(home, Targets[i].RelDir, "SKILL.md")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("target %s: failed to read installed file: %v", Targets[i].Name, err)
		}
		if string(content) != string(Content) {
			t.Errorf("target %s: installed content does not match embedded content", Targets[i].Name)
		}
	}
}

func TestInstallUpdatesExisting(t *testing.T) {
	home := t.TempDir()

	// First install
	Install(home)

	// Second install should report updated
	results := Install(home)
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("target %s: unexpected error: %v", Targets[i].Name, r.Err)
		}
		if !r.Updated {
			t.Errorf("target %s: expected update, got new install", Targets[i].Name)
		}
	}
}

func TestUninstallRemovesFiles(t *testing.T) {
	home := t.TempDir()

	// Install first
	Install(home)

	// Uninstall
	results := Uninstall(home)
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("target %s: unexpected error: %v", Targets[i].Name, r.Err)
		}
		if !r.Found {
			t.Errorf("target %s: expected found=true after install", Targets[i].Name)
		}

		dir := filepath.Join(home, Targets[i].RelDir)
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Errorf("target %s: directory still exists after uninstall", Targets[i].Name)
		}
	}
}

func TestUninstallWhenNotInstalled(t *testing.T) {
	home := t.TempDir()

	results := Uninstall(home)
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("target %s: unexpected error: %v", Targets[i].Name, r.Err)
		}
		if r.Found {
			t.Errorf("target %s: expected found=false when not installed", Targets[i].Name)
		}
	}
}
