---
name: ygg
description: Git worktree manager using the `ygg` CLI. Use when the user wants to create, list, switch, remove, or clean git worktrees. Triggers on phrases like "new worktree", "create a tree", "switch worktree", "ygg new", "ygg switch", "ygg clean", or any git worktree workflow management task.
---

# ygg — Git Worktree Helper

ygg manages git worktrees stored in `.worktrees/` inside the repo root. Each worktree gets its own branch and directory, enabling isolated parallel work.

## Commands

```bash
ygg new <name>       # Create worktree + branch from default branch, open subshell/Zellij tab
ygg list             # List all worktrees (* = current, [modified] = uncommitted changes)
ygg switch <name>    # Switch to an existing worktree (alias: sw)
ygg remove [name]    # Remove a worktree; omit name when inside the worktree (alias: rm)
ygg remove -f [name] # Force remove even with uncommitted/unmerged changes
ygg clean            # Remove worktrees whose branches are merged to default
ygg clean --dry-run  # Preview what clean would remove
ygg clean --force    # Clean without confirmation prompts
```

## Typical Workflow

1. `ygg new <feature-name>` — start isolated work on a feature
2. Do work, commit, open PR
3. After merge: `ygg clean` to remove merged worktrees

## Key Behaviors

- `ygg new` fetches latest from origin, bases the branch on `main`/`master`, and copies untracked files from the main worktree
- Sets `$YGG_WORKTREE` env var inside the shell for prompt integration
- Zellij detected via `$ZELLIJ`: opens a named tab instead of a nested subshell
- `ygg switch` inside an existing ygg shell emits a `cd` command for the wrapper to evaluate
- `ygg remove` without a name removes the current worktree (if inside one) and returns you to main

## When Helping the User

- Always use `ygg new <name>` to create worktrees — never raw `git worktree add`
- Suggest `ygg clean` after merging multiple PRs to tidy up
- Use `ygg list` to check what worktrees exist before switching or removing
- Prefer `ygg remove` over `ygg remove -f` unless the user explicitly wants to discard unmerged work
