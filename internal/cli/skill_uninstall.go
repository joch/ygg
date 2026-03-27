package cli

import (
	"os"

	"github.com/joch/ygg/internal/skill"
	"github.com/spf13/cobra"
)

var skillUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the ygg skill for AI coding agents",
	RunE:  runSkillUninstall,
}

func init() {
	skillCmd.AddCommand(skillUninstallCmd)
}

func runSkillUninstall(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		errorMsg("Failed to determine home directory: %v", err)
		return err
	}

	var firstErr error
	results := skill.Uninstall(homeDir)
	for i, r := range results {
		t := skill.Targets[i]
		if r.Err != nil {
			errorMsg("Failed to uninstall for %s: %v", t.Name, r.Err)
			if firstErr == nil {
				firstErr = r.Err
			}
			continue
		}
		if !r.Found {
			info("Skill not installed for %s", t.Name)
			continue
		}
		success("Skill removed from %s (~/%s)", t.Name, t.RelDir)
	}

	return firstErr
}
