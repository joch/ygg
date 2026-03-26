# Zellij Tab Integration

## Summary

When running inside a zellij session, ygg should create/focus zellij tabs instead of spawning subshells. This is auto-detected via the `ZELLIJ` environment variable with no user configuration required. Outside zellij, behavior is unchanged (subshells).

## Detection

- `os.Getenv("ZELLIJ") != ""` indicates a zellij session
- No flags or config needed — detection is automatic

## New Package: `internal/zellij/zellij.go`

Two exported functions:

### `InZellij() bool`

Returns true if `ZELLIJ` env var is set.

### `OpenTab(dir, name string) error`

Opens or focuses a zellij tab for the given worktree:

1. Run `zellij action query-tab-names` to get existing tab names
2. If a tab named `name` exists: run `zellij action go-to-tab-name <name>` to focus it
3. If no tab exists: run `zellij action new-tab --name <name> --cwd <dir>` to create one

Tab names match the worktree name (e.g., `my-feature`).

## Changes to Existing Code

### `internal/cli/new.go` — `runNew()`

After worktree creation, before calling `shell.Spawn()`:

```
if zellij.InZellij() {
    return zellij.OpenTab(wt.Path, wt.Name)
}
```

The existing nested-shell `cd` path (when `InYggShell()` is true) remains unchanged and takes priority — it runs before the zellij check.

### `internal/cli/switch.go` — `runSwitch()`

Same pattern: check `zellij.InZellij()` before `shell.Spawn()`.

### No changes to:

- `internal/shell/shell.go` — subshell logic stays as-is
- `internal/cli/remove.go` / `clean.go` — tab cleanup is not in scope
- `internal/cli/root.go` — no new flags

## Error Handling

If zellij commands fail (e.g., `zellij` binary not found despite env var being set), fall back to `shell.Spawn()` with a warning printed via the existing `info()` helper.

## README Update

Add a "Zellij" section documenting auto-detection and tab behavior.

## Testing

- Unit test for `InZellij()` by manipulating env var
- Integration tested manually inside a zellij session
