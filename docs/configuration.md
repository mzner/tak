# Configuration

tak uses two optional config files, merged together with local overriding global.

## Per-repo config: `.tak.yml`

Created by `tak init` in your repo root.

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

# Tmux pane layout for tak open (configure interactively with `tak layout`)
tmux:
  layout: main-vertical   # even-vertical, even-horizontal, main-vertical, main-horizontal, tiled
  panes:
    - name: editor
      command: $EDITOR
    - name: dev
      command: pnpm dev
    - name: shell
      command: ""          # empty = plain shell

# Lifecycle hooks — run automatically after tak add
hooks:
  post_create:
    - type: copy           # copy file/directory from main worktree to new
      from: .env
      to: .env             # defaults to same as 'from' if omitted
    - type: symlink        # create symlink pointing to main worktree
      from: node_modules
      to: node_modules
    - type: command        # run a shell command in the new worktree
      command: pnpm install
      env:                 # optional environment variables
        NODE_ENV: development
      work_dir: "."        # optional subdirectory to run in
```

## Global config: `~/.config/tak/config.yml`

Optional. Sets defaults for all repositories.

```yaml
# Default worktree location (overridden by per-repo worktree_base)
worktree_base: ~/projects/worktrees

# Registered repos for cross-repo access (managed by `tak repo add`)
repos:
  web: ~/projects/web
  api: ~/projects/api
  ocis: ~/projects/ocis
```

## Merge Logic

1. Start with defaults (sibling dirs, no prefix, no pins, no hooks)
2. Apply global config values (if file exists)
3. Apply local config values (if file exists)
4. Local wins for any key present in both

Note: hooks and tmux layout are per-repo only (not in global config).

## State: `.tak/state.json`

**You don't edit this file.** tak manages it automatically.

It caches which worktrees exist, when they were created, and which branch they were created from. If you delete it, tak rebuilds it from `git worktree list` — but you'll lose age and base-branch info (pins are safe in `.tak.yml`).

## Worktree Path Resolution

Default (no `worktree_base`):
```
~/projects/web + branch "feature/auth" → ~/projects/web--feature--auth
```

With `worktree_base: ~/projects/worktrees`:
```
~/projects/web + branch "feature/auth" → ~/projects/worktrees/web--feature--auth
```

Slug rules:
- `/` in branch names → `--` in directory names
- Spaces/special chars → `-`
- Lowercased

## View Current Config

```bash
tak config   # shows both file paths and their contents
```
