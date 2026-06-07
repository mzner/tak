# Configuration

tak uses two optional config files, merged together with local overriding global.

## Per-repo config: `.tak.yml`

Created by `tak init` in your repo root. Checked into git or gitignored — your choice.

```yaml
# Where to create worktrees
# Empty (default) = sibling directories (../repo--branch)
# Set a path to use a shared directory for all worktrees
worktree_base: ""

# Auto-prepend to short branch names
# With prefix "feature/", running `tak add auth` creates branch "feature/auth"
branch_prefix: ""

# Pinned worktrees — excluded from tak gc
# Managed by `tak pin` / `tak unpin`, or edit directly
pins:
  - feature/auth
  - long-running/experiment
```

## Global config: `~/.config/tak/config.yml`

Optional. Sets defaults for all repositories.

```yaml
# Default worktree location (overridden by per-repo config)
worktree_base: ~/worktrees

# Repository map (for future multi-repo features)
repos:
  web: ~/projects/web
  ocis: ~/projects/ocis
```

## Merge Logic

1. Start with defaults (sibling dirs, no prefix, no pins)
2. Apply global config values (if file exists)
3. Apply local config values (if file exists)
4. Local wins for any key present in both

## State: `.tak/state.json`

**You don't edit this file.** tak manages it automatically.

It caches which worktrees exist and when they were created. If you delete it, tak rebuilds it from `git worktree list` with no data loss (pins are in `.tak.yml`, not here).

## Worktree Path Resolution

Default (no `worktree_base`):
```
~/projects/web + branch "feature/auth" -> ~/projects/web--feature--auth
```

With `worktree_base: ~/worktrees`:
```
~/projects/web + branch "feature/auth" -> ~/worktrees/web--feature--auth
```

Slug rules:
- `/` in branch names -> `--` in directory names
- Spaces/special chars -> `-`
- Lowercased
