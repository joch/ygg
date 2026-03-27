package skill

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed SKILL.md
var Content []byte

// Target represents an agent skill directory.
type Target struct {
	Name   string // Display name (e.g. "Claude Code")
	RelDir string // Directory relative to home (e.g. ".claude/skills/ygg")
}

// Targets lists all supported agent skill directories.
var Targets = []Target{
	{Name: "Claude Code", RelDir: filepath.Join(".claude", "skills", "ygg")},
	{Name: "Codex", RelDir: filepath.Join(".agents", "skills", "ygg")},
}

// InstallResult holds the outcome of installing to one target.
type InstallResult struct {
	Updated bool  // true if file already existed and was overwritten
	Err     error
}

// UninstallResult holds the outcome of uninstalling from one target.
type UninstallResult struct {
	Found bool  // true if the skill was installed
	Err   error
}

// Install writes the embedded skill file to all target directories.
func Install(homeDir string) []InstallResult {
	results := make([]InstallResult, len(Targets))
	for i, t := range Targets {
		dir := filepath.Join(homeDir, t.RelDir)
		path := filepath.Join(dir, "SKILL.md")

		_, err := os.Stat(path)
		updated := err == nil

		if err := os.MkdirAll(dir, 0o755); err != nil {
			results[i] = InstallResult{Err: fmt.Errorf("create directory: %w", err)}
			continue
		}

		if err := os.WriteFile(path, Content, 0o644); err != nil {
			results[i] = InstallResult{Err: fmt.Errorf("write file: %w", err)}
			continue
		}

		results[i] = InstallResult{Updated: updated}
	}
	return results
}

// Uninstall removes the skill directory from all targets.
func Uninstall(homeDir string) []UninstallResult {
	results := make([]UninstallResult, len(Targets))
	for i, t := range Targets {
		dir := filepath.Join(homeDir, t.RelDir)

		if _, err := os.Stat(dir); err != nil {
			if os.IsNotExist(err) {
				results[i] = UninstallResult{Found: false}
			} else {
				results[i] = UninstallResult{Found: false, Err: fmt.Errorf("check directory: %w", err)}
			}
			continue
		}

		if err := os.RemoveAll(dir); err != nil {
			results[i] = UninstallResult{Found: true, Err: fmt.Errorf("remove directory: %w", err)}
			continue
		}

		results[i] = UninstallResult{Found: true}
	}
	return results
}
