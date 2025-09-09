package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ygg",
	Short: "Git worktree helper",
	Long:  `Ygg is a simple CLI tool for managing git worktrees.`,
	SilenceUsage: true,
}

func Execute() error {
	return rootCmd.Execute()
}

// InYggShell returns true if we're inside a ygg-spawned subshell
func InYggShell() bool {
	return os.Getenv("YGG_SHELL") == "1"
}

func success(format string, args ...interface{}) {
	green := color.New(color.FgGreen)
	green.Print("✓ ")
	fmt.Printf(format+"\n", args...)
}

func errorMsg(format string, args ...interface{}) {
	red := color.New(color.FgRed)
	red.Print("✗ ")
	fmt.Printf(format+"\n", args...)
}

func info(format string, args ...interface{}) {
	blue := color.New(color.FgBlue)
	blue.Print("ℹ ")
	fmt.Printf(format+"\n", args...)
}
