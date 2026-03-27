package cli

import (
	"github.com/spf13/cobra"
)

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage the ygg agent skill",
	Long:  `Install or uninstall the ygg skill for AI coding agents (Claude Code, Codex).`,
}

func init() {
	rootCmd.AddCommand(skillCmd)
}
