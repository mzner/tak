# tak

Git worktree manager with pinning, tmux integration, and lifecycle tools.

tak makes git worktrees easy to create, navigate, and clean up. Pin long-lived worktrees, jump between them with tmux, and garbage-collect stale ones.

<!-- TODO: Add demo GIF here (record with https://github.com/charmbracelet/vhs) -->
<!-- ![tak demo](./docs/demo.gif) -->

## Why tak?

Git worktrees let you work on multiple branches simultaneously without stashing or switching. But managing them by hand is tedious — you have to remember paths, manually clean up, and set up your dev environment every time.

tak handles all of that:

| Without tak | With tak |
|---|---|
| `git worktree add ../web--feature--auth -b feature/auth` | `tak add feature/auth` |
| Remember the path, `cd` manually | `tak cd feature/auth` |
| Forget to clean up merged branches | `tak gc --merged` |
| Manually open tmux, split panes, run commands | `tak open` (uses your layout config) |
| Accidentally delete pinned worktrees | `tak pin` protects them |

## Install

```bash
brew install mzner/tap/tak
```

Or build from source:

```bash
go install github.com/mzner/tak@latest
```

## Quick Start

```bash
# 1. Initialize tak in your repo
tak init

# 2. Set up shell integration (for tak cd)
eval "$(tak shell-init zsh)"   # add to ~/.zshrc

# 3. Create a worktree and open it in tmux
tak add feature/auth -o

# 4. Pin it so gc won't clean it up
tak pin

# 5. List all worktrees
tak ls

# 6. Jump between worktrees
tak cd feature/auth

# 7. Health check
tak doctor

# 8. Clean up merged branches
tak gc --merged
```

## Shell Integration

Required for `tak cd` to change your directory. Add to your shell rc file:

```bash
# .zshrc or .bashrc
eval "$(tak shell-init zsh)"
```

```fish
# config.fish
tak shell-init fish | source
```

Without this, `tak cd` prints the path but can't change your shell's working directory.

## Commands

| Command | Description |
|---------|-------------|
| `tak add <branch> [-o] [-p] [-f base]` | Create a worktree (`-o` opens in tmux, `-p` pins, `-f` sets base branch) |
| `tak rm [branch...] [-F]` | Remove worktree(s) and branch — interactive if no arg |
| `tak ls [-s]` | List all worktrees (`-s` includes dirty/clean status) |
| `tak info [branch]` | Show worktree details (base, ahead/behind, age) |
| `tak exec <branch> -- <cmd>` | Run a command in a worktree without cd'ing |
| `tak cd [branch]` | Change to a worktree directory — interactive if no arg |
| `tak open [branch]` | Open/switch to tmux window — interactive if no arg |
| `tak pin [branch]` | Pin a worktree (no arg = current) |
| `tak unpin [branch]` | Unpin a worktree |
| `tak doctor` | Health check all worktrees |
| `tak gc [-m] [-n]` | Clean up broken worktrees (`-m` includes merged, `-n` dry run) |
| `tak layout` | Configure tmux pane layout (interactive wizard) |
| `tak config` | Show config file paths and contents |
| `tak init` | Initialize tak in a repo |
| `tak completion <shell>` | Generate shell completion script |
| `tak shell-init <shell>` | Print shell hook for zsh/bash/fish |

## Configuration

tak uses two config files. Per-repo settings override global settings.

### Per-repo: `.tak.yml`

Created by `tak init`. Lives in your repo root.

```yaml
worktree_base: ""         # empty = sibling dirs (default)
branch_prefix: ""         # auto-prepend to branch names (e.g. "feature/")

pins:
  - feature/auth

# Optional: tmux pane layout for tak open (configure with tak layout)
tmux:
  layout: main-vertical
  panes:
    - name: editor
      command: $EDITOR
    - name: dev
      command: pnpm dev
    - name: shell
      command: ""

# Optional: hooks run after tak add creates a worktree
hooks:
  post_create:
    - type: copy
      from: .env
      to: .env
    - type: symlink
      from: node_modules
      to: node_modules
    - type: command
      command: npm ci
```

**Hook types:**

| Type | Description | Fields |
|------|-------------|--------|
| `copy` | Copy file/directory from main worktree to new | `from`, `to` (defaults to `from`) |
| `symlink` | Create symlink from new worktree pointing to main | `from`, `to` (defaults to `from`) |
| `command` | Run a shell command in the new worktree | `command`, `env` (optional), `work_dir` (optional) |

### Global: `~/.config/tak/config.yml`

Optional. Sets defaults for all repos. Per-repo `.tak.yml` overrides these.

```yaml
worktree_base: ~/worktrees   # override default for all repos
repos:
  web: ~/projects/web
  ocis: ~/projects/ocis
```

## How It Works

- Worktrees are created as sibling directories by default: `~/projects/web` → `~/projects/web--feature--auth`
- `tak rm` removes the worktree and deletes the branch (keeps it if there are unmerged commits, unless `-F`)
- `tak open` uses the `tmux` config from `.tak.yml` to create pane layouts, or a plain window if unconfigured
- Pins are stored in `.tak.yml` — recoverable config, not ephemeral state
- State cache (`.tak/state.json`) is rebuilt automatically if deleted
- Per-repo `.tak.yml` overrides global `~/.config/tak/config.yml` for any key present in both
- All git/tmux interaction is via shell commands — no heavy dependencies

## Contributing

See [docs/contributing.md](docs/contributing.md).

## License

MIT
