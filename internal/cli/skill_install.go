package cli

import (
	"os"

	"github.com/joch/ygg/internal/skill"
	"github.com/spf13/cobra"
)

var skillInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the ygg skill for AI coding agents",
	RunE:  runSkillInstall,
}

func init() {
	skillCmd.AddCommand(skillInstallCmd)
}

func runSkillInstall(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		errorMsg("Failed to determine home directory: %v", err)
		return err
	}

	var firstErr error
	results := skill.Install(homeDir)
	for i, r := range results {
		t := skill.Targets[i]
		if r.Err != nil {
			errorMsg("Failed to install for %s: %v", t.Name, r.Err)
			if firstErr == nil {
				firstErr = r.Err
			}
			continue
		}
		if r.Updated {
			info("Skill already installed for %s, updating...", t.Name)
		}
		success("Skill installed to %s (~/%s)", t.Name, t.RelDir)
	}

	return firstErr
}
