package shell

import (
	"os"
	"os/exec"
	"path/filepath"
)

const yggWrapper = `
ygg() {
  local out
  out=$(command ygg "$@")
  local ret=$?
  if [[ $out == cd\ * ]]; then
    eval "$out"
  else
    echo "$out"
  fi
  return $ret
}
`

// Spawn starts an interactive subshell in the given directory.
// It sets YGG_SHELL=1 so nested ygg commands can detect they're in a subshell.
// It sets YGG_WORKTREE to the worktree name for prompt integration.
func Spawn(dir, worktreeName string) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	shellName := filepath.Base(shell)

	// Build the command based on shell type
	var cmd *exec.Cmd
	switch shellName {
	case "zsh", "bash":
		cmd = exec.Command(shell, "-i")
	default:
		cmd = exec.Command(shell)
	}

	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set environment
	env := os.Environ()
	env = append(env, "YGG_SHELL=1")
	env = append(env, "YGG_WORKTREE="+worktreeName)
	env = append(env, "YGG_WRAPPER="+yggWrapper)
	cmd.Env = env

	return cmd.Run()
}
